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
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// newCopyObjectReq - Create a new HTTP request for PUT object with copy-
func newCopyObjectReq(sourceBucketName, sourceObjectName, destBucketName, destObjectName string) (Request, error) {
	var copyObjectReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName
	copyObjectReq.bucketName = destBucketName
	copyObjectReq.objectName = destObjectName

	// Body will be set by the server so don't upload any body here.
	reader := bytes.NewReader([]byte(""))
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}
	// Fill request headers.
	// Content-MD5 should never be set for CopyObject API.
	copyObjectReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	copyObjectReq.customHeader.Set("x-amz-copy-source", url.QueryEscape(sourceBucketName+"/"+sourceObjectName))
	copyObjectReq.customHeader.Set("User-Agent", appUserAgent)

	return copyObjectReq, nil
}

// copyObjectVerify - Verify that the response returned matches what is expected.
func copyObjectVerify(res *http.Response, expectedStatusCode int, expectedError ErrorResponse) error {
	if err := verifyHeaderCopyObject(res.Header); err != nil {
		return err
	}
	if err := verifyBodyCopyObject(res.Body, expectedError); err != nil {
		return err
	}
	if err := verifyStatusCopyObject(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	return nil
}

// verifyHeaderscopyObject - verify that the header returned matches what is expected.
func verifyHeaderCopyObject(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyBodycopyObject - verify that the body returned is a valid CopyObject Result or error.
func verifyBodyCopyObject(resBody io.Reader, expectedError ErrorResponse) error {
	if expectedError.Message == "" { // Verify that the body returned is a valid copyobject result.
		copyObjRes := copyObjectResult{}
		if err := xmlDecoder(resBody, &copyObjRes); err != nil {
			return err
		}
		return nil
	}
	// otherwise verify that the error returned is what is expected.
	receivedError := ErrorResponse{}
	if err := xmlDecoder(resBody, &receivedError); err != nil {
		return err
	}
	// Validate the message returned.
	if receivedError.Message != expectedError.Message {
		err := fmt.Errorf("Unexpected Error Message Received: wanted %s, got %s", expectedError.Message, receivedError.Message)
		return err
	}
	// Validate the error code returned.
	if receivedError.Code != expectedError.Code {
		err := fmt.Errorf("Unexpected Error Code Received: wanted %s, got %s", expectedError.Code, receivedError.Code)
		return err
	}
	return nil
}

// verifyStatusCopyObject - verify that the status returned matches what is expected.
func verifyStatusCopyObject(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// Test a PUT object request with the copy header set.
func mainCopyObject(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] CopyObject:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// All copy-object tests happen in s3verify created buckets
	// on s3verify created objects.
	sourceBucketName := s3verifyBuckets[0].Name
	destBucketName := s3verifyBuckets[1].Name
	sourceObject := s3verifyObjects[0]

	// TODO: create tests designed to fail.

	// Test same name source and dest objects.
	destObject := &ObjectInfo{
		Key: sourceObject.Key,
	}
	copyObjects = append(copyObjects, destObject)
	// Spin scanBar
	scanBar(message)
	// Create a new request.
	req, err := newCopyObjectReq(sourceBucketName, sourceObject.Key, destBucketName, destObject.Key)
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
	if err := copyObjectVerify(res, http.StatusOK, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}

	// Test different named source and dest objects.
	destObjectDifName := &ObjectInfo{
		Key: sourceObject.Key + "-copy",
	}
	// Spin scanBar
	scanBar(message)
	copyObjects = append(copyObjects, destObjectDifName)
	// Create a new request.
	difNameReq, err := newCopyObjectReq(sourceBucketName, sourceObject.Key, destBucketName, destObjectDifName.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute request.
	difNameRes, err := config.execRequest("PUT", difNameReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := copyObjectVerify(difNameRes, http.StatusOK, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}

	// Test a failed copy. The source will not exist.
	invalidKeyError := ErrorResponse{
		Code:    "NoSuchKey",
		Message: "The specified key does not exist.",
	}
	destObjectFailed := &ObjectInfo{
		Key: randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify-DNE-"),
	}
	// Spin scanBar
	scanBar(message)
	// No need to append this object will not be uploaded on a compatible server.
	// Create a new request.
	invalidReq, err := newCopyObjectReq(sourceBucketName, randString(60, rand.NewSource(time.Now().UnixNano()), ""), destBucketName, destObjectFailed.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute request.
	invalidRes, err := config.execRequest("PUT", invalidReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response failed.
	if err := copyObjectVerify(invalidRes, http.StatusNotFound, invalidKeyError); err != nil {
		printMessage(message, err)
		return false
	}

	// Test a failed copy. The source bucket will not exist.
	invalidSourceBucketError := ErrorResponse{
		Code:    "NoSuchBucket",
		Message: "The specified bucket does not exist",
	}

	destObjectNoSourceBucket := &ObjectInfo{
		Key: sourceObject.Key + "DNE",
	}
	// This object will not be created.
	invalidSourceBucketReq, err := newCopyObjectReq(sourceBucketName+"dne", sourceObject.Key, destBucketName, destObjectNoSourceBucket.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the response.
	invalidSourceBucketRes, err := config.execRequest("PUT", invalidSourceBucketReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the request failed correctly.
	if err := copyObjectVerify(invalidSourceBucketRes, http.StatusNotFound, invalidSourceBucketError); err != nil {
		printMessage(message, err)
		return false
	}

	// Test a failed copy. The dest bucket will not exist.
	invalidDestBucketError := ErrorResponse{
		Code:    "NoSuchBucket",
		Message: "The specified bucket does not exist",
	}

	destObjectNoDestBucket := &ObjectInfo{
		Key: sourceObject.Key + "DNE",
	}
	// This object will not be created.
	invalidDestBucketReq, err := newCopyObjectReq(sourceBucketName, sourceObject.Key, destBucketName+"dne", destObjectNoDestBucket.Key)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the response.
	invalidDestBucketRes, err := config.execRequest("PUT", invalidDestBucketReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the request failed correctly.
	if err := copyObjectVerify(invalidDestBucketRes, http.StatusNotFound, invalidDestBucketError); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
