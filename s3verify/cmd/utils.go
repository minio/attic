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
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/minio/mc/pkg/console"
)

const (
	appName       = "s3verify"
	letterIdxBits = 6                    // 6 bits to represetn a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting into 63 bits.
	letterBytes   = "abcdefghijklmnopqrstuvwxyz01234569"
)

// User Agent should always follow the style below.
// Please open an issue to discuss any new changes here.
//
//		Minio (OS; ARCH) LIB/VER APP/VER
const (
	appUserAgentPrefix = "Minio (" + runtime.GOOS + "; " + runtime.GOARCH + ") "
	appUserAgent       = appUserAgentPrefix + appName + "/" + globalS3verifyVersion
)

// List of success status.
var successStatus = []int{
	http.StatusOK,
	http.StatusNoContent,
	http.StatusPartialContent,
}

// List of known, valid headers taken from http://docs.aws.amazon.com/AmazonS3/latest/API/RESTCommonResponseHeaders.html
var validResponseHeaders = map[string]struct{}{
	"accept-ranges":       struct{}{},
	"cache-control":       struct{}{},
	"content-length":      struct{}{},
	"content-type":        struct{}{},
	"connection":          struct{}{},
	"content-disposition": struct{}{},
	"content-language":    struct{}{},
	"date":                struct{}{},
	"etag":                struct{}{},
	"expires":             struct{}{},
	"last-modified":       struct{}{},
	"location":            struct{}{},
	"server":              struct{}{},
	"vary":                struct{}{},
	"x-amz-delete-marker": struct{}{},
	"x-amz-id-2":          struct{}{},
	"x-amz-request-id":    struct{}{},
	"x-amz-version-id":    struct{}{},
	"x-amz-bucket-region": struct{}{},
}

// printMessage - Print test pass/fail messages with errors.
func printMessage(message string, err error) {
	// Erase the old progress line.
	console.Eraseline()
	if err != nil {
		message += strings.Repeat(" ", messageWidth-len([]rune(message))) + "[FAIL]\n" + err.Error()
		console.Println(message)
	} else {
		message += strings.Repeat(" ", messageWidth-len([]rune(message))) + "[OK]"
		console.Println(message)
	}
}

// verifyHostReachable - Execute a simple get request against the provided endpoint to make sure its reachable.
func verifyHostReachable(endpoint, region string) error {
	targetURL, err := makeTargetURL(endpoint, "", "", region, nil)
	if err != nil {
		return err
	}
	client := &http.Client{
		// Only give server 3 seconds to complete the request.
		Timeout: 3000 * time.Millisecond,
	}
	req := &http.Request{
		Method: "GET",
		URL:    targetURL,
	}
	if _, err := client.Do(req); err != nil {
		return err
	}
	return nil
}

// xmlDecoder provide decoded value in xml.
func xmlDecoder(body io.Reader, v interface{}) error {
	d := xml.NewDecoder(body)
	return d.Decode(v)
}

// jsonDecoder provide decoded value in json.
func jsonDecoder(body io.Reader, v interface{}) error {
	d := json.NewDecoder(body)
	return d.Decode(v)
}

// closeResponse close non nil response with any response Body.
// convenient wrapper to drain any remaining data on response body.
//
// Subsequently this allows golang http RoundTripper
// to re-use the same connection for future requests. (Connection pooling).
func closeResponse(res *http.Response) {
	// Callers should close resp.Body when done reading from it.
	// If resp.Body is not closed, the Client's underlying RoundTripper
	// (typically Transport) may not be able to re-use a persistent TCP
	// connection to the server for a subsequent "keep-alive" request.
	if res != nil && res.Body != nil {
		// Drain any remaining Body and then close the connection.
		// Without this closing connection would disallow re-using
		// the same connection for future uses.
		// - http://stackoverflow.com/a/17961593/4465767
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}
}

// randString generates random names.
func randString(n int, src rand.Source, prefix string) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return prefix + string(b[0:30-len(prefix)])
}

// Check if the endpoint is for an AWS S3 server.
func isAmazonEndpoint(endpointURL *url.URL) bool {
	if endpointURL == nil {
		return false
	}
	if endpointURL.Host == "s3.amazonaws.com" || endpointURL.Host == "s3.cn-north-1.amazonaws.com.cn" {
		return true
	}
	return false
}

// Generate a new URL from the user provided endpoint.
func makeTargetURL(endpoint, bucketName, objectName, region string, queryValues url.Values) (*url.URL, error) {
	targetURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if isAmazonEndpoint(targetURL) { // Change host to reflect the region.
		targetURL.Host = getS3Endpoint(region)
	}
	targetURL.Path = "/"
	if bucketName != "" {
		targetURL.Path = "/" + bucketName + "/" + objectName // Use path style requests only.
	}
	if len(queryValues) > 0 { // If there are query values include them.
		targetURL.RawQuery = queryValues.Encode()
	}
	return targetURL, nil
}

// Verify the date field of an HTTP response is formatted with HTTP time format.
func verifyDate(respDateStr string) error {
	_, err := time.Parse(http.TimeFormat, respDateStr)
	if err != nil {
		err = fmt.Errorf("Invalid time format received, expected http.TimeFormat")
		return err
	}
	return nil
}

// Verify all standard headers in an HTTP response.
func verifyStandardHeaders(header http.Header) error {
	for headerName, values := range map[string][]string(header) {
		if _, ok := validResponseHeaders[strings.ToLower(headerName)]; !ok {
			return fmt.Errorf("Invalid response header received: %s with values: %v", headerName, values)
		}
	}

	// Check the date header.
	respDateStr := header.Get("Date")
	if err := verifyDate(respDateStr); err != nil {
		return err
	}
	return nil
}

// Generate MD5 and SHA256 for an input readseeker.
func computeHash(reader io.ReadSeeker) (md5Sum, sha256Sum []byte, contentLength int64, err error) {
	// MD5 and SHA256 hasher.
	var hashMD5, hashSHA256 hash.Hash
	// MD5 and SHA256 hasher.
	hashMD5 = md5.New()
	hashSHA256 = sha256.New()
	hashWriter := io.MultiWriter(hashMD5, hashSHA256)

	// If no buffer is provided, no need to allocate just use io.Copy
	contentLength, err = io.Copy(hashWriter, reader)
	if err != nil {
		return nil, nil, 0, err
	}
	// Seek back to beginning location.
	if _, err := reader.Seek(0, 0); err != nil {
		return nil, nil, 0, err
	}
	// Finalize md5sum and sha256sum.
	md5Sum = hashMD5.Sum(nil)
	sha256Sum = hashSHA256.Sum(nil)

	return md5Sum, sha256Sum, contentLength, nil
}
