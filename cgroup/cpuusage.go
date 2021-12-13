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

type TotalCPUUsage interface {
	SystemSeconds() float64
	UserSeconds() float64
}

func NewCPUUsage(controlGroupMode ControlGroupMode, mountPointPrefix string, cgSubpath string) (TotalCPUUsage, error) {
	fs, err := NewDefaultFS(controlGroupMode, mountPointPrefix)
	if err != nil {
		return nil, err
	}

	if fs.cgroupUnified == MountModeUnified || fs.cgroupUnified == MountModeHybrid {
		ret, err2 := fs.NewCPUStat(cgSubpath)
		if ret == nil {
			return nil, err2
		} else {
			return ret, nil
		}
	} else {
		ret, err2 := fs.NewCPUAcct(cgSubpath)
		if ret == nil {
			return nil, err2
		} else {
			return ret, nil
		}
	}
}
