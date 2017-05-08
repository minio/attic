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
	"time"

	"github.com/minio/cli"
	"github.com/minio/mc/pkg/httptracer"
)

// ServerConfig - container for all the user passed server info
// and a reusable http.Client
type ServerConfig struct {
	Access   string
	Secret   string
	Endpoint string
	Region   string
	Client   *http.Client
}

// newServerConfig - new server config.
func newServerConfig(ctx *cli.Context) (*ServerConfig, error) {
	// Set config fields from either flags or env. variables.
	serverCfg := &ServerConfig{
		Access:   ctx.String("access"),
		Secret:   ctx.String("secret"),
		Endpoint: ctx.String("url"),
		Region:   ctx.String("region"),
		Client: &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		},
	}
	// Region was not provided, we try to set a default region instead.
	if !ctx.IsSet("region") {
		endpointURL, err := url.Parse(serverCfg.Endpoint)
		if err != nil {
			return nil, err
		}
		serverCfg.Region = defaultRegion(endpointURL)
	}

	if ctx.Bool("verbose") || ctx.GlobalBool("verbose") {
		// Set up new tracer.
		serverCfg.Client.Transport = httptracer.GetNewTraceTransport(newTraceV4(), http.DefaultTransport)
	}
	return serverCfg, nil
}

// defaultRegion - returns default region for a given URL.
func defaultRegion(endpointURL *url.URL) string {
	// If this is an AmazonHost default the region to us-west-1.
	if isAmazonEndpoint(endpointURL) {
		return "us-west-1"
	}
	return globalDefaultRegion
}
