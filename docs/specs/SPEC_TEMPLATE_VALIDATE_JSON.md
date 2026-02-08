# OpsKit `template validate --json` 输出契约规范

## 1. 目标与范围

本规范定义 `opskit template validate --json` 的机器可读输出契约，供：

- CI 门禁
- 自动化模板验收脚本
- 模板库发布流程

使用范围仅限 `template validate --json` 的 `stdout` JSON。

## 2. 调用约定

示例：

```bash
opskit template validate --vars-file examples/vars/demo-server-audit.json --json assets/templates/demo-server-audit.json
```

行为约定：

- `stdout` 输出 JSON
- 校验成功退出码为 `0`
- 校验失败退出码为 `2`（前置条件不满足）
- 调用方应同时读取 `stdout` 与进程退出码

## 3. 顶层字段（`schemaVersion = v1`）

字段：

- `command`：固定 `opskit template validate`
- `schemaVersion`：当前固定 `v1`
- `template`：本次校验输入的模板引用（ID 或文件路径）
- `valid`：是否通过校验（`true|false`）
- `errorCount`：错误总数（当前实现固定为 `0` 或 `1`）
- `issues`：错误列表

`issues[]` 字段：

- `path`：错误定位路径（如 `template.vars.ENV`、`template.stages.A.checks[0].params.hostname`）
- `code`：稳定错误码（如 `template_var_required`、`template_unresolved_var`）
- `message`：原始错误描述
- `advice`：修复建议

## 3.1 错误码清单（v1）

当前实现已使用的稳定错误码：

- `template_invalid`：未归类的模板校验失败
- `template_unknown_id`：模板 ID 未注册
- `vars_file_invalid`：`--vars-file` 解析失败
- `template_file_not_found`：模板文件不存在
- `template_file_permission_denied`：模板文件不可读
- `template_unknown_field`：模板 JSON 含未知字段
- `template_json_trailing_content`：模板 JSON 存在多余内容
- `template_unresolved_var`：模板阶段参数存在未解析变量
- `template_stage_invalid`：阶段 ID 非 `A..F`
- `template_severity_invalid`：`params.severity` 非法
- `template_var_invalid`：变量定义不合法（默认/枚举/分组等）
- `template_var_required`：缺少必填变量
- `template_var_type_mismatch`：变量类型不匹配
- `template_var_enum_mismatch`：变量值不在枚举中

兼容策略：

- `v1` 内允许新增错误码，不删除既有错误码语义
- 调用方应基于已知错误码做分类，未知错误码归入 `template_invalid` 类别处理

## 4. 示例

成功：

```json
{
  "command": "opskit template validate",
  "schemaVersion": "v1",
  "template": "assets/templates/demo-server-audit.json",
  "valid": true,
  "errorCount": 0,
  "issues": []
}
```

失败：

```json
{
  "command": "opskit template validate",
  "schemaVersion": "v1",
  "template": "/no/such/template.json",
  "valid": false,
  "errorCount": 1,
  "issues": [
    {
      "path": "template.file",
      "code": "template_file_not_found",
      "message": "open /no/such/template.json: no such file or directory",
      "advice": "check template path and file permissions"
    }
  ]
}
```

## 5. 兼容性策略

- `v1` 内仅允许追加字段，不改变现有字段语义
- 错误码可以扩展；现有错误码语义不得反向改变
- 调用方应忽略未知字段，避免因新增字段导致解析失败

## 6. 调用方建议

- 以进程退出码作为主判定（`0` 通过，非 `0` 失败）
- 同时校验 `valid` 与 `errorCount`，做交叉检查
- 对 `issues[].code` 进行分组统计，跟踪模板质量趋势
