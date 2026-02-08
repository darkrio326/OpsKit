# hello-service offline package demo

Chinese doc: `README.zh-CN.md`

This folder helps verify the deploy template flow:
`sha256_verify -> untar -> render unit -> install unit -> daemon-reload -> enable/start`.

## Template validate (JSON mode)

Success example:

```bash
go run ./cmd/opskit template validate --json templates/builtin/single-service-deploy.json
```

Failure example:

```bash
go run ./cmd/opskit template validate --json /no/such/template.json
```

Expected failure fields:

- `valid=false`
- `errorCount>0`
- `issues[0].code=template_file_not_found`

## 1) Build offline tar.gz and SHA256

```bash
OUT=/var/lib/opskit
mkdir -p "$OUT/packages"
cp examples/hello-service/hello-service.sh "$OUT/packages/hello-service.sh"
COPYFILE_DISABLE=1 tar -C "$OUT/packages" -czf "$OUT/packages/hello-service.tar.gz" hello-service.sh
SHA=$(sha256sum "$OUT/packages/hello-service.tar.gz" | awk '{print $1}')
echo "$SHA"
```

## 2) Run deploy template (real systemd host)

```bash
opskit run C \
  --template single-service-deploy \
  --output /var/lib/opskit \
  --vars "PACKAGE_SHA256=${SHA},SERVICE_NAME=hello-service,SERVICE_UNIT=hello-service.service,SERVICE_PORT=18080,SERVICE_EXEC=/var/lib/opskit/hello-service/release/hello-service.sh,SYSTEMD_UNIT_DIR=/etc/systemd/system"
```

## 3) Verify

```bash
systemctl status hello-service.service
ss -ltn | grep 18080
opskit run D --template single-service-deploy --output /var/lib/opskit
opskit run E --template single-service-deploy --output /var/lib/opskit
opskit accept --template single-service-deploy --output /var/lib/opskit
opskit handover --output /var/lib/opskit
```

## Local non-root dry-run style (no real systemd)

Use `--dry-run` or point `SYSTEMD_UNIT_DIR` to a writable folder.
