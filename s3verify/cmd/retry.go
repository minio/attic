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
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MaxRetry is the maximum number of retries before stopping.
var MaxRetry = 5

// TCPretry holds all the errors that can and should be retried.
var TCPretry = []string{"i/o timeout", "net/http: TLS handshake timeout", "connection reset by peer", "read: operation timed out"}

// newRetryTimer creates a timer with binomial increasing delays
// until the maximum retry attempts are reached.
func newRetryTimer(maxRetry int, unit time.Duration, doneCh <-chan struct{}) <-chan int {
	attemptCh := make(chan int)

	binomialWait := func(attempt int) time.Duration {
		return unit * time.Duration(1<<uint(attempt-1))
	}

	go func(done <-chan struct{}) {
		defer close(attemptCh)
		for i := 0; i < maxRetry; i++ {
			attemptCh <- i + 1
			timer := time.NewTimer(binomialWait(i))
			select {
			case <-timer.C:
			case <-done:
				return
			}
		}
	}(doneCh)
	return attemptCh
}

// isNetErrorRetryable - is network error retryable.
func isNetErrorRetryable(err error) bool {
	switch err.(type) {
	case net.Error:
		switch err.(type) {
		case *net.DNSError, *net.OpError, net.UnknownNetworkError:
			return true
		case *url.Error:
			// For a URL error, where it replies back "connection closed"
			// retry again.
			if strings.Contains(err.Error(), "Connection closed by foreign host") {
				return true
			}
		default:
			for _, errString := range TCPretry {
				if strings.Contains(err.Error(), errString) {
					// If error is retriable, retry.
					return true
				}
			}
		}
	}
	return false
}

// List of AWS S3 error codes which are retryable.
var retryableS3Codes = map[string]struct{}{
	"RequestError":          {},
	"RequestTimeout":        {},
	"Throttling":            {},
	"ThrottlingException":   {},
	"RequestLimitExceeded":  {},
	"RequestThrottled":      {},
	"InternalError":         {},
	"ExpiredToken":          {},
	"ExpiredTokenException": {},
	// Add more AWS S3 codes here.
}

// isS3CodeRetryable - is s3 error code retryable.
func isS3CodeRetryable(s3Code string) (ok bool) {
	_, ok = retryableS3Codes[s3Code]
	return ok
}

// List of HTTP status codes which are retryable.
var retryableHTTPStatusCodes = map[int]struct{}{
	429: {}, // http.StatusTooManyRequests is not part of the Go 1.5 library, yet
	http.StatusInternalServerError: {},
	http.StatusBadGateway:          {},
	http.StatusServiceUnavailable:  {},
	// Add more HTTP status codes here.
}

// isHTTPStatusRetryable - is HTTP error code retryable.
func isHTTPStatusRetryable(httpStatusCode int) (ok bool) {
	_, ok = retryableHTTPStatusCodes[httpStatusCode]
	return ok
}
