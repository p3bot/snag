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

// Catalog is an open agentdex index. Callers outside this package do not import agentdex.
type Catalog struct {
	idx *agentdex.Index
}

// Agent is a catalogued agent with skills-path data resolved by agentdex.
type Agent struct {
	ID    string
	inner agentdex.Agent
}

// Roots returns Primary / Native / Shared roots at loc.
func (a Agent) Roots(loc Location) AgentRoots {
	return rootsAt(a.inner, loc)
}

// OpenConfig is the test seam for catalog dir, cache, HOME, and binary lookup.
// Production Open uses agentdex defaults.
type OpenConfig struct {
	CatalogDir string
	CacheDir   string
	Home       string
	SearchDirs []string
	LookPath   func(string) (string, error)
}

var testOpen *OpenConfig

// SetTestOpen installs OpenConfig for subsequent Open calls. Pass nil to clear.
// The returned function restores the previous config.
func SetTestOpen(c *OpenConfig) func() {
	prev := testOpen
	testOpen = c
	return func() { testOpen = prev }
}

func hasSkillsConcept(a agentdex.Agent) bool {
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

func rootsAt(a agentdex.Agent, loc Location) AgentRoots {
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

func wrapAgents(in []agentdex.Agent) []Agent {
	out := make([]Agent, len(in))
	for i, a := range in {
		out[i] = Agent{ID: a.ID, inner: a}
	}
	return out
}

// Open constructs a Catalog. workingDir is applied last so Getwd vs "/" is
// not overwritten by test seams that also set a working directory.
func Open(workingDir string) (*Catalog, error) {
	opts := make([]agentdex.Option, 0, 8)
	if testOpen != nil {
		if testOpen.CatalogDir != "" {
			opts = append(opts, agentdex.WithCatalogDir(testOpen.CatalogDir))
		}
		if testOpen.CacheDir != "" {
			opts = append(opts, agentdex.WithCacheDir(testOpen.CacheDir))
		}
		if testOpen.Home != "" {
			home := testOpen.Home
			opts = append(opts, agentdex.WithEnvLookup(func(k string) (string, bool) {
				if k == "HOME" {
					return home, true
				}
				return "", false
			}))
		}
		if testOpen.LookPath != nil {
			opts = append(opts, agentdex.WithLookPath(testOpen.LookPath))
		}
		if len(testOpen.SearchDirs) > 0 {
			opts = append(opts, agentdex.WithSearchDirs(testOpen.SearchDirs...))
		}
	}
	opts = append(opts, agentdex.WithWorkingDir(workingDir))
	idx, err := agentdex.Open(opts...)
	if err != nil {
		return nil, MapCatalogError(err)
	}
	return &Catalog{idx: idx}, nil
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
func DefaultSet(ctx context.Context, cat *Catalog) ([]Agent, error) {
	res, err := cat.idx.Agents.List(ctx, agentdex.AgentQuery{
		Installed: true,
		Enrich:    agentdex.EnrichNone,
	})
	if err != nil {
		return nil, MapCatalogError(err)
	}
	var out []agentdex.Agent
	for _, a := range res.Items {
		if hasSkillsConcept(a) {
			out = append(out, a)
		}
	}
	return wrapAgents(out), nil
}

// ResolveExplicit loads each id via Get (paths resolve even when !Found).
func ResolveExplicit(ctx context.Context, cat *Catalog, ids []string) ([]Agent, error) {
	out := make([]agentdex.Agent, 0, len(ids))
	for _, id := range ids {
		d, err := cat.idx.Agents.Get(ctx, id, agentdex.AgentGetQuery{Enrich: agentdex.EnrichNone})
		if err != nil {
			if errors.Is(err, agentdex.ErrAgentUnknown) {
				return nil, fmt.Errorf("%w %q", ErrUnknownAgent, id)
			}
			return nil, MapCatalogError(err)
		}
		if !hasSkillsConcept(d.Agent) {
			return nil, fmt.Errorf("%w: %q", ErrNoSkillsConcept, id)
		}
		out = append(out, d.Agent)
	}
	return wrapAgents(out), nil
}

// NoWritablePathError returns ErrNoWritablePath wrapped with the agent id.
func NoWritablePathError(agentID string) error {
	return fmt.Errorf("%w under the named install rule: %q", ErrNoWritablePath, agentID)
}
