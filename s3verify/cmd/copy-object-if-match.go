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
)

// newCopyObjectIfMatchReq - Create a new HTTP request for a PUT copy object.
func newCopyObjectIfMatchReq(sourceBucketName, sourceObjectName, destBucketName, destObjectName, ETag string) (Request, error) {
	var copyObjectIfMatchReq = Request{
		customHeader: http.Header{},
	}
	// Set the bucketName and objectName
	copyObjectIfMatchReq.bucketName = destBucketName
	copyObjectIfMatchReq.objectName = destObjectName

	// The body will be set by the server so calculate SHA from an empty body.
	reader := bytes.NewReader([]byte(""))
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}
	// Fill in the request header.
	copyObjectIfMatchReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	// Content-MD5 should not be set for CopyObject request.
	// Content-Length should not be set for CopyObject request.
	copyObjectIfMatchReq.customHeader.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))
	copyObjectIfMatchReq.customHeader.Set("x-amz-copy-source-if-match", ETag)
	copyObjectIfMatchReq.customHeader.Set("User-Agent", appUserAgent)

	// Body will be set server side so set to nil here.
	copyObjectIfMatchReq.contentBody = nil

	return copyObjectIfMatchReq, nil
}

// copyObjectIfMatchVerify - Verify that the response returned matches what is expected.
func copyObjectIfMatchVerify(res *http.Response, expectedStatusCode int, expectedError ErrorResponse) error {
	if err := verifyBodyCopyObjectIfMatch(res.Body, expectedError); err != nil {
		return err
	}
	if err := verifyHeaderCopyObjectIfMatch(res.Header); err != nil {
		return err
	}
	if err := verifyStatusCopyObjectIfMatch(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	return nil
}

// verifyHeaderCopyObjectIfMatch - Verify that the header returned matches what is expected.
func verifyHeaderCopyObjectIfMatch(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyBodyCopyObjectIfMatch - Verify that the body returned matches what is expected.jK;
func verifyBodyCopyObjectIfMatch(resBody io.Reader, expectedError ErrorResponse) error {
	if expectedError.Message != "" { // Error is expected. Verify error returned matches.
		errResponse := ErrorResponse{}
		err := xmlDecoder(resBody, &errResponse)
		if err != nil {
			return err
		}
		if errResponse.Message != expectedError.Message {
			err = fmt.Errorf("Unexpected error message: wanted %v, got %v", expectedError.Message, errResponse.Message)
			return err
		}
	} else { // Error unexpected. Body should be a copyobjectresult.
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

	}
	return nil
}

// verifyStatusCopyObjectIfMatch - Verify that the status returned matches what is expected.
func verifyStatusCopyObjectIfMatch(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// Test the PUT Object Copy with If-Match header is set.
func mainCopyObjectIfMatch(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] CopyObject (If-Match)", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// All copy-object-if-match tests take place in
	// s3verify created buckets on s3verify created objects.
	sourceBucketName := s3verifyBuckets[0].Name
	destBucketName := s3verifyBuckets[1].Name
	sourceObject := s3verifyObjects[0]

	// Create bad ETag.
	badETag := "1234567890"

	destObject := &ObjectInfo{
		Key: sourceObject.Key + "if-match",
	}
	expectedError := ErrorResponse{
		Code:    "PreconditionFailed",
		Message: "At least one of the pre-conditions you specified did not hold",
	}
	// Spin scanBar
	scanBar(message)
	// Create a new valid PUT object copy request.
	req, err := newCopyObjectIfMatchReq(sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, sourceObject.ETag)
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
	if err := copyObjectIfMatchVerify(res, http.StatusOK, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}

	// Create a new invalid PUT object copy request.
	badReq, err := newCopyObjectIfMatchReq(sourceBucketName, sourceObject.Key, destBucketName, destObject.Key, badETag)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the request.
	badRes, err := config.execRequest("PUT", badReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(badRes)
	// Verify the request failed as expected.
	if err := copyObjectIfMatchVerify(badRes, http.StatusPreconditionFailed, expectedError); err != nil {
		printMessage(message, err)
		return false
	}
	// Save the copied object.
	copyObjects = append(copyObjects, destObject)
	// Test passed.
	printMessage(message, nil)
	return true
}
