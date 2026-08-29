// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/p3bot/agentdex"
	"github.com/spf13/pflag"

	"github.com/p3bot/snag/internal/logger"
	"github.com/p3bot/snag/internal/skill"
)

const skillCatalogSchema = `package catalog

import "struct"

#KnownAgent: {
	name:         string & !=""
	bin:          string & !=""
	description?: string
	config: {
		global: string & !=""
		local?: string & !=""
	}
	skills?: {
		global?: #SkillsScope
		local?:  #SkillsScope
		struct.MinFields(1)
	}
	version?: {
		args: [string, ...string]
		pattern?: string
	}
	agnostic: bool | *false
	if !agnostic {
		provider: [string, ...string]
	}
	homepage?: string
}

#SkillsScope: {
	agents?:       string & !=""
	native?:       string & !=""
	alternatives?: [string & !="", ...(string & !="")]
	struct.MinFields(1)
}

agents: [=~"^[a-z0-9]+(-[a-z0-9]+)*$"]: #KnownAgent
`

func writeSkillCatalog(t *testing.T, agentsBody string) string {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join("cue.mod", "module.cue"), "module: \"github.com/p3bot/agentdex/catalog@v1\"\nlanguage: {\n\tversion: \"v0.16.0\"\n}\n")
	write("schema.cue", skillCatalogSchema)
	write("agents.cue", "package catalog\n\n"+agentsBody+"\n")
	return dir
}

func skillFixtureBins(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range names {
		p := filepath.Join(dir, n)
		if err := os.WriteFile(p, []byte("#!/bin/sh\necho v1.0.0\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func skillEnvHome(home string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		if k == "HOME" {
			return home, true
		}
		return "", false
	}
}

func resetCLIFlags() {
	urlFile = ""
	flagOutput = ""
	outputDir = ""
	flagFormat = "md"
	timeout = DefaultTimeout
	waitFor = ""
	port = 9222
	closeTab = false
	forceHead = false
	openBrowser = false
	listTabs = false
	tab = ""
	allTabs = false
	killBrowser = false
	runDoctor = false
	showVersion = false
	info = false
	verbose = false
	debug = false
	userAgent = ""
	userDataDir = ""
	skillPrint = false
	skillInstall = nil
	skillList = false
	skillUninstall = nil
	skillLocal = false
	resetFlagSet := func(fs *pflag.FlagSet) {
		if fs == nil {
			return
		}
		fs.VisitAll(func(f *pflag.Flag) {
			f.Changed = false
			switch f.Name {
			case "help", "version":
				_ = f.Value.Set("false")
			}
		})
	}
	resetFlagSet(rootCmd.Flags())
	resetFlagSet(rootCmd.PersistentFlags())
}

func runSkillCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	resetCLIFlags()
	var out, errW bytes.Buffer
	logger.SetDefault(logger.NewWithWriter(logger.LevelNormal, &errW, false))
	t.Cleanup(func() {
		logger.SetDefault(logger.New(logger.LevelNormal))
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errW)
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	return out.String(), errW.String(), err
}

func useSkillFixture(t *testing.T, home, wd string, installed ...string) {
	t.Helper()
	const (
		binAlpha = "snag-fixture-alpha"
		binBeta  = "snag-fixture-beta"
		binGamma = "snag-fixture-gamma"
		binDelta = "snag-fixture-delta"
	)
	body := `
agents: "alpha-cli": {
	name: "Alpha CLI"
	bin:  "` + binAlpha + `"
	config: {global: "~/.alpha", local: ".alpha"}
	skills: {
		global: {native: "~/.alpha/skills"}
		local:  {native: ".alpha/skills"}
	}
	provider: ["anthropic"]
}
agents: "beta-tool": {
	name: "Beta Tool"
	bin:  "` + binBeta + `"
	config: {global: "~/.beta"}
	provider: ["openai"]
}
agents: "gamma-agent": {
	name: "Gamma Agent"
	bin:  "` + binGamma + `"
	config: {global: "~/.gamma", local: ".gamma"}
	skills: {
		global: {
			agents: "~/.agents/skills"
			alternatives: ["~/.claude/skills"]
		}
		local: {
			agents: ".agents/skills"
			alternatives: [".claude/skills"]
		}
	}
	provider: ["google"]
}
agents: "delta-agent": {
	name: "Delta Agent"
	bin:  "` + binDelta + `"
	config: {global: "~/.delta", local: ".delta"}
	skills: {
		global: {agents: "~/.agents/skills"}
		local:  {agents: ".agents/skills"}
	}
	provider: ["openai"]
}
agents: "epsilon-alt": {
	name: "Epsilon Alt Only"
	bin:  "snag-fixture-epsilon"
	config: {global: "~/.epsilon", local: ".epsilon"}
	skills: {
		global: {alternatives: ["~/.epsilon/skills"]}
		local:  {alternatives: [".epsilon/skills"]}
	}
	provider: ["openai"]
}
`
	catalogDir := writeSkillCatalog(t, body)
	binDir := skillFixtureBins(t, installed...)
	look := func(file string) (string, error) {
		for _, n := range installed {
			if file == n {
				return filepath.Join(binDir, n), nil
			}
		}
		return "", exec.ErrNotFound
	}
	prevOpts := skillAgentdexOpts
	prevGetwd := skillGetwd
	skillAgentdexOpts = []agentdex.Option{
		agentdex.WithCatalogDir(catalogDir),
		agentdex.WithCacheDir(t.TempDir()),
		agentdex.WithEnvLookup(skillEnvHome(home)),
		agentdex.WithLookPath(look),
		agentdex.WithSearchDirs(binDir),
	}
	skillGetwd = func() (string, error) { return wd, nil }
	t.Cleanup(func() {
		skillAgentdexOpts = prevOpts
		skillGetwd = prevGetwd
	})
}

func TestSkillIsNotASubcommand(t *testing.T) {
	if rootCmd.Flags().Lookup("skill") == nil {
		t.Fatal("expected --skill flag")
	}
	for _, c := range rootCmd.Commands() {
		switch c.Name() {
		case "skill", "skills":
			t.Fatalf("skill must not be a subcommand: %s", c.Name())
		}
	}
}

func TestSkillFalseDoesNotEnterSkillMode(t *testing.T) {
	out, _, err := runSkillCLI(t, "--skill=false", "--timeout", "0", "example.com")
	if out == skill.Text() {
		t.Fatal("--skill=false must not print the embed")
	}
	if err == nil {
		t.Fatal("expected fetch-path validation error, not success")
	}
	if strings.Contains(err.Error(), "skill mode") {
		t.Fatalf("--skill=false must not be a skill-mode error: %v", err)
	}
}

func TestSkillLocalFalseIsGlobal(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha")

	out, _, err := runSkillCLI(t, "--skill-install", "--local=false")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".alpha", "skills", "snag", "SKILL.md")
	if strings.TrimSpace(out) != want {
		t.Fatalf("out = %q want global %q", out, want)
	}
	if _, err := os.Stat(filepath.Join(wd, ".alpha", "skills", "snag", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatal("--local=false must not write project-local skills")
	}

	_, _, err = runSkillCLI(t, "--local=false", "--timeout", "0", "example.com")
	if err == nil {
		t.Fatal("expected fetch-path validation error, not success")
	}
	if strings.Contains(err.Error(), "--local") {
		t.Fatalf("--local=false must not require a skill flag: %v", err)
	}
}

func TestSkillPrintsEmbed(t *testing.T) {
	out, errOut, err := runSkillCLI(t, "--skill")
	if err != nil {
		t.Fatalf("snag --skill: %v", err)
	}
	if out != skill.Text() {
		t.Fatalf("stdout is not skill.Text() (got %d bytes, want %d)", len(out), len(skill.Text()))
	}
	if strings.Contains(errOut, "Error:") {
		t.Errorf("print should not error, stderr=%q", errOut)
	}
}

func TestSkillPrintIgnoresCatalog(t *testing.T) {
	prev := skillAgentdexOpts
	skillAgentdexOpts = nil
	t.Cleanup(func() { skillAgentdexOpts = prev })
	out, _, err := runSkillCLI(t, "--skill")
	if err != nil {
		t.Fatal(err)
	}
	if out != skill.Text() {
		t.Fatal("print must not require agentdex")
	}
}

func TestSkillFlagPlusPositional(t *testing.T) {
	_, _, err := runSkillCLI(t, "--skill", "example.com")
	if err == nil {
		t.Fatal("expected usage-class failure")
	}
}

func TestSkillInstallPositionalIsAddress(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha")
	_, _, err := runSkillCLI(t, "--skill-install", "grok")
	if err == nil {
		t.Fatal("expected positional grok to fail")
	}
	if !strings.Contains(err.Error(), "URL arguments") && !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("want positional conflict, got %v", err)
	}
	_, _, err = runSkillCLI(t, "--skill-uninstall", "grok")
	if err == nil {
		t.Fatal("expected positional grok to fail uninstall")
	}
}

func TestSkillInstallTrimsAgentID(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd)
	out, _, err := runSkillCLI(t, "--skill-install= alpha-cli ")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".alpha", "skills", "snag", "SKILL.md")
	if strings.TrimSpace(out) != want {
		t.Fatalf("out = %q want %q", out, want)
	}
}

func TestSkillInstallCSVIsOneID(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha")
	_, _, err := runSkillCLI(t, "--skill-install=grok,claude-code")
	if err == nil {
		t.Fatal("expected unknown single id")
	}
	if !errors.Is(err, skill.ErrUnknownAgent) {
		t.Fatalf("want unknown agent for comma id, got %v", err)
	}
	if strings.Contains(err.Error(), "grok\"") && !strings.Contains(err.Error(), "grok,claude-code") {
		t.Fatalf("comma must stay in one id, got %v", err)
	}
}

func TestSkillInstallStarIsNamedID(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha")
	_, _, err := runSkillCLI(t, "--skill-install=*")
	if err == nil || !errors.Is(err, skill.ErrUnknownAgent) {
		t.Fatalf("want unknown agent for typed *, got %v", err)
	}
}

func TestSkillOperationConflicts(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"doctor", []string{"--skill", "--doctor"}, "--doctor"},
		{"kill-browser", []string{"--skill", "--kill-browser"}, "--kill-browser"},
		{"list-tabs", []string{"--skill", "--list-tabs"}, "--list-tabs"},
		{"open-browser", []string{"--skill", "--open-browser"}, "--open-browser"},
		{"info", []string{"--skill", "--info"}, "--info"},
		{"tab", []string{"--skill", "--tab=1"}, "--tab"},
		{"all-tabs", []string{"--skill", "--all-tabs"}, "--all-tabs"},
		{"url-file", []string{"--skill", "--url-file", "urls.txt"}, "--url-file"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, _, err := runSkillCLI(t, tc.args...)
			if err == nil {
				t.Fatal("expected conflict")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %s in error, got %v", tc.want, err)
			}
			if out == skill.Text() {
				t.Fatal("must not print the embed on a conflict")
			}
		})
	}
}

func TestSkillBareArgIsAddress(t *testing.T) {
	out, _, err := runSkillCLI(t, "--timeout", "0", "skill")
	if out == skill.Text() {
		t.Fatal("snag skill must not print the embed")
	}
	if err == nil {
		t.Fatal("expected fetch-path validation error, not success")
	}
	if strings.Contains(err.Error(), "skill mode") {
		t.Fatalf("positional skill must not be a skill-mode error: %v", err)
	}
}

func TestSkillInstallDefaultPrimary(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha", "snag-fixture-gamma")

	out, errOut, err := runSkillCLI(t, "--skill-install")
	if err != nil {
		t.Fatalf("install: %v\nstderr=%s", err, errOut)
	}
	alphaFile := filepath.Join(home, ".alpha", "skills", "snag", "SKILL.md")
	gammaFile := filepath.Join(home, ".agents", "skills", "snag", "SKILL.md")
	lines := nonEmptyLines(out)
	if len(lines) != 2 {
		t.Fatalf("want 2 install paths, got %q", out)
	}
	if lines[0] != gammaFile || lines[1] != alphaFile {
		t.Fatalf("paths = %v want alphabetical [%s, %s]", lines, gammaFile, alphaFile)
	}
	for _, p := range []string{alphaFile, gammaFile} {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != skill.Text() {
			t.Errorf("%s content mismatch", p)
		}
	}
}

func TestSkillInstallNamedNativeElseShared(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-gamma")

	out, _, err := runSkillCLI(t, "--skill-install=delta-agent")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".agents", "skills", "snag", "SKILL.md")
	if strings.TrimSpace(out) != want {
		t.Fatalf("out = %q want %q", out, want)
	}
	out, _, err = runSkillCLI(t, "--skill-install=alpha-cli")
	if err != nil {
		t.Fatal(err)
	}
	wantAlpha := filepath.Join(home, ".alpha", "skills", "snag", "SKILL.md")
	if strings.TrimSpace(out) != wantAlpha {
		t.Fatalf("alpha out = %q want %q", out, wantAlpha)
	}
}

func TestSkillInstallRepeatedNamedIDs(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd)
	out, _, err := runSkillCLI(t, "--skill-install=alpha-cli", "--skill-install=delta-agent")
	if err != nil {
		t.Fatal(err)
	}
	lines := nonEmptyLines(out)
	if len(lines) != 2 {
		t.Fatalf("want 2 paths, got %q", out)
	}
	alpha := filepath.Join(home, ".alpha", "skills", "snag", "SKILL.md")
	shared := filepath.Join(home, ".agents", "skills", "snag", "SKILL.md")
	if lines[0] != shared || lines[1] != alpha {
		t.Fatalf("paths = %v want [%s, %s]", lines, shared, alpha)
	}
}

func TestSkillInstallMixBareAndNamed(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha")
	_, _, err := runSkillCLI(t, "--skill-install", "--skill-install=grok")
	if err == nil {
		t.Fatal("expected mix failure")
	}
	if !strings.Contains(err.Error(), "cannot combine valueless") {
		t.Fatalf("got %v", err)
	}
	_, _, err = runSkillCLI(t, "--skill-uninstall", "--skill-uninstall=grok")
	if err == nil {
		t.Fatal("expected uninstall mix failure")
	}
}

func TestSkillInstallLocal(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha")

	out, _, err := runSkillCLI(t, "--skill-install", "--local")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(wd, ".alpha", "skills", "snag", "SKILL.md")
	if strings.TrimSpace(out) != want {
		t.Fatalf("out = %q want %q", out, want)
	}
	if _, err := os.Stat(filepath.Join(home, ".alpha", "skills", "snag", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatal("global should not be written with --local")
	}
}

func TestSkillInstallEmptyDefaultSet(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd)
	_, _, err := runSkillCLI(t, "--skill-install")
	if err == nil || !errors.Is(err, skill.ErrEmptyAgentSet) {
		t.Fatalf("err = %v want empty set", err)
	}
}

func TestSkillUninstallEmptyDefaultSet(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd)
	_, _, err := runSkillCLI(t, "--skill-uninstall")
	if err == nil || !errors.Is(err, skill.ErrEmptyAgentSet) {
		t.Fatalf("err = %v want empty set", err)
	}
}

func TestSkillInstallUnknownAndNoSkills(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha")

	_, _, err := runSkillCLI(t, "--skill-install=no-such-agent")
	if err == nil || !errors.Is(err, skill.ErrUnknownAgent) {
		t.Fatalf("unknown: %v", err)
	}
	_, _, err = runSkillCLI(t, "--skill-install=beta-tool")
	if err == nil || !errors.Is(err, skill.ErrNoSkillsConcept) {
		t.Fatalf("no-skills: %v", err)
	}
}

func TestSkillInstallNamedNoWritablePath(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha")

	_, _, err := runSkillCLI(t, "--skill-install=epsilon-alt")
	if err == nil || !errors.Is(err, skill.ErrNoWritablePath) {
		t.Fatalf("err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".epsilon", "skills", "snag", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatal("named install must not write when Native and Shared are empty")
	}
}

func TestSkillInstallDedupeShared(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd)
	out, _, err := runSkillCLI(t, "--skill-install=gamma-agent", "--skill-install=delta-agent")
	if err != nil {
		t.Fatal(err)
	}
	lines := nonEmptyLines(out)
	if len(lines) != 1 {
		t.Fatalf("want 1 path for shared de-dupe, got %q", out)
	}
	want := filepath.Join(home, ".agents", "skills", "snag", "SKILL.md")
	if lines[0] != want {
		t.Fatalf("path = %q want %q", lines[0], want)
	}
}

func TestSkillList(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha", "snag-fixture-gamma", "snag-fixture-delta")

	out, errOut, err := runSkillCLI(t, "--skill-list")
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Fatalf("empty inventory want empty stdout, got %q", out)
	}
	if !strings.Contains(errOut, "not installed") {
		t.Fatalf("empty inventory want stderr note, got %q", errOut)
	}

	if _, _, err := runSkillCLI(t, "--skill-install"); err != nil {
		t.Fatal(err)
	}
	out, _, err = runSkillCLI(t, "--skill-list")
	if err != nil {
		t.Fatal(err)
	}
	lines := nonEmptyLines(out)
	if len(lines) != 2 {
		t.Fatalf("list lines = %q", out)
	}
	shared := filepath.Join(home, ".agents", "skills", "snag", "SKILL.md")
	alpha := filepath.Join(home, ".alpha", "skills", "snag", "SKILL.md")
	parts0 := strings.SplitN(lines[0], "\t", 2)
	parts1 := strings.SplitN(lines[1], "\t", 2)
	if parts0[0] != shared || parts1[0] != alpha {
		t.Fatalf("list order = %v want alphabetical [%s, %s]", lines, shared, alpha)
	}
	if len(parts0) < 2 || parts0[1] != "delta-agent,gamma-agent" {
		t.Fatalf("shared agents = %q want delta-agent,gamma-agent", parts0)
	}
}

func TestSkillListEmptyDefaultSet(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd)
	out, errOut, err := runSkillCLI(t, "--skill-list")
	if err != nil {
		t.Fatalf("list empty set: %v", err)
	}
	if out != "" {
		t.Fatalf("want empty stdout, got %q", out)
	}
	if !strings.Contains(errOut, "not installed") {
		t.Fatalf("want stderr note, got %q", errOut)
	}
}

func TestSkillListLocal(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha")
	if _, _, err := runSkillCLI(t, "--skill-install", "--local"); err != nil {
		t.Fatal(err)
	}
	out, _, err := runSkillCLI(t, "--skill-list", "--local")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(wd, ".alpha", "skills", "snag", "SKILL.md")
	if !strings.Contains(out, want) {
		t.Fatalf("list --local missing %s in %q", want, out)
	}
	out, errOut, err := runSkillCLI(t, "--skill-list")
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Fatalf("global list should be empty stdout after local-only install, got %q", out)
	}
	if !strings.Contains(errOut, "not installed") {
		t.Fatalf("global list should note not installed, got %q", errOut)
	}
}

func TestSkillUninstallLocal(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha")
	if _, _, err := runSkillCLI(t, "--skill-install", "--local"); err != nil {
		t.Fatal(err)
	}
	localDir := filepath.Join(wd, ".alpha", "skills", "snag")
	out, _, err := runSkillCLI(t, "--skill-uninstall", "--local")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "removed\t"+localDir) {
		t.Fatalf("want removed local, got %q", out)
	}
	if _, err := os.Stat(localDir); !os.IsNotExist(err) {
		t.Fatal("local skill dir should be gone")
	}
}

func TestSkillUninstallRepeatedNamedIDs(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-gamma", "snag-fixture-delta")
	if _, _, err := runSkillCLI(t, "--skill-install"); err != nil {
		t.Fatal(err)
	}
	sharedDir := filepath.Join(home, ".agents", "skills", "snag")
	out, _, err := runSkillCLI(t, "--skill-uninstall=gamma-agent", "--skill-uninstall=delta-agent")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "removed\t"+sharedDir) {
		t.Fatalf("want removed, got %q", out)
	}
}

func TestSkillUninstallMultiTenantAbsent(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-gamma", "snag-fixture-delta")
	sharedDir := filepath.Join(home, ".agents", "skills", "snag")
	out, _, err := runSkillCLI(t, "--skill-uninstall=gamma-agent")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "absent\t"+sharedDir) {
		t.Fatalf("want absent when blocked and missing, got %q", out)
	}
	if !strings.Contains(out, "delta-agent") {
		t.Fatalf("want delta blocker on absent line, got %q", out)
	}
	if strings.Contains(out, "kept\t"+sharedDir) {
		t.Fatalf("must not report kept for missing skill, got %q", out)
	}
}

func TestSkillUninstallMultiTenantKeep(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-gamma", "snag-fixture-delta")
	if _, _, err := runSkillCLI(t, "--skill-install"); err != nil {
		t.Fatal(err)
	}
	sharedDir := filepath.Join(home, ".agents", "skills", "snag")

	out, _, err := runSkillCLI(t, "--skill-uninstall=gamma-agent")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "kept\t"+sharedDir) {
		t.Fatalf("want kept shared, got %q", out)
	}
	if !strings.Contains(out, "delta-agent") {
		t.Fatalf("want delta blocker, got %q", out)
	}
	if _, err := os.Stat(filepath.Join(sharedDir, "SKILL.md")); err != nil {
		t.Fatal("shared skill should remain")
	}

	out, _, err = runSkillCLI(t, "--skill-uninstall=gamma-agent", "--skill-uninstall=delta-agent")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "removed\t"+sharedDir) {
		t.Fatalf("want removed, got %q", out)
	}
	if _, err := os.Stat(sharedDir); !os.IsNotExist(err) {
		t.Fatal("shared dir should be gone")
	}
}

func TestSkillUninstallPurity(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha")
	if _, _, err := runSkillCLI(t, "--skill-install=alpha-cli"); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, ".alpha", "skills", "snag")
	if err := os.WriteFile(filepath.Join(dir, "extra.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _, err := runSkillCLI(t, "--skill-uninstall=alpha-cli")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "kept\t"+dir) {
		t.Fatalf("want kept for extra, got %q", out)
	}
	if !strings.Contains(out, "directory is not pure (expected only SKILL.md)") {
		t.Fatalf("want purity reason, got %q", out)
	}
}

func TestSkillCatalogInvalid(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	bad := writeSkillCatalog(t, `agents: "x": { name: "", bin: "y", config: {global: "~/.x"}, provider: ["openai"] }`)
	prevOpts := skillAgentdexOpts
	prevGetwd := skillGetwd
	skillAgentdexOpts = []agentdex.Option{
		agentdex.WithCatalogDir(bad),
		agentdex.WithCacheDir(t.TempDir()),
		agentdex.WithEnvLookup(skillEnvHome(home)),
		agentdex.WithLookPath(func(string) (string, error) { return "", exec.ErrNotFound }),
	}
	skillGetwd = func() (string, error) { return wd, nil }
	t.Cleanup(func() {
		skillAgentdexOpts = prevOpts
		skillGetwd = prevGetwd
	})
	_, _, err := runSkillCLI(t, "--skill-install")
	if err == nil {
		t.Fatal("expected catalog failure")
	}
	if !strings.Contains(err.Error(), "snag --skill") {
		t.Fatalf("message should point at manual install: %v", err)
	}
}

func TestSkillLocalRequiresWorkingDirectory(t *testing.T) {
	home := t.TempDir()
	wd := t.TempDir()
	useSkillFixture(t, home, wd, "snag-fixture-alpha")
	prevGetwd := skillGetwd
	skillGetwd = func() (string, error) { return "", errors.New("gone") }
	t.Cleanup(func() { skillGetwd = prevGetwd })

	_, _, err := runSkillCLI(t, "--skill-install", "--local")
	if err == nil || !strings.Contains(err.Error(), "working directory required for --local") {
		t.Fatalf("want local Getwd error, got %v", err)
	}

	out, _, err := runSkillCLI(t, "--skill-install")
	if err != nil {
		t.Fatalf("global install should run when Getwd fails: %v", err)
	}
	want := filepath.Join(home, ".alpha", "skills", "snag", "SKILL.md")
	if strings.TrimSpace(out) != want {
		t.Fatalf("global out = %q want %q", out, want)
	}
}

func TestSkillVersionWins(t *testing.T) {
	out, _, err := runSkillCLI(t, "--version", "--skill")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "name: snag") {
		t.Fatal("must not print skill body when --version wins")
	}
}

func TestCLI_VersionWinsOverDoctor(t *testing.T) {
	stdout, stderr, err := runSnag("--version", "--doctor")
	assertNoError(t, err)
	out := stdout + stderr
	if !strings.Contains(out, "snag version") {
		t.Fatalf("expected version, got %s", out)
	}
	if strings.Contains(out, "OS/Arch") {
		t.Fatal("doctor must not run when --version is set")
	}
}

func TestCLI_VersionIgnoresMissingURLFile(t *testing.T) {
	stdout, stderr, err := runSnag("--version", "--url-file", filepath.Join(t.TempDir(), "missing.txt"))
	assertNoError(t, err)
	out := stdout + stderr
	if !strings.Contains(out, "snag version") {
		t.Fatalf("expected version, got %s", out)
	}
}

func TestSkillHelpWins(t *testing.T) {
	out, errOut, err := runSkillCLI(t, "--help", "--skill")
	if err != nil {
		t.Fatalf("help: %v stderr=%s", err, errOut)
	}
	combined := out + errOut
	if !strings.Contains(combined, "USAGE") {
		t.Fatalf("help should win, got %q", combined)
	}
}

func TestCLI_SkillPrintBinary(t *testing.T) {
	stdout, stderr, err := runSnag("--skill")
	if err != nil {
		t.Fatalf("binary --skill: %v stderr=%s", err, stderr)
	}
	if stdout != skill.Text() {
		t.Fatalf("binary stdout is not embed (got %d want %d)", len(stdout), len(skill.Text()))
	}
}

func TestCLI_SkillExtraPositionalBinary(t *testing.T) {
	_, _, err := runSnag("--skill", "example.com")
	assertError(t, err)
	assertExitCode(t, err, 1)
}

func TestSkillIgnoresFetchModifiers(t *testing.T) {
	out, _, err := runSkillCLI(t, "--skill", "--format", "html", "-o", "x.md", "--timeout", "5")
	if err != nil {
		t.Fatal(err)
	}
	if out != skill.Text() {
		t.Fatal("fetch modifiers must be ignored in skill print")
	}
}

func TestCLI_SkillDoctorConflictBinary(t *testing.T) {
	_, _, err := runSnag("--skill", "--doctor")
	assertError(t, err)
	assertExitCode(t, err, 1)
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
