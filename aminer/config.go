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
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/minio/cli"
	"github.com/minio/minio/pkg/quick"
)

// configV1
type configV1 struct {
	Version    string
	TID        string
	Production bool
}

// cached variables should *NEVER* be accessed directly from outside this file.
var cache sync.Pool

// isConfigExist returns err if config file doesn't exist
func isConfigExist() bool {
	configFile, err := getConfigPath()
	if err != nil {
		return false
	}
	if _, err := os.Stat(configFile); err != nil {
		return false
	}
	return true
}

func getConfigPath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	// For windows the path is slightly different
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(u.HomeDir, "miner\\miner.json"), nil
	default:
		return filepath.Join(u.HomeDir, ".miner/miner.json"), nil
	}
}

func loadConfigV1() (*configV1, error) {
	configFile, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	// Cached in private global variable.
	if v := cache.Get(); v != nil { // Use previously cached config.
		return v.(quick.Config).Data().(*configV1), nil
	}
	conf := new(configV1)
	qconf, err := quick.New(conf)
	if err != nil {
		return nil, err
	}
	err = qconf.Load(configFile)
	if err != nil {
		return nil, err
	}
	cache.Put(qconf)
	return qconf.Data().(*configV1), nil
}

func newConfigV1() *configV1 {
	conf := new(configV1)
	conf.Version = "0.0.1"
	conf.TID = ""
	return conf
}

func writeConfig() error {
	conf := newConfigV1()
	configFile, err := getConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(configFile), 0700); err != nil {
		return err
	}
	qconf, err := quick.New(conf)
	if err != nil {
		return err
	}
	return qconf.Save(configFile)
}

// newUUID generates a random UUID according to RFC 4122
func newUUID() string {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		panic(err)
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

func runConfigCmd(c *cli.Context) {
	if len(c.Args()) > 1 || c.Args().First() == "help" {
		cli.ShowCommandHelpAndExit(c, "config", 1) // last argument is exit code
	}
	switch {
	case strings.TrimSpace(c.Args().First()) == "generate":
		if !isConfigExist() {
			if err := writeConfig(); err != nil {
				log.Fatal(err)
			}
		}
	}
}
