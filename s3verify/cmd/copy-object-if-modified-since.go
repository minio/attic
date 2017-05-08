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
	"net/url"
	"time"
)

// newCopyObjectIfModifiedSinceReq - Create a new HTTP request for CopyObject with the x-amz-copy-source-if-modified-since header set.
func newCopyObjectIfModifiedSinceReq(sourceBucketName, sourceObjectName, destBucketName, destObjectName string, lastModified time.Time) (Request, error) {
	// Create a new HTTP request for a CopyObject.
	var copyObjectIfModifiedSinceReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName
	copyObjectIfModifiedSinceReq.bucketName = destBucketName
	copyObjectIfModifiedSinceReq.objectName = destObjectName

	// Copying is done by the server so no body is sent in the request.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}
	// Set the header.
	copyObjectIfModifiedSinceReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	copyObjectIfModifiedSinceReq.customHeader.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))
	copyObjectIfModifiedSinceReq.customHeader.Set("x-amz-copy-source-if-modified-since", lastModified.Format(http.TimeFormat))
	copyObjectIfModifiedSinceReq.customHeader.Set("User-Agent", appUserAgent)

	return copyObjectIfModifiedSinceReq, nil
}

// copyObjectIfModifiedSinceVerify - verify the response returned matches what is expected.
func copyObjectIfModifiedSinceVerify(res *http.Response, expectedStatusCode int, expectedError ErrorResponse) error {
	if err := verifyStatusCopyObjectIfModifiedSince(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyCopyObjectIfModifiedSince(res.Body, expectedError); err != nil {
		return err
	}
	if err := verifyHeaderCopyObjectIfModifiedSince(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyStatusCopyObjectIfModifiedSince - verify the status returned matches what is expected.
func verifyStatusCopyObjectIfModifiedSince(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyCopyObjectIfModifiedSince - verify the body returned matches what is expected.
func verifyBodyCopyObjectIfModifiedSince(resBody io.Reader, expectedError ErrorResponse) error {
	if expectedError.Message != "" {
		resError := ErrorResponse{}
		err := xmlDecoder(resBody, &resError)
		if err != nil {
			return err
		}
		if expectedError.Message != resError.Message {
			err := fmt.Errorf("Unexpected Error Message: wanted %v, got %v", expectedError.Message, resError.Message)
			return err
		}
		return nil
	}
	// Verify the body returned is a copyObject result.
	copyObjResult := copyObjectResult{}
	err := xmlDecoder(resBody, &copyObjResult)
	if err != nil {
		body, errR := ioutil.ReadAll(resBody)
		if errR != nil {
			return errR
		}
		err = fmt.Errorf("Unexpected Body Received: %v", string(body))
		return err
	}
	return nil
}

// verifyHeaderCopyObjectIfModifiedSince - verify the header returned matches what is expected.
func verifyHeaderCopyObjectIfModifiedSince(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainCopyObjectIfModifiedSince - test the CopyObject with if-modified-since header.
func mainCopyObjectIfModifiedSince(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] CopyObject (If-Modified-Since):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// All copy-object-if-modified-since tests happen in s3verify created buckets
	// on s3verify created objects.
	sourceBucketName := s3verifyBuckets[0].Name
	destBucketName := s3verifyBuckets[1].Name
	sourceObject := s3verifyObjects[0]

	// Set a date in the past.
	pastDate, err := time.Parse(http.TimeFormat, "Thu, 01 Jan 1970 00:00:00 GMT")
	if err != nil {
		printMessage(message, err)
		return false
	}
	destObject := &ObjectInfo{
		Key: sourceObject.Key + "if-modified-since",
	}
	copyObjects = append(copyObjects, destObject)
	expectedError := ErrorResponse{
		Code:    "PreconditionFailed",
		Message: "At least one of the pre-conditions you specified did not hold",
	}
	// Spin scanBar
	scanBar(message)
	// Create a new request with a valid date.
	req, err := newCopyObjectIfModifiedSinceReq(sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, pastDate)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := config.execRequest("PUT", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)
	// Verify the response is valid.
	if err := copyObjectIfModifiedSinceVerify(res, http.StatusOK, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Create a new request with an invalid date.
	badReq, err := newCopyObjectIfModifiedSinceReq(sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, time.Now().UTC().Add(2*time.Hour))
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	badRes, err := config.execRequest("PUT", badReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(badRes)
	// Spin scanBar
	scanBar(message)
	// Verify the bad request fails the right way.
	if err := copyObjectIfModifiedSinceVerify(badRes, http.StatusPreconditionFailed, expectedError); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	printMessage(message, err)
	return true
}
