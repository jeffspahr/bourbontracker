# Refactor Roadmap

## Goals
- Decouple inventory fetching from frontend deployment.
- Introduce efficient inventory history without storing full snapshots every run.
- Move auth into the app (OAuth) and migrate hosting to Cloudflare Workers.
- Support user-specific preferences and notifications.

## Phase 1: Decouple Inventory Fetching and Frontend Deploy
- Create a dedicated inventory refresh workflow (scheduled + on-demand).
- Keep frontend deploy workflow focused on static assets and latest inventory artifacts.
- Preserve current outputs (`inventory-va.json`, `inventory-nc.json`) and alerting flow.

## Phase 2: Incremental Inventory History
- Add a snapshot writer that stores daily inventory and optional deltas.
- Use object storage for raw history and a lightweight index for metadata.
- Prototype a minimal UI view for time-based changes (day-to-day diff, trend line).

## Phase 3: Auth and Hosting Migration
- Move OAuth into the frontend/API boundary; remove Cloudflare Access dependency.
- Migrate from Pages to Workers with API endpoints for inventory and history.

## Phase 4: User-Aware Features
- Introduce user profiles, preferences, and per-user filters.
- Extend alerts to honor user-specific rules and inventory interests.

## Out of Scope (for now)
- Changing tracker logic or scraping scope.
- Rebuilding the frontend UI.
