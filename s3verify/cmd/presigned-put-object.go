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
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// newPresignedPutObjectReq - Create a new Request for PUT object requests using presigned URLs.
func newPresignedPutObjectReq(config ServerConfig, bucketName, objectName string, expires time.Duration) (*url.URL, error) {
	// presignedPutObjectReq - a new request with presigned URL for PUT object requests.
	var presignedPutObjectReq = Request{
		customHeader: http.Header{},
	}

	// Indicate presigned.
	presignedPutObjectReq.presignURL = true

	// Set the bucketName and objectName.
	presignedPutObjectReq.bucketName = bucketName
	presignedPutObjectReq.objectName = objectName

	// Set the expiry.
	expireSeconds := int64(expires / time.Second)
	presignedPutObjectReq.expires = expireSeconds

	// Extract the url from the Request.
	req, err := config.newRequest("PUT", presignedPutObjectReq)
	if err != nil {
		return nil, err
	}

	return req.URL, nil
}

// presignedPutObjectVerify - verify the response returned matches what is expected.
func presignedPutObjectVerify(res *http.Response, expectedStatusCode int, expectedError ErrorResponse) error {
	if err := verifyStatusPresignedPutObject(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderPresignedPutObject(res.Header); err != nil {
		return err
	}
	if err := verifyBodyPresignedPutObject(res.Body, expectedError); err != nil {
		return err
	}
	return nil
}

// verifyStatusPresignedPutObject - verify the status returned matches what is expected.
func verifyStatusPresignedPutObject(respStatusCode int, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Code Received: wanted %d, got %d", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyHeaderPresignedPutObject - verify the header returned matches what is expected.
func verifyHeaderPresignedPutObject(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyBodyPresignedPutObject - verify the body returned matches what is expected.
func verifyBodyPresignedPutObject(resBody io.Reader, expectedError ErrorResponse) error {
	if expectedError.Message != "" {
		receivedError := ErrorResponse{}
		err := xmlDecoder(resBody, &receivedError)
		if err != nil {
			return err
		}
		if receivedError.Message != expectedError.Message {
			err := fmt.Errorf("Unexpected Error Message Received: wanted %s, got %s", expectedError.Message, receivedError.Message)
			return err
		}
		return nil
	}
	receivedBody, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	if !bytes.Equal(receivedBody, []byte{}) {
		err := fmt.Errorf("Unexpected Body Received: %q", receivedBody)
		return err
	}
	return nil
}

// THIS MIGHT CAUSE PROBLEMS HAVE TO CHECK LIST OBJECTS AFTER THIS IS DONE.
// mainPresignedPutObject - test the compatibility of the presigned PutObject API.
func mainPresignedPutObject(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] PutObject (Presigned):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// New objects will only be added to s3verify created buckets.
	bucketName := s3verifyBuckets[0].Name
	// Prefix this object differently to allow ListObjects to function easier.
	objectName := randString(60, rand.NewSource(time.Now().UnixNano()), "s3verify/presigned/object/00")

	presignedObject := &ObjectInfo{
		Key:  objectName,
		Body: []byte(randString(60, rand.NewSource(time.Now().UnixNano()), "a")),
	}
	reader := bytes.NewReader(presignedObject.Body)

	// Create a new presigned PUT URL.
	reqURL, err := newPresignedPutObjectReq(config, bucketName, objectName, time.Second*5)
	if err != nil {
		printMessage(message, err)
		return false
	}

	// Create a new http Request out of the URL.
	req, err := http.NewRequest("PUT", reqURL.String(), reader)
	if err != nil {
		printMessage(message, err)
		return false
	}

	// Execute the request.
	res, err := config.Client.Do(req)
	if err != nil {
		printMessage(message, err)
		return false
	}

	// Verify the response.
	if err := presignedPutObjectVerify(res, 200, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}

	// Store the newly created object.
	s3verifyObjects = append(s3verifyObjects, presignedObject)

	// Test passed.
	printMessage(message, nil)
	return true
}
