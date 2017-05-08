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
	"math/rand"
	"sync"
	"time"

	"github.com/minio/cli"
	"github.com/minio/mc/pkg/console"
)

var (
	globalVerbose       bool          // Used to decide whether or not http traces will be printed.
	globalDefaultRegion = "us-east-1" // Default all aws requests to us-east-1 unless told otherwise.
	globalTotalNumTest  int           // The total number of tests being run.
	globalRandom        *rand.Rand    // A global random seed used by retry code.
	globalSuffix        string        // The suffix to append to all s3verify created objects and buckets.
	globalExcludes      []string      // Test to exclude.
)

const (
	globalS3verifyVersion = "1.0.0" // Current s3verify version.
	globalNumTestObjects  = 1001
)

// lockedRandSource provides protected rand source, implements rand.Source interface.
type lockedRandSource struct {
	lk  sync.Mutex
	src rand.Source
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an
// int64.
func (r *lockedRandSource) Int63() (n int64) {
	r.lk.Lock()
	n = r.src.Int63()
	r.lk.Unlock()
	return
}

// Seed uses the provided seed value to initialize the generator to a
// deterministic state.
func (r *lockedRandSource) Seed(seed int64) {
	r.lk.Lock()
	r.src.Seed(seed)
	r.lk.Unlock()
}

// Separate out context.
func setGlobals(verbose bool, numTests int, suffix string, excludes []string) {
	globalTotalNumTest = numTests
	globalVerbose = verbose
	if globalVerbose {
		// Allow printing of traces.
		console.DebugPrint = true
	}
	globalRandom = rand.New(&lockedRandSource{src: rand.NewSource(time.Now().UTC().UnixNano())})
	globalSuffix = suffix
	globalExcludes = excludes
}

// Set any global flags here.
func setGlobalsFromContext(ctx *cli.Context) error {
	verbose := ctx.Bool("verbose") || ctx.GlobalBool("verbose")
	numTests := 0
	// Calculate the total number of tests being run.
	if ctx.Bool("extended") || ctx.GlobalBool("extended") {
		numTests = len(unpreparedTests)
	} else {
		// The length of unpreparedTests == preparedTests.
		for _, test := range unpreparedTests {
			if !test.Extended {
				numTests++
			}
		}
	}
	// Standard suffix.
	suffix := "tmp-bkt"
	if ctx.GlobalString("reuse") != "" {
		suffix = ctx.GlobalString("reuse")
	}
	excludes := ctx.StringSlice("exclude")
	setGlobals(verbose, numTests, suffix, excludes)

	return nil
}
