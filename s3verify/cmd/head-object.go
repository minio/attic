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
	"strconv"
	"time"
)

// newHeadObjectReq - Create a new HTTP request for a HEAD object.
func newHeadObjectReq(bucketName, objectName string) (Request, error) {
	// headObjectReq - an HTTP request for HEAD with no headers set.
	var headObjectReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	headObjectReq.bucketName = bucketName
	headObjectReq.objectName = objectName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because HEAD requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	headObjectReq.customHeader.Set("User-Agent", appUserAgent)
	headObjectReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	return headObjectReq, nil
}

// headObjectVerify - Verify that the response received matches what is expected.
func headObjectVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyStatusHeadObject(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderHeadObject(res.Header); err != nil {
		return err
	}
	if err := verifyBodyHeadObject(res.Body); err != nil {
		return err
	}
	return nil
}

// verifyStatusHeadObject - Verify that the status received matches what is expected.
func verifyStatusHeadObject(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyHeadObject - Verify that the body received is empty.
func verifyBodyHeadObject(resBody io.Reader) error {
	body, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Received: HEAD requests should not return a body, but got back: %v", string(body))
		return err
	}
	return nil
}

// verifyHeaderHeadObject - Verify that the header received matches what is exepected.
func verifyHeaderHeadObject(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	// TODO: add verification for ETag formation.
	return nil
}

// mainHeadObject - test the HeadObject API with no header set.
func mainHeadObject(config ServerConfig, curTest int, testObjects []*ObjectInfo, bucketName string) bool {
	message := fmt.Sprintf("[%02d/%d] HeadObject:", curTest, globalTotalNumTest)
	// All headobject tests are run in s3verify buckets on s3verify created objects.
	for _, object := range testObjects {
		// Spin scanBar
		scanBar(message)
		// Create a new HEAD object with no headers.
		req, err := newHeadObjectReq(bucketName, object.Key)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Execute the request.
		res, err := config.execRequest("HEAD", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Verify the response.
		if err := headObjectVerify(res, http.StatusOK); err != nil {
			printMessage(message, err)
			return false
		}
		// If the verification is valid then set the ETag, Size, and LastModified.
		// No need to canonicalize ETags because they will come back uncanonicalized every time.
		eTag := res.Header.Get("ETag")
		date, err := time.Parse(http.TimeFormat, res.Header.Get("Last-Modified")) // This will never error out because it has already been verified.
		if err != nil {
			printMessage(message, err)
			return false
		}
		size, err := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			printMessage(message, err)
			return false
		}
		object.Size = size
		object.ETag = eTag
		object.LastModified = date
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}

// mainHeadObjectUnPrepared - Test for HeadObject API when the environment was not previously created.
func mainHeadObjectUnPrepared(config ServerConfig, curTest int) bool {
	bucketName := s3verifyBuckets[0].Name
	testObjects := s3verifyObjects
	return mainHeadObject(config, curTest, testObjects, bucketName)
}

// mainHeadObjectPrepared - Test for HeadObject API when the environment was prepared.
func mainHeadObjectPrepared(config ServerConfig, curTest int) bool {
	bucketName := preparedBuckets[0].Name
	testObjects := preparedObjects
	return mainHeadObject(config, curTest, testObjects, bucketName)
}
