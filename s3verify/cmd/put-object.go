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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// Store all objects that are uploaded by s3verify tests.
var s3verifyObjects = []*ObjectInfo{}

// Store all objects that are uploaded through the preparing operation.
var preparedObjects = []*ObjectInfo{}

// Store all objects that were copied.
var copyObjects = []*ObjectInfo{}

// newPutObjectReq - Create a new HTTP request for PUT object.
func newPutObjectReq(bucketName, objectName string, objectData []byte) (Request, error) {
	// An HTTP request for a PUT object.
	var putObjectReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	putObjectReq.bucketName = bucketName
	putObjectReq.objectName = objectName

	// Compute md5Sum and sha256Sum from the input data.
	reader := bytes.NewReader(objectData)
	md5Sum, sha256Sum, contentLength, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	putObjectReq.customHeader.Set("Content-MD5", base64.StdEncoding.EncodeToString(md5Sum))
	putObjectReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	putObjectReq.customHeader.Set("User-Agent", appUserAgent)

	putObjectReq.contentLength = contentLength
	// Set the body to the data held in objectData.
	putObjectReq.contentBody = reader

	return putObjectReq, nil
}

// calculateSignedChunkLength - calculates the length of chunk metadata
func calculateSignedChunkLength(chunkDataSize int64) int64 {
	return int64(len(fmt.Sprintf("%x", chunkDataSize))) +
		17 + // ";chunk-signature="
		64 + // e.g. "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2"
		2 + // CRLF
		chunkDataSize +
		2 // CRLF
}

// calculateStreamContentLength - calculates the length of the overall stream (data + metadata)
func calculateStreamContentLength(dataLen, chunkSize int64) int64 {
	if dataLen <= 0 {
		return 0
	}
	chunksCount := int64(dataLen / chunkSize)
	remainingBytes := int64(dataLen % chunkSize)
	streamLen := int64(0)
	streamLen += chunksCount * calculateSignedChunkLength(chunkSize)
	if remainingBytes > 0 {
		streamLen += calculateSignedChunkLength(remainingBytes)
	}
	streamLen += calculateSignedChunkLength(0)
	return streamLen
}

// newPutObjectStreamingReq - Create a new HTTP streaming request for PUT object.
func newPutObjectStreamingReq(config ServerConfig, bucketName, objectName string, objectData []byte) (Request, error) {
	// An HTTP request for a PUT object.
	var putObjectReq = Request{
		customHeader:  http.Header{},
		streamingSign: true,
		chunkSize:     64 * 1024,
	}

	// Set the bucketName and objectName.
	putObjectReq.bucketName = bucketName
	putObjectReq.objectName = objectName

	dataLen := int64(len(objectData))
	contentLength := calculateStreamContentLength(dataLen, putObjectReq.chunkSize)

	putObjectReq.customHeader.Set("User-Agent", appUserAgent)
	putObjectReq.customHeader.Set("x-amz-content-sha256", "STREAMING-AWS4-HMAC-SHA256-PAYLOAD")
	putObjectReq.customHeader.Set("content-encoding", "aws-chunked")
	putObjectReq.customHeader.Set("x-amz-decoded-content-length", strconv.FormatInt(dataLen, 10))
	putObjectReq.customHeader.Set("Content-Length", strconv.FormatInt(contentLength, 10))

	// Set the body to the data held in objectData.
	reader := bytes.NewReader(objectData)
	putObjectReq.contentBody = reader
	putObjectReq.contentLength = contentLength

	return putObjectReq, nil
}

// putObjectVerify - Verify the response matches what is expected.
func putObjectVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyHeaderPutObject(res.Header); err != nil {
		return err
	}
	if err := verifyStatusPutObject(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyPutObject(res.Body); err != nil {
		return err
	}
	return nil
}

// verifyStatusPutObject - Verify that the res.StatusCode code matches what is expected.
func verifyStatusPutObject(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyPutObject - Verify that the body returned matches what is uploaded.
func verifyBodyPutObject(resBody io.Reader) error {
	body, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	// A PUT request should give back an empty body.
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Received: expected empty body but received: %v", string(body))
		return err
	}
	return nil
}

// verifyHeaderPutObject - Verify that the header returned matches what is expected.
func verifyHeaderPutObject(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainPutObjectPrepared - Test the PutObject API in a prepared environment.
func mainPutObjectPrepared(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] PutObject:", curTest, globalTotalNumTest)
	// Use the last bucket created by s3verify itself.
	bucket := s3verifyBuckets[0]
	// Spin scanBar
	scanBar(message)
	// Since the use of --prepare will have set up enough objects for future tests
	// only add one more additional object.
	object := &ObjectInfo{
		Key:  "s3verify/made/put/object",
		Body: []byte(randString(60, rand.NewSource(time.Now().UnixNano()), "")),
	}
	// Spin scanBar
	scanBar(message)
	// Create a new request.
	req, err := newPutObjectReq(bucket.Name, object.Key, object.Body)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := config.execRequest("PUT", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	defer closeResponse(res)
	// Verify the response.
	if err := putObjectVerify(res, http.StatusOK); err != nil {
		printMessage(message, err)
		return false
	}
	// Store this object in the global objects list.
	s3verifyObjects = append(s3verifyObjects, object)
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}

// Test a PUT object request with no special headers set. This adds one object to each of the test buckets.
func mainPutObjectUnPrepared(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] PutObject:", curTest, globalTotalNumTest)
	// TODO: create tests designed to fail.
	bucket := s3verifyBuckets[0]
	// Spin scanBar
	scanBar(message)
	// TODO: need to update to 1001 once this is production ready.
	// Upload 1001 objects with 1 byte each to check the ListObjects API with.
	for i := 0; i < globalNumTestObjects; i++ {
		// Spin scanBar
		scanBar(message)
		object := &ObjectInfo{}
		object.Key = "s3verify/put/object/" + strconv.Itoa(i)
		// Create 60 bytes worth of random data for each object.
		body := randString(60, rand.NewSource(time.Now().UnixNano()), "")
		object.Body = []byte(body)
		// Create a new request.
		req, err := newPutObjectReq(bucket.Name, object.Key, object.Body)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Execute the request.
		res, err := config.execRequest("PUT", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Verify the response.
		if err := putObjectVerify(res, http.StatusOK); err != nil {
			printMessage(message, err)
			return false
		}
		// Add the new object to the list of objects.
		s3verifyObjects = append(s3verifyObjects, object)
		// Spin scanBar
		scanBar(message)
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}

// Test a PUT object streaming request with no special headers set. This adds one object to each of the test buckets.
func mainPutObjectStream(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] PutObject (Streaming):", curTest, globalTotalNumTest)
	// TODO: create tests designed to fail.
	bucket := s3verifyBuckets[0]
	// Spin scanBar
	scanBar(message)
	object := &ObjectInfo{}
	// Only need to upload one new object.
	object.Key = "s3verify/put/object/stream"
	// Create 60 bytes worth of random data for each object.
	body := randString(60, rand.NewSource(time.Now().UnixNano()), "")
	object.Body = []byte(body)
	// Create a new request.
	req, err := newPutObjectStreamingReq(config, bucket.Name, object.Key, object.Body)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the request.
	res, err := config.execRequest("PUT", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Verify the response.
	if err := putObjectVerify(res, http.StatusOK); err != nil {
		printMessage(message, err)
		return false
	}
	// Add the new object to the list of objects.
	s3verifyObjects = append(s3verifyObjects, object)

	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
