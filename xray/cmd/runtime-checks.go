/*
 * Copyright (c) 2017 XRay, Inc. <https://www.minio.io>
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

import (
	"fmt"
	"runtime"

	"github.com/hashicorp/go-version"
)

// check if minimum Go version is met.
func checkGoVersion() {
	// Current version.
	curVersion, err := version.NewVersion(runtime.Version()[2:])
	if err != nil {
		rlog.Fatalln("Unable to determine current go version.", err)
	}

	// Prepare version constraint.
	constraints, err := version.NewConstraint(minGoVersion)
	if err != nil {
		rlog.Fatalln("Unable to check go version.")
	}

	// Check for minimum version.
	if !constraints.Check(curVersion) {
		rlog.Fatalln(fmt.Sprintf("Please recompile Xray with Golang version %s.", minGoVersion))
	}
}
