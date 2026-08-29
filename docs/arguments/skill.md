# `--skill`, `--skill-install`, `--skill-list`, `--skill-uninstall`, `--local`

**Status:** Complete (2026-08-29)

#### Validation Rules

**Skill verbs (mutually exclusive):**

- `--skill` — boolean, print the embedded skill contract
- `--skill-install[=id]` — optional value, repeatable; no CSV split
- `--skill-list` — boolean, no agent ids
- `--skill-uninstall[=id]` — optional value, repeatable; no CSV split

A following token is a positional URL, not an agent id. Use `--skill-install=grok`, not `--skill-install grok`.

`--skill-install` (no value) and `--skill-install=id` cannot appear on the same line (same for uninstall).

**`--local`:**

- Boolean
- Allowed only with `--skill-install`, `--skill-list`, or `--skill-uninstall`
- Default is global
- Requires a working directory; if Getwd fails: error `working directory required for --local`
- Global skill operations still run if Getwd fails (working dir `/` is passed to agentdex so `~` expands)

**Multiple skill verbs:**

- Combining any two of `--skill`, `--skill-install`, `--skill-list`, `--skill-uninstall` → error

#### Behavior

**`--skill`:**

- Prints the embedded `SKILL.md` body to stdout
- Does not open agentdex; catalog failure cannot block print

**`--skill-install`:**

- No value: write Primary `snag/SKILL.md` for the default agent set (Found and has a skills concept)
- `=id` (repeatable): Native if set, else Shared; Found is not required
- De-dupe by absolute path; write once per path; atomic replace (same-directory temp + rename)
- Stdout: one written `…/snag/SKILL.md` path per line, sorted
- Empty default set: usage-class error

**`--skill-list`:**

- Default set only
- Stdout: TSV `path\tagents` (claimers comma-separated, sorted)
- Empty inventory: empty stdout, `not installed` on stderr, exit 0

**`--skill-uninstall`:**

- Path universe is candidates(Primary, Native, Shared) for S
- Multi-tenant blockers R keep a path (reported, not an error)
- Purity: remove only a directory that contains exactly `SKILL.md` with frontmatter `name: snag`
- Stdout: TSV `removed|absent|kept\tdir`. Third column when kept (purity reason) or when multi-tenant blockers apply, including absent (`absent\tdir\tblockers`)

**Install is user-initiated only.** Never implied by fetch or other commands.

**Catalog failure:** fail closed; tell the user to run `snag --skill` and place the file manually. Do not invent paths.

#### Precedence

1. `--help` (highest)
2. `--version`
3. Skill mode vs `--doctor` / `--kill-browser` / content operations: usage-class error (neither side runs)
4. Skill verbs among themselves: mutually exclusive

`--verbose` / `--debug` work with skill modes.

Fetch modifiers (`--format`, `-o`, `-d`, `--timeout`, `--wait-for`, `--port`, and similar) are ignored, same as `--kill-browser`.

#### Interaction Matrix

**Higher priority:**

| Combination | Behavior |
| ----------- | -------- |
| `--skill` + `--help` | Help, exit 0 |
| `--skill` + `--version` | Version, exit 0 |

**Conflicts (usage-class error, exit 1):**

| Combination | Error |
| ----------- | ----- |
| Two skill verbs | Mutually exclusive skill flags |
| Skill verb + URL positional | Conflicting operations |
| `--skill-install grok` | `grok` is a URL positional |
| `--skill-install` + `--skill-install=id` | Do not union default-set and named path rules |
| Skill verb + `--doctor` | Conflicting operations |
| Skill verb + `--kill-browser` | Conflicting operations |
| Skill verb + `--list-tabs` | Conflicting operations |
| Skill verb + `--open-browser` | Conflicting operations |
| Skill verb + `--info` | Conflicting operations |
| Skill verb + `--tab` | Conflicting operations |
| Skill verb + `--all-tabs` | Conflicting operations |
| Skill verb + `--url-file` | Conflicting operations |
| `--local` without install/list/uninstall | `--local` requires a skill install, list, or uninstall flag |
| `--local` + `--skill` | `--local` is not valid with print |

**Works with skill modes:**

| Combination | Behavior |
| ----------- | -------- |
| Skill verb + `--verbose` | Verbose logging |
| Skill verb + `--debug` | Debug logging |
| `--skill-install` + `--local` | Project-local roots |
| `--skill-install=id` repeated | All named ids apply (path de-dupe) |

**Silently ignored:**

`--format`, `--output`, `--output-dir`, `--timeout`, `--wait-for`, `--port`, `--close-tab`, `--force-headless`, `--user-agent`, `--user-data-dir`

#### Examples

```bash
snag --skill
snag --skill-install
snag --skill-install=grok --skill-install=claude-code
snag --skill-list
snag --skill-uninstall
snag --skill-uninstall=grok --local
```
