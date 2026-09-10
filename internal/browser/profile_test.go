// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
)

func TestDefaultLaunchProfile_LinuxXDG(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux path rules")
	}

	t.Setenv("HOME", "/home/testuser")
	t.Setenv("XDG_STATE_HOME", "/custom/state")

	got, err := DefaultLaunchProfile()
	if err != nil {
		t.Fatal(err)
	}
	want := "/custom/state/snag/chrome"
	if got != want {
		t.Errorf("DefaultLaunchProfile() = %q, want %q", got, want)
	}
}

func TestDefaultLaunchProfile_LinuxEmptyAndRelativeXDG(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux path rules")
	}

	home := "/home/testuser"
	t.Setenv("HOME", home)
	want := filepath.Join(home, ".local", "state", "snag", "chrome")

	t.Setenv("XDG_STATE_HOME", "")
	got, err := DefaultLaunchProfile()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("empty XDG_STATE_HOME: got %q, want %q", got, want)
	}

	t.Setenv("XDG_STATE_HOME", "relative/state")
	got, err = DefaultLaunchProfile()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("relative XDG_STATE_HOME: got %q, want %q", got, want)
	}
}

func TestDefaultLaunchProfile_DarwinIgnoresXDG(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin path rules")
	}

	t.Setenv("HOME", "/Users/testuser")
	t.Setenv("XDG_STATE_HOME", "/custom/state")

	got, err := DefaultLaunchProfile()
	if err != nil {
		t.Fatal(err)
	}
	want := "/Users/testuser/Library/Application Support/snag/chrome"
	if got != want {
		t.Errorf("DefaultLaunchProfile() = %q, want %q", got, want)
	}
}

func TestDefaultLaunchProfile_DoesNotCreate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, "state"))

	path, err := DefaultLaunchProfile()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("DefaultLaunchProfile must not create %s: %v", path, err)
	}
}

func TestResolveLaunchProfile_Temp(t *testing.T) {
	got, err := ResolveLaunchProfile("/some/path", true)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Ephemeral {
		t.Error("temp profile should be ephemeral")
	}
	if got.Path != "" {
		t.Errorf("ephemeral path = %q, want empty", got.Path)
	}
	if got.Exists {
		t.Error("ephemeral profile should not report exists")
	}
}

func TestResolveLaunchProfile_OverrideStatOnly(t *testing.T) {
	dir := t.TempDir()
	got, err := ResolveLaunchProfile(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != dir {
		t.Errorf("path = %q, want %q", got.Path, dir)
	}
	if !got.Exists {
		t.Error("existing directory should exist")
	}

	file := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err = ResolveLaunchProfile(file, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Exists {
		t.Error("a file is not a launch profile directory")
	}

	missing := filepath.Join(dir, "missing")
	got, err = ResolveLaunchProfile(missing, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Exists {
		t.Error("missing path should not exist")
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("ResolveLaunchProfile must not mkdir %s", missing)
	}
}

func TestResolveLaunchProfile_DefaultDoesNotCreate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, "state"))

	got, err := ResolveLaunchProfile("", false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Ephemeral {
		t.Error("default profile is not ephemeral")
	}
	if got.Exists {
		t.Error("uncreated default profile should not exist")
	}
	if _, err := os.Stat(got.Path); !os.IsNotExist(err) {
		t.Fatalf("ResolveLaunchProfile must not create %s", got.Path)
	}
}

func TestEnsureUserDataDir_CreatesOwnerOnly(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "snag", "chrome")
	if err := EnsureUserDataDir(dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := info.Mode().Perm()
	if got != 0o700 {
		t.Errorf("profile dir mode = %04o, want 0700", got)
	}
}

func TestApplyLaunchProfile_TempLeavesRodDir(t *testing.T) {
	l := launcher.New()
	rodDir := l.Get(flags.UserDataDir)
	if rodDir == "" {
		t.Fatal("rod default user-data-dir should be set")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, "state"))

	bm := &BrowserManager{tempProfile: true}
	got, err := bm.applyLaunchProfile(l)
	if err != nil {
		t.Fatal(err)
	}
	if dir := got.Get(flags.UserDataDir); dir != rodDir {
		t.Errorf("temp profile user-data-dir = %q, want rod default %q", dir, rodDir)
	}

	def, err := DefaultLaunchProfile()
	if err != nil {
		t.Fatal(err)
	}
	if rodDir == def {
		t.Fatal("rod default dir must not be the persistent snag profile")
	}
	if _, err := os.Stat(def); !os.IsNotExist(err) {
		t.Fatalf("--temp-profile must not create %s", def)
	}
}

func TestApplyLaunchProfile_DefaultCreatesUnderHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, "state"))

	want, err := DefaultLaunchProfile()
	if err != nil {
		t.Fatal(err)
	}

	bm := &BrowserManager{}
	l, err := bm.applyLaunchProfile(launcher.New())
	if err != nil {
		t.Fatal(err)
	}
	if bm.userDataDir != want {
		t.Errorf("manager userDataDir = %q, want %q", bm.userDataDir, want)
	}
	if got := l.Get(flags.UserDataDir); got != want {
		t.Errorf("launcher user-data-dir = %q, want %q", got, want)
	}
	info, err := os.Stat(want)
	if err != nil {
		t.Fatalf("default profile should be created: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("default profile mode = %04o, want 0700", perm)
	}
}

func TestApplyLaunchProfile_CreatesMissing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing-profile")
	bm := &BrowserManager{userDataDir: dir}
	l, err := bm.applyLaunchProfile(launcher.New())
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("created profile mode = %04o, want 0700", perm)
	}
	if got := l.Get(flags.UserDataDir); got != dir {
		t.Errorf("launcher user-data-dir = %q, want %q", got, dir)
	}
}

func TestEnsureUserDataDir_DoesNotChmodExisting(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "existing")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureUserDataDir(dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := info.Mode().Perm()
	if got != 0o755 {
		t.Errorf("existing dir mode = %04o, want 0755 (unchanged)", got)
	}
}

func TestProfileLockPresent(t *testing.T) {
	if profileLockPresent("") {
		t.Fatal("empty dir must not look up SingletonLock in cwd")
	}
	dir := t.TempDir()
	if profileLockPresent(dir) {
		t.Fatal("missing lock should be false")
	}

	if err := os.WriteFile(filepath.Join(dir, chromeSingletonLock), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if !profileLockPresent(dir) {
		t.Fatal("regular SingletonLock file should count")
	}
}

func TestProfileLockPresent_DanglingSymlink(t *testing.T) {
	dir := t.TempDir()
	if err := os.Symlink("missing-target", filepath.Join(dir, chromeSingletonLock)); err != nil {
		t.Fatal(err)
	}
	if !profileLockPresent(dir) {
		t.Fatal("dangling SingletonLock symlink should count")
	}
}

func TestExtraInstanceSuggestion(t *testing.T) {
	got := extraInstanceSuggestion(9222)
	if !strings.Contains(got, "--port 9223") {
		t.Errorf("default port suggestion = %q, want --port 9223", got)
	}
	if !strings.Contains(got, "--temp-profile") || !strings.Contains(got, "--force-headless") {
		t.Errorf("default port suggestion = %q, want isolation flags", got)
	}

	got = extraInstanceSuggestion(0)
	if !strings.Contains(got, "--port 9223") {
		t.Errorf("unset port suggestion = %q, want --port 9223", got)
	}

	got = extraInstanceSuggestion(9223)
	if !strings.Contains(got, "--port 9223") {
		t.Errorf("custom port suggestion = %q, want --port 9223", got)
	}
	if strings.Contains(got, "--port 9224") {
		t.Errorf("custom port suggestion bumped the port: %q", got)
	}

	got = extraInstanceSuggestion(9333)
	if !strings.Contains(got, "--port 9333") {
		t.Errorf("non-default port suggestion = %q, want --port 9333", got)
	}
}

func TestWrapLaunchError_ProfileLock(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, chromeSingletonLock), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	bm := &BrowserManager{userDataDir: dir, port: 9222}
	err := bm.wrapLaunchError(errors.New("rod boom"))
	if !errors.Is(err, ErrProfileInUse) {
		t.Fatalf("got %v, want ErrProfileInUse", err)
	}
	if !strings.Contains(err.Error(), dir) {
		t.Errorf("error %q should name the profile dir", err)
	}
	if !strings.Contains(err.Error(), "rod boom") {
		t.Errorf("error %q should wrap the launch failure", err)
	}
}

func TestWrapLaunchError_NoLock(t *testing.T) {
	bm := &BrowserManager{userDataDir: t.TempDir()}
	err := bm.wrapLaunchError(errors.New("rod boom"))
	if errors.Is(err, ErrProfileInUse) {
		t.Fatal("unlocked profile should not be ErrProfileInUse")
	}
}

func TestWrapLaunchError_TempProfileIgnoresLock(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, chromeSingletonLock), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	bm := &BrowserManager{userDataDir: dir, tempProfile: true}
	err := bm.wrapLaunchError(errors.New("rod boom"))
	if errors.Is(err, ErrProfileInUse) {
		t.Fatal("temp profile should not report in-use")
	}
}
