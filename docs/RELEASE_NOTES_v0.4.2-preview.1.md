# OpsKit v0.4.2-preview.1（草稿）

> 预览版本（进行中）。主线聚焦 M4：在设计冻结约束内扩展 deploy 模板能力。

## Planned

### Added

- 新增发布计划：`docs/RELEASE_PLAN_v0.4.2-preview.1.md`
- 新增当前预览发布说明：`docs/RELEASE_NOTES_v0.4.2-preview.1.md`

### Changed

- 根目录当前 release 文档入口切换到 `v0.4.2-preview.1`
- `v0.4.1-preview.1` 计划与说明归档到 `docs/releases/`
- 文档入口（README/ROADMAP/docs index/GITHUB_RELEASE）同步滚动到新预览版

## Focus

- 通用 deploy-manage 模板骨架收敛
- MinIO/ELK/PowerJob 模板可交付性增强
- 模板门禁与发布流程继续收敛

## Notes

- 仍遵循 `v0.4.x` 设计冻结：不新增阶段、不改 `executil` 语义。
- 模板必须满足 `DELIVERY_GATE`，失败场景也要可交付。

## Recommended Verification

```bash
go test ./...
scripts/release-check.sh --with-offline-validate
scripts/template-delivery-check.sh --clean
```
