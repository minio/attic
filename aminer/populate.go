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
	"log"
	"os"
	"strings"

	"github.com/minio/cli"
	"gopkg.in/mgo.v2/bson"
)

func runPopulateCmd(c *cli.Context) {
	if len(c.Args()) > 1 || c.Args().First() == "help" {
		cli.ShowCommandHelpAndExit(c, "populate", 1) // last argument is exit code
	}
	if strings.TrimSpace(c.Args().First()) == "" {
		cli.ShowCommandHelpAndExit(c, "populate", 1) // last argument is exit code
	}
	s := connectToMongo(c)
	defer s.Close()
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
		// do not insert if already exists
		err = db.Find(bson.M{"http.request.remoteaddr": message.HTTP.Request.RemoteAddr}).One(nil)
		if err != nil {
			err = db.Insert(&message)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
