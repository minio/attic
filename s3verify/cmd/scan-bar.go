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
	"fmt"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/minio/mc/pkg/console"
)

// Set up global cursor channel.
var cursorCh = cursorAnimate()

// Set up a constant width to follow.
const (
	messageWidth = 50
)

/******************************** Scan Bar ************************************/
// fixateScanBar truncates long text to fit within the terminal size.
func fixateScanBar(text string, width int) string {
	if len([]rune(text)) > width {
		// Trim text to fit within the screen
		trimSize := len([]rune(text)) - width + 3 //"..."
		if trimSize < len([]rune(text)) {
			text = "..." + text[trimSize:]
		}
	}
	return text
}

// Progress bar function report objects being scaned.
type scanBarFunc func(string)

// scanBarFactory returns a progress bar function to report URL scanning.
func scanBarFactory() scanBarFunc {
	prevLineSize := 0
	var termWidth int
	termWidth, e := pb.GetTerminalWidth()
	if e != nil {
		// Underlying terminal may not support this operation,
		// falling back to default 80 width.
		termWidth = 80
	}

	return func(message string) {
		scanPrefix := fmt.Sprintf("%s", message)
		padding := messageWidth - len([]rune(scanPrefix))

		message = fixateScanBar(message, termWidth-len([]rune(scanPrefix))-1)
		barText := scanPrefix + strings.Repeat(" ", padding) + string(<-cursorCh)

		if prevLineSize != 0 { // erase previous line
			console.PrintC("\r")
		}
		console.PrintC(barText + " \b ") // Remove only the cursor from the text.
		prevLineSize = len([]rune(barText))
	}
}
