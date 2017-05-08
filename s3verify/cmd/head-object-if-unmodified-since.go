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
	"time"
)

// newHeadObjectIfUnModifiedReq - Create a new HTTP request for HEAD object with if-unmodified-since header set.
func newHeadObjectIfUnModifiedSinceReq(bucketName, objectName string, lastModified time.Time) (Request, error) {
	// headObjectIfUnModifiedReq - a new HTTP request for HEAD object with if-unmodified-since header set.
	var headObjectIfUnModifiedSinceReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName
	headObjectIfUnModifiedSinceReq.bucketName = bucketName
	headObjectIfUnModifiedSinceReq.objectName = objectName

	// No body is sent with a HEAD request.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}
	headObjectIfUnModifiedSinceReq.customHeader.Set("If-Unmodified-Since", lastModified.Format(http.TimeFormat))
	headObjectIfUnModifiedSinceReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	headObjectIfUnModifiedSinceReq.customHeader.Set("User-Agent", appUserAgent)

	return headObjectIfUnModifiedSinceReq, nil
}

// headObjectIfUnModifiedSinceVerify - verify the response returned matches what is expected.
func headObjectIfUnModifiedSinceVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyStatusHeadObjectIfUnModifiedSince(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyHeadObjectIfUnModifiedSince(res.Body); err != nil {
		return err
	}
	if err := verifyHeaderHeadObjectIfUnModifiedSince(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyStatusHeadObjectIfUnModifiedSince - verify the status returned matches what is expected.
func verifyStatusHeadObjectIfUnModifiedSince(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyHeadObjectIfUnModifiedSince - verify the body returned is empty.
func verifyBodyHeadObjectIfUnModifiedSince(resBody io.Reader) error {
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

// verifyHeaderHeadObjectIfUnModifiedSince - verify that the header returned matches what is expected.
func verifyHeaderHeadObjectIfUnModifiedSince(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainHeadObjectIfUnModifiedSince - HEAD object with if-unmodified-since header set test.
func mainHeadObjectIfUnModifiedSince(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] HeadObject (If-Unmodified-Since):", curTest, globalTotalNumTest)
	scanBar(message)
	// Create a date in the past to use.
	lastModified, err := time.Parse(http.TimeFormat, "Thu, 01 Jan 1970 00:00:00 GMT")
	if err != nil {
		printMessage(message, err)
		return false
	}
	// All headobject if-unmodified-since tests happen in s3verify created buckets
	// on s3verify created objects.
	bucketName := s3verifyBuckets[0].Name
	object := s3verifyObjects[0]
	// Create a new request.
	req, err := newHeadObjectIfUnModifiedSinceReq(bucketName, object.Key, object.LastModified)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Perform the request.
	res, err := config.execRequest("HEAD", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)
	// Verify the request succeeds as expected.
	if err := headObjectIfUnModifiedSinceVerify(res, http.StatusOK); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Create a bad request.
	badReq, err := newHeadObjectIfUnModifiedSinceReq(bucketName, object.Key, lastModified)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Perform the bad request.
	badRes, err := config.execRequest("HEAD", badReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(badRes)
	// Spin scanBar
	scanBar(message)
	// Verify the response failed.
	if err := headObjectIfUnModifiedSinceVerify(badRes, http.StatusPreconditionFailed); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
