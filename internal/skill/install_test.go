// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package skill_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/p3bot/snag/internal/skill"
)

func TestWriteInstallAndPresent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	written, err := skill.WriteInstall(root)
	if err != nil {
		t.Fatal(err)
	}
	want := skill.FilePath(root)
	if written != want {
		t.Fatalf("written = %q want %q", written, want)
	}
	if !skill.Present(root) {
		t.Fatal("Present = false after install")
	}
	data, err := os.ReadFile(written)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != skill.Text() {
		t.Fatalf("content mismatch: got %d bytes want %d", len(data), len(skill.Text()))
	}
	name, err := skill.FrontmatterName(data)
	if err != nil {
		t.Fatal(err)
	}
	if name != skill.ID {
		t.Fatalf("name = %q want %q", name, skill.ID)
	}
}

func TestWriteInstallRootIsFile(t *testing.T) {
	dir := t.TempDir()
	fileRoot := filepath.Join(dir, "notadir")
	if err := os.WriteFile(fileRoot, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := skill.WriteInstall(fileRoot); err == nil {
		t.Fatal("expected error when skills root is a file")
	}
}

func TestWriteInstallFailedWriteLeavesExisting(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can write into 0555 directories")
	}
	root := filepath.Join(t.TempDir(), "skills")
	written, err := skill.WriteInstall(root)
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(written)
	if err != nil {
		t.Fatal(err)
	}
	dir := skill.DirPath(root)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	if _, err := skill.WriteInstall(root); err == nil {
		t.Fatal("expected write failure against read-only skill dir")
	}
	got, err := os.ReadFile(written)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatal("failed write truncated existing SKILL.md")
	}
}

func TestRemoveOwned(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	if _, err := skill.WriteInstall(root); err != nil {
		t.Fatal(err)
	}
	res, err := skill.RemoveOwned(root)
	if err != nil {
		t.Fatal(err)
	}
	if res != skill.UninstallRemoved {
		t.Fatalf("result = %v want removed", res)
	}
	if skill.Present(root) {
		t.Fatal("still present after remove")
	}
	res, err = skill.RemoveOwned(root)
	if err != nil {
		t.Fatal(err)
	}
	if res != skill.UninstallAbsent {
		t.Fatalf("result = %v want absent", res)
	}
}

func TestRemoveOwnedKeepsExtraAndWrongName(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	if _, err := skill.WriteInstall(root); err != nil {
		t.Fatal(err)
	}
	path := skill.FilePath(root)
	if err := os.WriteFile(path, []byte("---\nname: snag\n---\n\n# edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := skill.RemoveOwned(root)
	if err != nil || res != skill.UninstallRemoved {
		t.Fatalf("edited body: res=%v err=%v", res, err)
	}

	if _, err := skill.WriteInstall(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill.DirPath(root), "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err = skill.RemoveOwned(root)
	if err != nil {
		t.Fatal(err)
	}
	if res != skill.UninstallKeptExtra {
		t.Fatalf("extra file: res=%v want kept extra", res)
	}
	if !skill.Present(root) {
		t.Fatal("extra-file dir should remain")
	}

	root2 := filepath.Join(t.TempDir(), "skills2")
	if err := os.MkdirAll(skill.DirPath(root2), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill.FilePath(root2), []byte("---\nname: other\n---\n\n# x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err = skill.RemoveOwned(root2)
	if err != nil {
		t.Fatal(err)
	}
	if res != skill.UninstallKeptNotOurs {
		t.Fatalf("wrong name: res=%v want not ours", res)
	}
	if !skill.Present(root2) {
		t.Fatal("wrong-name dir should remain")
	}

	root3 := filepath.Join(t.TempDir(), "skills3")
	if err := os.MkdirAll(skill.DirPath(root3), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill.FilePath(root3), []byte("---\nname: [broken\n---\n\n# x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err = skill.RemoveOwned(root3)
	if err != nil {
		t.Fatal(err)
	}
	if res != skill.UninstallKeptUnreadable {
		t.Fatalf("broken YAML: res=%v want unreadable", res)
	}
	if skill.ReasonDetail(res) != "skill frontmatter unreadable" {
		t.Fatalf("reason = %q", skill.ReasonDetail(res))
	}
	if !skill.Present(root3) {
		t.Fatal("unreadable dir should remain")
	}

	root4 := filepath.Join(t.TempDir(), "skills4")
	if err := os.MkdirAll(skill.DirPath(root4), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill.FilePath(root4), []byte("---\nthis is not yaml\n---\n\n# x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err = skill.RemoveOwned(root4)
	if err != nil {
		t.Fatal(err)
	}
	if res != skill.UninstallKeptUnreadable {
		t.Fatalf("non-YAML fence: res=%v want unreadable", res)
	}
	if skill.ReasonDetail(res) != "skill frontmatter unreadable" {
		t.Fatalf("reason = %q", skill.ReasonDetail(res))
	}
	if !skill.Present(root4) {
		t.Fatal("non-YAML dir should remain")
	}
}

func TestDedupeAndJoin(t *testing.T) {
	got := skill.DedupePaths([]string{"/b", "", "/a", "/b", "/a/"})
	if len(got) != 2 {
		t.Fatalf("dedupe = %v", got)
	}
	if skill.JoinAgents([]string{"z", "a", "m"}) != "a,m,z" {
		t.Fatalf("JoinAgents = %q", skill.JoinAgents([]string{"z", "a", "m"}))
	}
	if !strings.HasPrefix(skill.ReasonDetail(skill.UninstallKeptNotOurs), "frontmatter") {
		t.Fatal("reason detail missing")
	}
	ids := skill.DedupeIDs([]string{"b", "a", "b", ""})
	if len(ids) != 2 || ids[0] != "b" || ids[1] != "a" {
		t.Fatalf("DedupeIDs = %v", ids)
	}
}
