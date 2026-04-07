# Contributing to Chinwag

Thanks for your interest! Chinwag is a small open-source project — bug
reports, ideas and pull requests are all welcome.

## Reporting bugs / requesting features

Please use the issue forms:

- 🐛 [**Bug report**](https://github.com/seanmcn/chinwag/issues/new?template=bug_report.yml)
- ✨ [**Feature request**](https://github.com/seanmcn/chinwag/issues/new?template=feature_request.yml)

When filing a bug, **never paste real chat content**. The forms ask you to
confirm this — please redact names, numbers, and any text from the
conversation before sharing logs or screenshots.

## Setup

You'll need:

- **Go 1.26+** — for the parser, analysis engine, CLI, and Wails backend.
- **Node 20+** — for the React frontend.
- **Wails CLI** — only needed if you're touching the desktop app:
  ```sh
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
  On macOS you also need Xcode Command Line Tools (`xcode-select --install`).
  See <https://wails.io/docs/gettingstarted/installation> for Linux/Windows
  prerequisites.

Then:

```sh
git clone https://github.com/seanmcn/chinwag
cd chinwag
go test ./...                    # backend sanity check
cd app/frontend && npm ci        # frontend deps
```

## Dev loop

**CLI** — fast inner loop for parser/analyse changes:

```sh
go run ./cmd/chinwag testdata/sample_chat.txt
go run ./cmd/chinwag --format json testdata/sample_chat.txt | jq .Messages
```

**Desktop app** — live-reloads frontend, restarts Go on save:

```sh
cd app && wails dev
```

**Tests**:

```sh
go test ./...
```

**Docker (CLI)** — the repo ships a `Dockerfile` for the CLI. The README
points users at the prebuilt `ghcr.io/seanmcn/chinwag` image; to build and
run it locally:

```sh
docker build -t chinwag .
docker run --rm -v "$PWD/data:/data" chinwag /data/chat.zip
```

## Project layout

```
cmd/chinwag/            CLI entrypoint (text/JSON output)
internal/parser/        WhatsApp export parser (iOS, Android, .zip)
internal/analyse/       Stats, conversations, insights, rating, sentiment
app/                    Wails desktop app (Go backend + React/TS frontend)
app/app.go              IPC methods bound to the frontend
app/frontend/src/       React components, charts, format helpers
.github/workflows/      CI and tagged release pipelines
testdata/               Tiny sample chat for tests
```

## Code style

- **Go**: `gofmt`, idiomatic standard library only — the parser and analyse
  packages have zero external dependencies and we'd like to keep it that way.
- **Frontend**: TypeScript, React function components with hooks. The dark
  palette and component classes live in `app/frontend/src/App.css`; reuse the
  helpers in `app/frontend/src/format.ts` (`comma`, `fmtSec`, `fmtHour`,
  `cmpA/cmpB`, etc.) instead of reimplementing them.
- **No new dependencies** without a good reason — especially in the frontend.

## Commits

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(app): add multi-chat switcher
fix(app): reset scroll on tab change
docs: rewrite README around desktop app
ci: drop Windows signing placeholder
chore: bump frontend deps
feat(cli)!: drop --addr flag    # ! marks a breaking change
```

Common scopes: `app` (Wails / frontend), `cli`, `analyse`, `parser`, `ci`,
`docs`. Scope is optional but appreciated when the change is localised.

Write the body in the imperative ("add X", not "added X"), explain the *why*
not the *what*, and keep subject lines under ~72 chars.

## Pull requests

- Branch from `main`.
- Keep PRs small and focused — one logical change per PR.
- Make sure CI is green before requesting review:
  - `go test ./...`
  - `cd app/frontend && npm run build`
- Fill in the test plan in the PR template so the reviewer knows what you
  exercised manually (especially for UI work).
- If you've changed the dashboard, a screenshot in the PR body is gold.

## Releases

Releases are cut by pushing a `v*` tag:

```sh
git tag v0.4.0
git push origin v0.4.0
```

This triggers `.github/workflows/release.yml` which builds the macOS app
(signed + notarised), the Windows app (unsigned for now), and standalone
CLI binaries for Linux/macOS/Windows, then attaches them to a GitHub
release.

## License

Chinwag is MIT licensed. By contributing you agree your contributions will
be released under the same license.
