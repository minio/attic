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

// newHeadObjectIfModifiedSinceReq - Create a new HTTP request for HEAD object with if-modified-since header set.
func newHeadObjectIfModifiedSinceReq(bucketName, objectName string, lastModified time.Time) (Request, error) {
	// headObjectIfModifiedSinceReq - a new HTTP request for HEAD object with if-modified-since header set.
	var headObjectIfModifiedSinceReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	headObjectIfModifiedSinceReq.bucketName = bucketName
	headObjectIfModifiedSinceReq.objectName = objectName

	// No body is sent with HEAD request.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}
	headObjectIfModifiedSinceReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	headObjectIfModifiedSinceReq.customHeader.Set("If-Modified-Since", lastModified.Format(http.TimeFormat))
	headObjectIfModifiedSinceReq.customHeader.Set("User-Agent", appUserAgent)

	return headObjectIfModifiedSinceReq, nil
}

// headObjectIfModifiedSinceVerify - verify the response returned matches what is expected.
func headObjectIfModifiedSinceVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyStatusHeadObjectIfModifiedSince(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderHeadObjectIfModifiedSince(res.Header); err != nil {
		return err
	}
	if err := verifyBodyHeadObjectIfModifiedSince(res.Body); err != nil {
		return err
	}
	return nil
}

// verifyStatusHeadObjectIfModifiedSince - verify the status returned matches what is expected.
func verifyStatusHeadObjectIfModifiedSince(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyHeadObjectIfModifiedSince - verify the body returned is empty.
func verifyBodyHeadObjectIfModifiedSince(resBody io.Reader) error {
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

// verifyHeaderHeadObjectIfModifiedSince - verify the header returned matches what is expected.
func verifyHeaderHeadObjectIfModifiedSince(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainHeadObjectIfModifiedSince - test the HeadObject with the If-Modified-Since.
func mainHeadObjectIfModifiedSince(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] HeadObject (If-Modified-Since):", curTest, globalTotalNumTest)
	lastModified, err := time.Parse(http.TimeFormat, "Thu, 01 Jan 1970 00:00:00 GMT")
	if err != nil {
		printMessage(message, err)
		return false
	}
	// All headobject if-modified-since tests happen in s3verify created buckets
	// on s3verify created objects.
	bucketName := s3verifyBuckets[0].Name
	object := s3verifyObjects[0]
	// Spin scanBar
	scanBar(message)
	// Create a new request.
	req, err := newHeadObjectIfModifiedSinceReq(bucketName, object.Key, lastModified)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := config.execRequest("HEAD", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := headObjectIfModifiedSinceVerify(res, http.StatusOK); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Create a bad request.
	badReq, err := newHeadObjectIfModifiedSinceReq(bucketName, object.Key, object.LastModified.Add(time.Hour*2))
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the bad request.
	badRes, err := config.execRequest("HEAD", badReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(badRes)
	// Spin scanBar
	scanBar(message)
	// Verify the bad request failed as expected.
	if err := headObjectIfModifiedSinceVerify(badRes, 304); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	printMessage(message, nil)
	return true
}
