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
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// newGetObjcetIfModifiedSinceReq - Create a new HTTP request to perform.
func newGetObjectIfModifiedSinceReq(bucketName, objectName string, lastModified time.Time) (Request, error) {
	var getObjectIfModifiedReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	getObjectIfModifiedReq.bucketName = bucketName
	getObjectIfModifiedReq.objectName = objectName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because GET requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	getObjectIfModifiedReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	getObjectIfModifiedReq.customHeader.Set("If-Modified-Since", lastModified.Format(http.TimeFormat))
	getObjectIfModifiedReq.customHeader.Set("User-Agent", appUserAgent)

	return getObjectIfModifiedReq, nil
}

// verifyGetObjectIfModifiedSince - Verify that the response matches what is expected.
func verifyGetObjectIfModifiedSince(res *http.Response, expectedBody []byte, expectedStatusCode int) error {
	if err := verifyHeaderGetObjectIfModifiedSince(res.Header); err != nil {
		return err
	}
	if err := verifyBodyGetObjectIfModifiedSince(res.Body, expectedBody); err != nil {
		return err
	}
	if err := verifyStatusGetObjectIfModifiedSince(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	return nil
}

// verifyBodyGetObjectIfModifiedSince - Verify that the response body matches what is expected.
func verifyBodyGetObjectIfModifiedSince(resBody io.Reader, expectedBody []byte) error {
	body, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, expectedBody) {
		err := fmt.Errorf("Unexpected Body Received: wanted %v, got %v", string(expectedBody), string(body))
		return err
	}
	return nil
}

// verifyStatusGetObjectIfModifiedSince - Verify that the response status matches what is expected.
func verifyStatusGetObjectIfModifiedSince(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyHeaderGetObjectIfModifiedSince - Verify that the response header matches what is expected.
func verifyHeaderGetObjectIfModifiedSince(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// Test the compatibility of the GET object API when using the If-Modified-Since header.
func mainGetObjectIfModifiedSince(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] GetObject (If-Modified-Since):", curTest, globalTotalNumTest)
	// Set a date in the past.
	pastDate, err := time.Parse(http.TimeFormat, "Thu, 01 Jan 1970 00:00:00 GMT")
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// All getobject if-modified-since tests happen in s3verify created buckets
	// on s3verify created objects.
	bucketName := s3verifyBuckets[0].Name
	for _, object := range s3verifyObjects {
		// Spin scanBar
		scanBar(message)
		// Create new GET object request.
		req, err := newGetObjectIfModifiedSinceReq(bucketName, object.Key, object.LastModified)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Perform the request.
		res, err := config.execRequest("GET", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Verify the response...these checks do not check the headers yet.
		if err := verifyGetObjectIfModifiedSince(res, []byte(""), http.StatusNotModified); err != nil {
			printMessage(message, err)
			return false
		}
		// Create an acceptable request.
		goodReq, err := newGetObjectIfModifiedSinceReq(bucketName, object.Key, pastDate)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Execute the response that should give back a body.
		goodRes, err := config.execRequest("GET", goodReq)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(goodRes)
		// Verify that the past date gives back the data.
		if err := verifyGetObjectIfModifiedSince(goodRes, object.Body, http.StatusOK); err != nil {
			printMessage(message, err)
			return false
		}
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
