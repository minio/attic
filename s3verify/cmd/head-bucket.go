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

// newHeadBucketReq - Create a new HTTP request for the HeadBucket API.
func newHeadBucketReq(bucketName string) (Request, error) {
	// headBucketReq - a new HTTP request for the HeadBucket API.
	var headBucketReq = Request{
		customHeader: http.Header{},
	}
	// Set the bucketName.
	headBucketReq.bucketName = bucketName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because HEAD requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	headBucketReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	headBucketReq.customHeader.Set("User-Agent", appUserAgent)

	return headBucketReq, nil
}

// headBucketVerify - verify the response returned matches what is expected.
func headBucketVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyBodyHeadBucket(res.Body); err != nil {
		return err
	}
	if err := verifyStatusHeadBucket(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderHeadBucket(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyBodyHeadBucket - verify the body returned matches what is expected.
func verifyBodyHeadBucket(resBody io.Reader) error {
	// Verify that the body returned is empty.
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

// verifyHeaderHeadBucket - verify the header returned matches what is expected.
func verifyHeaderHeadBucket(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyStatusHeadBucket - verify the status returned matches what is expected.
func verifyStatusHeadBucket(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// mainHeadBucket - test the HeadBucket API.
func mainHeadBucket(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] HeadBucket:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	bucketName := s3verifyBuckets[0].Name
	// Create a new HeadBucket request.
	req, err := newHeadBucketReq(bucketName)
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
	if err := headBucketVerify(res, http.StatusOK); err != nil {
		printMessage(message, err)
		return false
	}
	// Test passed.
	printMessage(message, nil)
	return true
}
