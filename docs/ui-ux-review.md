# Mobile and desktop UI/UX review

## Changes

The filter values were clipped because the mobile rules added 20px of vertical padding to controls that already had a 44px height and 44px line-height. Horizontal filter groups also competed with labels sized to 100% width.

The updated single-file UI uses normal text line-height, 46px minimum control heights, labels above controls, explicit text colors, and responsive wrapping. The existing Cask Watch branding is retained.

| Area | Improvement |
| --- | --- |
| Filter layout | Closed select values remain readable; desktop inputs align; phone fields stack and tablet fields share space. |
| Map space | Map has a useful minimum height; the page scrolls on short screens; selected chips have a bounded scroll area. |
| Product picker | Native checkbox labels, named removal buttons, visible focus, Arrow Down entry, Space selection, and Escape/Done focus restoration. |
| Search feedback | Empty searches explain that no products match instead of silently hiding the picker. |
| Region changes | Product selections are reconciled with newly loaded inventory; stale responses cannot replace newer results. |
| Loading/recovery | Old markers clear while loading; unavailable inventory controls stay disabled after failures; retry restores controls and clears stale inventory errors. |
| Empty inventory | An explanatory state includes a reset action; freshness never renders Invalid Date. |
| Statistics | Store identity includes state; freshness uses loaded inventory even when filters match nothing; counts use locale formatting. |
| Readability | Higher contrast controls, larger touch targets, safe-area spacing, reduced-motion support, and a map skip link. |
| Map failures | Configuration/authentication failures show a short user-facing message and retry action. |

## Verification

- `node --test test/ui-regression.test.cjs`: 5 passing tests; no dependencies required.
- `git diff --check`: passed.
- HTML duplicate-ID and label-target checks, inline JavaScript parsing, and Python fixture-server syntax checks: passed.
- Browser viewport checks: 320×568, 390×844, 768×1024, 1024×768, and 1440×900. No horizontal document overflow in the checked layouts; phone controls measure 46px high and the map retains at least 360px height.
- Browser interactions: search/no-match feedback, checkbox selection, chip removal, region changes, empty listing counts, Arrow Down/Space/Escape, and Done returning focus to search.
- Fixture-backed testing uses the existing NC sample inventory and empty VA sample. Map calls are also covered with test doubles in the regression suite.

To reproduce locally with sample data, run `python3 test/ui-server.py 8001` and open `http://localhost:8001`. The server substitutes test inventory responses and leaves local inventory files untouched. Google Maps still uses the normal local configuration.

## Verification limits and future work

Google Maps rejected the local configuration, so live tiles and marker popups were not verified end to end. Repeat the map smoke test on an authorized origin. Viewport testing also does not substitute for a physical iPhone/Android test with the software keyboard open.

Useful future improvements, outside this focused repair, include shareable filter URLs and a searchable store list as an alternative to the map. The existing All Regions flow can also silently omit a region when just one inventory file returns an HTTP error; a partial-data notice would make that clearer.
