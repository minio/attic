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
	"os"
	"path/filepath"

	"github.com/minio/cli"
	"github.com/minio/minio-xl/pkg/probe"
	"github.com/minio/minio-xl/pkg/xl"
)

var (
	xlSubCommands = []cli.Command{
		{
			Name:        "make",
			Description: "make a xl",
			Action:      makeXLMain,
			CustomHelpTemplate: `NAME:
  minio-xl xl {{.Name}} - {{.Description}}

USAGE:
  minio-xl xl {{.Name}} XL-NAME [DISKS...]

EXAMPLES:
  1. Make a xl with 4 exports
      $ minio-xl xl {{.Name}} mongodb-backup /mnt/export1 /mnt/export2 /mnt/export3 /mnt/export4

  2. Make a xl with 16 exports
      $ minio-xl xl {{.Name}} operational-data /mnt/export1 /mnt/export2 /mnt/export3 /mnt/export4 /mnt/export5 \
       /mnt/export6 /mnt/export7 /mnt/export8 /mnt/export9 /mnt/export10 /mnt/export11 \
       /mnt/export12 /mnt/export13 /mnt/export14 /mnt/export15 /mnt/export16
`,
		},
	}

	xlCmd = cli.Command{
		Name:        "xl",
		Usage:       "Create and manage a xl configuration",
		Subcommands: xlSubCommands,
	}
)

func makeXLMain(c *cli.Context) {
	if !c.Args().Present() || c.Args().First() == "help" {
		cli.ShowCommandHelpAndExit(c, "make", 1)
	}
	xlName := c.Args().First()
	if c.Args().First() != "" {
		if !xl.IsValidXL(xlName) {
			Fatalf("Invalid xlname %s\n", xlName)
		}
	}
	var disks []string
	for _, disk := range c.Args().Tail() {
		if _, err := isUsable(disk); err != nil {
			Fatalln(err.Trace())
		}
		disks = append(disks, disk)
	}
	for _, disk := range disks {
		if err := os.MkdirAll(filepath.Join(disk, xlName), 0700); err != nil {
			Fatalln(probe.NewError(err))
		}
	}

	hostname, err := os.Hostname()
	if err != nil {
		Fatalln(probe.NewError(err))
	}
	xlConfig := &xl.Config{}
	xlConfig.Version = "0.0.1"
	xlConfig.XLName = xlName
	xlConfig.NodeDiskMap = make(map[string][]string)
	// keep it in exact order as it was specified, do not try to sort disks
	xlConfig.NodeDiskMap[hostname] = disks
	// default cache is unlimited
	xlConfig.MaxSize = 512000000

	if err := xl.SaveConfig(xlConfig); err != nil {
		Fatalln(err.Trace())
	}

	Infoln("Success!")
}
