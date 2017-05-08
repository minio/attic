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
	"time"
)

const expirationDateFormat = "2006-01-02T15:04:05.999Z"

// TODO: so far this function only creates valid policies.
// Should create some invalid cases to check for error messages.

// newPostPolicyBytes - creates a bare bones postpolicy string with key and bucket matches.
func newPostPolicyBytes(credential, bucketName, objectKey string, expiration time.Time) []byte {
	t := time.Now().UTC()
	// Add the expiration date.
	expirationStr := fmt.Sprintf(`"expiration": "%s"`, expiration.Format(expirationDateFormat))
	// Add the bucket condition, only accept buckets equal to the one passed.
	bucketConditionStr := fmt.Sprintf(`["eq", "$bucket", "%s"]`, bucketName)
	// Add the key condition, only accept keys equal to the one passed.
	keyConditionStr := fmt.Sprintf(`["eq", "$key", "%s"]`, objectKey)
	// Add the algorithm condition, only accept AWS SignV4 Sha256.
	algorithmConditionStr := `["eq", "$x-amz-algorithm", "AWS4-HMAC-SHA256"]`
	// Add the date condition, only accept the current date.
	dateConditionStr := fmt.Sprintf(`["eq", "$x-amz-date", "%s"]`, t.Format(iso8601DateFormat))
	// Add the credential string, only accept the credential passed.
	credentialConditionStr := fmt.Sprintf(`["eq", "$x-amz-credential", "%s"]`, credential)

	// Combine all conditions into one string.
	conditionStr := fmt.Sprintf(`"conditions":[%s, %s, %s, %s, %s]`, bucketConditionStr, keyConditionStr, algorithmConditionStr, dateConditionStr, credentialConditionStr)
	retStr := "{"
	retStr = retStr + expirationStr + ","
	retStr = retStr + conditionStr
	retStr = retStr + "}"

	return []byte(retStr)
}
