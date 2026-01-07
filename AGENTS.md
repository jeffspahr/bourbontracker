# Repository Guidelines

## Project Structure & Module Organization
- `cmd/`: Go entry points (`cmd/tracker`, `cmd/alerter`, `cmd/nc-scraper`).
- `pkg/`: Core tracker logic and region implementations (e.g., `pkg/va/abc`, `pkg/nc/wake`).
- `index.html`: Single-file web UI (CSS + JS + map logic).
- `cloudflare/` and `k8s/`: Deployment docs and Kubernetes manifests.
- `config.example.js`, `subscriptions.example.json`: Local config templates; copy to `config.js`/`subscriptions.json`.
- `test/`: Sample inventory and subscription JSONs for manual comparisons.
- `screenshots/`, `logo*.svg`, `favicon.svg`: UI assets.

## Build, Test, and Development Commands
- `go build -o tracker ./cmd/tracker`: Build the tracker binary.
- `./tracker -va -wake`: Run both VA ABC and Wake County scrapers.
- `./tracker -output my-inventory.json`: Custom output file name.
- `python3 -m http.server 8000`: Serve the UI locally after generating `inventory.json`.
- `docker buildx build --platform linux/amd64,linux/arm64 -t bourbontracker .`: Build multi-arch image.
- `docker run --rm -v $(pwd):/root ghcr.io/jeffspahr/bourbontracker:latest`: Run container and emit `inventory.json`.

## Coding Style & Naming Conventions
- Go code follows standard Go conventions; format with `gofmt`.
- Package names are lowercase; exported identifiers use `CamelCase`.
- Keep UI logic in `index.html` consistent with Go normalization rules (see `pkg/tracker`).

## Testing Guidelines
- No automated test suite in this repo.
- Use `test/` JSON fixtures to diff outputs when changing tracker logic.
- Smoke test: run `./tracker` and open `index.html` via the local server to verify map filters.

## Commit & Pull Request Guidelines
- Commit messages follow Conventional Commits (e.g., `feat(ui): add filter` or `build(deps): bump ...`).
- PRs should include a concise description, linked issue (if any), and screenshots for UI changes.

## Security & Configuration Tips
- Do not commit secrets; use `config.js` and `subscriptions.json` locally.
- Google Maps API keys are injected during deployment; keep local keys out of git history.
