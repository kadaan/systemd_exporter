package cgroup

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"testing"
)

const (
	testFixturesUnified = "fixtures/cgroup-unified"
	testFixturesLegacy = "fixtures/cgroup-legacy"
)

func TestMountModeParsing(t *testing.T) {
	// This test cannot (easily) use test fixtures, because it relies on being
	// able to call Statfs on mounted filesystems. So we only run inside
	// system where we expect to find cgroupfs mounted in a mode systemd expects.
	// For now, that's only inside TravisCI, but in future we may expand to run
	// this by default on certain Linux systems
	if _, inTravisCI := os.LookupEnv("TRAVIS"); inTravisCI == false {
		return
	}

	if _, err := NewDefaultFS(); err != nil {
		t.Errorf("expected success determining mount type inside of travis CI: %s", err)
	}
}

func TestCgUnifiedCached(t *testing.T) {
	// Build some functions we will use to simulate various cgroup mounting scenarios
	noCgroupMount := func(path string, stat *unix.Statfs_t) error {
		// No fs present on /sys/fs/cgroup/
		return errors.New("boo")
	}
	unknownCgroupMount := func(path string, stat *unix.Statfs_t) error {
		// Unknown fs type present on /sys/fs/cgroup/
		stat.Type = 0x0
		return nil
	}
	unifiedMount := func(path string, stat *unix.Statfs_t) error {
		// unified fs present
		switch path {
		case DefaultMountPoint:
			stat.Type = cgroup2SuperMagic
			return nil
		default:
			return errors.New("pretend path not found")
		}
	}
	hybridMountSystemdV232 := func(path string, stat *unix.Statfs_t) error {
		switch path {
		case DefaultMountPoint:
			stat.Type = tmpFsMagic
		case SystemdMountPoint:
			stat.Type = cgroup2SuperMagic
		}
		return nil
	}
	hybridMountSystemdV233 := func(path string, stat *unix.Statfs_t) error {
		switch path {
		case DefaultMountPoint:
			stat.Type = tmpFsMagic
		case UnifiedMountPoint:
			stat.Type = cgroup2SuperMagic
		case SystemdMountPoint:
			stat.Type = cgroupSuperMagic
		}
		return nil
	}
	legacyMount := func(path string, stat *unix.Statfs_t) error {
		switch path {
		case DefaultMountPoint:
			stat.Type = tmpFsMagic
		case UnifiedMountPoint:
			return errors.New("pretend unified path not found")
		case SystemdMountPoint:
			stat.Type = cgroupSuperMagic
		}
		return nil
	}
	missingSystemdFolder := func(path string, stat *unix.Statfs_t) error {
		switch path {
		case DefaultMountPoint:
			stat.Type = tmpFsMagic
		case UnifiedMountPoint:
			return errors.New("pretend unified path not found")
		case SystemdMountPoint:
			return errors.New("pretend we cannot stat systemd dir")
		}
		return nil
	}
	unknownSystemdFolderMountType := func(path string, stat *unix.Statfs_t) error {
		switch path {
		case DefaultMountPoint:
			stat.Type = tmpFsMagic
		case UnifiedMountPoint:
			return errors.New("pretend unified path not found")
		case SystemdMountPoint:
			stat.Type = 0x0
		}
		return nil
	}

	tables := []struct {
		name         string
		statFn       func(string, *unix.Statfs_t) error
		expectedMode MountMode
		errExpected  bool
	}{
		{"NoCgroupMount", noCgroupMount, MountModeUnknown, true},
		{"UnknownCgroupMountType", unknownCgroupMount, MountModeUnknown, true},
		{"LegacyMount", legacyMount, MountModeLegacy, false},
		{"HybridMount, v232", hybridMountSystemdV232, MountModeHybrid, false},
		{"HybridMount, v233+", hybridMountSystemdV233, MountModeHybrid, false},
		{"MissingSystemdFolder", missingSystemdFolder, MountModeUnknown, true},
		{"UnknownSystemdFolderType", unknownSystemdFolderMountType, MountModeUnknown, true},
		{"UnifiedMount", unifiedMount, MountModeUnified, false},
	}

	for _, table := range tables {
		statfsFunc = table.statFn
		mode, _, _, err := cgUnifiedCached()
		if table.errExpected && err == nil {
			t.Errorf("%s: expected an err, but got mode %s with no error", table.name, mode)
		}
		if !table.errExpected && err != nil {
			t.Errorf("%s: expected no error, but got mode %s with err: %s", table.name, mode, err)
		}
		if mode != table.expectedMode {
			t.Errorf("%s: expected mode %s but got mode %s", table.name, table.expectedMode, mode)
		}
	}
}

func TestNewFS(t *testing.T) {
	if _, err := NewFS(MountModeUnknown, "foobar", ""); err == nil {
		t.Error("NewFS should have failed with non-existing path")
	}

	if _, err := NewFS(MountModeUnknown, "", "cgroups_test.go"); err == nil {
		t.Error("want NewFS to fail if mount point is not a dir")
	}

	if _, err := NewFS(MountModeUnknown, testFixturesUnified, testFixturesLegacy); err != nil {
		t.Error("want NewFS to succeed if mount point exists")
	}
}

func getUnifiedFixtures(t *testing.T) FS {
	fs, err := NewFS(MountModeUnified, testFixturesUnified, "")
	if err != nil {
		t.Fatal("Unable to create unified test fixtures")
	}
	return fs
}

func getHybridFixtures(t *testing.T) FS {
	fs, err := NewFS(MountModeHybrid, testFixturesUnified, testFixturesLegacy)
	if err != nil {
		t.Fatal("Unable to create hybrid test fixtures")
	}
	return fs
}

func getLegacyFixtures(t *testing.T) FS {
	fs, err := NewFS(MountModeLegacy, "", testFixturesLegacy)
	if err != nil {
		t.Fatal("Unable to create legacy test fixtures")
	}
	return fs
}

func TestCgSubpathCPU(t *testing.T) {
	controller := "cpu"
	subpath := "/system.slice"
	suffix := "cpu.stat"

	fs := getHybridFixtures(t)

	fs.cgroupUnified = MountModeUnknown
	if _, err := fs.cgGetPath(controller, subpath, suffix); err == nil {
		t.Errorf("should not be able to determine path with unknown mount mode: %s", err)
	}

	verifyControllerPath(t, MountModeLegacy, controller, subpath, suffix, testFixturesLegacy+"/cpu/system.slice/cpu.stat")
	verifyControllerPath(t, MountModeHybrid, controller, subpath, suffix, testFixturesUnified+"/system.slice/cpu.stat")
	verifyControllerPath(t, MountModeUnified, controller, subpath, suffix, testFixturesUnified+"/system.slice/cpu.stat")
}

func TestCgSubpathMemory(t *testing.T) {
	controller := "mem"
	subpath := "/system.slice"
	suffix := "memory.stat"

	fs := getHybridFixtures(t)

	fs.cgroupUnified = MountModeUnknown
	if _, err := fs.cgGetPath(controller, subpath, suffix); err == nil {
		t.Errorf("should not be able to determine path with unknown mount mode: %s", err)
	}

	verifyControllerPath(t, MountModeLegacy, controller, subpath, suffix, testFixturesLegacy+"/mem/system.slice/memory.stat")
	verifyControllerPath(t, MountModeHybrid, controller, subpath, suffix, testFixturesLegacy+"/mem/system.slice/memory.stat")
	verifyControllerPath(t, MountModeUnified, controller, subpath, suffix, testFixturesUnified+"/system.slice/memory.stat")
}

func verifyControllerPath(t *testing.T, mountMode MountMode, controller string, subpath string, suffix string, expected string) {
	fs := getHybridFixtures(t)
	fs.cgroupUnified = mountMode

	path, err := fs.cgGetPath(controller, subpath, suffix)
	if err != nil {
		t.Errorf("should be able to determine path with %s mount mode: %s", mountMode, err)
	}

	if path != expected {
		t.Errorf("bad response for %s. Wanted %s, got %s", mountMode, expected, path)
	}
}
