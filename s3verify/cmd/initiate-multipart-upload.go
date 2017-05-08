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
	"net/http"
	"net/url"
)

// Holds all the objects to be uploaded by a multipart request.
var multipartObjects = []*ObjectInfo{
	// An object that will have more than 5MB of data to be uploaded as part of a multipart upload.
	&ObjectInfo{
		Key:         "s3verify/multipart/object",
		ContentType: "application/octet-stream",
		// Body: to be set dynamically,
		// UploadID: to be set dynamically,
	},
	&ObjectInfo{
		Key:         "s3verify/multipart/abort",
		ContentType: "application/octet-stream",
		// Body: to be set dynamically,
		// UploadID: to be set dynamically,
	},
}

// newInitiateMultipartUploadReq - Create a new HTTP request for the initiate-multipart-upload API.
func newInitiateMultipartUploadReq(bucketName, objectName string) (Request, error) {
	// Initialize url queries.
	urlValues := make(url.Values)
	urlValues.Set("uploads", "")
	// An HTTP request for a multipart upload.
	var initiateMultipartUploadReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName
	initiateMultipartUploadReq.bucketName = bucketName
	initiateMultipartUploadReq.objectName = objectName

	// Set the query values.
	initiateMultipartUploadReq.queryValues = urlValues

	// No body is sent with initiate-multipart requests.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	initiateMultipartUploadReq.customHeader.Set("User-Agent", appUserAgent)
	initiateMultipartUploadReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	return initiateMultipartUploadReq, nil
}

// initiateMultipartUploadVerify - verify that the response returned matches what is expected.
func initiateMultipartUploadVerify(res *http.Response, expectedStatusCode int) (string, error) {
	uploadID, err := verifyBodyInitiateMultipartUpload(res.Body)
	if err != nil {
		return uploadID, err
	}
	if err := verifyHeaderInitiateMultipartUpload(res.Header); err != nil {
		return uploadID, err
	}
	if err := verifyStatusInitiateMultipartUpload(res.StatusCode, expectedStatusCode); err != nil {
		return uploadID, err
	}
	return uploadID, nil
}

// verifyStatusInitiateMultipartUpload - verify that the status returned matches what is expected.
func verifyStatusInitiateMultipartUpload(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyInitiateMultipartUpload - verify that the body returned matches what is expected.
func verifyBodyInitiateMultipartUpload(resBody io.Reader) (string, error) {
	resInitiateMultipartUpload := initiateMultipartUploadResult{}
	if err := xmlDecoder(resBody, &resInitiateMultipartUpload); err != nil {
		return "", err
	}
	// Body was sent set the object UploadID.
	uploadID := resInitiateMultipartUpload.UploadID
	return uploadID, nil
}

// verifyHeaderInitiateMultipartUpload - verify that the header returned matches what is expected.
func verifyHeaderInitiateMultipartUpload(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainInitiateMultipartUpload - initiate multipart upload test.
func mainInitiateMultipartUpload(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] Multipart (Initiate-Upload):", curTest, globalTotalNumTest)
	// Spin scanBar.
	scanBar(message)
	// All initiate-multipart tests happen in s3verify created buckets.
	bucketName := s3verifyBuckets[0].Name
	// Get the bucket to upload to and the objectName to call the new upload.
	for _, object := range multipartObjects {
		// Spin scanBar
		scanBar(message)
		// Create a new InitiateMultiPartUpload request.
		req, err := newInitiateMultipartUploadReq(bucketName, object.Key)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Execute the request.
		res, err := config.execRequest("POST", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Verify the response and get the uploadID.
		uploadID, err := initiateMultipartUploadVerify(res, http.StatusOK)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Set the uploadId of the uploaded object.
		object.UploadID = uploadID
		// Spin scanBar
		scanBar(message)
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
