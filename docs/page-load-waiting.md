# When is a page ready to fetch?

**Status:** Current (2026-09-02)
**Code:** `internal/fetch/fetch.go`, `internal/browser/page.go`
**Flags:** [`--timeout`](arguments/timeout.md), [`--wait-for`](arguments/wait-for.md)

This document records the detexian.com hang, the research into how other tools wait, the options we rejected, and the wait policy snag uses now.

## The problem

`snag https://detexian.com/` produced nothing on stdout. The process sat in “Waiting for page to stabilize…” until killed.

The site is a Canva-exported SPA (`export_website`) with animation and a live websocket to `wss://www.canva.com/_stream`. It does not open a second tab in headless Chrome. A JavaScript `alert` was not the cause either, though `alert` is a separate hang class (see [JavaScript dialogs](#javascript-dialogs)).

HTML from curl is an empty `<div id="root"></div>` plus a large JSON bootstrap. After a few seconds in a real browser the root is hydrated and the page has hundreds of kilobytes of DOM. Markdown conversion of that DOM is a few kilobytes of real copy, ending at the site footer.

## What snag used to wait for

After `Navigate`, fetch called Rod `WaitStable`.

Rod `WaitStable(d)` waits for **all three** of:

1. `window.onload` (`WaitLoad`)
2. no matching HTTP for duration `d` (`WaitRequestIdle`)
3. the DOM snapshot unchanged for duration `d` (`WaitDOMStable` with 0% diff)

`d` is **least idle time**, not a deadline. Rod’s own comment on `WaitDOMStable`: *“Be careful, d is not the max wait timeout, it's the least stable time.”*

snag passed `StabilizeTimeout = 3s` as `d`. The constant name and older notes treated it as a 3-second timeout. It was not. On a page whose DOM never stops changing (CSS/JS animation, canvas, live widgets), step 3 never completes, so `WaitStable` never returns.

Rod `WaitRequestIdle` already ignores websocket, EventSource, media, image, and font by default. The hang on detexian was DOM freeze, not the websocket.

Rod `Navigate` only sends CDP `Page.navigate` and returns. That is not commit. `title` can still be `about:blank`. Load and hydration happen after that. Without a later wait, snag would extract the empty shell. After Navigate, wait until the frame URL leaves `about:blank`, then wait for that document's `window.onload`. Do not use Rod `WaitNavigation(Load)`: enabling lifecycle events replays `load` for `about:blank`, and matching the first load extracts a head-only shell.

## Rejected fix: dump after 3 seconds

The first patch bounded `WaitStable` at 3 seconds, then printed whatever HTML existed and exited 0.

That stops the hang. It also **cuts off a page that is still loading**. A script that blocks `onload` for 4 seconds would be missing from stdout. `--timeout` (default 30s) only covered `Navigate`, not the rest of readiness.

Chrome’s `--dump-dom --timeout` does the same “print anyway” thing. Playwright, Puppeteer, Selenium, and Crawl4AI do not: if the load condition is not met in time, they **fail**.

We do not dump a half page.

## What “loading” means

| Still loading (must wait, or fail) | Not loading (must not block forever) |
| ---------------------------------- | ------------------------------------ |
| Navigation not committed | CSS/JS animation |
| `window.onload` not fired | Open websocket / EventSource |
| Document subresources (scripts, CSS) still in flight | Analytics / heartbeat XHR after load |
| SPA hydrate XHR that started during load | DOM attributes changing on already-rendered nodes |

There is no universal “the page is fully done” event. Playwright’s navigations guide says so explicitly: after `load`, pages keep fetching and hydrating; readiness depends on the page.

snag is a fetcher, not a test runner. It does not know which button or selector means “ready” unless the user passes `--wait-for`.

## How other systems wait

Sources: official docs (snag/Kagi), Playwright / Crawl4AI / Firecrawl trees cloned under `/tmp` (2026-09-02).

| System | Default “ready” | Network idle | DOM freeze | If the clock runs out |
| ------ | ---------------- | ------------ | ---------- | --------------------- |
| Playwright `page.goto` | `waitUntil: "load"` | Exists; **DISCOURAGED** on the Page API | No | Fail |
| Puppeteer `page.goto` | `waitUntil: "load"`, default 30s | Optional `networkidle0` / `networkidle2`; hangs on sockets and polling | No | Fail |
| Selenium | `pageLoadStrategy: normal` → `document.readyState === "complete"` | No | No | Fail (page load timeout) |
| Crawl4AI | `wait_until="domcontentloaded"`, then `body` attached | No | No | Fail (`page_timeout` 60s) |
| Firecrawl (open-source path) | Playwright `domcontentloaded`, then optional extra `waitFor` ms | One replay path uses `networkidle0` | Hosted “smart wait” is closed source | Job timeout |
| Chrome `--dump-dom` | Scripts ran | `--virtual-time-budget` can exit early when nothing is pending | No | **Dump anyway** |
| Rod `WaitStable` (old snag) | `load` **and** HTTP quiet **and** DOM snapshot unchanged | Yes (long-lived types ignored) | **Yes** | No deadline unless the caller adds one |

Playwright default in `packages/playwright-core/src/server/frames.ts`: `options.waitUntil === undefined ? 'load'`. `networkidle` is 500ms with no in-flight connections in the frame subtree; the public docs tell tests not to use it and to assert a locator instead.

Puppeteer `WaitForOptions.waitUntil` default is `'load'`, `timeout` default `30000`.

Crawl4AI `CrawlerRunConfig.wait_until = "domcontentloaded"`, then `page.goto(..., timeout=page_timeout)`, optional `wait_for` (CSS or JS), then `delay_before_return_html` default **0.1s**. Their `smart_wait` is not a content heuristic: it is “CSS selector or JS predicate”.

Firecrawl `scrape-replay.ts`: `page.goto(..., { waitUntil: "domcontentloaded" })` then optional `waitForTimeout`. Engine “waterfall” in `scrapeURL/index.ts` is about switching scrape engines, not page idle.

Nobody except Rod `WaitStable` uses “the DOM stopped moving” as the default meaning of loaded. That helper is for clicking a moving button, not for “take the HTML now”.

Test tools then wait for **a specific element**. Fetch-to-markdown tools wait for a **browser load event**, optionally a selector, optionally a short extra pause.

## Options we considered

1. **Load + HTTP idle, ignore DOM freeze.** `--timeout` is the ceiling. If `load` has not fired, fail. If HTTP never quiets after load (polling), still print. `--wait-for` for pages that paint after `load`.
2. **Same, but fail unless HTTP also went idle.** Polling sites fail even though the document is in.
3. **Keep DOM freeze, bound with `--timeout`, fail.** detexian fails every time.
4. **Guess “has content”** (body text length, `#root` not empty). Wrong on empty pages, canvas-only pages, and login walls.
5. **3s dump.** Hang fixed, slow loads clipped. Rejected.

**Choice: option 1.**

Same as Playwright/Puppeteer/Selenium on `load` + fail on timeout + no DOM freeze. Extra HTTP-quiet wait after load because snag cannot default to “wait for this locator”. That extra wait is **best-effort**: it must not fail the fetch. Crawl4AI is looser (`domcontentloaded` + 0.1s sleep). We wait for full `load`, which is closer to Playwright/Puppeteer.

## Current policy

For `snag <url>` (and each URL in a batch):

1. One context with deadline `--timeout` (default 30s) covers navigate + load + HTTP idle.
2. `Navigate`.
3. Wait until the main frame URL is no longer `about:blank`, then wait for that document's `window.onload` (Rod `WaitLoad`). Do not use Rod `WaitNavigation(Load)`: it returns on about:blank's replayed `load` and fetch extracts a head-only shell (detexian.com). If the deadline hits first, **error** `page load timeout exceeded` (exit 1, nothing on stdout).
4. After load, poll until in-flight `fetch` / `XMLHttpRequest` have been quiet for `HTTPIdle` (500ms), or until the same deadline. A document hook counts those calls (installed with the dialog stubs). Websockets and other long-lived traffic are not counted. If HTTP never quiets, **log verbose and continue**.
5. If `--wait-for` is set, wait for that selector (its own `--timeout` budget, as before).
6. Extract HTML and convert.

`--tab` / `--all-tabs` do not navigate. They still honour `--wait-for` + `--timeout`. They do not run this load pipeline.

`--open-browser` with URLs only navigates; it does not wait for load or extract.

## Code map

| Piece | Where |
| ----- | ----- |
| Deadline + sequence | `internal/fetch/fetch.go` `PageFetcher.Fetch` |
| `Navigate` / `WaitLoad` / `WaitHTTPIdle` | `internal/browser/page.go` |
| 500ms quiet constant | `internal/browser/browser.go` `HTTPIdle` |
| Dialog stub and fetch/XHR idle hook | `internal/browser/page.go` `suppressJSDialogs` |

`WaitLoad` after `Navigate`: poll until the URL is not `about:blank`, then Rod `WaitLoad`. Then `WaitHTTPIdle` polls `window.__snagHTTPIdle`.

## JavaScript dialogs

`alert` / `confirm` / `prompt` block the renderer. CDP `Eval` and `HTML()` then hang even after a load timeout.

snag is a passive observer. Dialogs are dismissed (Cancel), not accepted (OK). Playwright’s default with no listener is the same: auto-dismiss.

On every wrapped page, before navigation:

- `EvalOnNewDocument` stubs `window.alert` (no-op), `confirm` (returns `false`), `prompt` (returns `null`), and wraps `fetch` / `XMLHttpRequest` so idle wait can see in-flight calls
- a CDP listener for dialogs that still open natively: **dismiss** `alert` / `confirm` / `prompt`; **accept** `beforeunload` so snag can leave the tab (close, next URL)
- one `Page.handleJavaScriptDialog` dismiss after the listener is running, because attaching does not re-fire `javascriptDialogOpening` for a dialog already on screen (`--tab`). Ignore “no dialog”. An already-open `beforeunload` is cancelled too.

A one-shot Rod `HandleDialog` loop raced: an empty event, `No dialog is showing`, then the listener stopped while a real dialog remained. Do not go back to that.

## Tests

| Test | What it guards |
| ---- | -------------- |
| `TestBrowser_NeverIdlePage` | DOM that mutates forever (`testdata/never-idle.html`) must still return content; snag must not hang |
| `TestBrowser_NeverIdleHTTP` | Short-lived HTTP that never quiets (`testdata/never-idle-http.html`) must still return content; snag must not hang |
| `TestBrowser_SlowSubresource` | A script that blocks `onload` for 4s must appear in stdout (`Loaded after delay`); the old 3s dump would miss it |
| `TestBrowser_InFlightXHR` | `fetch()` started during parse that finishes after `load` must appear in stdout |
| `TestBrowser_LoadNeverFires` | A blocking subresource that prevents `load` must error; no half page on stdout |
| `TestBrowser_StreamedHTMLBody` | HTML that flushes `<head>` then the body later must include the body; about:blank `load` plus HTTP idle would not |
| `TestBrowser_JSDialogDismissed` | `alert()` on load (`testdata/js-dialog.html`) must not block fetch |
| `TestBrowser_JSConfirmCancelled` | `confirm()` is dismissed (`testdata/js-confirm.html`); page takes the Cancel branch |
| Existing timeout / `--wait-for` tests | Load timeout still errors; selector wait still uses `--timeout` |

`TestBrowser_SlowSubresource` uses an in-test HTTP server: HTML plus `/slow.js` that sleeps 4s then writes into `#slot`.

## Limits that remain

- A SPA that fires `load`, goes HTTP-quiet, then paints 2s later can still yield a shell. Use `--wait-for`.
- Continuous XHR/`fetch` faster than 500ms can consume the rest of `--timeout` before we give up idle and print. Bounded, not infinite. Websocket-only chatter is not counted.
- Content in a `window.open` popup is not followed. detexian does not do that in headless. User can `--list-tabs` / `--tab`.
- `--wait-for` is still a separate timeout budget, not remaining time from navigate+load.

## What we are not doing

- Waiting for 0% DOM diff
- Printing on timeout
- Failing the fetch because HTTP never went idle
- A “body has text” heuristic
- Following new tabs by default

## See also

- [Playwright Page.goto `waitUntil`](https://playwright.dev/docs/api/class-page#page-goto) (`networkidle` discouraged)
- [Playwright navigations](https://playwright.dev/docs/navigations) (no general “fully loaded” event)
- [Puppeteer WaitForOptions](https://pptr.dev/api/puppeteer.waitforoptions) (default `load`, 30s)
- [Selenium pageLoadStrategy](https://www.selenium.dev/documentation/webdriver/drivers/options/#pageloadstrategy)
- [Crawl4AI page interaction](https://docs.crawl4ai.com/core/page-interaction/)
- Rod `WaitStable` / `WaitRequestIdle` in `github.com/go-rod/rod` v0.116.2 `page.go`
