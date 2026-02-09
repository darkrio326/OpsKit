# SPEC: `opskit template list --json`

## 1. 目标

定义 `opskit template list --json` 的机器可读输出，供 CI/脚本做模板可发现性检查。
输出中的 `templates[].ref` 可直接用于 `opskit template validate <ref>`。

## 2. 命令

```bash
opskit template list --json
```

成功退出码：`0`  
失败退出码：`1`（执行失败）或 `2`（参数错误）

## 3. JSON 契约（v1）

```json
{
  "command": "opskit template list",
  "schemaVersion": "v1",
  "count": 4,
  "templates": [
    {
      "ref": "generic-manage-v1",
      "source": "builtin/default-manage.json",
      "templateId": "generic-manage-v1",
      "name": "Generic Manage v1",
      "mode": "manage",
      "serviceScope": "single-service",
      "tags": ["manage", "single-service", "builtin"]
    },
    {
      "ref": "single-service-deploy-v1",
      "aliases": ["single-service-deploy"],
      "source": "builtin/single-service-deploy.json",
      "templateId": "single-service-deploy-v1",
      "name": "Single Service Deploy v1",
      "mode": "deploy",
      "serviceScope": "single-service",
      "tags": ["deploy", "single-service", "builtin"]
    },
    {
      "ref": "demo-elk-deploy",
      "source": "assets/templates/demo-elk-deploy.json",
      "templateId": "demo-elk-deploy-v1",
      "name": "Demo ELK Deploy v1",
      "mode": "deploy",
      "serviceScope": "multi-service",
      "tags": ["deploy", "multi-service", "demo"]
    }
  ]
}
```

## 4. 字段说明

- `command`: 固定为 `opskit template list`
- `schemaVersion`: 当前固定 `v1`
- `count`: `templates` 数组元素个数
- `templates[].ref`: 模板引用名（CLI 可用）
- `templates[].aliases`: 可选别名列表
- `templates[].source`: 内置模板来源文件
- `templates[].source`: 模板来源（如 `builtin/...` 或 `assets/templates/...`）
- `templates[].templateId`: 模板 JSON 中的 `id`
- `templates[].name`: 模板名称
- `templates[].mode`: 模板模式（`manage|deploy`）
- `templates[].serviceScope`: 服务范围标签（如 `single-service|multi-service`）
- `templates[].tags`: 分类标签列表（建议至少包含 `mode` 和 `serviceScope`）

备注：

- 结果至少包含 builtin 模板
- 若检测到本地 `assets/templates/*.json`，会一并加入列表

## 5. 稳定性要求

- v1 阶段仅追加字段，不移除已发布字段
- `count` 必须与 `templates` 数组长度一致
- `ref` 需唯一，避免脚本歧义
