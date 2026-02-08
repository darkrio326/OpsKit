# demo-runtime-baseline

`demo-runtime-baseline.json` is a de-productionized template for baseline runtime checks.
It focuses on generic server readiness and evidence output, without middleware deployment.

## Scope

- Lifecycle stages: `A` / `D` / `F`
- Check classes: system info, mount, time sync, DNS, disk/memory/load
- Evidence classes: file hash, directory hash, command output

## Variables

- `INSTALL_ROOT` (`group=paths`, required): output root (state/reports)
- `EVIDENCE_DIR` (`group=paths`, required): evidence output directory
- `ROOT_MOUNT` (`group=runtime`, default=`/`): mount path for disk checks
- `DNS_HOST` (`group=network`, default=`localhost`): hostname for DNS resolve check
- `PROFILE` (`group=runtime`, default=`baseline`): label used in evidence filenames

## Validate

```bash
go run ./cmd/opskit template validate --vars-file examples/vars/demo-runtime-baseline.json --json assets/templates/demo-runtime-baseline.json
```

## Run (optional)

```bash
go run ./cmd/opskit run A --template assets/templates/demo-runtime-baseline.json --vars-file examples/vars/demo-runtime-baseline.json --output /tmp/opskit-demo-runtime
go run ./cmd/opskit run D --template assets/templates/demo-runtime-baseline.json --vars-file examples/vars/demo-runtime-baseline.json --output /tmp/opskit-demo-runtime
go run ./cmd/opskit accept --template assets/templates/demo-runtime-baseline.json --vars-file examples/vars/demo-runtime-baseline.json --output /tmp/opskit-demo-runtime
```

Expected output:

- `/tmp/opskit-demo-runtime/state/*.json`
- `/tmp/opskit-demo-runtime/reports/*.html`
- `/tmp/opskit-demo-runtime/evidence/*-hash.json`
