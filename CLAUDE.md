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