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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/minio/cli"
	"gopkg.in/mgo.v2/bson"
)

func runFindCmd(c *cli.Context) {
	if len(c.Args()) > 1 || c.Args().First() == "help" {
		cli.ShowCommandHelpAndExit(c, "find", 1) // last argument is exit code
	}
	s := connectToMongo(c)
	defer s.Close()
	var result LogMessage
	switch {
	case strings.ToUpper(c.Args().First()) == "GET":
		iter := db.Find(bson.M{"http.request.method": "GET"}).Iter()
		for iter.Next(&result) {
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
			for _, supportedBin := range supportedBinaries {
				if strings.HasSuffix(requestURI, supportedBin) {
					if c.GlobalBool("json") {
						type resultJSON struct {
							Method     string
							RemoteAddr string
							RequestURI string
						}
						resultBytes, _ := json.Marshal(resultJSON{
							Method: result.HTTP.Request.Method,
							RemoteAddr: result.HTTP.Request.RemoteAddr,
							RequestURI: result.HTTP.Request.RequestURI,
						})
						fmt.Println(string(resultBytes))
						continue
					}
					fmt.Print(result.HTTP.Request.Method)
					fmt.Print("    ")
					fmt.Print(result.HTTP.Request.RemoteAddr)
					fmt.Print("    ")
					fmt.Print(result.HTTP.Request.RequestURI)
					fmt.Println("    ")
				}
			}
		}
	case strings.ToUpper(c.Args().First()) == "HEAD":
		iter := db.Find(bson.M{"http.request.method": "HEAD"}).Iter()
		for iter.Next(&result) {
			filters := strings.Split(c.GlobalString("filter"), ",")
			for _, filter := range filters {
				if strings.Contains(result.HTTP.Request.RemoteAddr, filter) {
					continue
				}
			}
			if !strings.HasPrefix(result.HTTP.Request.RequestURI, "/updates/2015") {
				continue
			}
			fmt.Print(result.HTTP.Request.Method)
			fmt.Print("    ")
			fmt.Print(result.HTTP.Request.RemoteAddr)
			fmt.Print("    ")
			fmt.Print(result.HTTP.Request.RequestURI)
			fmt.Println("    ")
		}
	}
}
