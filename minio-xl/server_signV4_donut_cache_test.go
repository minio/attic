/*
 * Minio Cloud Storage, (C) 2014 Minio, Inc.
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

package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"encoding/hex"
	"encoding/xml"
	"net/http"
	"net/http/httptest"

	"github.com/minio/minio-xl/pkg/xl"
	. "gopkg.in/check.v1"
)

type MyAPIXLCacheSuite struct {
	root            string
	req             *http.Request
	body            io.ReadSeeker
	accessKeyID     string
	secretAccessKey string
}

var _ = Suite(&MyAPIXLCacheSuite{})

var testAPIXLCacheServer *httptest.Server

func (s *MyAPIXLCacheSuite) SetUpSuite(c *C) {
	root, err := ioutil.TempDir(os.TempDir(), "api-")
	c.Assert(err, IsNil)
	s.root = root

	conf := &xl.Config{}
	conf.Version = "0.0.1"
	conf.MaxSize = 100000
	xl.SetXLConfigPath(filepath.Join(root, "xl.json"))
	perr := xl.SaveConfig(conf)
	c.Assert(perr, IsNil)

	accessKeyID, perr := generateAccessKeyID()
	c.Assert(perr, IsNil)
	secretAccessKey, perr := generateSecretAccessKey()
	c.Assert(perr, IsNil)

	authConf := &AuthConfig{}
	authConf.Users = make(map[string]*AuthUser)
	authConf.Users[string(accessKeyID)] = &AuthUser{
		Name:            "testuser",
		AccessKeyID:     string(accessKeyID),
		SecretAccessKey: string(secretAccessKey),
	}
	s.accessKeyID = string(accessKeyID)
	s.secretAccessKey = string(secretAccessKey)

	SetAuthConfigPath(root)
	perr = SaveConfig(authConf)
	c.Assert(perr, IsNil)

	minioAPI := getNewAPI(false)
	httpHandler := getAPIHandler(false, minioAPI)
	go startTM(minioAPI)
	testAPIXLCacheServer = httptest.NewServer(httpHandler)
}

func (s *MyAPIXLCacheSuite) TearDownSuite(c *C) {
	os.RemoveAll(s.root)
	testAPIXLCacheServer.Close()
}

func (s *MyAPIXLCacheSuite) newRequest(method, urlStr string, contentLength int64, body io.ReadSeeker) (*http.Request, error) {
	t := time.Now().UTC()
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-amz-date", t.Format(iso8601Format))
	if method == "" {
		method = "POST"
	}

	// add Content-Length
	req.ContentLength = contentLength

	// add body
	switch {
	case body == nil:
		req.Body = nil
	default:
		req.Body = ioutil.NopCloser(body)
	}

	// save for subsequent use
	hash := func() string {
		switch {
		case body == nil:
			return hex.EncodeToString(sum256([]byte{}))
		default:
			sum256Bytes, _ := sum256Reader(body)
			return hex.EncodeToString(sum256Bytes)
		}
	}
	hashedPayload := hash()
	req.Header.Set("x-amz-content-sha256", hashedPayload)

	var headers []string
	vals := make(map[string][]string)
	for k, vv := range req.Header {
		if _, ok := ignoredHeaders[http.CanonicalHeaderKey(k)]; ok {
			continue // ignored header
		}
		headers = append(headers, strings.ToLower(k))
		vals[strings.ToLower(k)] = vv
	}
	headers = append(headers, "host")
	sort.Strings(headers)

	var canonicalHeaders bytes.Buffer
	for _, k := range headers {
		canonicalHeaders.WriteString(k)
		canonicalHeaders.WriteByte(':')
		switch {
		case k == "host":
			canonicalHeaders.WriteString(req.URL.Host)
			fallthrough
		default:
			for idx, v := range vals[k] {
				if idx > 0 {
					canonicalHeaders.WriteByte(',')
				}
				canonicalHeaders.WriteString(v)
			}
			canonicalHeaders.WriteByte('\n')
		}
	}

	signedHeaders := strings.Join(headers, ";")

	req.URL.RawQuery = strings.Replace(req.URL.Query().Encode(), "+", "%20", -1)
	encodedPath := getURLEncodedName(req.URL.Path)
	// convert any space strings back to "+"
	encodedPath = strings.Replace(encodedPath, "+", "%20", -1)

	//
	// canonicalRequest =
	//  <HTTPMethod>\n
	//  <CanonicalURI>\n
	//  <CanonicalQueryString>\n
	//  <CanonicalHeaders>\n
	//  <SignedHeaders>\n
	//  <HashedPayload>
	//
	canonicalRequest := strings.Join([]string{
		req.Method,
		encodedPath,
		req.URL.RawQuery,
		canonicalHeaders.String(),
		signedHeaders,
		hashedPayload,
	}, "\n")

	scope := strings.Join([]string{
		t.Format(yyyymmdd),
		"us-east-1",
		"s3",
		"aws4_request",
	}, "/")

	stringToSign := authHeaderPrefix + "\n" + t.Format(iso8601Format) + "\n"
	stringToSign = stringToSign + scope + "\n"
	stringToSign = stringToSign + hex.EncodeToString(sum256([]byte(canonicalRequest)))

	date := sumHMAC([]byte("AWS4"+s.secretAccessKey), []byte(t.Format(yyyymmdd)))
	region := sumHMAC(date, []byte("us-east-1"))
	service := sumHMAC(region, []byte("s3"))
	signingKey := sumHMAC(service, []byte("aws4_request"))

	signature := hex.EncodeToString(sumHMAC(signingKey, []byte(stringToSign)))

	// final Authorization header
	parts := []string{
		authHeaderPrefix + " Credential=" + s.accessKeyID + "/" + scope,
		"SignedHeaders=" + signedHeaders,
		"Signature=" + signature,
	}
	auth := strings.Join(parts, ", ")
	req.Header.Set("Authorization", auth)

	return req, nil
}

func (s *MyAPIXLCacheSuite) TestDeleteBucket(c *C) {
	request, err := s.newRequest("DELETE", testAPIXLCacheServer.URL+"/mybucket", 0, nil)
	c.Assert(err, IsNil)

	client := &http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusMethodNotAllowed)
}

func (s *MyAPIXLCacheSuite) TestDeleteObject(c *C) {
	request, err := s.newRequest("DELETE", testAPIXLCacheServer.URL+"/mybucket/myobject", 0, nil)
	c.Assert(err, IsNil)
	client := &http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusMethodNotAllowed)
}

func (s *MyAPIXLCacheSuite) TestNonExistantBucket(c *C) {
	request, err := s.newRequest("HEAD", testAPIXLCacheServer.URL+"/nonexistantbucket", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusNotFound)
}

func (s *MyAPIXLCacheSuite) TestEmptyObject(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/emptyobject", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/emptyobject/object", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/emptyobject/object", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	var buffer bytes.Buffer
	responseBody, err := ioutil.ReadAll(response.Body)
	c.Assert(err, IsNil)
	c.Assert(true, Equals, bytes.Equal(responseBody, buffer.Bytes()))
}

func (s *MyAPIXLCacheSuite) TestBucket(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/bucket", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("HEAD", testAPIXLCacheServer.URL+"/bucket", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)
}

func (s *MyAPIXLCacheSuite) TestObject(c *C) {
	buffer := bytes.NewReader([]byte("hello world"))
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/testobject", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/testobject/object", int64(buffer.Len()), buffer)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/testobject/object", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	responseBody, err := ioutil.ReadAll(response.Body)
	c.Assert(err, IsNil)
	c.Assert(responseBody, DeepEquals, []byte("hello world"))

}

func (s *MyAPIXLCacheSuite) TestMultipleObjects(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/multipleobjects", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/multipleobjects/object", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response, "NoSuchKey", "The specified key does not exist.", http.StatusNotFound)

	//// test object 1

	// get object
	buffer1 := bytes.NewReader([]byte("hello one"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/multipleobjects/object1", int64(buffer1.Len()), buffer1)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/multipleobjects/object1", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	// verify response data
	responseBody, err := ioutil.ReadAll(response.Body)
	c.Assert(err, IsNil)
	c.Assert(true, Equals, bytes.Equal(responseBody, []byte("hello one")))

	buffer2 := bytes.NewReader([]byte("hello two"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/multipleobjects/object2", int64(buffer2.Len()), buffer2)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/multipleobjects/object2", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	// verify response data
	responseBody, err = ioutil.ReadAll(response.Body)
	c.Assert(err, IsNil)
	c.Assert(true, Equals, bytes.Equal(responseBody, []byte("hello two")))

	buffer3 := bytes.NewReader([]byte("hello three"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/multipleobjects/object3", int64(buffer3.Len()), buffer3)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/multipleobjects/object3", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	// verify object
	responseBody, err = ioutil.ReadAll(response.Body)
	c.Assert(err, IsNil)
	c.Assert(true, Equals, bytes.Equal(responseBody, []byte("hello three")))
}

func (s *MyAPIXLCacheSuite) TestNotImplemented(c *C) {
	request, err := s.newRequest("GET", testAPIXLCacheServer.URL+"/bucket/object?policy", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusNotImplemented)

}

func (s *MyAPIXLCacheSuite) TestHeader(c *C) {
	request, err := s.newRequest("GET", testAPIXLCacheServer.URL+"/bucket/object", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)

	verifyError(c, response, "NoSuchKey", "The specified key does not exist.", http.StatusNotFound)
}

func (s *MyAPIXLCacheSuite) TestPutBucket(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/put-bucket", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("x-amz-acl", "private")

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/put-bucket-slash/", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("x-amz-acl", "private")

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)
}

func (s *MyAPIXLCacheSuite) TestPutObject(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/put-object", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("x-amz-acl", "private")

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	buffer1 := bytes.NewReader([]byte("hello world"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/put-object/object", int64(buffer1.Len()), buffer1)
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)
}

func (s *MyAPIXLCacheSuite) TestListBuckets(c *C) {
	request, err := s.newRequest("GET", testAPIXLCacheServer.URL+"/", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	var results ListBucketsResponse
	decoder := xml.NewDecoder(response.Body)
	err = decoder.Decode(&results)
	c.Assert(err, IsNil)
}

func (s *MyAPIXLCacheSuite) TestNotBeAbleToCreateObjectInNonexistantBucket(c *C) {
	buffer1 := bytes.NewReader([]byte("hello world"))
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/innonexistantbucket/object", int64(buffer1.Len()), buffer1)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response, "NoSuchBucket", "The specified bucket does not exist.", http.StatusNotFound)
}

func (s *MyAPIXLCacheSuite) TestHeadOnObject(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/headonobject", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("x-amz-acl", "private")

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	buffer1 := bytes.NewReader([]byte("hello world"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/headonobject/object1", int64(buffer1.Len()), buffer1)
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("HEAD", testAPIXLCacheServer.URL+"/headonobject/object1", 0, nil)
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)
}

func (s *MyAPIXLCacheSuite) TestHeadOnBucket(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/headonbucket", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("x-amz-acl", "private")

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("HEAD", testAPIXLCacheServer.URL+"/headonbucket", 0, nil)
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)
}

func (s *MyAPIXLCacheSuite) TestXMLNameNotInBucketListJson(c *C) {
	request, err := s.newRequest("GET", testAPIXLCacheServer.URL+"/", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("Accept", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	byteResults, err := ioutil.ReadAll(response.Body)
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(string(byteResults), "XML"), Equals, false)
}

func (s *MyAPIXLCacheSuite) TestXMLNameNotInObjectListJson(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/xmlnamenotinobjectlistjson", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("Accept", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/xmlnamenotinobjectlistjson", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("Accept", "application/json")

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	byteResults, err := ioutil.ReadAll(response.Body)
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(string(byteResults), "XML"), Equals, false)
}

func (s *MyAPIXLCacheSuite) TestContentTypePersists(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/contenttype-persists", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	buffer1 := bytes.NewReader([]byte("hello world"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/contenttype-persists/one", int64(buffer1.Len()), buffer1)
	delete(request.Header, "Content-Type")
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("HEAD", testAPIXLCacheServer.URL+"/contenttype-persists/one", 0, nil)
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.Header.Get("Content-Type"), Equals, "application/octet-stream")

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/contenttype-persists/one", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)
	c.Assert(response.Header.Get("Content-Type"), Equals, "application/octet-stream")

	buffer2 := bytes.NewReader([]byte("hello world"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/contenttype-persists/two", int64(buffer2.Len()), buffer2)
	delete(request.Header, "Content-Type")
	request.Header.Add("Content-Type", "application/json")
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("HEAD", testAPIXLCacheServer.URL+"/contenttype-persists/two", 0, nil)
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.Header.Get("Content-Type"), Equals, "application/octet-stream")

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/contenttype-persists/two", 0, nil)
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.Header.Get("Content-Type"), Equals, "application/octet-stream")
}

func (s *MyAPIXLCacheSuite) TestPartialContent(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/partial-content", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	buffer1 := bytes.NewReader([]byte("Hello World"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/partial-content/bar", int64(buffer1.Len()), buffer1)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	// prepare request
	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/partial-content/bar", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Range", "bytes=6-7")

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusPartialContent)
	partialObject, err := ioutil.ReadAll(response.Body)
	c.Assert(err, IsNil)

	c.Assert(string(partialObject), Equals, "Wo")
}

func (s *MyAPIXLCacheSuite) TestListObjectsHandlerErrors(c *C) {
	request, err := s.newRequest("GET", testAPIXLCacheServer.URL+"/objecthandlererrors-.", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/objecthandlererrors", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response, "NoSuchBucket", "The specified bucket does not exist.", http.StatusNotFound)

	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/objecthandlererrors", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("x-amz-acl", "private")

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/objecthandlererrors?max-keys=-2", 0, nil)
	c.Assert(err, IsNil)
	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response, "InvalidArgument", "Argument maxKeys must be an integer between 0 and 2147483647.", http.StatusBadRequest)
}

func (s *MyAPIXLCacheSuite) TestPutBucketErrors(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/putbucket-.", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("x-amz-acl", "private")

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/putbucket", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("x-amz-acl", "private")

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/putbucket", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("x-amz-acl", "private")

	response, err = client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response, "BucketAlreadyExists", "The requested bucket name is not available.", http.StatusConflict)

	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/putbucket?acl", 0, nil)
	c.Assert(err, IsNil)
	request.Header.Add("x-amz-acl", "unknown")

	response, err = client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response, "NotImplemented", "A header you provided implies functionality that is not implemented.", http.StatusNotImplemented)
}

func (s *MyAPIXLCacheSuite) TestGetObjectErrors(c *C) {
	request, err := s.newRequest("GET", testAPIXLCacheServer.URL+"/getobjecterrors", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response, "NoSuchBucket", "The specified bucket does not exist.", http.StatusNotFound)

	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/getobjecterrors", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/getobjecterrors/bar", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response, "NoSuchKey", "The specified key does not exist.", http.StatusNotFound)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/getobjecterrors-./bar", 0, nil)
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response, "InvalidBucketName", "The specified bucket is not valid.", http.StatusBadRequest)

}

func (s *MyAPIXLCacheSuite) TestGetObjectRangeErrors(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/getobjectrangeerrors", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	buffer1 := bytes.NewReader([]byte("Hello World"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/getobjectrangeerrors/bar", int64(buffer1.Len()), buffer1)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/getobjectrangeerrors/bar", 0, nil)
	request.Header.Add("Range", "bytes=7-6")
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response, "InvalidRange", "The requested range cannot be satisfied.", http.StatusRequestedRangeNotSatisfiable)
}

func (s *MyAPIXLCacheSuite) TestObjectMultipartAbort(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/objectmultipartabort", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, 200)

	request, err = s.newRequest("POST", testAPIXLCacheServer.URL+"/objectmultipartabort/object?uploads", 0, nil)
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	decoder := xml.NewDecoder(response.Body)
	newResponse := &InitiateMultipartUploadResponse{}

	err = decoder.Decode(newResponse)
	c.Assert(err, IsNil)
	c.Assert(len(newResponse.UploadID) > 0, Equals, true)
	uploadID := newResponse.UploadID

	buffer1 := bytes.NewReader([]byte("hello world"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/objectmultipartabort/object?uploadId="+uploadID+"&partNumber=1", int64(buffer1.Len()), buffer1)
	c.Assert(err, IsNil)

	response1, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response1.StatusCode, Equals, http.StatusOK)

	buffer2 := bytes.NewReader([]byte("hello world"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/objectmultipartabort/object?uploadId="+uploadID+"&partNumber=2", int64(buffer2.Len()), buffer2)
	c.Assert(err, IsNil)

	response2, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response2.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("DELETE", testAPIXLCacheServer.URL+"/objectmultipartabort/object?uploadId="+uploadID, 0, nil)
	c.Assert(err, IsNil)

	response3, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response3.StatusCode, Equals, http.StatusNoContent)
}

func (s *MyAPIXLCacheSuite) TestBucketMultipartList(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/bucketmultipartlist", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, 200)

	request, err = s.newRequest("POST", testAPIXLCacheServer.URL+"/bucketmultipartlist/object?uploads", 0, nil)
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	decoder := xml.NewDecoder(response.Body)
	newResponse := &InitiateMultipartUploadResponse{}

	err = decoder.Decode(newResponse)
	c.Assert(err, IsNil)
	c.Assert(len(newResponse.UploadID) > 0, Equals, true)
	uploadID := newResponse.UploadID

	buffer1 := bytes.NewReader([]byte("hello world"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/bucketmultipartlist/object?uploadId="+uploadID+"&partNumber=1", int64(buffer1.Len()), buffer1)
	c.Assert(err, IsNil)

	response1, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response1.StatusCode, Equals, http.StatusOK)

	buffer2 := bytes.NewReader([]byte("hello world"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/bucketmultipartlist/object?uploadId="+uploadID+"&partNumber=2", int64(buffer2.Len()), buffer2)
	c.Assert(err, IsNil)

	response2, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response2.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/bucketmultipartlist?uploads", 0, nil)
	c.Assert(err, IsNil)

	response3, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response3.StatusCode, Equals, http.StatusOK)

	decoder = xml.NewDecoder(response3.Body)
	newResponse3 := &ListMultipartUploadsResponse{}
	err = decoder.Decode(newResponse3)
	c.Assert(err, IsNil)
	c.Assert(newResponse3.Bucket, Equals, "bucketmultipartlist")
}

func (s *MyAPIXLCacheSuite) TestObjectMultipartList(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/objectmultipartlist", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, 200)

	request, err = s.newRequest("POST", testAPIXLCacheServer.URL+"/objectmultipartlist/object?uploads", 0, nil)
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	decoder := xml.NewDecoder(response.Body)
	newResponse := &InitiateMultipartUploadResponse{}

	err = decoder.Decode(newResponse)
	c.Assert(err, IsNil)
	c.Assert(len(newResponse.UploadID) > 0, Equals, true)
	uploadID := newResponse.UploadID

	buffer1 := bytes.NewReader([]byte("hello world"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/objectmultipartlist/object?uploadId="+uploadID+"&partNumber=1", int64(buffer1.Len()), buffer1)
	c.Assert(err, IsNil)

	response1, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response1.StatusCode, Equals, http.StatusOK)

	buffer2 := bytes.NewReader([]byte("hello world"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/objectmultipartlist/object?uploadId="+uploadID+"&partNumber=2", int64(buffer2.Len()), buffer2)
	c.Assert(err, IsNil)

	response2, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response2.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/objectmultipartlist/object?uploadId="+uploadID, 0, nil)
	c.Assert(err, IsNil)

	response3, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response3.StatusCode, Equals, http.StatusOK)

	request, err = s.newRequest("GET", testAPIXLCacheServer.URL+"/objectmultipartlist/object?max-parts=-2&uploadId="+uploadID, 0, nil)
	c.Assert(err, IsNil)

	response4, err := client.Do(request)
	c.Assert(err, IsNil)
	verifyError(c, response4, "InvalidArgument", "Argument maxParts must be an integer between 1 and 10000.", http.StatusBadRequest)
}

func (s *MyAPIXLCacheSuite) TestObjectMultipart(c *C) {
	request, err := s.newRequest("PUT", testAPIXLCacheServer.URL+"/objectmultiparts", 0, nil)
	c.Assert(err, IsNil)

	client := http.Client{}
	response, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, 200)

	request, err = s.newRequest("POST", testAPIXLCacheServer.URL+"/objectmultiparts/object?uploads", 0, nil)
	c.Assert(err, IsNil)

	client = http.Client{}
	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	decoder := xml.NewDecoder(response.Body)
	newResponse := &InitiateMultipartUploadResponse{}

	err = decoder.Decode(newResponse)
	c.Assert(err, IsNil)
	c.Assert(len(newResponse.UploadID) > 0, Equals, true)
	uploadID := newResponse.UploadID

	buffer1 := bytes.NewReader([]byte("hello world"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/objectmultiparts/object?uploadId="+uploadID+"&partNumber=1", int64(buffer1.Len()), buffer1)
	c.Assert(err, IsNil)

	client = http.Client{}
	response1, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response1.StatusCode, Equals, http.StatusOK)

	buffer2 := bytes.NewReader([]byte("hello world"))
	request, err = s.newRequest("PUT", testAPIXLCacheServer.URL+"/objectmultiparts/object?uploadId="+uploadID+"&partNumber=2", int64(buffer2.Len()), buffer2)
	c.Assert(err, IsNil)

	client = http.Client{}
	response2, err := client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response2.StatusCode, Equals, http.StatusOK)

	// complete multipart upload
	completeUploads := &xl.CompleteMultipartUpload{
		Part: []xl.CompletePart{
			{
				PartNumber: 1,
				ETag:       response1.Header.Get("ETag"),
			},
			{
				PartNumber: 2,
				ETag:       response2.Header.Get("ETag"),
			},
		},
	}

	completeBytes, err := xml.Marshal(completeUploads)
	c.Assert(err, IsNil)

	request, err = s.newRequest("POST", testAPIXLCacheServer.URL+"/objectmultiparts/object?uploadId="+uploadID, int64(len(completeBytes)), bytes.NewReader(completeBytes))
	c.Assert(err, IsNil)

	response, err = client.Do(request)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)
}

func verifyError(c *C, response *http.Response, code, description string, statusCode int) {
	data, err := ioutil.ReadAll(response.Body)
	c.Assert(err, IsNil)
	errorResponse := APIErrorResponse{}
	err = xml.Unmarshal(data, &errorResponse)
	c.Assert(err, IsNil)
	c.Assert(errorResponse.Code, Equals, code)
	c.Assert(errorResponse.Message, Equals, description)
	c.Assert(response.StatusCode, Equals, statusCode)
}
