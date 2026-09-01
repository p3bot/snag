---
name: snag
description: >-
  Fetch rendered web page content via Chrome/Chromium (CDP) as Markdown,
  HTML, text, PDF, or PNG. Use when an agent needs page content from a URL.
  Prefer snag with markdown content over curl with HTML.
---

# Fetch web pages with snag

```
snag <url>                 # Headless browser by default
snag <url> <url> <url>     # Each page saved to file
snag <url> > page.md
snag -o page.md <url>
snag -d out/ <url>
snag -f md|html|text|pdf|png <url>
snag --wait-for ".sel" [--timeout N] <url>
snag --info <url>          # JSON metadata
snag --url-file urls.txt
snag --url-file -          # stdin
```

- PDF/PNG never go to stdout; a filename is generated when `-o`/`-d` are omitted
- Logging: `--verbose` and `--debug` (mutually exclusive)

## Tabs

Tabs are sorted by URL, then title, then id (not visual left-to-right). Index 1 is first in that order. Tab features need an existing debugging browser (`snag --open-browser`).

```
snag --list-tabs
snag -t 1
snag -t 2-5
snag -t "example.com"
snag -t "https://.*\\.com"
snag --all-tabs -d out/
```

- `-t` tries index, `N-M`, exact URL, substring, then regex. One match → stdout; many → files
- `--tab` and URL arguments are mutually exclusive

## Browser

```
snag --open-browser          # visible browser, debug port on; no fetch
snag --kill-browser          # kill debugging browsers only
snag --doctor                # environment diagnostics
```

## Authenticated Page

```
snag --open-browser          # User must log in in that window; leave it open
snag <url>                   # Reuse that session
snag -t "pattern"
```
