# Agent 调度规则

你是主 agent，负责任务拆解、选择合适的子 agent、汇总结果。

## 自动选择规则

- 需要写代码、修 bug、改配置、补测试代码：调用 [code-implementer](.claude/agents/build-migration-runner.md)
- 需要运行测试、查看测试是否通过、分析测试失败：调用 [test-runner](.claude/agents/test-runner.md)。
- 需要构建、类型检查、lint、数据库 migration、初始化数据库、seed 数据：调用 [build-migration-runner](.claude/agents/build-migration-runner.md)。

## 串行流程

当用户提出完整开发任务时，默认按以下顺序执行：

1. code-implementer 实现或修复。
2. test-runner 运行相关测试。
3. build-migration-runner 执行构建、类型检查或 migration 验证。
4. 主 agent 汇总最终结果。

## 权限边界

- test-runner 不允许修改任何文件。
- build-migration-runner 不允许修改源码文件。
- build-migration-runner 执行数据库 migration、初始化或 seed 前，必须确认目标环境是 local/development。
- 如果 test-runner 或 build-migration-runner 发现失败，需要改代码，则交回 code-implementer

# 前端组件复用规范

实现 UI 功能前，必须先复用项目已有封装，禁止重复造轮子：

1. 动手前先搜索现有实现，再决定是否新建。搜索范围：
   `frontend/src/lib/ui/`、`frontend/src/lib/components/`、
   `frontend/src/lib/components/files/`、`frontend/src/lib/services/`、
   `frontend/src/lib/utils/`、`frontend/src/lib/` 下 `*.ts`。

2. 弹窗：使用 `frontend/src/lib/dialog.ts` 提供的命令式 API
   （confirmDelete / confirmAction / promptInput / pinInput），
   底层 dialog 原语在 `lib/ui/dialog`。需要弹窗时 import 这些，不要新建 Dialog 组件。

3. 基础交互（下拉/抽屉/吐司/日期选择/工具提示等）：用 `frontend/src/lib/ui/` 下对应原语，
   不要重写。

4. 文件列表相关 UI：先查 `frontend/src/lib/components/files/`（toolbar、表格、网格、
   动作菜单、各类业务 Dialog、上传面板等）。

5. 状态管理：用 `frontend/src/lib/services/` 下的单例服务（fileManager/lockManager/
   previewManager/settingsManager），不要新起同类状态。

6. 上传链路已封装在 `frontend/src/lib/upload-*.ts` 与 `upload-manager.svelte.ts`，
   不要新写上传逻辑。

7. 仅当确认不存在可复用项时才新建；新建组件放入对应目录，遵循现有命名与 props 风格。