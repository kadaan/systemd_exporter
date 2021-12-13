// Copyright © 2021 Joel Baranick <jbaranick@gmail.com>
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
// 	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cgroup

import (
	"bufio"
	"bytes"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

type CPUStat struct {
	SystemMicrosec uint64
	UserMicrosec   uint64
}

func (c CPUStat) UserSeconds() float64 {
	return float64(c.UserMicrosec) / float64(time.Second.Microseconds())
}

func (c CPUStat) SystemSeconds() float64 {
	return float64(c.SystemMicrosec) / float64(time.Second.Microseconds())
}

// NewCPUUsage will locate and read the kernel's cpu accounting info for
// the provided systemd cgroup subpath.
func (fs FS) NewCPUStat(cgSubpath string) (*CPUStat, error) {
	cgPath, err := fs.cgGetPath("cpu", cgSubpath, "cpu.stat")
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get cpu controller path")
	}

	// Example cpu.stat
	// usage_usec 291912970
	// user_usec 238552676
	// system_usec 53360293
	b, err := ReadFileNoStat(cgPath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read file %s", cgPath)
	}

	scanner := bufio.NewScanner(bytes.NewReader(b))
	if ok := scanner.Scan(); !ok {
		return nil, errors.Errorf("unable to scan file %s", cgPath)
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrapf(err, "unable to scan file %s", cgPath)
	}
	var user, sys uint64
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, errors.Wrapf(err, "unable to scan file %s", cgPath)
		}
		text := scanner.Text()
		vals := strings.Split(text, " ")
		if len(vals) != 2 {
			return nil, errors.Errorf("unable to parse contents of file %s", cgPath)
		}
		val, err := strconv.ParseUint(vals[1], 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse %s as uint64 (from %s)", vals[1], cgPath)
		}
		if vals[0] == "user_usec" {
			user = val
		}
		if vals[0] == "system_usec" {
			sys = val
		}
	}
	if user == 0 && sys == 0 {
		return nil, nil
	}
	cpuStat := CPUStat{
		UserMicrosec:   user,
		SystemMicrosec: sys,
	}

	return &cpuStat, nil
}
