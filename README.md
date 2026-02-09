# OpsKit

OpsKit is a **server lifecycle operations and acceptance tool for offline / trusted-computing environments**, covering A-F stages: Preflight / Baseline / Deploy / Operate / Recover / Accept-Handover.

Chinese documentation: `README.zh-CN.md`

Current version: `v0.4.2-preview.1` (M4 deploy-template expansion preview)

## Current Capabilities (Milestone 3)

- Generic inspection workflow: stages A and D can run independently with status aggregation and severity grading
- Verifiable evidence package: `accept` generates manifest + hashes + reports + snapshots
- UI status page: reads `state/*.json` and shows overall status, stage status, and artifact entry points
- Template-driven execution: supports templates with variable rendering (builtin + external)
- Template variable validation: required/type/enum/default/example/group
- Unified command executor: `executil` is the only external command execution/audit entry
- Concurrency safety: global lock prevents concurrent runs (lock conflict returns exit code `4`)

## Out of Scope / Not Promised

- Production-grade one-click middleware deployment templates (only demo templates are included)
- Customer-specific templates, environment adaptation, and on-site scripts
- Login/account system (RBAC, user management)
- Multi-node orchestration and distributed coordination

## Usage Modes

- No template: temporary takeover / troubleshooting / acceptance patch-up; no template-delivery commitment
- Standard template (recommended): select a builtin/demo template by server role and run A/D/Accept with auditable outputs
- Custom template: for special projects; must pass `docs/product-design/09-模板设计指南.md` and `docs/product-design/DELIVERY_GATE.md`

## Quick Start (Shortest Path)

1. Build release binaries (Linux):

```bash
mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/opskit-linux-arm64 ./cmd/opskit
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/opskit-linux-amd64 ./cmd/opskit
```

2. Run the minimum local chain (A / D / Accept):

```bash
go build -o opskit ./cmd/opskit
./opskit run A --template templates/builtin/default-manage.json --output ./.tmp/opskit-demo
./opskit run D --template templates/builtin/default-manage.json --output ./.tmp/opskit-demo
./opskit accept --template templates/builtin/default-manage.json --output ./.tmp/opskit-demo
./opskit status --output ./.tmp/opskit-demo
./opskit status --output ./.tmp/opskit-demo --json
```

Machine-readable template validation (recommended before template integration):

```bash
./opskit template validate --vars-file examples/vars/demo-server-audit.json --json assets/templates/demo-server-audit.json
```

3. Start UI and view status:

```bash
./opskit web --output ./.tmp/opskit-demo --listen 127.0.0.1:18080 --status-interval 15s
```

Open in browser: `http://127.0.0.1:18080`

See also:

- `docs/deployment/DOCKER_KYLIN_V10_DEPLOY.md`
- `docs/getting-started/KYLIN_V10_OFFLINE_RELEASE.md`
- `docs/getting-started/KYLIN_V10_OFFLINE_VALIDATION.md`
- `scripts/kylin-offline-validate.sh`
- `scripts/generic-readiness-check.sh`

## Optional Validation (Kylin V10 Docker)

```bash
make -C examples/generic-manage docker-kylin-e2e
```

## Repository Layout

```text
assets/templates/          # Demo templates (de-productionized)
docs/                      # Design, specs, release docs
internal/                  # Core implementation (engine/state/plugins/...)
cmd/opskit/                # CLI entrypoint
templates/builtin/         # Builtin templates
web/ui/                    # Development UI assets
```

## Roadmap (Milestone 4-6)

- Milestone 4: template catalog expansion (ELK and other examples, template acceptance rules)
- Milestone 5: deeper Recover/Operate capability (policy-based recovery, broader generic checks)
- Milestone 6: delivery hardening (multi-format handover, template repository workflow, multi-instance research)

See `ROADMAP.md` and `docs/architecture/ROADMAP.md`.

## Documentation and Release Entry Points

- Specs and security: `docs/specs/README.md`
- Product design (consolidated): `docs/product-design/README.md`
- Template delivery gate: `docs/product-design/DELIVERY_GATE.md`
- GitHub release guide: `docs/GITHUB_RELEASE.md`
- Release planning guide: `docs/RELEASE_PLANNING_GUIDE.md`
- Current stable release notes: `docs/releases/notes/RELEASE_NOTES_v0.3.7.md`
- Current preview release notes: `docs/RELEASE_NOTES_v0.4.2-preview.1.md`
- Current preview release plan: `docs/RELEASE_PLAN_v0.4.2-preview.1.md`
- Changelog: `CHANGELOG.md`
- Security boundary: `SECURITY.md`
- License: `LICENSE` (Apache-2.0)
- Chinese README: `README.zh-CN.md`

## Pre-Release Gates (v0.4.2-preview.1)

Standard gate set:

```bash
scripts/template-validate-check.sh --clean
scripts/release-check.sh --with-offline-validate
scripts/template-delivery-check.sh --clean
```

Recommended before real server validation:

```bash
scripts/generic-readiness-check.sh --clean
```

Add `release-check summary.json` contract check:

```bash
scripts/generic-readiness-check.sh --with-release-json-contract --clean
```

Strict mode (generic + offline validation require all-zero exits):

```bash
scripts/generic-readiness-check.sh --generic-strict --offline-strict --clean
```

Strict offline gate (all offline validation exits must be `0`):

```bash
scripts/release-check.sh --with-offline-validate --offline-strict-exit
```

Build release artifacts:

```bash
scripts/release.sh --version v0.4.2-preview.1 --clean
```

Release output includes: Linux binaries (amd64/arm64), `checksums.txt`, `release-metadata.json`.

## Disclaimer

Current version (`v0.4.2-preview.1`) is intended for **generic capability validation and acceptance rehearsal in intranet/offline environments**. Production usage requires separate security, reliability, and compliance evaluation.
