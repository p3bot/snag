// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package skill

import (
	"context"
	"errors"
	"fmt"

	"github.com/p3bot/agentdex"
)

var (
	// ErrUnknownAgent is an agent id absent from the catalog.
	ErrUnknownAgent = errors.New("unknown agent")

	// ErrNoSkillsConcept is a catalogued agent with no skills path data.
	ErrNoSkillsConcept = errors.New("agent has no skills concept")

	// ErrNoWritablePath is a named install with no Native or Shared root.
	ErrNoWritablePath = errors.New("agent has no writable skills path")

	// ErrEmptyAgentSet is install/uninstall with no agents in the default set.
	ErrEmptyAgentSet = errors.New("no installed agents with a skills concept")
)

// Location selects global vs project-local skill roots from agentdex.
type Location int

const (
	// LocationGlobal uses Detection.Skills.Global.
	LocationGlobal Location = iota
	// LocationLocal uses Detection.Skills.Local (working-directory base).
	LocationLocal
)

// AgentRoots is one agent's absolute skill roots at a location scope.
// Empty strings mean that role is unset for the agent.
type AgentRoots struct {
	ID      string
	Found   bool
	Primary string
	Native  string
	Shared  string
}

// HasSkillsConcept reports whether agentdex resolved any skills path data.
func HasSkillsConcept(a agentdex.Agent) bool {
	return !skillsPathsZero(a.Detection.Skills)
}

func skillsPathsZero(sk agentdex.SkillsPaths) bool {
	return skillsScopeZero(sk.Global) && skillsScopeZero(sk.Local)
}

func skillsScopeZero(sc agentdex.SkillsScope) bool {
	return sc.Agents.Path == "" && !sc.Agents.Exists &&
		sc.Native.Path == "" && !sc.Native.Exists &&
		len(sc.Alternatives) == 0 &&
		sc.Primary.Path == "" && !sc.Primary.Exists
}

// RootsAt extracts Primary / Native / Shared (agents role) absolute paths
// for the given location. Paths are cleaned; empty means unset.
func RootsAt(a agentdex.Agent, loc Location) AgentRoots {
	sc := a.Detection.Skills.Global
	if loc == LocationLocal {
		sc = a.Detection.Skills.Local
	}
	return AgentRoots{
		ID:      a.ID,
		Found:   a.Detection.Found,
		Primary: CleanAbs(sc.Primary.Path),
		Native:  CleanAbs(sc.Native.Path),
		Shared:  CleanAbs(sc.Agents.Path),
	}
}

// Candidates returns unique non-empty absolute paths among Primary, Native,
// and Shared (agents role).
func Candidates(r AgentRoots) []string {
	return DedupePaths([]string{r.Primary, r.Native, r.Shared})
}

// InstallRoot selects the skills root for an install path rule.
// named=false → Primary; named=true → Native if set else Shared.
func InstallRoot(r AgentRoots, named bool) string {
	if !named {
		return r.Primary
	}
	if r.Native != "" {
		return r.Native
	}
	return r.Shared
}

// OpenIndex constructs an agentdex Index for skill path operations.
// workingDir is applied last so the CLI's Getwd vs "/" fallback is not
// overwritten by test seams that also pass WithWorkingDir.
func OpenIndex(workingDir string, extra ...agentdex.Option) (*agentdex.Index, error) {
	opts := make([]agentdex.Option, 0, 1+len(extra))
	opts = append(opts, extra...)
	opts = append(opts, agentdex.WithWorkingDir(workingDir))
	return agentdex.Open(opts...)
}

// MapCatalogError turns agentdex catalog sentinels into a user-facing error
// that points at manual install via `snag --skill`. Other errors pass through.
func MapCatalogError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, agentdex.ErrCatalogUnavailable):
		return fmt.Errorf("catalog unavailable: run 'snag --skill' and install the skill manually into your agent's skills directory (%w)", err)
	case errors.Is(err, agentdex.ErrCatalogInvalid):
		return fmt.Errorf("catalog invalid: run 'snag --skill' and install the skill manually into your agent's skills directory (%w)", err)
	default:
		return err
	}
}

// DefaultSet returns agents with Found and a skills concept.
func DefaultSet(ctx context.Context, idx *agentdex.Index) ([]agentdex.Agent, error) {
	res, err := idx.Agents.List(ctx, agentdex.AgentQuery{
		Installed: true,
		Enrich:    agentdex.EnrichNone,
	})
	if err != nil {
		return nil, MapCatalogError(err)
	}
	out := make([]agentdex.Agent, 0, len(res.Items))
	for _, a := range res.Items {
		if HasSkillsConcept(a) {
			out = append(out, a)
		}
	}
	return out, nil
}

// ResolveExplicit loads each id via Get (paths resolve even when !Found).
func ResolveExplicit(ctx context.Context, idx *agentdex.Index, ids []string) ([]agentdex.Agent, error) {
	out := make([]agentdex.Agent, 0, len(ids))
	for _, id := range ids {
		d, err := idx.Agents.Get(ctx, id, agentdex.AgentGetQuery{Enrich: agentdex.EnrichNone})
		if err != nil {
			if errors.Is(err, agentdex.ErrAgentUnknown) {
				return nil, fmt.Errorf("%w %q", ErrUnknownAgent, id)
			}
			return nil, MapCatalogError(err)
		}
		if !HasSkillsConcept(d.Agent) {
			return nil, fmt.Errorf("%w: %q", ErrNoSkillsConcept, id)
		}
		out = append(out, d.Agent)
	}
	return out, nil
}

// NoWritablePathError returns ErrNoWritablePath wrapped with the agent id.
func NoWritablePathError(agentID string) error {
	return fmt.Errorf("%w under the named install rule: %q", ErrNoWritablePath, agentID)
}
