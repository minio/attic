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
	"math/rand"
	"net/url"
	"strconv"
	"time"

	"github.com/minio/minio-go"
)

// prepareBucket - Uses minio-go library to create new testing bucket for use by s3verify.
func prepareBucket(region string, client *minio.Client) (string, error) {
	reuseMessage := "Reusing test bucket"
	bucketName := "s3verify-" + globalSuffix
	preparedBucket := BucketInfo{
		Name: bucketName,
	}

	// Check to see if the desired bucket already exists.
	bucketExists, err := client.BucketExists(bucketName)
	if err != nil {
		printMessage(reuseMessage, err)
		return "", err
	}
	// Exit successfully if bucket already exists.
	if bucketExists {
		// Store the existing bucket for testing.
		preparedBuckets = append(preparedBuckets, preparedBucket)
		// Don't print anything for successfully reusable environments.
		return bucketName, nil
	}
	createMessage := "Creating test bucket"
	// Spin scanBar
	scanBar(createMessage)
	// Create the new testing bucket.
	if err := client.MakeBucket(bucketName, region); err != nil {
		printMessage(createMessage, err)
		return "", err
	}
	// Store the created bucket for testing.
	preparedBuckets = append(preparedBuckets, preparedBucket)
	// Spin scanBar
	scanBar(createMessage)
	// Bucket preparation passed.
	printMessage(createMessage, nil)
	return bucketName, nil
}

// TODO: see if parallelization has a place here.

// prepareObjects - Uses minio-go library to create 1001 new testing objects for use by s3verify.
func prepareObjects(client *minio.Client, bucketName string) error {
	createMessage := "Creating test objects"
	// First check that the bucketName does not already contain the correct number of s3verify objects.
	var objCount int
	doneCh := make(chan struct{})
	// Require all objects stored in the prepared bucket to be of the form 's3verify/put/object/#'.
	objectInfoCh := client.ListObjects(bucketName, "s3verify/put/object/", true, doneCh)
	for obj := range objectInfoCh {
		objCount++
		preparedObject := &ObjectInfo{
			Key: obj.Key,
		}
		// Store the already created object for testing.
		preparedObjects = append(preparedObjects, preparedObject)
	}
	if objCount == globalNumTestObjects {
		//  Don't print anything for successfully prepared environments.
		return nil
	}
	// Upload 1001 objects specifically for the list-objects tests.
	for i := objCount; i < globalNumTestObjects; i++ {
		// Spin scanBar
		scanBar(createMessage)
		randomData := randString(60, rand.NewSource(time.Now().UnixNano()), "")
		objectKey := "s3verify/put/object/" + globalSuffix + strconv.Itoa(i)
		byteData := []byte(randomData)
		preparedObject := &ObjectInfo{
			Key:  objectKey,
			Body: byteData,
		}
		// Create 60 bytes worth of random data for each object.
		reader := bytes.NewReader(byteData)
		_, err := client.PutObject(bucketName, objectKey, reader, "application/octet-stream")
		if err != nil {
			printMessage(createMessage, err)
			return err
		}
		// Store the created object for testing.
		preparedObjects = append(preparedObjects, preparedObject)

		// Spin scanBar
		scanBar(createMessage)
	}
	// Spin scanBar
	scanBar(createMessage)
	// Object preparation passed.
	printMessage(createMessage, nil)
	return nil
}

// TODO: Create function using minio-go to upload 1001 parts of a multipart operation.

// mainReuseS3Verify - Create one new buckets and 1001 objects for s3verify to use in the test.
func mainReuseS3Verify(config ServerConfig) error {
	// Extract necessary values from the config.
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return err
	}
	region := config.Region
	isSecure := hostURL.Scheme == "https"
	client, err := minio.New(hostURL.Host, config.Access, config.Secret, isSecure)
	if err != nil {
		return err
	}
	// Create testing bucket if it doesn't already exist.
	validBucketName, err := prepareBucket(region, client)
	if err != nil {
		return err
	}
	// Use the first newly created bucket to store all the objects.
	if err := prepareObjects(client, validBucketName); err != nil {
		return err
	}
	return nil
}
