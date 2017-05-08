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
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	//
	s3verifyBuckets = []BucketInfo{}
	// Make two random buckets below in the test.
	preparedBuckets = []BucketInfo{}
	// See http://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html for all bucket naming restrictions.
	invalidBuckets = []BucketInfo{
		BucketInfo{
			Name: "s3", // Bucket names must be at least 3 chars long.
		},
		BucketInfo{
			Name: "babcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwzyz", // Bucket names must be less than 63 chars long. This is only on regions other than us-east-1.
		},
		BucketInfo{
			Name: "S3verify", // Bucket names must start with a lowercase letter or a number.
		},
		BucketInfo{
			Name: "192.168.5.4", // Bucket names must not be formatted as an IP address.
		},
		BucketInfo{
			Name: "s3..verify", // Bucket names can not have adjacent periods in them.
		},
		BucketInfo{
			Name: ".s3verify", // Bucket names can not start with periods.
		},
		BucketInfo{
			Name: "s3verify.", // Bucket names can not end with periods.
		},
	}
)

// newPutBucketReq - Create a new Make bucket request.
func newPutBucketReq(region, bucketName string) (Request, error) {
	// putBucketReq - hardcode the static portions of a new Make Bucket request.
	var putBucketReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName
	putBucketReq.bucketName = bucketName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body for requests that do not use regions other than us-east-1.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	putBucketReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	// Set req URL, Header and Body if necessary.
	if region != globalDefaultRegion { // Must set the request elements for non us-east-1 regions.
		bucketConfig := createBucketConfiguration{}
		bucketConfig.Location = region
		bucketConfigBytes, err := xml.Marshal(bucketConfig)
		if err != nil {
			return Request{}, err
		}
		bucketConfigBuffer := bytes.NewReader(bucketConfigBytes)
		_, sha256Sum, contentLength, err := computeHash(bucketConfigBuffer)
		if err != nil {
			return Request{}, err
		}
		// Set the body.
		putBucketReq.contentBody = bucketConfigBuffer
		putBucketReq.contentLength = contentLength
		// Fill request headers and URL.
		putBucketReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	}
	// Set the bucketName.
	putBucketReq.bucketName = bucketName
	// Set the user-agent.
	putBucketReq.customHeader.Set("User-Agent", appUserAgent)

	return putBucketReq, nil
}

// putBucketVerify - Check the response Body, Header, Status for AWS S3 compliance.
func putBucketVerify(res *http.Response, bucketName string, expectedStatusCode int, expectedError ErrorResponse) error {
	// Previous attempt to create bucket succeeded, treat it as a good condition.
	if res.StatusCode == http.StatusConflict {
		return nil
	}
	if err := verifyStatusPutBucket(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderPutBucket(res.Header, bucketName, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyPutBucket(res.Body, expectedError); err != nil {
		return err
	}
	return nil
}

// verifyStatusPutBucket - Check the response status for AWS S3 compliance.
func verifyStatusPutBucket(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyPutBucket - Check the response body for AWS S3 compliance.
func verifyBodyPutBucket(resBody io.Reader, expectedError ErrorResponse) error {
	if expectedError.Message != "" {
		resError := ErrorResponse{}
		err := xmlDecoder(resBody, &resError)
		if err != nil {
			return err
		}
		if resError.Message != expectedError.Message {
			err := fmt.Errorf("Unexpected Error Message: wanted %v, got %v", expectedError.Message, resError.Message)
			return err
		}
		return nil
	}
	// Read the body to make sure it is empty.
	body, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	// There is no body returned by a Put Bucket request.
	if string(body) != "" {
		err := fmt.Errorf("Unexpected Body: %v", string(body))
		return err
	}
	return nil
}

// verifyHeaderPutBucket - Check the response header for AWS S3 compliance.
func verifyHeaderPutBucket(header http.Header, bucketName string, expectedStatusCode int) error {
	if expectedStatusCode == http.StatusOK {
		location := header.Get("Location")
		if location != "http://"+bucketName+".s3.amazonaws.com/" && location != "/"+bucketName {
			// TODO: wait for Minio server to fix endpoint detection.
			err := fmt.Errorf("Unexpected Location: got %v", location)
			return err
		}
	}
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// mainPutBucket- entry point for the putBucket test with valid names.
func mainPutBucket(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] PutBucket (Valid Names):", curTest, globalTotalNumTest)
	// Spin the scanBar scanBar(message)
	// Four new buckets are created on the same host regardless of whether or not the test has been prepared.
	for i := 0; i < 4; i++ {
		validBucket := BucketInfo{
			Name: "s3verify-" + globalSuffix + strconv.Itoa(i),
		}
		// Spin the scanBar
		scanBar(message)
		// Create a new Make bucket request.
		customPutBucketReq, err := newPutBucketReq(config.Region, validBucket.Name)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin the scanBar
		scanBar(message)
		// Execute the request.
		res, err := config.execRequest("PUT", customPutBucketReq)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Spin the scanBar
		scanBar(message)
		// Check the responses Body, Status, Header.
		if err := putBucketVerify(res, validBucket.Name, http.StatusOK, ErrorResponse{}); err != nil {
			printMessage(message, err)
			return false
		}
		// Save the newly created bucket.
		s3verifyBuckets = append(s3verifyBuckets, validBucket)
		// Spin the scanBar
		scanBar(message)
	}
	printMessage(message, nil)
	return true
}

// mainPutBucketInvalid - entry point for testing putbucket API with invalid names.
func mainPutBucketInvalid(config ServerConfig, curTest int) bool {
	// Test invalid names. This cannot be separated yet into its own test because of the way --prepared is laid out currently.
	message := fmt.Sprintf("[%02d/%d] PutBucket (Invalid Names):", curTest, globalTotalNumTest)
	expectedError := ErrorResponse{
		Message: "The specified bucket is not valid.",
	}
	// Test that all invalid names fail correctly.
	for _, bucket := range invalidBuckets {
		// Spin scanBar
		scanBar(message)
		// Create a new PUT bucket request.
		customPutBucketReq, err := newPutBucketReq(config.Region, bucket.Name)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		res, err := config.execRequest("PUT", customPutBucketReq)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Spin scanBar
		scanBar(message)
		// Verify that the request failed as predicted.
		if err := putBucketVerify(res, bucket.Name, 400, expectedError); err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
