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
	"fmt"
	"strings"
)

type ControlGroupMode int8

const (
	_ ControlGroupMode = iota
	Auto
	Legacy
	Hybrid
	UnifiedV232
	Unified
)

const _ControlGroupModeName = "AutoLegacyHybridUnifiedV232Unified"

var _ControlGroupModeIndex = [...]uint8{0, 4, 10, 16, 27, 34}

const _ControlGroupModeLowerName = "autolegacyhybridunifiedv232unified"

func (i ControlGroupMode) String() string {
	i -= 1
	if i < 0 || i >= ControlGroupMode(len(_ControlGroupModeIndex)-1) {
		return fmt.Sprintf("ControlGroupMode(%d)", i+1)
	}
	return _ControlGroupModeName[_ControlGroupModeIndex[i]:_ControlGroupModeIndex[i+1]]
}

var _ControlGroupModeValues = []ControlGroupMode{Auto, Legacy, Hybrid, Unified}

var _ControlGroupModeNameToValueMap = map[string]ControlGroupMode{
	_ControlGroupModeName[0:4]:        Auto,
	_ControlGroupModeLowerName[0:4]:   Auto,
	_ControlGroupModeName[4:10]:       Legacy,
	_ControlGroupModeLowerName[4:10]:  Legacy,
	_ControlGroupModeName[10:16]:      Hybrid,
	_ControlGroupModeLowerName[10:16]: Hybrid,
	_ControlGroupModeName[16:27]:      UnifiedV232,
	_ControlGroupModeLowerName[16:27]: UnifiedV232,
	_ControlGroupModeName[27:34]:      Unified,
	_ControlGroupModeLowerName[27:34]: Unified,
}

var _ControlGroupModeNames = []string{
	_ControlGroupModeName[0:4],
	_ControlGroupModeName[4:10],
	_ControlGroupModeName[10:16],
	_ControlGroupModeName[16:23],
}

// ControlGroupModeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func ControlGroupModeString(s string) (ControlGroupMode, error) {
	if val, ok := _ControlGroupModeNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _ControlGroupModeNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to ControlGroupMode values", s)
}

// ControlGroupModeValues returns all values of the enum
func ControlGroupModeValues() []ControlGroupMode {
	return _ControlGroupModeValues
}

// ControlGroupModeStrings returns a slice of all String values of the enum
func ControlGroupModeStrings() []string {
	strs := make([]string, len(_ControlGroupModeNames))
	copy(strs, _ControlGroupModeNames)
	return strs
}

// IsAControlGroupMode returns "true" if the value is listed in the enum definition. "false" otherwise
func (i ControlGroupMode) IsAControlGroupMode() bool {
	for _, v := range _ControlGroupModeValues {
		if i == v {
			return true
		}
	}
	return false
}
