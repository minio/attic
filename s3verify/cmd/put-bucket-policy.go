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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// s3verifyPolicies - maintain a list of all access policies used by s3verify.
var s3verifyPolicies = []BucketAccessPolicy{}

// newPutBucketPolicyReq - create a new PutBucketPolicyReq
func newPutBucketPolicyReq(bucketName string, bucketPolicy BucketAccessPolicy) (Request, error) {
	//
	var putBucketPolicyReq = Request{
		customHeader: http.Header{},
	}
	// Set the Request bucketName
	putBucketPolicyReq.bucketName = bucketName

	// Set the query values.
	urlValues := make(url.Values)
	urlValues.Set("policy", "")

	putBucketPolicyReq.queryValues = urlValues

	// Set the body.
	policyBytes, err := json.Marshal(&bucketPolicy)
	if err != nil {
		return Request{}, err
	}

	reader := bytes.NewReader(policyBytes)
	md5Sum, sha256Sum, contentLength, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}
	putBucketPolicyReq.contentBody = reader
	putBucketPolicyReq.contentLength = contentLength
	// Set the header values.
	putBucketPolicyReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	putBucketPolicyReq.customHeader.Set("Content-MD5", base64.StdEncoding.EncodeToString(md5Sum))
	putBucketPolicyReq.customHeader.Set("User-Agent", appUserAgent)

	return putBucketPolicyReq, nil
}

// putBucketPolicyVerify - verify the response returned matches what is expected.
func putBucketPolicyVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyBodyPutBucketPolicy(res.Body); err != nil {
		return err
	}
	if err := verifyStatusPutBucketPolicy(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderPutBucketPolicy(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyHeaderPutBucketPolicy - verify the header returned matches what is expected.
func verifyHeaderPutBucketPolicy(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// verifyStatusPutBucketPolicy - verify the status returned matches what is expected.
func verifyStatusPutBucketPolicy(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpexcted Status Code Received: wanted %d, got %d", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyPutBucketPolicy - verify the body returned is empty.
func verifyBodyPutBucketPolicy(resBody io.Reader) error {
	body, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	// Verify that the body received is empty.
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Received: %v", string(body))
		return err
	}
	return nil
}

// mainPutBucketPolicy - entry point for the putbucketpolicy test.
func mainPutBucketPolicy(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] PutBucketPolicy:", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)

	// List of different bucketPolicies to create.
	policies := []BucketPolicy{
		BucketPolicyReadWrite, // Set the second bucket to have both read/write permissions.
		BucketPolicyReadOnly,  // Set the third bucket to only read.
		BucketPolicyWriteOnly, // Set the last bucket to only write.
	}

	// Set a policy for all buckets created by s3verify.
	for i, bucket := range s3verifyBuckets[:3] {
		// Spin scanBar
		scanBar(message)

		bucketName := bucket.Name
		// Gather the policy you wish to create.
		setPolicy := policies[i]
		// Create the statements required.
		statements := SetPolicy([]Statement{}, setPolicy, bucketName, "")
		bucketPolicy := BucketAccessPolicy{
			Version:    "2008-10-17",
			Statements: statements,
		}
		// Add this new policy to the list of s3verify activated policies.
		s3verifyPolicies = append(s3verifyPolicies, bucketPolicy)
		// Create a new request to add the bucket policy to the bucket.
		req, err := newPutBucketPolicyReq(bucketName, bucketPolicy)
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
		// Verify the response.
		if err := putBucketPolicyVerify(res, http.StatusNoContent); err != nil {
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
