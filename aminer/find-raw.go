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
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/minio/cli"
)

func runFindRawCmd(c *cli.Context) {
	if len(c.Args()) > 1 || c.Args().First() == "help" {
		cli.ShowCommandHelpAndExit(c, "find-raw", 1) // last argument is exit code
	}
	if strings.TrimSpace(c.Args().First()) == "" {
		cli.ShowCommandHelpAndExit(c, "find-raw", 1) // last argument is exit code
	}
	f, err := os.Open(strings.TrimSpace(c.Args().First()))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	var message LogMessage
	for scanner.Scan() {
		// ignore errors purposefully here, since json.Unmarshal fails for io.ReadCloser
		json.Unmarshal([]byte(scanner.Text()), &message)
		filters := strings.Split(c.GlobalString("filter"), ",")
		var skip bool
		for _, filter := range filters {
			if strings.Contains(message.HTTP.Request.RemoteAddr, filter) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		requestURI := message.HTTP.Request.RequestURI
		for _, supportedBin := range supportedBinaries {
			if strings.HasSuffix(requestURI, supportedBin) {
				if c.GlobalBool("json") {
					type resultJSON struct {
						Method     string
						RemoteAddr string
						RequestURI string
					}
					resultBytes, _ := json.Marshal(resultJSON{
						Method: message.HTTP.Request.Method,
						RemoteAddr: message.HTTP.Request.RemoteAddr,
						RequestURI: message.HTTP.Request.RequestURI,
					})
					fmt.Println(string(resultBytes))
					continue
				}
				fmt.Print(message.HTTP.Request.Method)
				fmt.Print("    ")
				fmt.Print(message.HTTP.Request.RemoteAddr)
				fmt.Print("    ")
				fmt.Print(message.HTTP.Request.RequestURI)
				fmt.Println("    ")
			}
		}
	}
}
