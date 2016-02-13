/*
 * Minio Cloud Storage, (C) 2015 Minio, Inc.
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
	"net/http"
	"runtime"
)

type xlRPCService struct{}

func (s *xlRPCService) ListNodes(r *http.Request, arg *XLArg, rep *ListNodesRep) error {
	rep.Nodes = []struct {
		Hostname string `json:"hostname"`
		Address  string `json:"address"`
		ID       string `json:"id"`
	}{
		{
			Hostname: "localhost",
			Address:  "192.168.1.102:9000",
			ID:       "6F27CB16-493D-40FA-B035-2A2E5646066A",
		},
	}
	return nil
}

// Usage bytes
const (
	PB = 1024 * 1024 * 1024 * 1024
	TB = 1024 * 1024 * 1024 * 1024
	GB = 1024 * 1024 * 1024
)

func (s *xlRPCService) StorageStats(r *http.Request, arg *XLArg, rep *StorageStatsRep) error {
	rep.Buckets = []BucketStats{{"bucket1", 4 * TB}, {"bucket2", 120 * TB}, {"bucket3", 45 * TB}}
	return nil
}

func (s *xlRPCService) RebalanceStats(r *http.Request, arg *XLArg, rep *RebalanceStatsRep) error {
	rep.State = make(map[string]string)
	rep.State["bucket1/obj1"] = "inProgress"
	rep.State["bucket2/obj2"] = "finished"
	rep.State["bucket3/obj3"] = "errored"
	rep.State["bucket4/obj4"] = "unknownState"
	return nil
}

func (s *xlRPCService) Version(r *http.Request, arg *ServerArg, rep *XLVersionRep) error {
	rep.Version = "0.1.0"
	rep.Architecture = runtime.GOARCH
	rep.OperatingSystem = runtime.GOOS
	return nil
}
