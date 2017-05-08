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
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// newCompleteMultipartUploadReq - Create a new Request for complete-multipart API.
func newCompleteMultipartUploadReq(bucketName, objectName, uploadID string, complete *completeMultipartUpload) (Request, error) {
	// completeMultipartUploadReq - a new Request for complete-multipart API.
	var completeMultipartUploadReq = Request{
		customHeader: http.Header{},
	}
	// Set the bucketName and objectName
	completeMultipartUploadReq.bucketName = bucketName
	completeMultipartUploadReq.objectName = objectName

	// Initialize url queries.
	urlValues := make(url.Values)
	urlValues.Set("uploadId", uploadID)
	completeMultipartUploadReq.queryValues = urlValues

	completeMultipartUploadBytes, err := xml.Marshal(complete)
	if err != nil {
		return Request{}, err
	}

	reader := bytes.NewReader(completeMultipartUploadBytes)
	// Compute sha256Sum and contentLength.
	_, sha256Sum, contentLength, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the Body, Header, ContentLength of the request.
	completeMultipartUploadReq.contentLength = contentLength
	completeMultipartUploadReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	completeMultipartUploadReq.customHeader.Set("User-Agent", appUserAgent)
	completeMultipartUploadReq.contentBody = reader

	return completeMultipartUploadReq, nil
}

// TODO: So far only valid multipart requests are used. Implement tests that SHOULD fail.
//
// completeMultipartUploadVerify - verify tthat the response returned matches what is expected.
func completeMultipartUploadVerify(res *http.Response, expectedStatusCode int, bucketName, objectKey string) error {
	if err := verifyStatusCompleteMultipartUpload(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyCompleteMultipartUpload(res.Body, bucketName, objectKey); err != nil {
		return err
	}
	if err := verifyHeaderCompleteMultipartUpload(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyStatusCompleteMultipartUpload - verify the status returned matches what is expected.
func verifyStatusCompleteMultipartUpload(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyCompleteMultipartUpload - verify the body returned matches what is expected.
func verifyBodyCompleteMultipartUpload(resBody io.Reader, bucketName, objectKey string) error {
	resCompleteMultipartUploadResult := completeMultipartUploadResult{}
	if err := xmlDecoder(resBody, &resCompleteMultipartUploadResult); err != nil {
		return err
	}
	if resCompleteMultipartUploadResult.Bucket != bucketName {
		return fmt.Errorf("Wrong bucket in Complete Multipart XML result, expected: %s, received: %s",
			resCompleteMultipartUploadResult.Bucket, bucketName)
	}
	if resCompleteMultipartUploadResult.Key != objectKey {
		return fmt.Errorf("Wrong key in Complete Multipart XML result, expected: %s, received: %s",
			resCompleteMultipartUploadResult.Key, objectKey)
	}
	return nil
}

// verifyHeaderCompleteMultipartUpload - verify the header returned matches what is expected.
func verifyHeaderCompleteMultipartUpload(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainCompleteMultipartUpload - Complete Multipart Upload API test.
func mainCompleteMultipartUpload(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] Multipart (Complete-Upload):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	bucketName := s3verifyBuckets[0].Name
	object := multipartObjects[0]
	// Create a new completeMultipartUpload request.
	req, err := newCompleteMultipartUploadReq(bucketName, object.Key, object.UploadID, complMultipartUploads[0])
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := config.execRequest("POST", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := completeMultipartUploadVerify(res, http.StatusOK, bucketName, object.Key); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	printMessage(message, nil)
	return true
}
