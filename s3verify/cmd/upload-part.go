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
	crand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Store parts to be listed. Parts are grouped by object in sub-lists
var objectParts = [2][]objectPart{}

// Complete multipart upload.
var complMultipartUploads = []*completeMultipartUpload{
	&completeMultipartUpload{
	// To be filled out by the test.
	},
	&completeMultipartUpload{
	// To be filled out by the test.
	},
}

// newUploadPartReq - Create a new HTTP request for an upload part request.
func newUploadPartReq(bucketName, objectName, uploadID string, partNumber int, partData []byte) (Request, error) {
	// Create a new request for uploading a part.
	var uploadPartReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	uploadPartReq.bucketName = bucketName
	uploadPartReq.objectName = objectName

	// Set the query values.
	urlValues := make(url.Values)
	urlValues.Set("partNumber", strconv.Itoa(partNumber))
	urlValues.Set("uploadId", uploadID)
	uploadPartReq.queryValues = urlValues

	// Compute md5sum, sha256Sum and contentlength.
	reader := bytes.NewReader(partData)
	md5Sum, sha256Sum, contentLength, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the Header values and Body of request.
	uploadPartReq.contentBody = reader
	uploadPartReq.contentLength = contentLength
	uploadPartReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	uploadPartReq.customHeader.Set("Content-MD5", base64.StdEncoding.EncodeToString(md5Sum))
	uploadPartReq.customHeader.Set("User-Agent", appUserAgent)

	return uploadPartReq, nil
}

// uploadPartVerify - verify that the response returned matches what is expected.
func uploadPartVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyBodyUploadPart(res.Body); err != nil {
		return err
	}
	if err := verifyStatusUploadPart(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderUploadPart(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyBodyUploadPart - verify that that body returned is empty.
func verifyBodyUploadPart(resBody io.Reader) error {
	body, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	if !bytes.Equal(body, []byte("")) { // Body for PUT responses should be empty.
		err := fmt.Errorf("Unexpected Body Received: %v", string(body))
		return err
	}
	return nil
}

// verifyStatusUploadPart - verify that the status returned matches what is expected.
func verifyStatusUploadPart(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyHeaderUploadPart - verify that the header returned matches what is expected.
func verifyHeaderUploadPart(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainUploadPart - upload part test.
func mainUploadPart(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] Multipart (Upload-Part):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// All multipart objects created by s3verify will be stored in s3verify buckets.
	bucketName := s3verifyBuckets[0].Name
	// TODO: upload more than one part for at least one object.
	for i, object := range multipartObjects { // Upload 1 5MB or smaller part per object.
		// Spin scanBar
		scanBar(message)
		part := objectPart{}
		// Create some random data at most 5MB to upload via multipart operations.
		objectData := make([]byte, rand.Intn(1<<20)+4*1024*1024)
		part.PartNumber = 1
		part.Data = objectData
		part.Size = int64(len(objectData))
		_, err := io.ReadFull(crand.Reader, objectData)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Create a new multipart upload part request.
		req, err := newUploadPartReq(bucketName, object.Key, object.UploadID, 1, objectData)
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
		if err := uploadPartVerify(res, http.StatusOK); err != nil {
			printMessage(message, err)
			return false
		}
		// Update the ETag of the part.
		part.ETag = strings.TrimPrefix(res.Header.Get("ETag"), "\"")
		part.ETag = strings.TrimSuffix(part.ETag, "\"")
		// Store the parts to be listed in the list-multipart-uploads test.
		objectParts[i] = append(objectParts[i], part)
		// Test cleared store the uploaded parts to be completed/aborted.
		var complPart completePart
		complPart.ETag = part.ETag
		complPart.PartNumber = part.PartNumber
		// Save the completed part into the complMultiPartUpload struct.
		complMultipartUploads[i].Parts = append(complMultipartUploads[i].Parts, complPart)
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}

// mainReuploadPart - reupload the first part of a multipart upload operation
// initiated in previous tests
func mainReuploadPart(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] Multipart (Reupload-Part):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// All multipart objects created by s3verify will be stored in s3verify buckets.
	bucketName := s3verifyBuckets[0].Name
	object := multipartObjects[0]
	// Spin scanBar
	scanBar(message)
	part := objectPart{}
	// Create some random data at most 5MB to upload via multipart operations.
	objectData := make([]byte, rand.Intn(1<<20)+4*1024*1024)
	part.PartNumber = 1
	part.Size = int64(len(objectData))
	part.Data = objectData
	_, err := io.ReadFull(crand.Reader, objectData)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Create a new multipart upload part request.
	req, err := newUploadPartReq(bucketName, object.Key, object.UploadID, 1, objectData)
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
	if err := uploadPartVerify(res, http.StatusOK); err != nil {
		printMessage(message, err)
		return false
	}

	// At this point, we need to update data in global objectParts
	// and complMultipartUploads to make further tests work as expected
	var complPart completePart
	part.ETag = strings.TrimPrefix(res.Header.Get("ETag"), "\"")
	part.ETag = strings.TrimSuffix(part.ETag, "\"")
	complPart.ETag = part.ETag
	complPart.PartNumber = part.PartNumber

	complMultipartUploads[0].Parts[0] = complPart
	objectParts[0][0] = part

	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
