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

// newAbortMultipartUploadReq - Create a new HTTP request for an abort multipart API.
func newAbortMultipartUploadReq(bucketName, objectName, uploadID string) (Request, error) {
	// abortMultipartUploadReq - a new HTTP request for an abort multipart.
	var abortMultipartUploadReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName
	abortMultipartUploadReq.bucketName = bucketName
	abortMultipartUploadReq.objectName = objectName

	// Set the query values.
	urlValues := make(url.Values)
	urlValues.Set("uploadId", uploadID)
	abortMultipartUploadReq.queryValues = urlValues

	// Set the header.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}
	abortMultipartUploadReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	abortMultipartUploadReq.customHeader.Set("User-Agent", appUserAgent)

	return abortMultipartUploadReq, nil
}

// abortMultipartUploadVerify - verify the response returned matches what is expected.
func abortMultipartUploadVerify(res *http.Response, expectedStatusCode int, expectedError ErrorResponse) error {
	if err := verifyBodyAbortMultipartUploadVerify(res.Body, expectedError); err != nil {
		return err
	}
	if err := verifyStatusAbortMultipartUpload(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderAbortMultipartUpload(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyHeaderAbortMultipartUpload - verify the header returned matches what is expected.
func verifyHeaderAbortMultipartUpload(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyBodyAbortMultipartUploadVerify - verify the body returned has either an error or is empty.
func verifyBodyAbortMultipartUploadVerify(resBody io.Reader, expectedError ErrorResponse) error {
	if expectedError.Message != "" {
		resError := ErrorResponse{}
		err := xmlDecoder(resBody, &resError)
		if err != nil {
			return err
		}
		if resError.Message != expectedError.Message {
			err := fmt.Errorf("Unexpected Error Response: wanted %v, got %v", expectedError.Message, resError.Message)
			return err
		}
	}
	return nil
}

// verifyStatusAbortMultipartUpload - verify the status returned matches what is expected.
func verifyStatusAbortMultipartUpload(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// TODO: This test does not yet test for tests that should fail. Until there is a workaround for the way
// AWS maintains the uploadIDs for several hours there is no sure way to test for the right error messages.  // As of now though it is known there is a bug within the Minio Server that returns a shortened form of the
// error AWS is said to return.

// mainAbortMultipartUpload - abort multipart upload API test.
func mainAbortMultipartUpload(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] Multipart (Abort Upload):", curTest, globalTotalNumTest)
	scanBar(message)
	// All multipart operations take place in the s3verify created buckets.
	bucketName := s3verifyBuckets[0].Name
	validObject := multipartObjects[1] // This multipart has not been completed and will instead be aborted.
	// Spin scanBar
	scanBar(message)
	// Create a new request.
	req, err := newAbortMultipartUploadReq(bucketName, validObject.Key, validObject.UploadID)
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
	// Verify that the response went through.
	if err := abortMultipartUploadVerify(res, 204, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
