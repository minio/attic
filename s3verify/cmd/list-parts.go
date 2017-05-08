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

// newListPartsReq - Create a new HTTP request for the ListParts API.
func newListPartsReq(bucketName, objectName, uploadID string) (Request, error) {
	// listPartsReq - a new HTTP request for ListParts.
	var listPartsReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	listPartsReq.bucketName = bucketName
	listPartsReq.objectName = objectName

	// Create new url queries.
	urlValues := make(url.Values)
	urlValues.Set("uploadId", uploadID)
	listPartsReq.queryValues = urlValues

	// No body is sent with GET requests.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the requests URL and Header values.
	listPartsReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	listPartsReq.customHeader.Set("User-Agent", appUserAgent)

	return listPartsReq, nil
}

// listPartsVerify - verify that the returned response matches what is expected.
func listPartsVerify(res *http.Response, expectedStatusCode int, expectedList listObjectPartsResult) error {
	if err := verifyStatusListParts(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyListParts(res.Body, expectedList); err != nil {
		return err
	}
	if err := verifyHeaderListParts(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyStatusListParts - verify that the status returned matches what is expected.
func verifyStatusListParts(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyListParts - verify that the returned body matches whats expected.
func verifyBodyListParts(resBody io.Reader, expectedList listObjectPartsResult) error {
	result := listObjectPartsResult{}
	err := xmlDecoder(resBody, &result)
	if err != nil {
		return err
	}
	totalParts := 0
	for _, part := range expectedList.ObjectParts {
		for _, resPart := range result.ObjectParts {
			if part.PartNumber == resPart.PartNumber && "\""+part.ETag+"\"" == resPart.ETag {
				totalParts++
			}
		}
	}
	if totalParts != 1 {
		err := fmt.Errorf("Incorrect number of parts listed: wanted 1, got %v", totalParts)
		return err
	}
	return nil
}

// verifyHeaderListParts - verify the header returned matches what is expected.
func verifyHeaderListParts(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainListParts - Entry point for the ListParts API test.
func mainListParts(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] Multipart (List-Parts):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// All multipart objects are stored in s3verify created buckets so only list parts in those buckets.
	bucketName := s3verifyBuckets[0].Name
	// TODO: eventually separate tests will be needed here when during prepare we concurrently upload
	// 1001 parts for the list parts test.

	object := multipartObjects[0]
	// Create a handcrafted ListObjectsPartsResult
	expectedList := listObjectPartsResult{
		Bucket:      bucketName,
		Key:         object.Key,
		UploadID:    object.UploadID,
		ObjectParts: objectParts[0],
	}
	// Create a new ListParts request.
	req, err := newListPartsReq(bucketName, object.Key, object.UploadID)
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
	if err := listPartsVerify(res, http.StatusOK, expectedList); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, err)
	return true
}
