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
	"math/rand"
	"net/http"
	"time"
)

// newRemoveObjectReq - Create a new DELETE object HTTP request.
func newRemoveObjectReq(bucketName, objectName string) (Request, error) {
	var removeObjectReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	removeObjectReq.bucketName = bucketName
	removeObjectReq.objectName = objectName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because DELETE requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	removeObjectReq.customHeader.Set("User-Agent", appUserAgent)
	removeObjectReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	return removeObjectReq, nil
}

// removeObjectVerify - Verify that the response returned matches what is expected.
func removeObjectVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyHeaderRemoveObject(res.Header); err != nil {
		return err
	}
	if err := verifyBodyRemoveObject(res.Body); err != nil {
		return err
	}
	if err := verifyStatusRemoveObject(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	return nil
}

// verifyHeaderRemoveObject - Verify that header returned matches what is expected.
func verifyHeaderRemoveObject(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyBodyRemoveObject - Verify that the body returned is empty.
func verifyBodyRemoveObject(resBody io.Reader) error {
	body, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Received: %v", string(body))
		return err
	}
	return nil
}

// verifyStatusRemoveObject - Verify that the status returned matches what is expected.
func verifyStatusRemoveObject(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %d, got %d", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// mainRemoveObjectExists - RemoveObject API test when object exists.
func mainRemoveObjectExists(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%d/%d] RemoveObject:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// Only remove objects from s3verify created buckets.
	// Only remove s3verify created objects.
	// First remove all PutObject test created objects.
	bucketName := s3verifyBuckets[0].Name
	for _, object := range s3verifyObjects {
		// Spin scanBar
		scanBar(message)
		// Create a new request.
		req, err := newRemoveObjectReq(bucketName, object.Key)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Execute the request.
		res, err := config.execRequest("DELETE", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Verify the response.
		if err := removeObjectVerify(res, http.StatusNoContent); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
	}
	// Remove all MultipartObject test objects.
	for _, object := range multipartObjects {
		// Spin scanBar
		scanBar(message)
		// Create a new request.
		req, err := newRemoveObjectReq(bucketName, object.Key)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Execute the request.
		res, err := config.execRequest("DELETE", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Verify the response.
		if err := removeObjectVerify(res, http.StatusNoContent); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
	}

	// Remove all copied objects. These exist in a different bucket.
	bucketName = s3verifyBuckets[1].Name
	for _, object := range copyObjects {
		// Spin scanBar
		scanBar(message)
		// Create a new request.
		req, err := newRemoveObjectReq(bucketName, object.Key)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Execute the request.
		res, err := config.execRequest("DELETE", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Verify the response.
		if err := removeObjectVerify(res, http.StatusNoContent); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
	}
	// Test passed.
	printMessage(message, nil)
	return true
}

// mainRemoveObjectDNE - Test the RemoveObject API when the object does not exist.
func mainRemoveObjectDNE(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] RemoveObject (Object DNE):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	bucketName := s3verifyBuckets[0].Name
	object := ObjectInfo{
		Key: randString(60, rand.NewSource(time.Now().UnixNano()), ""),
	}
	// Create a new request.
	req, err := newRemoveObjectReq(bucketName, object.Key)
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
	// Verify the response.
	if err := removeObjectVerify(res, http.StatusNoContent); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
