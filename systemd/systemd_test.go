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

package systemd

import (
	"github.com/coreos/go-systemd/dbus"
	"testing"
)

func TestParseUnitType(t *testing.T) {
	x := dbus.UnitStatus{
		Name:        "test.service",
		Description: "",
		LoadState:   "",
		ActiveState: "",
		SubState:    "",
		Followed:    "",
		Path:        "",
		JobId:       0,
		JobType:     "",
		JobPath:     "",
	}
	found := parseUnitType(x)
	if found != "service" {
		t.Errorf("Bad unit name parsing. Wanted %s got %s", "service", found)
	}

}
