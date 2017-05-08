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
	"reflect"
)

// newGetBucketPolicyReq - create a new request for the get-bucket-policy API.
func newGetBucketPolicyReq(bucketName string) (Request, error) {
	var getBucketPolicyReq = Request{
		customHeader: http.Header{},
	}

	// Set the request bucketName.
	getBucketPolicyReq.bucketName = bucketName

	// Set queryValues.
	urlValues := make(url.Values)
	urlValues.Set("policy", "")
	getBucketPolicyReq.queryValues = urlValues

	// The body of a GET request is always empty.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	getBucketPolicyReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	getBucketPolicyReq.customHeader.Set("User-Agent", appUserAgent)

	return getBucketPolicyReq, nil
}

// getBucketPolicyVerify - Verify the response returned matches what is expected.
func getBucketPolicyVerify(res *http.Response, expectedStatusCode int, expectedPolicy BucketAccessPolicy, expectedError ErrorResponse) error {
	if err := verifyStatusGetBucketPolicy(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderGetBucketPolicy(res.Header); err != nil {
		return err
	}
	if err := verifyBodyGetBucketPolicy(res.Body, expectedPolicy, expectedError); err != nil {
		return err
	}
	return nil
}

// verifyStatusGetBucketPolicy - verify the status returned matches what is expected.
func verifyStatusGetBucketPolicy(respStatusCode int, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %d, got %d", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyHeaderGetBucketPolicy - verify the header returned matches what is expected.
func verifyHeaderGetBucketPolicy(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyBodyGetBucketPolicy - verify the policy returned matches what is expected.
func verifyBodyGetBucketPolicy(resBody io.Reader, expectedPolicy BucketAccessPolicy, expectedError ErrorResponse) error {
	if len(expectedPolicy.Statements) != 0 {
		receivedPolicy := BucketAccessPolicy{}
		if err := jsonDecoder(resBody, &receivedPolicy); err != nil {
			return err
		}
		if !reflect.DeepEqual(receivedPolicy, expectedPolicy) {
			err := fmt.Errorf("Unexpected Bucket Policy Received: wanted %v, got %v", expectedPolicy, receivedPolicy)
			return err
		}
		return nil
	}
	receivedError := ErrorResponse{}
	if err := xmlDecoder(resBody, &receivedError); err != nil {
		return err
	}
	if receivedError.Message != expectedError.Message {
		err := fmt.Errorf("Unexpected Error Message: wanted %s, got %s", expectedError.Message, receivedError.Message)
		return err
	}
	if receivedError.Code != expectedError.Code {
		err := fmt.Errorf("Unexpected Error Code: wanted %s, got %s", expectedError.Code, receivedError.Code)
		return err
	}
	return nil
}

// mainGetBucketPolicy - Entry point for the get-bucket-policy test.
func mainGetBucketPolicy(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] GetBucketPolicy:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)

	// Test missing bucket
	expectedError := ErrorResponse{
		Message: "The bucket policy does not exist",
		Code:    "NoSuchBucketPolicy",
	}
	bucketName := s3verifyBuckets[3].Name
	// Create a new request.
	req, err := newGetBucketPolicyReq(bucketName)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the request.
	res, err := config.execRequest("GET", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Verify the response.
	if err := getBucketPolicyVerify(res, http.StatusNotFound, BucketAccessPolicy{}, expectedError); err != nil {
		printMessage(message, err)
		return false
	}

	// Test readwrite policy is set.
	bucketName = s3verifyBuckets[0].Name
	// Create a new request.
	readWriteReq, err := newGetBucketPolicyReq(bucketName)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	readWriteRes, err := config.execRequest("GET", readWriteReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := getBucketPolicyVerify(readWriteRes, http.StatusOK, s3verifyPolicies[0], ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test readonly policy is set.
	bucketName = s3verifyBuckets[1].Name
	// Create a new request.
	readOnlyReq, err := newGetBucketPolicyReq(bucketName)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	readOnlyRes, err := config.execRequest("GET", readOnlyReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := getBucketPolicyVerify(readOnlyRes, http.StatusOK, s3verifyPolicies[1], ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test writeonly policy is set.
	bucketName = s3verifyBuckets[2].Name
	// Create a new request.
	writeOnlyReq, err := newGetBucketPolicyReq(bucketName)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	writeOnlyRes, err := config.execRequest("GET", writeOnlyReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := getBucketPolicyVerify(writeOnlyRes, http.StatusOK, s3verifyPolicies[2], ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true

}
