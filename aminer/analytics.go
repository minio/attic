/*
 * Aminer (C) 2014, 2015 Minio, Inc.
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
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/minio/cli"
	"gopkg.in/mgo.v2/bson"
)

// SSLAnalytics
const (
	SSLAnalytics   = "https://ssl.google-analytics.com/collect"
	DebugAnalytics = "https://www.google-analytics.com/debug/collect"
)

// Global constants
const (
	AppName    = "aminer"
	AppVersion = "0.0.1"
)

// GetUserAgent - get an rfc formatted user agent
func GetUserAgent(name string, version string, comments ...string) string {
	// if no name and version is set we do not add new user agents
	if name != "" && version != "" {
		return name + "/" + version + " (" + strings.Join(comments, "; ") + ") "
	}
	return ""
}

func updateGoogleAnalytics(c *configV1, result LogMessage) error {
	var payload bytes.Buffer
	payload.WriteString("v=1")
	// Tracking id UA-XXXXXXXX-1
	payload.WriteString("&tid=" + c.TID)
	// User unique id UUID
	payload.WriteString("&cid=" + newUUID())
	// Type of hit
	payload.WriteString("&t=pageview")
	// Data source
	payload.WriteString("&ds=downloads")
	// data referrer
	payload.WriteString("&dr=" + mustURLEncodeName(result.HTTP.Request.Header.Get("Referer")))
	// Document hostname
	payload.WriteString("&dh=" + result.HTTP.Request.Host)
	// Document title
	payload.WriteString("&dt=downloads")
	// Document path
	payload.WriteString("&dp=" + mustURLEncodeName(result.HTTP.Request.RequestURI))
	// UserAgent override
	payload.WriteString("&ua=" + mustURLEncodeName(result.HTTP.Request.Header.Get("User-Agent")))
	// IP Override
	payload.WriteString("&uip=" + strings.Split(result.HTTP.Request.RemoteAddr, ":")[0])
	if !c.Production {
		req, err := http.NewRequest("GET", DebugAnalytics+"?"+payload.String(), nil)
		if err != nil {
			return err
		}
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New("Data was not uploaded error: " + resp.Status)
		}
		var b bytes.Buffer
		io.Copy(&b, resp.Body)
		fmt.Println(b.String())
		return nil
	}
	req, err := http.NewRequest("POST", SSLAnalytics, &payload)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", GetUserAgent(AppName, AppVersion, runtime.GOOS, runtime.GOARCH))

	client := http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func runAnalyticsCmd(c *cli.Context) {
	conf, err := loadConfigV1()
	if err != nil {
		log.Fatal(err.Error())
	}
	s := connectToMongo(c)
	defer s.Close()

	var result LogMessage
	iter := db.Find(bson.M{"http.request.method": "GET"}).Iter()
	for iter.Next(&result) {
		if time.Since(result.StartTime) < time.Duration(24*time.Hour) {
			filters := strings.Split(c.GlobalString("filter"), ",")
			var skip bool
			for _, filter := range filters {
				if strings.Contains(result.HTTP.Request.RemoteAddr, filter) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			requestURI := result.HTTP.Request.RequestURI
			if result.StatusMessage == "" || result.StatusMessage == "OK" {
				for _, supportedBin := range supportedBinaries {
					if strings.HasSuffix(requestURI, supportedBin) {
						if err := updateGoogleAnalytics(conf, result); err != nil {
							log.Fatal(err.Error())
						}
					}
				}
			}
		}
	}
}
