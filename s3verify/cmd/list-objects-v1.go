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
	"sort"
)

// newListObjectsV1Req - Create a new HTTP request for ListObjects V1.
func newListObjectsV1Req(bucketName string, parameters map[string]string) (Request, error) {
	// listObjectsV1Req - a new HTTP request for ListObjects V1.
	var listObjectsV1Req = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName.
	listObjectsV1Req.bucketName = bucketName

	// Set the query values.
	urlValues := make(url.Values)
	for k, v := range parameters {
		urlValues.Set(k, v)
	}
	listObjectsV1Req.queryValues = urlValues

	// No body is sent with GET requests.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	listObjectsV1Req.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	listObjectsV1Req.customHeader.Set("User-Agent", appUserAgent)

	return listObjectsV1Req, nil
}

// listObjectsV1Verify - verify the response returned matches what is expected.
func listObjectsV1Verify(res *http.Response, expectedStatusCode int, expectedList listBucketResult) error {
	if err := verifyStatusListObjectsV1(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyListObjectsV1(res.Body, expectedList); err != nil {
		return err
	}
	if err := verifyHeaderListObjectsV1(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyStatusListObjectsV1 - verify the status returned matches what is expected.
func verifyStatusListObjectsV1(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyObjectsListObjects - verify that the objects returned in a listobjects request match what is expected.
func verifyObjectsListObjects(receivedContents []ObjectInfo, expectedContents []ObjectInfo) error {
	if len(receivedContents) > 1000 {
		err := fmt.Errorf("Unexpected Number of Objects Listed: ListObjects V1 cannot return more than 1000 objects, got %v objects", len(receivedContents))
		return err
	}
	for i, expectedObject := range expectedContents {
		receivedObject := receivedContents[i]
		if receivedObject.ETag != expectedObject.ETag {
			err := fmt.Errorf("Incorrect ETag Received: wanted %s, got %s", expectedObject.ETag, receivedObject.ETag)
			return err
		} else if receivedObject.LastModified != expectedObject.LastModified {
			err := fmt.Errorf("Incorrect LastModified Received: wanted %v, got %v", expectedObject.LastModified, receivedObject.LastModified)
			return err
		} else if receivedObject.Key != expectedObject.Key {
			err := fmt.Errorf("Incorrect Key Received: wanted %s, got %s", expectedObject.Key, receivedObject.Key)
			return err
		}
	}
	return nil
}

// verifyBodyListObjectsV1 - verify the body returned matches what is expected.
func verifyBodyListObjectsV1(resBody io.Reader, expectedList listBucketResult) error {
	receivedList := listBucketResult{}
	err := xmlDecoder(resBody, &receivedList)
	if err != nil {
		return err
	}
	if receivedList.Name != expectedList.Name {
		err := fmt.Errorf("Unexpected Bucket Listed: wanted %v, got %v", expectedList.Name, receivedList.Name)
		return err
	}
	if receivedList.IsTruncated != expectedList.IsTruncated {
		err := fmt.Errorf("Unexpected Truncation: wanted %v, got %v", expectedList.IsTruncated, receivedList.IsTruncated)
		return err
	}
	if len(receivedList.Contents)+len(receivedList.CommonPrefixes) != len(expectedList.Contents)+len(expectedList.CommonPrefixes) {
		err := fmt.Errorf("Incorrect Number of Objects Listed: wanted %d objects and %d prefixes, got %d objects and %d prefixes",
			len(expectedList.Contents), len(expectedList.CommonPrefixes),
			len(receivedList.Contents), len(receivedList.CommonPrefixes))
		return err
	}
	if err := verifyObjectsListObjects(receivedList.Contents, expectedList.Contents); err != nil {
		return err
	}
	return nil
}

// verifyHeaderListObjectsV1 - verify the header returned matches what is expected.
func verifyHeaderListObjectsV1(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainListObjectsV1 - ListObjects V1 API test. This test is the same for both --prepared and non --prepared environments.
func mainListObjectsV1(config ServerConfig, curTest int, bucketName string, testObjects []*ObjectInfo) bool {
	message := fmt.Sprintf("[%02d/%d] ListObjects V1:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	objectInfo := ObjectInfos{}
	for _, object := range testObjects {
		objectInfo = append(objectInfo, *object)
	}
	sort.Sort(objectInfo)

	// Test for listobjects with no extra parameters (should only return the first 1000 objects).
	expectedList := listBucketResult{
		Name:        bucketName,        // Listing from the first bucket created that houses all objects.
		Contents:    objectInfo[:1000], // The first bucket created will house all the objects created by the PUT object test.
		IsTruncated: true,              // This list is shorter than the actual 1002 objects stored in testObjects.
	}
	// Create a new request.
	noParamReq, err := newListObjectsV1Req(bucketName, nil) // No extra parameters for the first test.
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	noParamRes, err := config.execRequest("GET", noParamReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(noParamRes)
	// Spin scanBar
	scanBar(message)
	// Verify the response.
	if err := listObjectsV1Verify(noParamRes, http.StatusOK, expectedList); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test for listobjects with maxkeys parameter set.
	expectedListMaxKeys := listBucketResult{
		Name:        bucketName,
		Contents:    objectInfo[:30], // Only return the first 30 objects.
		MaxKeys:     30,              // Only return the first 30 objects.
		IsTruncated: true,
	}
	// Store the parameters to be set by the request.
	maxKeysMap := map[string]string{
		"max-keys": "30", // 30 objects.
	}
	// Create a new request with max-keys set to 30.
	maxKeysReq, err := newListObjectsV1Req(bucketName, maxKeysMap) // MaxKeys set to 30.
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	maxKeysRes, err := config.execRequest("GET", maxKeysReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(maxKeysRes)
	// Spin scanBar
	scanBar(message)
	// Verify the max-keys parameter is respected.
	if err := listObjectsV1Verify(maxKeysRes, http.StatusOK, expectedListMaxKeys); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test for listobjects with prefix parameter set.
	expectedListPrefix := listBucketResult{
		Name: bucketName,
		// Should only return objects that were put during the put-object test.
		Contents:    objectInfo[:1000], // Will only get 1000 objects after the excluded object.
		Prefix:      "s3verify/put/object/",
		IsTruncated: true,
	}
	// Store the parameters.
	prefixMap := map[string]string{
		"prefix": "s3verify/put/object/",
	}

	prefixReq, err := newListObjectsV1Req(bucketName, prefixMap)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	prefixRes, err := config.execRequest("GET", prefixReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(prefixRes)
	// Verify the prefix parameter is respected.
	if err := listObjectsV1Verify(prefixRes, http.StatusOK, expectedListPrefix); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test for listobjects with delimiter parameter and prefix parameter set.
	expectedListDelimiterPrefix := listBucketResult{
		Name: bucketName,
		// No objects will be returned only the prefix.
		CommonPrefixes: []commonPrefix{commonPrefix{"s3verify/put/object/"}},
		Prefix:         "s3verify/put/",
		Delimiter:      "/",
		IsTruncated:    false,
	}

	// Store the parameters.
	prefixDelimiterMap := map[string]string{
		"delimiter": "/",
		"prefix":    "s3verify/put/",
	}

	prefixDelimiterReq, err := newListObjectsV1Req(bucketName, prefixDelimiterMap)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	prefixDelimRes, err := config.execRequest("GET", prefixDelimiterReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(prefixDelimRes)
	// Verify that delimiter and prefix parameters are respected.
	if err := listObjectsV1Verify(prefixDelimRes, http.StatusOK, expectedListDelimiterPrefix); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test for listobjects with max-keys set over 1000.
	expectedListMaxKeysOver := listBucketResult{
		Name: bucketName,
		// Should return only 1000 objects.
		Contents:    objectInfo[:1000], // Will only get 1000 objects after the excluded object.
		IsTruncated: true,
	}

	// Store the parameters.
	maxKeysOverMap := map[string]string{
		"max-keys": "1001",
	}
	// Spin scanBar
	scanBar(message)
	// Create the new request.
	maxKeysOverReq, err := newListObjectsV1Req(bucketName, maxKeysOverMap)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	maxKeysOverRes, err := config.execRequest("GET", maxKeysOverReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(maxKeysOverRes)
	// Verify that the max-keys parameter tops out at 1000 objects.
	if err := listObjectsV1Verify(maxKeysOverRes, http.StatusOK, expectedListMaxKeysOver); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test passed.
	printMessage(message, nil)
	return true
}

// mainListObjectsV1UnPrepared - Test the ListObjects V1 API in an unprepared environment.
func mainListObjectsV1UnPrepared(config ServerConfig, curTest int) bool {
	bucketName := s3verifyBuckets[0].Name
	return mainListObjectsV1(config, curTest, bucketName, s3verifyObjects)
}

// mainListObjectsV1Prepared - Test the ListObjects V1 API in a prepared environment.
func mainListObjectsV1Prepared(config ServerConfig, curTest int) bool {
	bucketName := preparedBuckets[0].Name
	return mainListObjectsV1(config, curTest, bucketName, preparedObjects)
}
