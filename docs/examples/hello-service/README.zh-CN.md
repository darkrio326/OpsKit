# hello-service 离线包示例

该目录用于验证 deploy 模板链路：
`sha256_verify -> untar -> render unit -> install unit -> daemon-reload -> enable/start`。

## 1) 构建离线 tar.gz 与 SHA256

```bash
OUT=/var/lib/opskit
mkdir -p "$OUT/packages"
cp examples/hello-service/hello-service.sh "$OUT/packages/hello-service.sh"
COPYFILE_DISABLE=1 tar -C "$OUT/packages" -czf "$OUT/packages/hello-service.tar.gz" hello-service.sh
SHA=$(sha256sum "$OUT/packages/hello-service.tar.gz" | awk '{print $1}')
echo "$SHA"
```

## 2) 运行 deploy 模板（真实 systemd 主机）

```bash
opskit run C \
  --template single-service-deploy \
  --output /var/lib/opskit \
  --vars "PACKAGE_SHA256=${SHA},SERVICE_NAME=hello-service,SERVICE_UNIT=hello-service.service,SERVICE_PORT=18080,SERVICE_EXEC=/var/lib/opskit/hello-service/release/hello-service.sh,SYSTEMD_UNIT_DIR=/etc/systemd/system"
```

## 3) 验证

```bash
systemctl status hello-service.service
ss -ltn | grep 18080
opskit run D --template single-service-deploy --output /var/lib/opskit
opskit run E --template single-service-deploy --output /var/lib/opskit
opskit accept --template single-service-deploy --output /var/lib/opskit
opskit handover --output /var/lib/opskit
```

## 本地非 root dry-run 方式（不操作真实 systemd）

使用 `--dry-run`，或把 `SYSTEMD_UNIT_DIR` 指向可写目录。
