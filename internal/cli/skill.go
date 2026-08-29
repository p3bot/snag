// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/p3bot/agentdex"
	"github.com/spf13/cobra"

	"github.com/p3bot/snag/internal/logger"
	"github.com/p3bot/snag/internal/skill"
)

// skillBareSentinel is NoOptDefVal for --skill-install / --skill-uninstall so a
// following token stays a positional URL. NUL cannot appear in argv, so it
// cannot collide with a typed --skill-install=id.
const skillBareSentinel = "\x00"

var (
	skillAgentdexOpts []agentdex.Option
	skillGetwd        = os.Getwd
)

func skillFlagsChanged(cmd *cobra.Command) bool {
	return skillPrint ||
		cmd.Flags().Changed("skill-install") ||
		skillList ||
		cmd.Flags().Changed("skill-uninstall")
}

func handleSkill(cmd *cobra.Command, args []string) error {
	print := skillPrint
	install := cmd.Flags().Changed("skill-install")
	list := skillList
	uninstall := cmd.Flags().Changed("skill-uninstall")
	local := skillLocal

	n := 0
	for _, on := range []bool{print, install, list, uninstall} {
		if on {
			n++
		}
	}
	if n > 1 {
		logger.Error("Cannot combine skill flags (mutually exclusive)")
		return fmt.Errorf("conflicting flags: skill modes are mutually exclusive")
	}

	if err := rejectSkillConflicts(cmd, args); err != nil {
		return err
	}

	if print {
		if local {
			logger.Error("Cannot use --local with --skill")
			return fmt.Errorf("conflicting flags: --local and --skill")
		}
		_, err := fmt.Fprint(cmd.OutOrStdout(), skill.Text())
		return err
	}

	if !install && !list && !uninstall {
		return fmt.Errorf("no skill mode selected")
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	switch {
	case install:
		ids, named, err := parseSkillAgentIDs(skillInstall, "skill-install")
		if err != nil {
			logger.Error("%s", err.Error())
			return err
		}
		return runSkillInstall(cmd, ctx, ids, named, local)
	case list:
		return runSkillList(cmd, ctx, local)
	default:
		ids, named, err := parseSkillAgentIDs(skillUninstall, "skill-uninstall")
		if err != nil {
			logger.Error("%s", err.Error())
			return err
		}
		return runSkillUninstall(cmd, ctx, ids, named, local)
	}
}

func rejectSkillConflicts(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		if strings.TrimSpace(arg) != "" {
			logger.Error("Cannot use a skill flag with URL arguments (conflicting operations)")
			return fmt.Errorf("conflicting flags: skill mode and URL arguments")
		}
	}

	conflicts := []struct {
		name string
		on   bool
	}{
		{"--doctor", runDoctor},
		{"--kill-browser", killBrowser},
		{"--list-tabs", listTabs},
		{"--open-browser", openBrowser},
		{"--info", info},
		{"--tab", cmd.Flags().Changed("tab")},
		{"--all-tabs", allTabs},
		{"--url-file", cmd.Flags().Changed("url-file")},
	}
	for _, c := range conflicts {
		if c.on {
			logger.Error("Cannot use a skill flag with %s (conflicting operations)", c.name)
			return fmt.Errorf("conflicting flags: skill mode and %s", c.name)
		}
	}
	return nil
}

func parseSkillAgentIDs(vals []string, flag string) (ids []string, named bool, err error) {
	var hasBare, hasNamed bool
	for _, v := range vals {
		if v == skillBareSentinel {
			hasBare = true
			continue
		}
		v = strings.TrimSpace(v)
		if v == "" {
			return nil, false, fmt.Errorf("empty agent id on --%s", flag)
		}
		hasNamed = true
		ids = append(ids, v)
	}
	if hasBare && hasNamed {
		return nil, false, fmt.Errorf("cannot combine valueless --%s with --%s=id", flag, flag)
	}
	if hasNamed {
		return skill.DedupeIDs(ids), true, nil
	}
	return nil, false, nil
}

func skillLocation(local bool) skill.Location {
	if local {
		return skill.LocationLocal
	}
	return skill.LocationGlobal
}

func openSkillIndex(local bool) (*agentdex.Index, error) {
	wd, err := skillGetwd()
	if err != nil {
		if local {
			return nil, fmt.Errorf("working directory required for --local: %w", err)
		}
		wd = "/"
	}
	return skill.OpenIndex(wd, skillAgentdexOpts...)
}

func runSkillInstall(cmd *cobra.Command, ctx context.Context, agentIDs []string, named, local bool) error {
	idx, err := openSkillIndex(local)
	if err != nil {
		return err
	}
	loc := skillLocation(local)

	var agents []agentdex.Agent
	if named {
		agents, err = skill.ResolveExplicit(ctx, idx, agentIDs)
		if err != nil {
			logger.Error("%s", err.Error())
			return err
		}
	} else {
		agents, err = skill.DefaultSet(ctx, idx)
		if err != nil {
			return err
		}
		if len(agents) == 0 {
			logger.Error("%s", skill.ErrEmptyAgentSet.Error())
			return skill.ErrEmptyAgentSet
		}
	}

	seen := make(map[string]struct{})
	var order []string
	for _, a := range agents {
		r := skill.RootsAt(a, loc)
		root := skill.InstallRoot(r, named)
		if root == "" {
			if named {
				err := skill.NoWritablePathError(a.ID)
				logger.Error("%s", err.Error())
				return err
			}
			continue
		}
		if _, ok := seen[root]; ok {
			continue
		}
		seen[root] = struct{}{}
		order = append(order, root)
	}
	if len(order) == 0 {
		logger.Error("%s", skill.ErrNoWritablePath.Error())
		return skill.ErrNoWritablePath
	}
	skill.SortPaths(order)

	for _, root := range order {
		written, err := skill.WriteInstall(root)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), written)
	}
	return nil
}

func runSkillList(cmd *cobra.Command, ctx context.Context, local bool) error {
	idx, err := openSkillIndex(local)
	if err != nil {
		return err
	}
	loc := skillLocation(local)

	agents, err := skill.DefaultSet(ctx, idx)
	if err != nil {
		return err
	}
	claimers := make(map[string][]string)
	for _, a := range agents {
		r := skill.RootsAt(a, loc)
		for _, p := range skill.Candidates(r) {
			if !skill.Present(p) {
				continue
			}
			claimers[p] = append(claimers[p], a.ID)
		}
	}
	paths := make([]string, 0, len(claimers))
	for p := range claimers {
		paths = append(paths, p)
	}
	skill.SortPaths(paths)
	if len(paths) == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "not installed")
		return nil
	}
	for _, p := range paths {
		fmt.Fprintln(cmd.OutOrStdout(), skill.FilePath(p)+"\t"+skill.JoinAgents(claimers[p]))
	}
	return nil
}

func runSkillUninstall(cmd *cobra.Command, ctx context.Context, agentIDs []string, named, local bool) error {
	idx, err := openSkillIndex(local)
	if err != nil {
		return err
	}
	loc := skillLocation(local)

	var sAgents []agentdex.Agent
	if named {
		sAgents, err = skill.ResolveExplicit(ctx, idx, agentIDs)
		if err != nil {
			logger.Error("%s", err.Error())
			return err
		}
	} else {
		sAgents, err = skill.DefaultSet(ctx, idx)
		if err != nil {
			return err
		}
		if len(sAgents) == 0 {
			logger.Error("%s", skill.ErrEmptyAgentSet.Error())
			return skill.ErrEmptyAgentSet
		}
	}

	var rAgents []agentdex.Agent
	if named {
		sIDs := make(map[string]struct{}, len(sAgents))
		for _, a := range sAgents {
			sIDs[a.ID] = struct{}{}
		}
		defaultSet, err := skill.DefaultSet(ctx, idx)
		if err != nil {
			return err
		}
		for _, a := range defaultSet {
			if _, inS := sIDs[a.ID]; !inS {
				rAgents = append(rAgents, a)
			}
		}
	}

	rClaim := make(map[string][]string)
	for _, a := range rAgents {
		r := skill.RootsAt(a, loc)
		for _, p := range skill.Candidates(r) {
			rClaim[p] = append(rClaim[p], a.ID)
		}
	}

	pathSet := make(map[string]struct{})
	for _, a := range sAgents {
		r := skill.RootsAt(a, loc)
		for _, p := range skill.Candidates(r) {
			pathSet[p] = struct{}{}
		}
	}
	paths := make([]string, 0, len(pathSet))
	for p := range pathSet {
		paths = append(paths, p)
	}
	skill.SortPaths(paths)

	out := cmd.OutOrStdout()
	for _, p := range paths {
		dir := skill.DirPath(p)
		if blockers := rClaim[p]; len(blockers) > 0 {
			sort.Strings(blockers)
			if !skill.Present(p) {
				fmt.Fprintf(out, "absent\t%s\t%s\n", dir, skill.JoinAgents(blockers))
			} else {
				fmt.Fprintf(out, "kept\t%s\t%s\n", dir, skill.JoinAgents(blockers))
			}
			continue
		}
		res, err := skill.RemoveOwned(p)
		if err != nil {
			return err
		}
		switch res {
		case skill.UninstallRemoved:
			fmt.Fprintf(out, "removed\t%s\n", dir)
		case skill.UninstallAbsent:
			fmt.Fprintf(out, "absent\t%s\n", dir)
		case skill.UninstallKeptExtra, skill.UninstallKeptNotOurs, skill.UninstallKeptUnreadable:
			fmt.Fprintf(out, "kept\t%s\t%s\n", dir, skill.ReasonDetail(res))
		default:
			fmt.Fprintf(out, "%s\t%s\n", res.String(), dir)
		}
	}
	return nil
}
