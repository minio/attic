/*
 * s3verify (C) 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/s3verify/signv4"
)

// Request - a custom version of the http.Request that uses an io.Reader
// rather than an io.ReadCloser for the body to reduce the number of type
// conversions that must happen for retry to function properly.
type Request struct {
	presignURL bool  // Indicates whether or not this will be a presigned http.Request.
	expires    int64 // Describes for how long the presigned URL will be valid for.

	streamingSign bool
	chunkSize     int64

	customHeader http.Header
	contentBody  io.Reader

	bucketName  string
	objectName  string
	queryValues url.Values

	contentLength int64
}

// execRequest - Executes an HTTP request creating an HTTP response and implements retry logic for predefined retryable errors.
func (c ServerConfig) execRequest(method string, customReq Request) (resp *http.Response, err error) {
	var isRetryable bool     // Indicates if request can be retried.
	var bodySeeker io.Seeker // io.Seeking for seeking.
	if customReq.contentBody != nil {
		// Check if body is seekable then it is retryable.
		bodySeeker, isRetryable = customReq.contentBody.(io.Seeker)
	}

	doneCh := make(chan struct{}, 1)
	defer func() {
		doneCh <- struct{}{}
	}()

	// Do not need the index.
	for _ = range newRetryTimer(MaxRetry, time.Second, doneCh) {
		if isRetryable {
			// Seek back to beginning for each attempt.
			if _, err := bodySeeker.Seek(0, 0); err != nil {
				// If seek failed no need to retry.
				return resp, err
			}
		}
		// Create a new request.
		var req *http.Request
		req, err = c.newRequest(method, customReq)
		if err != nil {
			errResponse := ToErrorResponse(err)
			if isS3CodeRetryable(errResponse.Code) {
				continue // Retry.
			}
			return nil, err
		}
		resp, err = c.Client.Do(req)
		if err != nil {
			// For supported network errors verify.
			if isNetErrorRetryable(err) {
				continue // Retry.
			}
			// For other errors there is no need to retry.
			return resp, err
		}
		// For any known successful http status, return quickly.
		for _, httpStatus := range successStatus {
			if httpStatus == resp.StatusCode {
				return resp, nil
			}
		}
		// Read the body to be saved later.
		errBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}
		// Save the body.
		errBodySeeker := bytes.NewReader(errBodyBytes)
		resp.Body = ioutil.NopCloser(errBodySeeker)

		// For errors verify if its retryable otherwise fail quickly.
		errResponse := ToErrorResponse(httpRespToErrorResponse(resp, customReq.bucketName, customReq.objectName))

		//Verify if error response code is retryable.
		if isS3CodeRetryable(errResponse.Code) {
			continue // Retry.
		}
		// Verify if http status code is retryable.
		if isHTTPStatusRetryable(resp.StatusCode) {
			continue // Retry.
		}

		// Save the body back again.
		errBodySeeker.Seek(0, 0) // Seek back to starting point.
		resp.Body = ioutil.NopCloser(errBodySeeker)

		// For all other cases break out of the retry loop.
		break
	}
	return resp, err
}

// newRequest - create an HTTP request out of a customRequest.
func (c ServerConfig) newRequest(method string, customReq Request) (req *http.Request, err error) {
	// Construct a new target URL.
	targetURL, err := makeTargetURL(c.Endpoint, customReq.bucketName, customReq.objectName, c.Region, customReq.queryValues)
	if err != nil {
		return nil, err
	}

	// Initialize a new HTTP request for the method.
	req, err = http.NewRequest(method, targetURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set the target URL.
	req.URL = targetURL

	// Set content body if available.
	if customReq.contentBody != nil {
		req.Body = ioutil.NopCloser(customReq.contentBody)
	}

	// Set all headers.
	for k, v := range customReq.customHeader {
		req.Header.Set(k, v[0])
	}

	// Set ContentLength.
	if customReq.contentLength > 0 {
		req.ContentLength = customReq.contentLength
	}

	// Sign the request if needed.
	if customReq.presignURL && method == "POST" { // These requests would have already been signed.
	} else if customReq.presignURL {
		// Presign the request.
		req = signv4.PreSignV4(*req, c.Access, c.Secret, c.Region, customReq.expires)
	} else if customReq.streamingSign {
		req = signv4.StreamingSignV4(*req, c.Access, c.Secret, c.Region, customReq.chunkSize)
	} else {
		// Else use regular signature v4.
		req = signv4.SignV4(*req, c.Access, c.Secret, c.Region)
	}

	return req, nil
}
