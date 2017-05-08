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
	"net/http"
	"net/url"
)

// newPutObjectCopyIfNoneMatchReq - Create a new HTTP request for a CopyObject with the if-none-match header set.
func newCopyObjectIfNoneMatchReq(sourceBucketName, sourceObjectName, destBucketName, destObjectName, ETag string) (Request, error) {
	var copyObjectIfNoneMatchReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName
	copyObjectIfNoneMatchReq.bucketName = destBucketName
	copyObjectIfNoneMatchReq.objectName = destObjectName

	// Body will be calculated by the server so no body needs to be sent in the request.
	reader := bytes.NewReader([]byte(""))
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}
	// Fill in the request header.
	copyObjectIfNoneMatchReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	copyObjectIfNoneMatchReq.customHeader.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))
	copyObjectIfNoneMatchReq.customHeader.Set("x-amz-copy-source-if-none-match", ETag)
	copyObjectIfNoneMatchReq.customHeader.Set("User-Agent", appUserAgent)

	return copyObjectIfNoneMatchReq, nil
}

// Verify that the response returned matches what is expected.
func copyObjectIfNoneMatchVerify(res *http.Response, expectedStatusCode int, expectedError ErrorResponse) error {
	if err := verifyStatusCopyObjectIfNoneMatch(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderCopyObjectIfNoneMatch(res.Header); err != nil {
		return err
	}
	if err := verifyBodyCopyObjectIfNoneMatch(res.Body, expectedError); err != nil {
		return err
	}
	return nil
}

// verifyStatusCopyIfNoneMatch - Verify that the response status matches what is expected.
func verifyStatusCopyObjectIfNoneMatch(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyCopyIfNoneMatch - Verify the body returned matches what is expected.
func verifyBodyCopyObjectIfNoneMatch(resBody io.Reader, expectedError ErrorResponse) error {
	if expectedError.Message != "" { // Error is expected.
		errResponse := ErrorResponse{}
		if err := xmlDecoder(resBody, &errResponse); err != nil {
			return err
		}
		if errResponse.Message != expectedError.Message {
			err := fmt.Errorf("Unexpected Error Response: wanted %v, got %v", expectedError.Message, errResponse.Message)
			return err
		}
		return nil
	}
	// Successful copy expected.
	copyObjRes := copyObjectResult{}
	if err := xmlDecoder(resBody, &copyObjRes); err != nil {
		return err
	}
	return nil
}

// verifyHeaderCopyIfNoneMatch - Verify that the header returned matches what is expected.
func verifyHeaderCopyObjectIfNoneMatch(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// Test the CopyObject API with the if-none-match header set.
func mainCopyObjectIfNoneMatch(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] CopyObject (If-None-Match):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)

	// All copy-object-if-none-match tests happen in s3verify created buckets
	// on s3verify created objects.
	sourceBucketName := s3verifyBuckets[0].Name
	destBucketName := s3verifyBuckets[1].Name
	sourceObject := s3verifyObjects[0]

	// Create unmatchable ETag.
	goodETag := "1234567890"

	destObject := &ObjectInfo{
		Key: sourceObject.Key + "if-none-match",
	}
	copyObjects = append(copyObjects, destObject)
	// Create an error for the case that is expected to fail.
	expectedError := ErrorResponse{
		Code:    "PreconditionFailed",
		Message: "At least one of the pre-conditions you specified did not hold",
	}
	// Create a successful copy request.
	req, err := newCopyObjectIfNoneMatchReq(sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, goodETag)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the response.
	res, err := config.execRequest("PUT", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Verify the response.
	if err = copyObjectIfNoneMatchVerify(res, http.StatusOK, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Create a bad copy request.
	badReq, err := newCopyObjectIfNoneMatchReq(sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, sourceObject.ETag)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the response.
	badRes, err := config.execRequest("PUT", badReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(badRes)
	// Verify the response errors out as it should.
	if err = copyObjectIfNoneMatchVerify(badRes, http.StatusPreconditionFailed, expectedError); err != nil {
		printMessage(message, err)
		return false
	}
	printMessage(message, nil)
	return true
}
