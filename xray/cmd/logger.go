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

 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package cmd

import (
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
)

var rlog = logrus.New()

// Get file, line, function name of the caller.
func callerSource() string {
	pc, file, line, success := runtime.Caller(2)
	if !success {
		file = "<unknown>"
		line = 0
	}
	file = path.Base(file)
	name := runtime.FuncForPC(pc).Name()
	name = strings.TrimPrefix(name, "github.com/minio/minio/cmd.")
	return fmt.Sprintf("[%s:%d:%s()]", file, line, name)
}

// errorIf synonymous with fatalIf but doesn't exit on error != nil
func errorIf(err error, msg string, data ...interface{}) {
	if err == nil {
		return
	}
	source := callerSource()
	fields := logrus.Fields{
		"source": source,
		"cause":  err.Error(),
	}
	rlog.WithFields(fields).Errorf(msg, data...)
}

// fatalIf wrapper function which takes error and prints jsonic error messages.
func fatalIf(err error, msg string, data ...interface{}) {
	if err == nil {
		return
	}
	source := callerSource()
	fields := logrus.Fields{
		"source": source,
		"cause":  err.Error(),
	}
	rlog.WithFields(fields).Fatalf(msg, data...)
}

func printf(msg string, data ...interface{}) {
	rlog.Printf(msg, data...)
}
