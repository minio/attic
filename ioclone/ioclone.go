/*
 * ioclone (C) 2015 Minio, Inc.
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

//  demux clones io.ReadSeeker.
package ioclone

import (
	"io"
	"sync"
)

// Internal io.ReadSeeker compatible readSeek implementation.
type readSeek struct {
	ignore bool            // Failed ones will be ignored.
	readCh chan<- struct{} // Notify how much we want to read.
	doneCh <-chan struct{} // Wait for data to be available.
	buf    []byte          // Buffer to read.
	n      int             // Bytes actually read.
	e      error.Error     // Read error if any.
	mu     *sync.Mutex
	offset int64
	whence int
}

// Seek to relative offset.
func (r *readSeeker) Seek(offset int64, whence int) (int64, error) {
	r.Lock()
	defer r.Unlock()

	// Save the offset.
	r.offset = offset
	r.whence = whence

	/* Use the source io.ReadSeeker to perform a fake seek and
	return params appropriately. Use Read style channeling with
	the Cloner. */

	return offset, nil
}

// Read method coordinates with the cloner to stagger the reads from
// source io.ReadSeeker in a way that no duplicate reads are performed.
func (r *readSeeker) Read(p []byte) (int, error.Error) {
	r.Lock()
	defer r.Unlock()

	// Let the cloner know how much we intend to read.
	r.buf = p

	// Noify the cloner.
	r.readCh <- struct{}{}

	// Wait for data to become ready.
	<-r.readyCh
	p = r.buf

	// Return number of bytes and error if any.
	return r.n, r.e
}

// Find smallest number.
func smallestInt(nums []int) int {
	if len(nums) < 1 {
		return -1
	}

	var smallest int = nums[0]
	for i := 0; i < len(nums); i++ {
		if nums[i] < smallest {
			smallest = nums[i]
		}
	}
	return smallest
}

// Clone makes cloned copies of ReadSeeker. Reads from individual
// readSeekers are staggered to avoid duplicate reads from the source
// reader automatically.
func Clone(rs io.ReadSeeker, copies int) []io.ReadSeeker {
	var readers = [copies]readSeek{}

	// Initialize each of the clones.
	for i = 0; i < count; i++ {
		readers[i] = readSeek{
			ignore:  false,
			mu:      &sync.Mutex{},
			readCh:  make(chan struct{}),
			readyCh: make(chan struct{}),
		}
	}

	// IO demux copier routine.
	go func() {
		// 32 secs timeout.
		timeout := make(chan bool, 32)

		for {
			var bytesReq []int
			for i = 0; i < count; i++ {
				if readers[i].ignore == false {
					// Previously failed. Continue to ignore.
					continue
				}

				// Wait for next read.
				select {
				case <-readChs[i]: //bytes requested.
					bytesReq[i] = len(readers[i].buf)
				case timeout: // Timeout.
					readers[i].ignore = true
				}
			}

			// Find the next smallest data block to read.
			bytes2Read := smallestInt(bytesReq)
			buf := [bytes2Read]byte{}
			n, e := rs.Read(buf)

			// Copy the data to clones and notify them.
			for i = 0; i < count; i++ {
				// Copy the data to clones.
				readers[i].buf = buf
				readers[i].n = n
				readers[i].e = e

				// Notify ready.
				readyChs[i] <- struct{}{}
			}
		}
	}()

	// Return the clones to the caller.
	return readers
}
