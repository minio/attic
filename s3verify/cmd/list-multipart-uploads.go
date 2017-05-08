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

// newListMultipartUploadsReq - Create a new HTTP request for List Multipart Uploads API.
func newListMultipartUploadsReq(bucketName string) (Request, error) {
	// listMultipartUploadsReq - a new HTTP request for the List Multipart Uploads API.
	var listMultipartUploadsReq = Request{
		customHeader: http.Header{},
	}

	// Set the query values.
	urlValues := make(url.Values)
	urlValues.Set("uploads", "")
	listMultipartUploadsReq.queryValues = urlValues

	// Set the bucketName.
	listMultipartUploadsReq.bucketName = bucketName

	// No body is sent with GET requests.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set Header values.
	listMultipartUploadsReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	listMultipartUploadsReq.customHeader.Set("User-Agent", appUserAgent)

	return listMultipartUploadsReq, nil
}

// listMultipartUploadsVerify - Verify that the response returned matches what is expected.
func listMultipartUploadsVerify(res *http.Response, expectedStatusCode int, expectedList listMultipartUploadsResult) error {
	if err := verifyHeaderListMultipartUploads(res.Header); err != nil {
		return err
	}
	if err := verifyBodyListMultipartUploads(res.Body, expectedList); err != nil {
		return err
	}
	if err := verifyStatusListMultipartUploads(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	return nil
}

// verifyHeaderListMultipartUploads - verify the header returned matches what is expected.
func verifyHeaderListMultipartUploads(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyStatusListMultipartUploads - verify the status returned matches what is expected.
func verifyStatusListMultipartUploads(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyListMultipartUploads - verify the body returned matches what is expected.
func verifyBodyListMultipartUploads(resBody io.Reader, expectedList listMultipartUploadsResult) error {
	receivedList := listMultipartUploadsResult{}
	err := xmlDecoder(resBody, &receivedList)
	if err != nil {
		return err
	}
	if len(receivedList.Uploads) != len(expectedList.Uploads) {
		err := fmt.Errorf("Unexpected Number of Uploads Listed: wanted %d, got %d", len(expectedList.Uploads), len(receivedList.Uploads))
		return err
	}
	totalUploads := 0
	for _, receivedUpload := range receivedList.Uploads {
		for _, expectedUpload := range expectedList.Uploads {
			if receivedUpload.Size == expectedUpload.Size &&
				receivedUpload.UploadID == expectedUpload.UploadID &&
				receivedUpload.Key == expectedUpload.Key {
				totalUploads++
			}
		}
	}
	// TODO: revisit this and the error and find a better way of checking.
	if totalUploads != len(expectedList.Uploads) {
		err := fmt.Errorf("Wrong MetaData Saved in Received List")
		return err
	}
	return nil
}

// mainListMultipartUploads - list-multipart-uplods API test.
func mainListMultipartUploads(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] Multipart (List-Uploads):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	uploads := []ObjectMultipartInfo{}
	// All multipart objects are stored in s3verify created buckets so only list on those.
	bucketName := s3verifyBuckets[0].Name
	for _, multipartObject := range multipartObjects {
		uploads = append(uploads, ObjectMultipartInfo{
			Key:      multipartObject.Key,
			UploadID: multipartObject.UploadID,
		})
	}
	expectedList := listMultipartUploadsResult{
		Bucket:  bucketName,
		Uploads: uploads,
	}
	// Spin scanBar
	scanBar(message)
	// Create a new request.
	req, err := newListMultipartUploadsReq(bucketName)
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
	defer closeResponse(res)
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := listMultipartUploadsVerify(res, http.StatusOK, expectedList); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
