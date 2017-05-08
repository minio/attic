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
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// newGetObjectRangeReq - Create a new GET object range request.
func newGetObjectRangeReq(bucketName, objectName, startRange, endRange string) (Request, error) {
	// getObjectRangeReq - a new HTTP request for a GET object with a specific range request.
	var getObjectRangeReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	getObjectRangeReq.bucketName = bucketName
	getObjectRangeReq.objectName = objectName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because GET requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	getObjectRangeReq.customHeader.Set("Range", "bytes="+startRange+"-"+endRange)
	getObjectRangeReq.customHeader.Set("User-Agent", appUserAgent)
	getObjectRangeReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	return getObjectRangeReq, nil
}

// Test a GET object request with a range header set.
func mainGetObjectRange(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] GetObject (Range):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	rand.Seed(time.Now().UnixNano())
	// All getobject tests happen in s3verify created buckets
	// on s3verify created objects.
	bucketName := s3verifyBuckets[0].Name
	testObject := s3verifyObjects[0]
	// Spin scanBar
	scanBar(message)

	// Test a random range.
	startRange := rand.Int63n(testObject.Size)
	endRange := rand.Int63n(int64(testObject.Size-startRange)) + startRange
	// Create new GET object range request...testing range.
	req, err := newGetObjectRangeReq(bucketName, testObject.Key, strconv.FormatInt(startRange, 10), strconv.FormatInt(endRange, 10))
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the request.
	res, err := config.execRequest("GET", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	defer closeResponse(res)
	bufRange := testObject.Body[startRange : endRange+1]
	// Verify the response...these checks do not check the headers yet.
	if err := getObjectVerify(res, bufRange, http.StatusPartialContent, nil, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)

	// Test an openended range. Expecting whole object back.
	startOpenRange := ""
	endOpenRange := strconv.FormatInt(testObject.Size, 10)
	// Create a new open range req.
	openRangeReq, err := newGetObjectRangeReq(bucketName, testObject.Key, startOpenRange, endOpenRange)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the open request.
	openRangeRes, err := config.execRequest("GET", openRangeReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Verify that the request failed as expected.
	if err := getObjectVerify(openRangeRes, testObject.Body, http.StatusPartialContent, nil, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}
	// Test a negative range request. Should give back whole object.
	startNegativeRange := "-5"
	endNegativeRange := "1"

	// Creatge a new negative range req.
	negativeRangeReq, err := newGetObjectRangeReq(bucketName, testObject.Key, startNegativeRange, endNegativeRange)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the negative request.
	negativeRangeRes, err := config.execRequest("GET", negativeRangeReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Verify that the request failed as expected.
	if err := getObjectVerify(negativeRangeRes, testObject.Body, http.StatusOK, nil, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}

	// Test an unended range. Expecting full object.
	startUnEndedRange := ""
	endUnEndedRange := strconv.FormatInt(testObject.Size, 10)

	// Creatge a new UnEnded range req.
	unEndedRangeReq, err := newGetObjectRangeReq(bucketName, testObject.Key, startUnEndedRange, endUnEndedRange)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Execute the UnEnded request.
	unEndedRangeRes, err := config.execRequest("GET", unEndedRangeReq)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Verify that the request failed as expected.
	if err := getObjectVerify(unEndedRangeRes, testObject.Body, http.StatusPartialContent, nil, ErrorResponse{}); err != nil {
		printMessage(message, err)
		return false
	}

	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
