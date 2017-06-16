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
	"encoding/hex"
	"errors"
	"strings"
	"unicode/utf8"
)

func mustURLEncodeName(name string) string {
	encodedName, _ := urlEncodeName(name)
	return encodedName
}

func urlEncodeName(name string) (string, error) {
	var encodedName string
	for _, s := range name {
		if 'A' <= s && s <= 'Z' || 'a' <= s && s <= 'z' || '0' <= s && s <= '9' { // Â§2.3 Unreserved characters (mark)
			encodedName = encodedName + string(s)
			continue
		}
		switch s {
		case '-', '_', '.', '~': // §2.3 Unreserved characters (mark)
			encodedName = encodedName + string(s)
			continue
		default:
			len := utf8.RuneLen(s)
			if len < 0 {
				return "", errors.New("invalid utf-8")
			}
			u := make([]byte, len)
			utf8.EncodeRune(u, s)
			for _, r := range u {
				hex := hex.EncodeToString([]byte{r})
				encodedName = encodedName + "%" + strings.ToUpper(hex)
			}
		}
	}
	return encodedName, nil
}
