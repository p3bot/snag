---
name: snag
description: >-
  Fetch rendered web page content via Chrome/Chromium (CDP) as Markdown,
  HTML, text, PDF, or PNG. Use when an agent needs page content from a URL
  or an existing browser tab. First positional is always a URL; no subcommands.
---

# Fetch web pages with snag

- First positional is a URL/address. There are no subcommands. `snag skill` fetches host `skill`.
- stdout = content (or a skill-mode path report). stderr = logs. Default stderr is silent for fetch progress; result/status lines still print (generated filename, `Saved to`, empty `--skill-list` `not installed`).
- Install is user-initiated only. Never install on fetch or any unrelated command.
- Skill verbs are flags with optional `=id` values, not `snag skill install [agents...]`.

## Fetch

```
snag <url>...
snag --url-file urls.txt
snag --url-file -          # stdin
snag -o page.md <url>
snag -d out/ <url>...
snag -f md|html|text|pdf|png <url>
snag --wait-for ".sel" [--timeout N] <url>
snag --info <url>          # JSON metadata
```

PDF/PNG never go to stdout; a filename is generated when `-o`/`-d` are omitted.

## Tabs

Tabs are sorted by URL, then title, then id (not visual left-to-right). Index 1 is first in that order. Tab features need an existing debugging browser (`snag --open-browser`).

```
snag --list-tabs
snag -t 1
snag -t "example.com"
snag -t "https://.*\\.com"
snag --all-tabs -d out/
```

`--tab` and URL arguments are mutually exclusive.

## Browser

```
snag --open-browser          # visible browser, debug port on; no fetch
snag --kill-browser          # kill debugging browsers only
snag --doctor                # environment diagnostics
```

`--kill-browser` errors on conflicting operations; fetch modifiers are ignored.

## Logging

`--verbose` and `--debug` (mutually exclusive). No `--quiet`.

## Exit codes

- 0 success
- 1 any error
- 130 SIGINT (user interrupt, not retryable)
- 143 SIGTERM (user interrupt, not retryable)

## Skill flags

```
snag --skill
snag --skill-install [--local]
snag --skill-install=id [--skill-install=id]... [--local]
snag --skill-list [--local]
snag --skill-uninstall [--local]
snag --skill-uninstall=id [--skill-uninstall=id]... [--local]
```

- `--skill` prints this file to stdout. No agentdex. Catalog failure does not block print.
- `--skill-install` with no value: Primary path for the default agent set (Found and has a skills concept).
- `--skill-install=id`: Native if set, else Shared. Repeat the flag; a comma is part of one id. A following token is a URL positional, not an agent id (`snag --skill-install grok` is usage-class).
- Do not mix a valueless `--skill-install` / `--skill-uninstall` with `=id` on the same line.
- `--skill-list` takes no ids. Default set only.
- `--local`: project skills roots. Requires a working directory. Global (default) still runs if Getwd fails.
- Skill flags are mutually exclusive with each other and with `--doctor`, `--kill-browser`, `--list-tabs`, `--open-browser`, `--info`, `--tab`, `--all-tabs`, `--url-file`, and URL positionals. `--help` and `--version` win. `--verbose` / `--debug` work. Fetch modifiers (`--format`, `-o`, `-d`, `--timeout`, `--wait-for`, `--port`, and similar) are ignored.
- Catalog unavailable/invalid: fail closed; run `snag --skill` and place the file manually. Errors do not invent paths.

Reports (stdout; paths sorted):

- `--skill-install`: one written `…/snag/SKILL.md` path per line
- `--skill-list`: TSV `path\tagent,agent` (claimers comma-separated, sorted). Empty inventory: empty stdout, `not installed` on stderr, exit 0
- `--skill-uninstall`: TSV `removed|absent|kept\tdir`. Third column when kept (purity reason) or when multi-tenant blockers apply, including absent (`absent\tdir\tblockers`). Purity reasons: `directory is not pure (expected only SKILL.md)`, `frontmatter name is not snag`, `skill frontmatter unreadable`
