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
	"time"
)

// newRemoveBucketReq - Fill in the dynamic fields of a DELETE request here.
func newRemoveBucketReq(bucketName string) (Request, error) {
	// removeBucketReq is a new DELETE bucket request.
	var removeBucketReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName.
	removeBucketReq.bucketName = bucketName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because DELETE requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	removeBucketReq.customHeader.Set("User-Agent", appUserAgent)
	removeBucketReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	return removeBucketReq, nil
}

// removeBucketVerify - Check a Response's Status, Headers, and Body for AWS S3 compliance.
func removeBucketVerify(res *http.Response, expectedStatusCode int, errorResponse ErrorResponse) error {
	if err := verifyHeaderRemoveBucket(res.Header); err != nil {
		return err
	}
	if err := verifyStatusRemoveBucket(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyRemoveBucket(res.Body, errorResponse); err != nil {
		return err
	}
	return nil
}

// TODO: right now only checks for correctly deleted buckets...need to add in checks for 'failed' tests.

// verifyHeaderRemoveBucket - Check that the responses headers match the expected headers for a given DELETE Bucket request.
func verifyHeaderRemoveBucket(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyBodyRemoveBucket - Check that the body of the response matches the expected body for a given DELETE Bucket request.
func verifyBodyRemoveBucket(resBody io.Reader, expectedError ErrorResponse) error {
	if expectedError.Message != "" { // Error is expected.
		errResponse := ErrorResponse{}
		err := xmlDecoder(resBody, &errResponse)
		if err != nil {
			return err
		}
		if errResponse.Message != expectedError.Message {
			err := fmt.Errorf("Unexpected Error Message: wanted %v, got %v", expectedError.Message, errResponse.Message)
			return err
		}
	}
	return nil
}

// verifyStatusRemoveBucket - Check that the status of the response matches the expected status for a given DELETE Bucket request.
func verifyStatusRemoveBucket(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode { // Successful DELETE request will result in 204 No Content.
		err := fmt.Errorf("Unexpected Status: wanted %d, got %d", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// mainRemoveBucketExists - test the removebucket API.
func mainRemoveBucketExists(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] RemoveBucket (Bucket Exists):", curTest, globalTotalNumTest)
	// Only remove s3verify created buckets.
	for _, bucket := range s3verifyBuckets {
		// Spin the scanBar
		scanBar(message)
		// Generate the new DELETE bucket request.
		req, err := newRemoveBucketReq(bucket.Name)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin the scanBar
		scanBar(message)
		// Perform the request.
		res, err := config.execRequest("DELETE", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Spin the scanBar
		scanBar(message)
		if err := removeBucketVerify(res, 204, ErrorResponse{}); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin the scanBar
		scanBar(message)
	}
	printMessage(message, nil)
	return true
}

// Test the RemoveBucket API when the bucket does not exist.
func mainRemoveBucketDNE(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] RemoveBucket (Bucket DNE):", curTest, globalTotalNumTest)
	// Generate a random bucketName.
	bucketName := randString(60, rand.NewSource(time.Now().UnixNano()), "")
	// Hardcode the expected error response.
	errResponse := ErrorResponse{
		Code:       "NoSuchBucket",
		Message:    "The specified bucket does not exist",
		BucketName: bucketName,
		Key:        "",
	}
	// Spin scanBar
	scanBar(message)
	// Generate a new DELETE bucket request for a bucket that does not exist.
	req, err := newRemoveBucketReq(bucketName)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// spin scanBar
	scanBar(message)
	// Perform the request.
	res, err := config.execRequest("DELETE", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)
	if err := removeBucketVerify(res, http.StatusNotFound, errResponse); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	printMessage(message, nil)
	return true
}

// Test the RemoveBucket API when the bucket is not empty.
func mainRemoveBucketNotEmpty(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] RemoveBucket (Bucket Not Empty):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// Attempt to remove a s3verify created bucket before the objects inside have been removed.
	bucketName := s3verifyBuckets[0].Name

	// Expected error response.
	errResponse := ErrorResponse{
		Code:    "BucketNotEmpty",
		Message: "The bucket you tried to delete is not empty",
	}

	// Spin scanBar
	scanBar(message)
	// Create a new DELETE request for a bucket that is not yet empty.
	req, err := newRemoveBucketReq(bucketName)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Execute the request.
	res, err := config.execRequest("DELETE", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)

	// Verify that the request failed.
	if err := removeBucketVerify(res, http.StatusConflict, errResponse); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
