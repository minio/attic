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
	"fmt"
	"net/url"
	"strings"

	"github.com/minio/minio-go"
)

// cleanObjects - use minio-go to remove any s3verify created objects.
func cleanObjects(client *minio.Client, bucketName string) error {
	message := fmt.Sprintf("CleanUp %s (Removing Objects):", bucketName)
	// Spin scanBar
	scanBar(message)

	doneCh := make(chan struct{})
	defer close(doneCh)

	// Only remove s3verify created objects.
	objectCh := client.ListObjects(bucketName, "s3verify/", true, doneCh)
	for object := range objectCh {
		// Spin scanBar
		scanBar(message)
		err := client.RemoveObject(bucketName, object.Key)
		if err != nil {
			// Do not stop on errors.
			continue
		}
	}
	printMessage(message, nil)
	return nil
}

// cleanBucket - use minio-go to cleanup any s3verify created buckets.
func cleanBucket(client *minio.Client, bucketName string) error {
	message := fmt.Sprintf("CleanUp %s (Removing Bucket):", bucketName)
	// Spin scanBar
	scanBar(message)
	if err := client.RemoveBucket(bucketName); err != nil {
		return err
	}
	printMessage(message, nil)
	return nil
}

// cleanS3verify - purges the given bucketName of objects then removes the bucket.
func cleanS3verify(config ServerConfig, bucketPrefix string) error {
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return err
	}
	secure := false
	if hostURL.Scheme == "https" {
		secure = true
	}
	// Extract only the host from the url.
	client, err := minio.New(hostURL.Host, config.Access, config.Secret, secure)
	if err != nil {
		return err
	}
	buckets, err := client.ListBuckets()
	if err != nil {
		return err
	}
	// Delete all s3verify objects and buckets.
	for _, bucket := range buckets {
		if strings.HasPrefix(bucket.Name, bucketPrefix) {
			if err := cleanObjects(client, bucket.Name); err != nil {
				return err
			}
			if err := cleanBucket(client, bucket.Name); err != nil {
				return err
			}
		}
	}
	return nil
}
