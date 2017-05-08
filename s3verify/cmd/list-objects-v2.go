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

// newListObjectsV2Req - Create a new HTTP request for ListObjects V2 API.
func newListObjectsV2Req(bucketName string, requestParameters map[string]string) (Request, error) {
	// listObjectsV2Req - a new HTTP request for ListObjects V2 API.
	var listObjectsV2Req = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName.
	listObjectsV2Req.bucketName = bucketName

	// Set URL query values.
	urlValues := make(url.Values)
	urlValues.Set("list-type", "2")
	for k, v := range requestParameters {
		urlValues.Set(k, v)
	}

	listObjectsV2Req.queryValues = urlValues

	// No body is sent with GET requests.
	reader := bytes.NewReader([]byte{})
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	listObjectsV2Req.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	listObjectsV2Req.customHeader.Set("User-Agent", appUserAgent)

	return listObjectsV2Req, nil
}

// listObjectsV2Verify - verify the response returned matches what is expected.
func listObjectsV2Verify(res *http.Response, expectedStatusCode int, expectedList listBucketV2Result) error {
	if err := verifyStatusListObjectsV2(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyListObjectsV2(res.Body, expectedList); err != nil {
		return err
	}
	if err := verifyHeaderListObjectsV2(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyHeaderListObjectsV2 - verify the heaer returned matches what is expected.
func verifyHeaderListObjectsV2(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyStatusListObjectsV2 - verify the status returned matches what is expected.
func verifyStatusListObjectsV2(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Status Received: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyListObjectsV2 - verify the objects listed match what is expected.
func verifyBodyListObjectsV2(resBody io.Reader, expectedList listBucketV2Result) error {
	receivedList := listBucketV2Result{}
	if err := xmlDecoder(resBody, &receivedList); err != nil {
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
		err := fmt.Errorf("Unexpected Number of Objects Listed: wanted %d objects and %d prefixes, got %d objects and %d prefixes",
			len(expectedList.Contents), len(expectedList.CommonPrefixes),
			len(receivedList.Contents), len(receivedList.CommonPrefixes))
		return err
	}
	if err := verifyObjectsListObjects(receivedList.Contents, expectedList.Contents); err != nil {
		return err
	}
	return nil
}

// mainListObjectsV2 - Entry point for the ListObjects V2 API test. This test is the same for --prepared environments and non --prepared.
func mainListObjectsV2(config ServerConfig, curTest int, bucketName string, testObjects []*ObjectInfo) bool {
	message := fmt.Sprintf("[%02d/%d] ListObjects V2:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	objectInfo := ObjectInfos{}
	for _, object := range testObjects {
		objectInfo = append(objectInfo, *object)
	}
	sort.Sort(objectInfo)

	expectedList := listBucketV2Result{
		Name:        bucketName,        // List only from the first bucket created because that is the bucket holding the objects.
		Contents:    objectInfo[:1000], // Will only return the first 1000 objects.
		IsTruncated: true,
	}
	// Create a new request.
	req, err := newListObjectsV2Req(bucketName, nil)
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
	// Verify the response.
	if err := listObjectsV2Verify(res, http.StatusOK, expectedList); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test for listobjects with start-after parameter set.
	expectedListStartAfter := listBucketV2Result{
		Name:        bucketName,
		Contents:    objectInfo[31:],
		IsTruncated: false,
	}

	// Store the parameters.
	startAfterMap := map[string]string{
		"start-after": objectInfo[30].Key,
	}

	// Create a new request.
	startAfterReq, err := newListObjectsV2Req(bucketName, startAfterMap)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute request.
	startAfterRes, err := config.execRequest("GET", startAfterReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(startAfterRes)
	// Verify the response
	if err := listObjectsV2Verify(startAfterRes, http.StatusOK, expectedListStartAfter); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test for listobjects with maxkeys parameter set.
	expectedListMaxKeys := listBucketV2Result{
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
	maxKeysReq, err := newListObjectsV2Req(bucketName, maxKeysMap) // MaxKeys set to 30.
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
	if err := listObjectsV2Verify(maxKeysRes, http.StatusOK, expectedListMaxKeys); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test for listobjects with prefix parameter set.
	expectedListPrefix := listBucketV2Result{
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

	prefixReq, err := newListObjectsV2Req(bucketName, prefixMap)
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
	if err := listObjectsV2Verify(prefixRes, http.StatusOK, expectedListPrefix); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test for listobjects with delimiter parameter and prefix parameter set.
	expectedListDelimiterPrefix := listBucketV2Result{
		Name: bucketName,
		// Should return no objects just the prefix.
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

	prefixDelimiterReq, err := newListObjectsV2Req(bucketName, prefixDelimiterMap)
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
	if err := listObjectsV2Verify(prefixDelimRes, http.StatusOK, expectedListDelimiterPrefix); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test for listobjects with max-keys set over 1000.
	expectedListMaxKeysOver := listBucketV2Result{
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
	maxKeysOverReq, err := newListObjectsV2Req(bucketName, maxKeysOverMap)
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
	if err := listObjectsV2Verify(maxKeysOverRes, http.StatusOK, expectedListMaxKeysOver); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test passed.
	printMessage(message, nil)
	return true
}

// mainListObjectsV2UnPrepared - Test the ListObjects V2 API in an unprepared environment.
func mainListObjectsV2UnPrepared(config ServerConfig, curTest int) bool {
	bucketName := s3verifyBuckets[0].Name
	return mainListObjectsV2(config, curTest, bucketName, s3verifyObjects)
}

// mainListObjectsV2Prepared - Test the ListObjects V2 API in a prepared environment.
func mainListObjectsV2Prepared(config ServerConfig, curTest int) bool {
	bucketName := preparedBuckets[0].Name
	return mainListObjectsV2(config, curTest, bucketName, preparedObjects)
}
