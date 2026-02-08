# generic-manage end-to-end demo

中文文档：`README.zh-CN.md`

This example validates the **generic (non-template-specific)** capability path
before moving to stack templates.

It runs:

1. template validate
2. install (no systemd install in local demo)
3. run AF
4. status
5. accept
6. handover

## Template validate (JSON mode)

Success example:

```bash
go run ./cmd/opskit template validate --json assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json
```

Failure example (for CI assertion):

```bash
go run ./cmd/opskit template validate --json /no/such/template.json
```

Expected failure fields:

- `valid=false`
- `errorCount>0`
- `issues[0].code=template_file_not_found`

CI gate script:

```bash
scripts/template-validate-check.sh --clean
```

## Quick start

```bash
BIN=./opskit OUTPUT=/tmp/opskit-generic ./examples/generic-manage/run-af.sh
```

Or run directly from source:

```bash
GOCACHE=/tmp/opskit-gocache BIN="go run ./cmd/opskit" OUTPUT=/tmp/opskit-generic ./examples/generic-manage/run-af.sh
```

## Dry-run mode

```bash
DRY_RUN=1 BIN=./opskit OUTPUT=/tmp/opskit-generic ./examples/generic-manage/run-af.sh
```

Strict exit mode (requires stage commands to return `exit=0`):

```bash
STRICT_EXIT=1 BIN=./opskit OUTPUT=/tmp/opskit-generic ./examples/generic-manage/run-af.sh
```

## Expected outputs

- `OUTPUT/state/overall.json`
- `OUTPUT/state/lifecycle.json`
- `OUTPUT/state/artifacts.json`
- `OUTPUT/status.json`
- `OUTPUT/summary.json`
- `OUTPUT/reports/*.html`
- `OUTPUT/bundles/*.tar.gz`
- `OUTPUT/ui/index.html`

Notes:

- In non-Linux or minimal environments, some checks may return WARN/FAILED.
- This is expected for generic capability validation; the script still prints
  output paths for inspection.
- `summary.json` now includes `result/reasonCode/recommendedAction` for gate decisions.

## Run in Kylin V10 Docker (clean runtime)

Use a clean Kylin V10 container to verify OpsKit generic capability flow:

See deployment guide: `docs/deployment/DOCKER_KYLIN_V10_DEPLOY.md`

```bash
OUTPUT=$PWD/.tmp/generic-e2e-kylin \
  ./examples/generic-manage/run-af-kylin-v10-docker.sh
```

Or via Make:

```bash
make -C examples/generic-manage docker-kylin-e2e
```

What this wrapper does:

- builds `opskit` for Linux (`amd64` or `arm64`, based on image arch)
- starts `kylinv10/kylin:b09`
- mounts host output dir to `/out`
- mounts dedicated host dirs to `/opt`, `/data`, `/logs` for mount checks
- runs the normal `run-af.sh` flow in container

You can override image/platform:

```bash
IMAGE=kylinv10/kylin:b09 \
DOCKER_PLATFORM=linux/amd64 \
OUTPUT=$PWD/.tmp/generic-e2e-kylin-amd64 \
./examples/generic-manage/run-af-kylin-v10-docker.sh
```

Dry-run in container:

```bash
DRY_RUN=1 make -C examples/generic-manage docker-kylin-e2e
```
