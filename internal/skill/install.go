// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package skill

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

// On-disk skill identity (directory segment and frontmatter name).
const (
	ID       = "snag"
	FileName = "SKILL.md"
)

// DirPath returns the skill directory under a skills root: <root>/snag.
func DirPath(skillsRoot string) string {
	return filepath.Join(skillsRoot, ID)
}

// FilePath returns the skill file path: <root>/snag/SKILL.md.
func FilePath(skillsRoot string) string {
	return filepath.Join(skillsRoot, ID, FileName)
}

// WriteInstall writes Text() to <skillsRoot>/snag/SKILL.md, creating
// directories as needed and replacing an existing file atomically.
// Hard-fails when skillsRoot exists and is not a directory.
func WriteInstall(skillsRoot string) (written string, err error) {
	if err := ensureSkillsRootDir(skillsRoot); err != nil {
		return "", err
	}
	dest := FilePath(skillsRoot)
	if err := writeAtomic(dest, []byte(Text()), 0o644); err != nil {
		return "", err
	}
	return dest, nil
}

func ensureSkillsRootDir(skillsRoot string) error {
	info, err := os.Stat(skillsRoot)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("skills root %s exists and is not a directory", skillsRoot)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("stat skills root %s: %w", skillsRoot, err)
	}
	return nil
}

func writeAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp for %s: %w", path, err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp for %s: %w", path, err)
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp for %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp for %s: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("install %s: %w", path, err)
	}
	return nil
}

// FrontmatterName returns the frontmatter name: value from a skill file, or ""
// when the fence or key is absent.
func FrontmatterName(data []byte) (string, error) {
	interior, present := splitFrontmatter(data)
	if !present {
		return "", nil
	}
	name, err := yamlMappingName(interior)
	if err != nil {
		return "", fmt.Errorf("parse skill frontmatter: %w", err)
	}
	return name, nil
}

func splitFrontmatter(data []byte) (interior []byte, present bool) {
	if !bytes.HasPrefix(data, []byte("---")) {
		return nil, false
	}
	rest := data[3:]
	switch {
	case bytes.HasPrefix(rest, []byte("\r\n")):
		rest = rest[2:]
	case bytes.HasPrefix(rest, []byte("\n")):
		rest = rest[1:]
	default:
		return nil, false
	}
	idx := bytes.Index(rest, []byte("\n---"))
	if idx < 0 {
		return nil, false
	}
	return rest[:idx], true
}

func yamlMappingName(interior []byte) (string, error) {
	var m map[string]any
	if err := yaml.Unmarshal(interior, &m); err != nil {
		return "", err
	}
	v, ok := m["name"]
	if !ok || v == nil {
		return "", nil
	}
	switch t := v.(type) {
	case string:
		return t, nil
	default:
		return fmt.Sprint(t), nil
	}
}

// UninstallResult is the outcome of attempting to remove one skill directory.
type UninstallResult int

const (
	// UninstallRemoved means the owned pure snag/ directory was deleted.
	UninstallRemoved UninstallResult = iota
	// UninstallAbsent means no skill directory was present at the path.
	UninstallAbsent
	// UninstallKeptExtra means the directory has entries other than only SKILL.md.
	UninstallKeptExtra
	// UninstallKeptNotOurs means SKILL.md frontmatter name is not snag.
	UninstallKeptNotOurs
	// UninstallKeptUnreadable means SKILL.md frontmatter could not be parsed.
	UninstallKeptUnreadable
)

// String returns a short label for reporting.
func (r UninstallResult) String() string {
	switch r {
	case UninstallRemoved:
		return "removed"
	case UninstallAbsent:
		return "absent"
	case UninstallKeptExtra, UninstallKeptNotOurs, UninstallKeptUnreadable:
		return "kept"
	default:
		return fmt.Sprintf("UninstallResult(%d)", int(r))
	}
}

// RemoveOwned applies the uninstall purity rules to D = skillsRoot/snag.
func RemoveOwned(skillsRoot string) (UninstallResult, error) {
	dir := DirPath(skillsRoot)
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return UninstallAbsent, nil
		}
		return 0, fmt.Errorf("stat %s: %w", dir, err)
	}
	if !info.IsDir() {
		return 0, fmt.Errorf("skill path %s exists and is not a directory", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", dir, err)
	}
	if len(entries) != 1 || entries[0].Name() != FileName || entries[0].IsDir() {
		return UninstallKeptExtra, nil
	}

	data, err := os.ReadFile(filepath.Join(dir, FileName))
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", filepath.Join(dir, FileName), err)
	}
	name, err := FrontmatterName(data)
	if err != nil {
		return UninstallKeptUnreadable, nil
	}
	if name != ID {
		return UninstallKeptNotOurs, nil
	}

	if err := os.RemoveAll(dir); err != nil {
		return 0, fmt.Errorf("remove %s: %w", dir, err)
	}
	return UninstallRemoved, nil
}

// Present reports whether <skillsRoot>/snag/SKILL.md exists as a regular file.
func Present(skillsRoot string) bool {
	info, err := os.Stat(FilePath(skillsRoot))
	return err == nil && info.Mode().IsRegular()
}

// ReasonDetail elaborates kept uninstall outcomes for stdout reporting.
func ReasonDetail(r UninstallResult) string {
	switch r {
	case UninstallKeptExtra:
		return "directory is not pure (expected only SKILL.md)"
	case UninstallKeptNotOurs:
		return "frontmatter name is not " + ID
	case UninstallKeptUnreadable:
		return "skill frontmatter unreadable"
	default:
		return ""
	}
}

// CleanAbs returns filepath.Clean(p) for non-empty paths (no symlink resolve).
func CleanAbs(p string) string {
	if p == "" {
		return ""
	}
	return filepath.Clean(p)
}

// DedupePaths returns unique non-empty cleaned paths in first-seen order.
func DedupePaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		p = CleanAbs(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

// SortPaths sorts paths lexicographically in place and returns the slice.
func SortPaths(paths []string) []string {
	sort.Strings(paths)
	return paths
}

// JoinAgents formats agent ids for a report column (comma-separated, sorted).
func JoinAgents(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	cp := append([]string(nil), ids...)
	sort.Strings(cp)
	return strings.Join(cp, ",")
}

// DedupeIDs returns unique ids in first-seen order, skipping empty values.
func DedupeIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
