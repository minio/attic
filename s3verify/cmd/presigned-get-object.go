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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// newGetObjectPresignedReq - create a new request for GetObject with a presigned URL.
func newGetObjectPresignedReq(config ServerConfig, bucketName, objectName string, expires time.Duration, requestParameters url.Values) (*url.URL, error) {
	// getObjectPresignedReq - represents a request for GetObject with a presigned URL
	var getObjectPresignedReq = Request{
		customHeader: http.Header{},
	}
	// Set the bucketName and objectName.
	getObjectPresignedReq.bucketName = bucketName
	getObjectPresignedReq.objectName = objectName

	// Identify as presigned.
	getObjectPresignedReq.presignURL = true

	// Set the expiration.
	expireSeconds := int64(expires / time.Second)
	getObjectPresignedReq.expires = expireSeconds

	// Set the request parameters.
	getObjectPresignedReq.queryValues = requestParameters

	req, err := config.newRequest("GET", getObjectPresignedReq)
	if err != nil {
		return nil, err
	}

	return req.URL, nil
}

// getObjectPresignedVerify - verify the response returned matches what is expected.
func getObjectPresignedVerify(res *http.Response, expectedStatusCode int, expectedBody []byte, expectedError ErrorResponse) error {
	if err := verifyBodyGetObjectPresigned(res.Body, expectedBody, expectedError); err != nil {
		return err
	}
	if err := verifyStatusGetObjectPresigned(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderGetObjectPresigned(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyBodyGetObjectPresigned - verify the body returned matches what is expected.
func verifyBodyGetObjectPresigned(resBody io.Reader, expectedBody []byte, expectedError ErrorResponse) error {
	if expectedError.Message != "" {
		receivedError := ErrorResponse{}
		err := xmlDecoder(resBody, &receivedError)
		if err != nil {
			return err
		}
		if expectedError.Message != receivedError.Message {
			err := fmt.Errorf("Unexpected Error Message Received: wanted %s, got %s", expectedError.Message, receivedError.Message)
			return err
		}
		return nil
	}
	// Capture the body stream.
	buf := new(bytes.Buffer)
	buf.ReadFrom(resBody)
	if !bytes.Equal(buf.Bytes(), expectedBody) {
		err := fmt.Errorf("Unexpected Body Received: wanted %q, got %q", expectedBody, buf.Bytes())
		return err
	}
	return nil
}

// verifyStatusGetObjectPresigned - verify the status returned matches what is expected.
func verifyStatusGetObjectPresigned(respStatusCode int, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyHeaderGetObjectPresigned - verify the header returned matches what is expected.
func verifyHeaderGetObjectPresigned(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainGetObjectPresigned - test the compliance of the GetObject API using presigned URLs.
func mainGetObjectPresigned(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] GetObject (Presigned):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// Save an expired presigned url for testing the error response.
	var expiredURL *url.URL
	// Presigned getobject will only be tested in s3verify created buckets
	// on s3verify created objects.
	bucketName := s3verifyBuckets[0].Name
	testObject := s3verifyObjects[0]
	// Spin scanBar
	scanBar(message)
	// Create a new presigned GetObject req.
	// TODO: so far these requests do not use request/response parameters.
	reqURL, err := newGetObjectPresignedReq(config, bucketName, testObject.Key, time.Second*5, nil)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Store the created URL and make sure it expires later.
	expiredURL = reqURL
	// Execute the request.
	res, err := config.Client.Get(reqURL.String())
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Verify the response.
	if err := getObjectPresignedVerify(res, http.StatusOK, testObject.Body, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Make sure the saved URL has expired.
	time.Sleep(time.Second * 5)
	// Create the expected error.
	expectedError := ErrorResponse{
		Message: "Request has expired",
	}
	// Spin scanBar
	scanBar(message)
	// Attempt to use the expired url.
	badRes, err := config.Client.Get(expiredURL.String())
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(badRes)
	// Verify that this badRes failed as expected.
	if err := getObjectPresignedVerify(badRes, http.StatusForbidden, testObject.Body, expectedError); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
