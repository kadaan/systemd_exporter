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

import "testing"

func TestNewCPUAcct(t *testing.T) {
	fs := getLegacyFixtures(t)
	cpu, err := fs.NewCPUAcct("/")
	if err != nil {
		t.Errorf("want NewCPUAcct('/') to succeed: %s", err)
		return
	}

	if len(cpu.CPUs) != 4 {
		t.Errorf("Wrong number of CPUs. Wanted %d got %d", 4, len(cpu.CPUs))
	}

	var expectedUser uint64 = 29531441016368
	if cpu.UsageUserNanosecs() != expectedUser {
		t.Errorf("Wrong user nanoseconds. Wanted %d got %d", expectedUser, cpu.UsageUserNanosecs())
	}

	var expectedSys uint64 = 619186701953
	if cpu.UsageSystemNanosecs() != expectedSys {
		t.Errorf("Wrong sys nanoseconds. Wanted %d got %d", expectedSys, cpu.UsageSystemNanosecs())
	}

	expectedTotal := expectedSys + expectedUser
	if cpu.UsageAllNanosecs() != expectedTotal {
		t.Errorf("Wrong total nanoseconds. Wanted %d got %d", expectedTotal, cpu.UsageAllNanosecs())
	}

	if _, err := fs.NewCPUAcct("foobar"); err == nil {
		t.Errorf("expected error getting cpu accounting info for bogus cgroup")
	}
}
