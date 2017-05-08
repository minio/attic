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
)

// newGetObjectIfMatchReq - Create a new HTTP request to perform.
func newGetObjectIfMatchReq(bucketName, objectName, ETag string) (Request, error) {
	var getObjectIfMatchReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName
	getObjectIfMatchReq.bucketName = bucketName
	getObjectIfMatchReq.objectName = objectName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because GET requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	getObjectIfMatchReq.customHeader.Set("If-Match", ETag)
	getObjectIfMatchReq.customHeader.Set("User-Agent", appUserAgent)
	getObjectIfMatchReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	return getObjectIfMatchReq, nil
}

// getObjectIfMatchVerify - Verify that the response matches what is expected.
func getObjectIfMatchVerify(res *http.Response, objectBody []byte, expectedStatusCode int, shouldFail bool) error {
	if err := verifyHeaderGetObjectIfMatch(res.Header); err != nil {
		return err
	}
	if err := verifyBodyGetObjectIfMatch(res.Body, objectBody, shouldFail); err != nil {
		return err
	}
	if err := verifyStatusGetObjectIfMatch(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	return nil
}

// verifyHeaderGetObjectIfMatch - Verify that the response header matches what is expected.
func verifyHeaderGetObjectIfMatch(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

//verifyBodyGetObjectIfMatch - Verify that the response body matches what is expected.
func verifyBodyGetObjectIfMatch(resBody io.Reader, objectBody []byte, shouldFail bool) error {
	if shouldFail {
		// Decode the supposed error response.
		errBody := ErrorResponse{}
		err := xmlDecoder(resBody, &errBody)
		if err != nil {
			return err
		}
		if errBody.Code != "PreconditionFailed" {
			err := fmt.Errorf("Unexpected Error Response: wanted PreconditionFailed, got %v", errBody.Code)
			return err
		}
	} else {
		// The body should be returned in full.
		body, err := ioutil.ReadAll(resBody)
		if err != nil {
			return err
		}
		if !shouldFail && !bytes.Equal(body, objectBody) { // Test should pass ensure body is what was uploaded.
			err := fmt.Errorf("Unexpected Body Received: wanted %v, got %v", string(objectBody), string(body))
			return err
		}
	}
	// Otherwise test failed / passed as expected.
	return nil
}

// verifyStatusGetObjectIfMatch - Verify that the response status matches what is expected.
func verifyStatusGetObjectIfMatch(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// Test the compatibility of the GET object API when using the If-Match header.
func mainGetObjectIfMatch(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] GetObject (If-Match):", curTest, globalTotalNumTest)
	// Set up an invalid ETag to test failed requests responses.
	invalidETag := "1234567890"
	// All getobject tests happen in s3verify created buckets
	// on s3verify created objects.
	bucketName := s3verifyBuckets[0].Name
	for _, object := range s3verifyObjects {
		// Spin scanBar
		scanBar(message)
		// Create new GET object If-Match request.
		req, err := newGetObjectIfMatchReq(bucketName, object.Key, object.ETag)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		res, err := config.execRequest("GET", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		defer closeResponse(res)
		// Verify the response...these checks do not check the headers yet.
		if err := getObjectIfMatchVerify(res, object.Body, http.StatusOK, false); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Create a bad GET object If-Match request.
		badReq, err := newGetObjectIfMatchReq(bucketName, object.Key, invalidETag)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		badRes, err := config.execRequest("GET", badReq)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		defer closeResponse(badRes)
		// Verify the request fails as expected.
		if err := getObjectIfMatchVerify(badRes, []byte(""), http.StatusPreconditionFailed, true); err != nil {
			printMessage(message, err)
			return false
		}
	}
	// Spin scanBar
	scanBar(message)
	printMessage(message, nil)
	return true
}
