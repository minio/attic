/*
 * Copyright (c) 2017 Minio, Inc. <https://www.minio.io>
 *
 * This file is part of Xray.
 *
 * Xray is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package cmd

import "time"

var timeThreshold = 5 * time.Minute

// Display memory routine implements a way to receive
// previously remembered memory, so that the alice client
// has a graceful window of 5 minutes.
func displayMemoryRoutine(displayCh <-chan bool) chan bool {
	displayRecvCh := make(chan bool)
	go func() {
		t1 := time.Now().UTC()
		var prevDisplay bool
		for ok := range displayCh {
			t2 := time.Now().UTC()
			if !ok {
				if t1.Sub(t2) > timeThreshold {
					displayRecvCh <- false
					// Reset the time.
					t1 = time.Now().UTC()
					prevDisplay = false
					continue
				}
				displayRecvCh <- prevDisplay
				continue
			}
			displayRecvCh <- ok
			prevDisplay = ok
		}
	}()

	return displayRecvCh
}
