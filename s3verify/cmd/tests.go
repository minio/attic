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

// Tests must be run in the following order:
// PutBucket,
// PutObject,
// ListBuckets,
// ListObjects,
// Multipart,
// HeadBucket,
// HeadObject,
// CopyObject,
// GetObject,
// RemoveObject,
// RemoveBucket,

// Tests are sorted into the following lists:
// preparedTests    -- tests that will use materials set up by the --prepare flag.
// unpreparedTests  -- tests that will be self-sufficient and create their own testing environment.

// Tests - holds all tests that must be run differently based on usage of the -- flag.
var preparedTests = []APItest{
	// Tests for PutBucket API.
	APItest{
		Test:     mainPutBucket,
		Extended: false, // PutBucket is not an extended API.
		Critical: false, // Because -- has been used this bucket is not necessary for future tests.
	},
	APItest{
		Test:     mainPutBucketInvalid,
		Extended: false, // PutBucket is not an extended API.
		Critical: false, // This test is not used for future tests.
	},

	// Tests for PutBucketPolicy API.
	APItest{
		Test:     mainPutBucketPolicy,
		Extended: false, // PutBucketPolicy is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for GetBucketPolicy API.
	APItest{
		Test:     mainGetBucketPolicy,
		Extended: false, // GetBucketPolicy is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for PutObject API.
	APItest{
		Test:     mainPutObjectPrepared,
		Extended: false, // PutObject is not an extended API.
		Critical: false, // Because -- has been used this object is not necessary for future tests.
	},

	// Tests for HeadBucket API.
	APItest{
		Test:     mainHeadBucket,
		Extended: false, // HeadBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for HeadObject API if an environment was prepared.
	APItest{
		Test:     mainHeadObjectPrepared,
		Extended: false, // HeadObject is not an extended API.
		Critical: true,  // This test affects future tests and must pass.
	},
	APItest{
		Test:     mainHeadObjectIfModifiedSince,
		Extended: true,  // HeadObject with if-modified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainHeadObjectIfUnModifiedSince,
		Extended: true,  // HeadObject with if-unmodified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainHeadObjectIfMatch,
		Extended: true,  // HeadObject with if-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainHeadObjectIfNoneMatch,
		Extended: true,  // HeadObject with if-none-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for ListBuckets API.
	APItest{
		Test:     mainListBuckets,
		Extended: false, // ListBuckets is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for ListObjects API.
	APItest{
		Test:     mainListObjectsV1Prepared,
		Extended: false, // ListObjects is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainListObjectsV2Prepared,
		Extended: false, // ListObjects is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for PutObject streaming API.
	APItest{
		Test:     mainPutObjectStream,
		Extended: false, // PutObject streaming v4 is not an extended API.
		Critical: false, // Because -- has been used this object is not necessary for future tests.
	},

	APItest{
		Test:     mainPresignedPutObject,
		Extended: false, // PutObject presigned is not an extended API.
		Critical: false, // This object is not needed for future tests.
	},

	// Tests for PostObject API.
	APItest{
		Test:     mainPostObject,
		Extended: false, // PostObject is not an extended API.
		Critical: true,  // This test does affect other tests.
	},
	// Tests for Multipart API.
	APItest{
		Test:     mainInitiateMultipartUpload,
		Extended: false, // Initiate Multipart test must be run even without extended flags being set.
		Critical: true,  // Initiate Multipart test must pass before other tests can be run.
	},
	APItest{
		Test:     mainUploadPart,
		Extended: false, // Upload Part test must be run even without extended flag being set.
		Critical: true,  // Upload Part test must pass before other tests can be run.
	},
	APItest{
		Test:     mainReuploadPart,
		Extended: false, // Upload Part test must be run even without extended flag being set.
		Critical: true,  // Upload Part test must pass before other tests can be run.
	},
	APItest{
		Test:     mainListParts,
		Extended: false, // List Part test must be run even without extended flag being set.
		Critical: false, // List Part test can fail without affecting other tests.
	},
	APItest{
		Test:     mainListMultipartUploads,
		Extended: false, // List Multipart Uploads test must be run without extended flag being set.
		Critical: false, // List Multipart Uploads test can fail without affecting other tests.
	},
	APItest{
		Test:     mainCompleteMultipartUpload,
		Extended: false, // Complete Multipart test must be run even without extended flag being set.
		Critical: true,  // Complete Multipart test can fail without affecting other tests.
	},
	APItest{
		Test:     mainAbortMultipartUpload,
		Extended: false, // Abort Multipart test must be run even without extended flag being set.
		Critical: false, // Abort Multipart test can fail without affecting other tests.
	},

	// Tests for CopyObject API.
	APItest{
		Test:     mainCopyObject,
		Extended: false, // CopyObject is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfModifiedSince,
		Extended: true,  // CopyObject with if-modified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfUnModifiedSince,
		Extended: true,  // CopyObject with if-unmodified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfMatch,
		Extended: true,  // CopyObject with if-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfNoneMatch,
		Extended: true,  // CopyObject with if-none-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for GetObject API.
	APItest{
		Test:     mainGetObject,
		Extended: false, // GetObject is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectMultipart,
		Extended: false, // GetObject is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectPresigned,
		Extended: false, // GetObject Presigned is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	APItest{
		Test:     mainGetObjectIfModifiedSince,
		Extended: true,  // GetObject with if-modified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfUnModifiedSince,
		Extended: true,  // GetObject with if-unmodified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfMatch,
		Extended: true,  // GetObject with if-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfNoneMatch,
		Extended: true,  // GetObject with if-none-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectRange,
		Extended: true,  // GetObject with range header is an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Test for RemoveBucket API. (needs to be before remove object)
	APItest{
		Test:     mainRemoveBucketNotEmpty,
		Extended: false, // RemoveBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Test for RemoveObject API.
	APItest{
		Test:     mainRemoveObjectExists,
		Extended: false, // RemoveObject is not an extended API.
		Critical: true,  // This test does affect future tests.
	},
	APItest{
		Test:     mainRemoveObjectDNE,
		Extended: false, // RemoveObject is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for RemoveBucket API.
	APItest{
		Test:     mainRemoveBucketExists,
		Extended: false, // RemoveBucket is not an extended API.
		Critical: true,  // Removing this bucket is necessary for a good test.
	},
	APItest{
		Test:     mainRemoveBucketDNE,
		Extended: false, // RemoveBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
}

// Tests - holds all tests that must be run differently based on usage of the -- flag.
var unpreparedTests = []APItest{
	// Tests for PutBucket API.
	APItest{
		Test:     mainPutBucket,
		Extended: false, // PutBucket is not an extended API.
		Critical: true,  // This test does affect future tests.
	},
	APItest{
		Test:     mainPutBucketInvalid,
		Extended: false, // PutBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for PutBucketPolicy API.
	APItest{
		Test:     mainPutBucketPolicy,
		Extended: false, // PutBucketPolicy is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for GetBucketPolicy API.
	APItest{
		Test:     mainGetBucketPolicy,
		Extended: false, // GetBucketPolicy is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for PutObject API.
	APItest{
		Test:     mainPutObjectUnPrepared,
		Extended: false, // PutObject is not an extended API.
		Critical: true,  // These objects are necessary for future tests.
	},

	// Tests for HeadBucket API.
	APItest{
		Test:     mainHeadBucket,
		Extended: false, // HeadBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for HeadObject API.
	APItest{
		Test:     mainHeadObjectUnPrepared,
		Extended: false, // HeadObject is not an extended API.
		Critical: true,  // This test affects future tests and must pass.
	},
	APItest{
		Test:     mainHeadObjectIfModifiedSince,
		Extended: true,  // HeadObject with if-modified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainHeadObjectIfUnModifiedSince,
		Extended: true,  // HeadObject with if-unmodified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainHeadObjectIfMatch,
		Extended: true,  // HeadObject with if-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainHeadObjectIfNoneMatch,
		Extended: true,  // HeadObject with if-none-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for ListBuckets API.
	APItest{
		Test:     mainListBuckets,
		Extended: false, // ListBuckets is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for ListObjects API.
	APItest{
		Test:     mainListObjectsV1UnPrepared,
		Extended: false, // ListObjects is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainListObjectsV2UnPrepared,
		Extended: false, // ListObjects is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for PutObject Streaming API.
	APItest{
		Test:     mainPutObjectStream,
		Extended: false, // PutObject Streaming V4 is not an extended API.
		Critical: true,  // These objects are necessary for future tests.
	},

	APItest{
		Test:     mainPresignedPutObject,
		Extended: false, // PutObject presigned is not an extended API.
		Critical: true,  // This object is necessary for future tests.
	},

	// Tests for PostObject API.
	APItest{
		Test:     mainPostObject,
		Extended: false, // PostObject is not an extended API.
		Critical: true,  // This test does affect other tests.
	},

	// Tests for Multipart API.
	APItest{
		Test:     mainInitiateMultipartUpload,
		Extended: false, // Initiate Multipart test must be run even without extended flags being set.
		Critical: true,  // Initiate Multipart test must pass before other tests can be run.
	},
	APItest{
		Test:     mainUploadPart,
		Extended: false, // Upload Part test must be run even without extended flag being set.
		Critical: true,  // Upload Part test must pass before other tests can be run.
	},
	APItest{
		Test:     mainReuploadPart,
		Extended: false, // Upload Part test must be run even without extended flag being set.
		Critical: true,  // Upload Part test must pass before other tests can be run.
	},
	APItest{
		Test:     mainListParts,
		Extended: false, // List Part test must be run even without extended flag being set.
		Critical: false, // List Part test can fail without affecting other tests.
	},
	APItest{
		Test:     mainListMultipartUploads,
		Extended: false, // List Multipart Uploads test must be run without extended flag being set.
		Critical: false, // List Multipart Uploads test can fail without affecting other tests.
	},
	APItest{
		Test:     mainCompleteMultipartUpload,
		Extended: false, // Complete Multipart test must be run even without extended flag being set.
		Critical: true,  // Complete Multipart test can fail without affecting other tests.
	},
	APItest{
		Test:     mainAbortMultipartUpload,
		Extended: false, // Abort Multipart test must be run even without extended flag being set.
		Critical: false, // Abort Multipart test can fail without affecting other tests.
	},

	// Tests for CopyObject API.
	APItest{
		Test:     mainCopyObject,
		Extended: false, // CopyObject is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfModifiedSince,
		Extended: true,  // CopyObject with if-modified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfUnModifiedSince,
		Extended: true,  // CopyObject with if-unmodified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfMatch,
		Extended: true,  // CopyObject with if-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfNoneMatch,
		Extended: true,  // CopyObject with if-none-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for GetObject API.
	APItest{
		Test:     mainGetObject,
		Extended: false, // GetObject is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	// Tests for GetObject API.
	APItest{
		Test:     mainGetObjectMultipart,
		Extended: false, // GetObject is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectPresigned,
		Extended: false, // GetObject Presigned is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfModifiedSince,
		Extended: true,  // GetObject with if-modified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfUnModifiedSince,
		Extended: true,  // GetObject with if-unmodified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfMatch,
		Extended: true,  // GetObject with if-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfNoneMatch,
		Extended: true,  // GetObject with if-none-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectRange,
		Extended: true,  // GetObject with range header is an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Test for RemoveBucket API. (needs to be before remove object)
	APItest{
		Test:     mainRemoveBucketNotEmpty,
		Extended: false, // RemoveBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Test for RemoveObject API.
	APItest{
		Test:     mainRemoveObjectExists,
		Extended: false, // Remove Object test must be run.
		Critical: true,  // Remove Object test must pass for future tests.
	},
	APItest{
		Test:     mainRemoveObjectDNE,
		Extended: false, // RemoveObject is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for RemoveBucket API.
	APItest{
		Test:     mainRemoveBucketExists,
		Extended: false, // RemoveBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainRemoveBucketDNE,
		Extended: false, // RemoveBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
}
