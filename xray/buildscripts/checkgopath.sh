#!/usr/bin/env bash
# Copyright (c) 2017 Minio, Inc. <https://www.minio.io>
#
# This file is part of Xray.
#
# Xray is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as
# published by the Free Software Foundation, either version 3 of the
# License, or (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program. If not, see <http://www.gnu.org/licenses/>.

main() {
    echo "Checking project is in GOPATH:"

    IFS=':' read -r -a paths <<< "$GOPATH"
    for path in "${paths[@]}"; do
        minio_path="$path/src/github.com/minio/xray"
        if [ -d "$minio_path" ]; then
            if [ "$minio_path" -ef "$PWD" ]; then
               exit 0
            fi
        fi
    done

    echo "ERROR"
    echo "Project not found in ${GOPATH}."
    echo "Follow instructions at https://github.com/minio/xray/blob/master/CONTRIBUTING.md#setup-your-minio-github-repository"
    exit 1
}

main
