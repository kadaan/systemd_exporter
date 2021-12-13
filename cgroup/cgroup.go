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
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

// FS is the pseudo-filesystem cgroupfs, which provides an interface to
// kernel data structures
type FS struct {
	// WARNING: We only read this data once at process start, systemd updates
	// may require restarting systemd-exporter
	cgroupUnified MountMode

	unifiedPath string
	legacyPath  string
}

// DefaultMountPoint is the common mount point of the cgroupfs filesystem
const DefaultMountPoint = "/sys/fs/cgroup"

const UnifiedMountPoint = "/sys/fs/cgroup/unified"

const SystemdMountPoint = "/sys/fs/cgroup/systemd"

// NewDefaultFS returns a new cgroup FS mounted under the default mountPoint.
// It will error if cgroup hierarchies are not laid out in a manner understood
// by systemd.
func NewDefaultFS(controlGroupMode ControlGroupMode, mountPointPrefix string) (FS, error) {
	mode, unifiedPath, legacyPath, err := cgUnifiedCached(controlGroupMode, mountPointPrefix)
	if err != nil || mode == MountModeUnknown {
		return FS{}, fmt.Errorf("could not determine cgroupfs mount mode: %s", err)
	}

	return newFS(mode, unifiedPath, legacyPath)
}

// NewFS returns a new cgroup FS mounted under the given mountPoint. It does not check
// the provided mount mode
func newFS(mountMode MountMode, unifiedPath string, legacyPath string) (FS, error) {
	if unifiedPath != "" {
		if err := checkMountPath(unifiedPath); err != nil {
			return FS{}, err
		}
	}

	if legacyPath != "" {
		if err := checkMountPath(legacyPath); err != nil {
			return FS{}, err
		}
	}

	return FS{mountMode, unifiedPath, legacyPath}, nil
}

func checkMountPath(mountPoint string) error {
	info, err := os.Stat(mountPoint)
	if err != nil {
		return fmt.Errorf("could not read %s: %s", mountPoint, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("mount point %s is not a directory", mountPoint)
	}
	return nil
}

// MountMode constants describe how the kernel has mounted various cgroup filesystems under /sys/fs/cgroup.
// Generally speaking, kernels using the cgroups-v1 API will have many cgroup controller hierarchies, each with
// their own fs and their own mount point. Kernels using cgroups-v2 API will only have the one unified hierarchy.
// To support back compatibility, kernels often mount both the v1 and v2 hierarchies at different points. Systemd
// has to know where the hierarchies are, so it inspects the mounts under /sys/fs/cgroup and decides what
// MountMode this kernel is using. See each constant for a description of that mode. This type corresponds to
// the unified_cache variable in systemd/src/basic/cgroup-util.c
type MountMode int8

const (
	// MountModeUnknown indicates we do not recognize the mount pattern of the cgroup filesystems in /sys/fs/cgroup.
	// systemd source calls this mode CGROUP_UNIFIED_UNKNOWN
	MountModeUnknown MountMode = iota
	// MountModeLegacy indicates both systemd and individual cgroups are using cgroup-v1 hierarchies. There is
	// typically one mount point per hierarchy, and no usage of the cgroup-v2 unified hierarchy.
	// systemd source calls this mode CGROUP_UNIFIED_NONE
	MountModeLegacy MountMode = iota
	// MountModeHybrid indicates the systemd controller is using cgroup-v2 unified hierarchy for organizing
	// processes, but all other cgroups are using cgroup-v1 legacy hierarchies.
	// systemd source calls this CGROUP_UNIFIED_SYSTEMD and also stores the unified_systemd_v232 flag
	MountModeHybrid MountMode = iota
	// MountModeUnified indicates cgroup-v2 API is in full usage and there are no cgroup-v1 hierarchies mounted.
	// Non-updated programs (e.g. container orchestrators such as docker/runc) that rely on cgroup-v1 mounts will break.
	// systemd source calls this CGROUP_UNIFIED_ALL
	MountModeUnified MountMode = iota
)

func (c MountMode) String() string {
	return [...]string{"unknown", "none", "systemd", "all"}[c]
}

// Values copied from https://github.com/torvalds/linux/blob/master/include/uapi/linux/magic.h
const (
	tmpFsMagic        = 0x01021994
	cgroupSuperMagic  = 0x27e0eb
	cgroup2SuperMagic = 0x63677270
)

// cgUnifiedCached checks the filesystem types mounted under /sys/fs/cgroup to determine
// which systemd layout (legacy/hybrid/unified) is in use.
// We do not bother to track unified_systemd_v232 as our usage does not
// depend on reading the systemd hierarchy directly, we only focus on reading
// the controllers. If you care if /sys/fs/cgroup/systemd is v1 or v2 you need
// to track this
// WARNING: We cache this data once at process start. Systemd updates
// may require restarting systemd-exporter
// Equivalent to systemd cgroup-util.c#cg_unified_cached
var statfsFunc = unix.Statfs

func cgUnifiedCached(mode ControlGroupMode, mountPrintPrefix string) (MountMode, string, string, error) {
	// if cgroupUnified != MountModeUnknown {
	// 	return cgroupUnified, nil
	// }
	defaultMountPoint := prefixMountPoint(mountPrintPrefix, DefaultMountPoint)
	unifiedMountPoint := prefixMountPoint(mountPrintPrefix, UnifiedMountPoint)
	systemdMountPoint := prefixMountPoint(mountPrintPrefix, SystemdMountPoint)

	switch mode {
	case Auto:
		break
	case Legacy:
		log.Debugf("Explicitly enabled cgroup on %s, legacy hierarchy", defaultMountPoint)
		return MountModeLegacy, "", defaultMountPoint, nil
	case UnifiedV232:
		log.Debugf("Explicitly enabled cgroup2 on %s, unified hierarchy for systemd controller (v232 variant)", systemdMountPoint)
		return MountModeHybrid, unifiedMountPoint, systemdMountPoint, nil
	case Unified:
		log.Debugf("Explicitly enabled cgroup2 on %s, unified hierarchy for systemd controller", unifiedMountPoint)
		return MountModeHybrid, defaultMountPoint, "", nil
	default:
		return MountModeUnknown, "", "", errors.Errorf("unknown control group mode %s", mode)
	}

	var fs unix.Statfs_t
	if mode == Auto {
		err := statfsFunc(defaultMountPoint, &fs)
		if err != nil {
			return MountModeUnknown, "", "", errors.Wrapf(err, "failed statfs(%s)", defaultMountPoint)
		}
	}

	if fs.Type == cgroup2SuperMagic || mode == Unified {
		if mode == Unified {
			log.Debugf("Found cgroup2 on %s, full unified hierarchy", defaultMountPoint)
		}
		return MountModeUnified, defaultMountPoint, "", nil
	}

	err := statfsFunc(defaultMountPoint, &fs)
	if err != nil {
		return MountModeUnknown, "", "", errors.Wrapf(err, "failed statfs(%s)", defaultMountPoint)
	}

	switch fs.Type {
	case cgroup2SuperMagic:
		log.Debugf("Found cgroup2 on %s, full unified hierarchy", defaultMountPoint)
		return MountModeUnified, defaultMountPoint, "", nil
	case tmpFsMagic:
		err := statfsFunc(unifiedMountPoint, &fs)

		// Ignore err, we expect path to be missing on v232
		if err == nil && fs.Type == cgroup2SuperMagic {
			log.Debugf("Found cgroup2 on %s, unified hierarchy for systemd controller", unifiedMountPoint)
			return MountModeHybrid, unifiedMountPoint, defaultMountPoint, nil
		}

		err = statfsFunc(systemdMountPoint, &fs)
		if err != nil {
			return MountModeUnknown, "", "", errors.Wrapf(err, "failed statfs(%s)", systemdMountPoint)
		}

		switch fs.Type {
		case cgroup2SuperMagic:
			log.Debugf("Found cgroup2 on %s, unified hierarchy for systemd controller (v232 variant)", systemdMountPoint)
			return MountModeHybrid, unifiedMountPoint, systemdMountPoint, nil
		case cgroupSuperMagic:
			log.Debugf("Found cgroup on %s, legacy hierarchy", defaultMountPoint)
			return MountModeLegacy, "", defaultMountPoint, nil
		default:
			return MountModeUnknown, "", "", errors.Errorf("unknown magic number %x for fstype returned by statfs(%s)", fs.Type, systemdMountPoint)
		}

	default:
		return MountModeUnknown, "", "", errors.Errorf("unknown magic number %x for fstype returned by statfs(%s)", fs.Type, defaultMountPoint)
	}
}

func prefixMountPoint(prefix string, mountPoint string) string {
	if len(prefix) > 0 {
		return filepath.Join(prefix, mountPoint)
	}
	return mountPoint
}

// cgGetPath returns the absolute path for a specific file in a specific controller
// in the specific cgroup denoted by the passed subpath.
// Input examples: ("cpu", "/system.slice", "cpuacct.usage_all")
func (fs FS) cgGetPath(controller string, subpath string, suffix string) (string, error) {
	// relevant systemd source code in cgroup-util.[h|c] specifically cg_get_path
	//  2. Joins controller name with base path

	if fs.cgroupUnified == MountModeUnknown {
		return "", errors.Errorf("Cannot determine path with unknown mounting hierarchy")
	}

	// TODO Ensure controller name is valid
	// TODO Convert controller name into guaranteed valid directory name
	dn := controller

	joined := ""
	switch fs.cgroupUnified {
	case MountModeLegacy:
		joined = filepath.Join(fs.legacyPath, dn, subpath, suffix)
	case MountModeHybrid:
		// cpu.stat exists in the unified tree
		if controller == "cpu" {
			joined = filepath.Join(fs.unifiedPath, subpath, suffix)
		} else {
			joined = filepath.Join(fs.legacyPath, dn, subpath, suffix)
		}
	case MountModeUnified:
		joined = filepath.Join(fs.unifiedPath, subpath, suffix)
	default:
		return "", errors.Errorf("unknown cgroup mount mode (e.g. unified mode) %d", fs.cgroupUnified)
	}
	return joined, nil
}
