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

// newCopyObjectIfUnModifiedSinceReq - Create a new HTTP request for CopyObject with if-unmodified-since header set.
func newCopyObjectIfUnModifiedSinceReq(sourceBucketName, sourceObjectName, destBucketName, destObjectName string, lastModified time.Time) (Request, error) {
	// copyObjectIfUnModifiedSinceReq - A new HTTP request for CopyObject with if-unmodified-since header set.
	var copyObjectIfUnModifiedSinceReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName
	copyObjectIfUnModifiedSinceReq.bucketName = destBucketName
	copyObjectIfUnModifiedSinceReq.objectName = destObjectName

	// Body will be set by the server so send with request with empty body.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}
	copyObjectIfUnModifiedSinceReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	copyObjectIfUnModifiedSinceReq.customHeader.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))
	copyObjectIfUnModifiedSinceReq.customHeader.Set("x-amz-copy-source-if-unmodified-since", lastModified.Format(http.TimeFormat))
	copyObjectIfUnModifiedSinceReq.customHeader.Set("User-Agent", appUserAgent)

	return copyObjectIfUnModifiedSinceReq, nil
}

// copyObjectIfUnModifiedSinceVerify - verify the returned response matches what is expected.
func copyObjectIfUnModifiedSinceVerify(res *http.Response, expectedStatusCode int, expectedError ErrorResponse) error {
	if err := verifyStatusCopyObjectIfUnModifiedSince(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyCopyObjectIfUnModifiedSince(res.Body, expectedError); err != nil {
		return err
	}
	if err := verifyHeaderCopyObjectIfUnModifiedSince(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyStatusCopyObjectIfUnModifiedSince - verify the status returned matches what is expected.
func verifyStatusCopyObjectIfUnModifiedSince(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyCopyObjectIfUnModifiedSince - verify the body returned matches what is expected.
func verifyBodyCopyObjectIfUnModifiedSince(resBody io.Reader, expectedError ErrorResponse) error {
	if expectedError.Message != "" {
		responseError := ErrorResponse{}
		err := xmlDecoder(resBody, &responseError)
		if err != nil {
			return err
		}
		if responseError.Message != expectedError.Message {
			err := fmt.Errorf("Unexpected Error Message Received: wanted %v, got %v", expectedError.Message, responseError.Message)
			return err
		}
		return nil
	}
	// Verify the body returned is a copyobject result.
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

// verifyHeaderCopyObjectIfUnModifiedSince - verify the header returned matches what is expected.
func verifyHeaderCopyObjectIfUnModifiedSince(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainCopyObjectIfUnModifiedSince - Entry point for the CopyObject if-unmodified-since test.
func mainCopyObjectIfUnModifiedSince(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] CopyObject (If-Unmodified-Since): ", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// All copy-object-if-unmodified-since tests happen in s3verify created buckets
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
		Key: sourceObject.Key + "if-unmodified-since",
	}
	// Expected error on failure.
	expectedError := ErrorResponse{
		Code:    "PreconditionFailed",
		Message: "At least one of the pre-conditions you specified did not hold",
	}
	// Create a new valid request.
	req, err := newCopyObjectIfUnModifiedSinceReq(sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, sourceObject.LastModified.Add(time.Hour*2))
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
	// Verify the response.
	if err := copyObjectIfUnModifiedSinceVerify(res, http.StatusOK, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Add the copied object to the copyObjects slice.
	copyObjects = append(copyObjects, destObject)
	// Spin scanBar
	scanBar(message)
	// Create a new invalid request.
	badReq, err := newCopyObjectIfUnModifiedSinceReq(sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, pastDate)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the bad request.
	badRes, err := config.execRequest("PUT", badReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(badRes)
	// Spin scanBar
	scanBar(message)
	// Verify the bad request fails with the proper error.
	if err := copyObjectIfUnModifiedSinceVerify(badRes, http.StatusPreconditionFailed, expectedError); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
