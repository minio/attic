/*
 * Copyright (c) 2017 Minio, Inc. <https://www.minio.io>
 *
 * This file is part of Xray.
 *
 * Xray is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package cmd

import (
	"net/url"
	"os"
	"strconv"
	"time"

	minio "github.com/minio/minio-go"
)

func genObjectName() string {
	uid := time.Now().UTC()
	return uid.Format("02-Jan-2006-MST/15hrs-04mins-05secs.jpg")
}

type minioConfig struct{}

func (c minioConfig) Endpoint() string {
	ep := os.Getenv("S3_ENDPOINT")
	if ep == "" {
		// Default
		ep = "play.minio.io:9000"
	}
	return ep
}

func (c minioConfig) AccessKey() string {
	ak := os.Getenv("ACCESS_KEY")
	if ak == "" {
		// Default
		ak = "Q3AM3UQ867SPQQA43P2F"
	}
	return ak
}

func (c minioConfig) SecretKey() string {
	sk := os.Getenv("SECRET_KEY")
	if sk == "" {
		// Default
		sk = "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	}
	return sk
}

func (c minioConfig) SSL() bool {
	return mustParseBool(os.Getenv("S3_SECURE"))
}

// Convert string to bool and always return true if any error
func mustParseBool(str string) bool {
	b, err := strconv.ParseBool(str)
	if err != nil {
		return true
	}
	return b
}

func (c minioConfig) BucketName() string {
	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		// Default
		bucketName = "alice"
	}
	return bucketName
}

func (c minioConfig) Region() string {
	region := os.Getenv("S3_REGION")
	if region == "" {
		// Default
		region = "us-east-1"
	}
	return region
}

// Generates new presigned PUT URL.
func (v *xrayHandlers) newPresignedURL(objName string) (*url.URL, error) {
	return v.minioClient.PresignedPutObject(globalMinioClntConfig.BucketName(),
		objName, time.Duration(48*time.Hour))
}

// Create a minio client to play.minio.io and make a bucket.
func newMinioClient() (*minio.Client, error) {
	// Initialize minio client instance.
	minioClient, err := minio.NewWithRegion(globalMinioClntConfig.Endpoint(), globalMinioClntConfig.AccessKey(),
		globalMinioClntConfig.SecretKey(), globalMinioClntConfig.SSL(),
		globalMinioClntConfig.Region())
	if err != nil {
		return nil, err
	}

	// Check to see if we already own this bucket (which happens if you run this twice)
	exists, err := minioClient.BucketExists(globalMinioClntConfig.BucketName())
	if err != nil {
		return nil, err
	}

	// Create the bucket if it doesn't exist yet.
	if !exists {
		err = minioClient.MakeBucket(globalMinioClntConfig.BucketName(), globalMinioClntConfig.Region())
		if err != nil {
			return nil, err
		}
	}

	return minioClient, nil

}
