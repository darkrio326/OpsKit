# OpsKit UI Screenshots

This folder stores release-friendly UI screenshots.

## Directory layout

- `latest/`: screenshot slots currently used by docs
- `releases/<version>/`: snapshot of screenshot slots for each release

Required slots:

- `ui-template-stage.png` (template selector + A-F cards)
- `ui-dashboard-evidence.png` (dashboard + evidence entry points)

## Update workflow

1. Replace files under `latest/` with real screenshots.
2. Run:
   - `scripts/screenshot-sync.sh --version <version>`
   - `scripts/screenshot-check.sh --version <version>`
3. Commit `latest/` and `releases/<version>/`.

## Redaction / safety

- Do not include customer IPs, hostnames, usernames, tokens, or secrets.
- Keep screenshots from demo or sanitized environments only.
