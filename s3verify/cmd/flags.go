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

import "github.com/minio/cli"

// Collection of flags currently supported by every command.
var globalFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "access, a",
		Usage: "Set AWS S3 access key",
		// Allow env. variables to be used as well as flags.
		EnvVar: "S3_ACCESS",
	},
	cli.StringFlag{
		Name:  "secret, s",
		Usage: "Set AWS S3 secret key",
		// Allow env. variables to be used as well as flags.
		EnvVar: "S3_SECRET",
	},
	cli.StringFlag{
		Name:  "region, r",
		Usage: `Set AWS S3 region`,
		// Allow env. variables to used as well as flags.
		EnvVar: "S3_REGION",
	},
	cli.StringFlag{
		Name:   "url, u",
		Usage:  "URL to S3 compatible server",
		Value:  "https://s3.amazonaws.com",
		EnvVar: "S3_URL",
	},
	cli.BoolFlag{
		Name:  "verbose, v",
		Usage: "Enable verbose output",
	},
	cli.BoolFlag{
		Name:  "extended",
		Usage: "Enable testing of extra S3 APIs",
	},
	cli.StringSliceFlag{
		Name:  "exclude, x",
		Usage: "Exclude tests by name",
		Value: &cli.StringSlice{},
	},
	cli.StringFlag{
		Name:  "reuse",
		Usage: `Prepare or reuse a testing environment`,
	},
	cli.StringFlag{
		Name:  "clean",
		Usage: `Remove anything suffixed by the passed id`,
	},
	cli.BoolFlag{
		Name:  "version",
		Usage: "Print version",
	},
}
