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
	"encoding/xml"
	"time"
)

// copyObjectResult container for copy object response.
type copyObjectResult struct {
	ETag         string
	LastModified string // time string format "2006-01-02T15:04:05.000Z"
}

// listAllMyBucketsResult container for listBuckets response.
type listAllMyBucketsResult struct {
	// Container for one or more buckets.
	Buckets buckets
	Owner   owner
}

type buckets struct {
	Bucket []BucketInfo
}

// initiator container for who initiated multipart upload
type initiator struct {
	ID          string
	DisplayName string
}

// owner container for bucket owner information.
type owner struct {
	DisplayName string
	ID          string
}

// commonPrefix container for prefix response.
type commonPrefix struct {
	Prefix string
}

// listBucketResult container for listObjects response.
type listBucketResult struct {
	// A response can contain CommonPrefixes only if you have
	// specified a delimiter.
	CommonPrefixes []commonPrefix
	// Metadata about each object returned.
	Contents  []ObjectInfo
	Delimiter string

	// Encoding type used to encode object keys in the response.
	EncodingType string

	// A flag that indicates whether or not ListObjects returned all of the results
	// that satisfied the search criteria.
	IsTruncated bool
	Marker      string
	MaxKeys     int64
	Name        string

	// When response is truncated (the IsTruncated element value in
	// the response is true), you can use the key name in this field
	// as marker in the subsequent request to get next set of objects.
	// Object storage lists objects in alphabetical order Note: This
	// element is returned only if you have delimiter request
	// parameter specified. If response does not include the NextMaker
	// and it is truncated, you can use the value of the last Key in
	// the response as the marker in the subsequent request to get the
	// next set of object keys.
	NextMarker string
	Prefix     string
}

// listBucketV2Result container for ListObjects V2 response.
type listBucketV2Result struct {
	// A response can contain CommonPrefixes only if you have
	// specified a delimiter.
	CommonPrefixes []commonPrefix
	// Metadata about each object returned.
	Contents  []ObjectInfo
	Delimiter string

	// Encoding type used to encode object keys in the response.
	EncodingType string

	// A flag that indicates whether or not ListObjects returned all of the results
	// that satisfied the search criteria.
	IsTruncated bool
	MaxKeys     int64
	Name        string

	// Hold the token that will be sent in the next request to fetch the next group of keys
	NextContinuationToken string

	ContinuationToken string
	Prefix            string

	// FetchOwner and StartAfter are currently not used
	FetchOwner string
	StartAfter string
}

// createBucketConfiguration container for bucket configuration.
type createBucketConfiguration struct {
	XMLName  xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ CreateBucketConfiguration" json:"-"`
	Location string   `xml:"LocationConstraint"`
}

// initiateMultipartUploadResult container for InitiateMultipartUpload
// response.
type initiateMultipartUploadResult struct {
	Bucket   string
	Key      string
	UploadID string `xml:"UploadId"`
}

// objectPart container for parts of a multipart upload.
type objectPart struct {
	// Part number identifies the part.
	PartNumber int

	// Date and time the part was uploaded.
	LastModified time.Time

	// Entity tag returned when the part was uploaded, usually
	// md5sum of the part.
	ETag string

	// Size of the uploaded part data.
	Size int64

	// Part data
	Data []byte
}

// completeMultipartUploadResult container for completed multipart
// upload response.
type completeMultipartUploadResult struct {
	Location string
	Bucket   string
	Key      string
	ETag     string
}

// listMultipartUploadsResult container for ListMultipartUploads result.
type listMultipartUploadsResult struct {
	Bucket         string
	KeyMarker      string
	UploadIDMarker string `xml:"UploadIdMarker"`
	NextKeyMarker  string
	EncodingType   string
	MaxUploads     int64
	IsTruncated    bool
	Uploads        []ObjectMultipartInfo `xml:"Upload"`
	Prefix         string
	Delimiter      string
	// A response can contain CommonPrefixes only if you specify a delimiter.
	CommonPrefixes []commonPrefix
}

// completePart sub container lists individual part numbers and their md5sum,
// part of completeMultipartUpload.
type completePart struct {
	XMLName xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ Part" json:"-"`

	// Part number identifies the part.
	PartNumber int
	ETag       string
}

// completeMultipartUpload container for completing multipart upload.
type completeMultipartUpload struct {
	XMLName xml.Name       `xml:"http://s3.amazonaws.com/doc/2006-03-0/ CompleteMultipartUpload" json:"-"`
	Parts   []completePart `xml:"Part"`
}

// listObjectPartsResult container for ListObjectParts response.
type listObjectPartsResult struct {
	Bucket               string
	Key                  string
	UploadID             string
	Initiator            initiator
	Owner                owner
	StorageClass         string
	PartNumberMarker     int
	NextPartNumberMarker int
	MaxParts             int
	IsTruncated          bool
	ObjectParts          []objectPart `xml:"Part"`

	EncodingType string
}
