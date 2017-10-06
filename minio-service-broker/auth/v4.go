package auth

// Implements minimal AWS-Signature-V4 features. Does not implement SHA256, presigned requests.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/textproto"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	SignV4Algorithm = "AWS4-HMAC-SHA256"
	DateFormat      = "20060102T150405Z"
	yyyymmdd        = "20060102"
)

// Sign-V4 calculation rule for single chunk is given here:
// http://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-header-based-auth.html

// These headers are ignored for signature calculation.
var ignoredHeaders = map[string]bool{
	"authorization":  true,
	"content-type":   true,
	"content-length": true,
	"user-agent":     true,
}

// CredentialsV4 - S3 creds for Sign-V4.
type CredentialsV4 struct {
	AccessKey string
	SecretKey string
	Region    string
}

// sum256 calculate sha256 sum for an input byte array.
func sum256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// sumHMAC calculate hmac between two input byte array.
func sumHMAC(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}

func sortQuery(encodedQuery string) string {
	m, _ := url.ParseQuery(encodedQuery)
	return strings.Replace(m.Encode(), "+", "%20", -1)
}

func encodePath(pathName string) string {
	var encodedPathname string
	for _, s := range pathName {
		if 'A' <= s && s <= 'Z' || 'a' <= s && s <= 'z' || '0' <= s && s <= '9' { // §2.3 Unrserved characters (mark)
			encodedPathname = encodedPathname + string(s)
			continue
		}
		switch s {
		case '-', '_', '.', '~', '/': // §2.3 Unreserved characters (mark)
			encodedPathname = encodedPathname + string(s)
			continue
		default:
			len := utf8.RuneLen(s)
			if len < 0 {
				// if utf8 cannot convert return the same string as is
				return pathName
			}
			u := make([]byte, len)
			utf8.EncodeRune(u, s)
			for _, r := range u {
				hex := hex.EncodeToString([]byte{r})
				encodedPathname = encodedPathname + "%" + strings.ToUpper(hex)
			}
		}
	}
	return encodedPathname
}

func getRequestSignedHeaders(r *http.Request) (h http.Header) {
	h = make(http.Header)
	authz := r.Header.Get("Authorization")
	signedHeadersField := strings.Split(authz, " ")
	if len(signedHeadersField) < 3 {
		return
	}
	signedHeadersList := strings.Split(signedHeadersField[2], "=")
	if len(signedHeadersField) < 2 {
		return
	}
	for _, header := range strings.Split(signedHeadersList[1], ";") {
		header = strings.TrimSuffix(header, ",")
		h.Set(header, r.Header.Get(header))
	}
	h.Set("Host", r.Host)
	return h
}

// Sign - Add Authorization header using AWS Signature-V4.
func (c CredentialsV4) Sign(r *http.Request) {
	date := time.Now().UTC()
	dateStr := date.Format(DateFormat)
	r.Header.Set("X-Amz-Date", dateStr)                      // Mandatory for V4 signature.
	r.Header.Set("Host", r.Host)                             // Host header at the ingress will be availabe as r.Host
	r.Header.Set("X-Amz-Content-Sha256", "UNSIGNED-PAYLOAD") // We don't compute SHA256 for the data.

	authz := c.authz(r.Method, encodePath(r.URL.Path), r.URL.Query().Encode(), r.Header)
	r.Header.Set("Authorization", authz)
}

// IsSigned - Return if the request is signed using correct credentials.
func (c CredentialsV4) IsSigned(r *http.Request) bool {
	gotAuthz := r.Header.Get("Authorization")
	headers := getRequestSignedHeaders(r)
	expectedAuthz := c.authz(r.Method, encodePath(r.URL.Path), r.URL.Query().Encode(), headers)
	return gotAuthz == expectedAuthz
}

// Return Authorization header.
func (c CredentialsV4) authz(method string, encodedResource string, encodedQuery string, headers http.Header) (authHeader string) {
	dateStr := headers.Get("X-Amz-Date")
	date, _ := time.Parse(DateFormat, dateStr)

	canonicalReq := getCanonicalRequest(method, encodedResource, encodedQuery, headers)
	stringToSign := strings.Join([]string{
		SignV4Algorithm,
		date.Format(DateFormat),
		getScope(c.Region, date),
		hex.EncodeToString(sum256([]byte(canonicalReq))),
	}, "\n")

	signingKey := getSigningKey(c.SecretKey, c.Region, date)

	credential := getCredential(c.AccessKey, c.Region, date)

	signedHeaders := getSignedHeaders(headers)

	signature := getSignature(signingKey, stringToSign)

	authHeader = strings.Join([]string{
		SignV4Algorithm + " Credential=" + credential,
		" SignedHeaders=" + signedHeaders,
		" Signature=" + signature,
	}, ",")

	return authHeader
}

// Canonical request is constructed as:
// <HTTPMethod>\n
// <CanonicalURI>\n
// <CanonicalQueryString>\n
// <CanonicalHeaders>\n
// <SignedHeaders>\n
// <HashedPayload>
func getCanonicalRequest(method string, encodedResource string, encodedQuery string, headers http.Header) string {
	return strings.Join([]string{
		method,
		encodedResource,
		sortQuery(encodedQuery),
		getCanonicalHeaders(headers),
		getSignedHeaders(headers),
		getHashedPayload(headers),
	}, "\n")
}

// Canonical headers is constructed as:
// Lowercase(<HeaderName1>)+":"+Trim(<value>)+"\n"
// Lowercase(<HeaderName2>)+":"+Trim(<value>)+"\n"
// ...
// Lowercase(<HeaderNameN>)+":"+Trim(<value>)+"\n"
func getCanonicalHeaders(headers http.Header) string {
	keys := signedHeaders(headers)
	var canonicalHeaders []string
	for _, key := range keys {
		canonicalHeaders = append(canonicalHeaders,
			key+":"+strings.Join(headers[textproto.CanonicalMIMEHeaderKey(key)], ","),
		)
	}
	return strings.Join(canonicalHeaders, "\n") + "\n"
}

func signedHeaders(headers http.Header) []string {
	var keys []string
	for key := range headers {
		lkey := strings.ToLower(key)
		if ignoredHeaders[lkey] {
			continue
		}
		keys = append(keys, lkey)
	}
	sort.Strings(keys)
	return keys
}

// SignedHeaders is an alphabetically sorted, semicolon-separated list of lowercase request header names.
// The request headers in the list are the same headers that you included in the CanonicalHeaders string.
func getSignedHeaders(headers http.Header) string {
	return strings.Join(signedHeaders(headers), ";")
}

// Sha256 sum of the payload.
func getHashedPayload(headers http.Header) string {
	hashedPayload := headers.Get("X-Amz-Content-Sha256")
	if hashedPayload == "" {
		hashedPayload = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	}
	return hashedPayload
}

func getScope(region string, date time.Time) string {
	scope := strings.Join([]string{
		date.Format(yyyymmdd),
		region,
		"s3",
		"aws4_request",
	}, "/")
	return scope
}

func getCredential(accessKey, region string, date time.Time) string {
	return accessKey + "/" + getScope(region, date)
}

func getSigningKey(secret, region string, date time.Time) []byte {
	dateKey := sumHMAC([]byte("AWS4"+secret), []byte(date.Format(yyyymmdd)))
	dateRegionKey := sumHMAC(dateKey, []byte(region))
	dateRegionServiceKey := sumHMAC(dateRegionKey, []byte("s3"))
	signingKey := sumHMAC(dateRegionServiceKey, []byte("aws4_request"))
	return signingKey
}

func getSignature(signingKey []byte, stringToSign string) string {
	return hex.EncodeToString(sumHMAC(signingKey, []byte(stringToSign)))
}
