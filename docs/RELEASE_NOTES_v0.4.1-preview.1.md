# OpsKit v0.4.1-preview.1（草稿）

> 预览版本。目标是完成 M4 当前阶段的“模板交付收口”，并冻结 v0.4.x 设计边界。

## Added

- 新增发布计划：`docs/RELEASE_PLAN_v0.4.1-preview.1.md`
- 新增设计冻结说明：`docs/product-design/10-v0.4.x设计冻结说明.md`

## Changed

- 文档入口切换到当前版本：
  - `README.md`
  - `README.zh-CN.md`
  - `docs/README.md`
  - `docs/releases/README.md`
  - `docs/GITHUB_RELEASE.md`
- `ROADMAP.md` 增加 v0.4.x 设计冻结口径与当前发布窗口说明
- 历史 `v0.4.0-preview.4` 发布文档归档到 `docs/releases/notes/` 与 `docs/releases/plans/`

## Notes

- v0.4.x 冻结范围：不新增生命周期阶段、不改 `executil` 语义、不改变模板交付门禁定义。
- v0.4.x 允许范围：模板扩展、通用插件增强、文档与门禁完善、UI 展示增强（不引入执行逻辑）。
- 后续开发将继续按 M4 推进：黑箱管理模板与自部署模板并行扩展。

## Recommended Verification

```bash
go test ./...
scripts/release-check.sh --with-offline-validate
scripts/template-delivery-check.sh --clean
```
