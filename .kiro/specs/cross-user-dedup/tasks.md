# Implementation Plan: Cross-User File Deduplication

## Overview

按依赖顺序实现跨用户文件去重功能：先建立数据库基础（迁移 + sqlc 查询），再实现核心 `DedupService`，然后修改各服务层（`UploadService`、`UploadSessionService`、`TaskService`），接着修改 Worker（`FFmpegProcessor`），最后更新 Handler 和依赖注入（`app.go`）。

## Tasks

- [x] 1. 数据库迁移与 sqlc 查询定义
  - [x] 1.1 创建迁移文件 `000002_cross_user_dedup.up.sql`
    - 新建 `file_store` 表：`id`（BIGSERIAL PK）、`sha256_hash`（VARCHAR(64) UNIQUE NOT NULL）、`storage_path`（TEXT NOT NULL）、`hls_output_dir`（TEXT）、`ref_count`（INTEGER NOT NULL DEFAULT 1 CHECK (ref_count >= 0)）、`created_at`（TIMESTAMPTZ NOT NULL DEFAULT NOW()）
    - 在 `file_store(sha256_hash)` 上创建索引 `idx_file_store_sha256`
    - 在 `tasks` 表上新增 `file_store_id`（BIGINT REFERENCES file_store(id) ON DELETE SET NULL）和 `sha256_hash`（VARCHAR(64)）列
    - 在 `tasks(user_id, sha256_hash)` 上创建复合索引 `idx_tasks_user_sha256`
    - _Requirements: 1.1, 1.2, 1.6, 10.3_

  - [x] 1.2 创建回滚迁移文件 `000002_cross_user_dedup.down.sql`
    - 删除 `tasks.file_store_id`、`tasks.sha256_hash` 列
    - 删除 `file_store` 表
    - _Requirements: 1.1_

  - [x] 1.3 创建 `backend/internal/db/query/file_store.sql` 查询文件
    - 实现 `UpsertFileStore`（`INSERT ... ON CONFLICT DO UPDATE SET ref_count = ref_count + 1 RETURNING *`）
    - 实现 `GetFileStoreForUpdate`（`SELECT * FROM file_store WHERE id=$1 FOR UPDATE`）
    - 实现 `DecrFileStoreRefCount`（`UPDATE ... SET ref_count = ref_count - 1 RETURNING *`）
    - 实现 `DeleteFileStore`（`DELETE FROM file_store WHERE id=$1`）
    - 实现 `UpdateFileStoreHLSDir`（`UPDATE file_store SET hls_output_dir=$2 WHERE id=$1`）
    - 实现 `GetFileStoreByTaskID`（通过 JOIN tasks 获取关联的 file_store 行）
    - _Requirements: 1.3, 1.4, 6.1, 9.1_

  - [x] 1.4 在 `backend/internal/db/query/tasks.sql` 中新增三个查询
    - 实现 `CreateTaskWithDedup`：INSERT 包含 `file_store_id`、`sha256_hash`、`status`、`output_dir`、`m3u8_url`、`duration_sec` 字段
    - 实现 `GetTaskByUserHash`：`SELECT slug, status FROM tasks WHERE user_id=$1 AND sha256_hash=$2 LIMIT 1`（仅查当前用户，不 JOIN file_store）
    - 实现 `UpdatePendingTasksCompleted`：批量将同一 `file_store_id` 下的 `pending` tasks 更新为 `completed`，排除原始 task
    - _Requirements: 3.2, 3.3, 6.2, 8.1, 8.4, 10.1, 10.2, 10.4_

  - [x] 1.5 运行 sqlc generate，更新 `backend/internal/db/sqlc/` 下的生成文件
    - 确认 `models.go` 中 `Task` 新增 `FileStoreID pgtype.Int8` 和 `Sha256Hash pgtype.Text` 字段
    - 确认新增 `FileStore` 结构体
    - 确认 `querier.go` 和各 `*.sql.go` 文件包含所有新查询
    - _Requirements: 1.1, 1.6_

- [x] 2. 实现 DedupService
  - [x] 2.1 创建 `backend/internal/service/dedup.go`，实现 `DedupService` 结构体和构造函数
    - 定义 `DedupService` 结构体（持有 `*pgxpool.Pool` 和 `zerolog.Logger`）
    - 定义 `FindOrCreateParams`、`FindOrCreateResult` 类型
    - 实现 `NewDedupService(pg *pgxpool.Pool, logger zerolog.Logger) *DedupService`
    - _Requirements: 9.1, 9.2_

  - [x] 2.2 实现 `DedupService.FindOrCreate` 方法
    - 开启事务，调用 `UpsertFileStore`（`ON CONFLICT DO UPDATE`）原子 upsert
    - 根据 `fs.RefCount > 1` 判断是否去重，根据 `fs.HlsOutputDir.Valid` 判断是否已完成
    - 若已完成（`alreadyDone=true`）：task status 设为 `completed`，填充 `output_dir`、`m3u8_url`、`duration_sec`
    - 若进行中（`deduped=true, alreadyDone=false`）：task status 设为 `pending`，`input_path` 使用 `fs.StoragePath`
    - 调用 `CreateTaskWithDedup` 插入 task 行，提交事务
    - _Requirements: 3.1, 3.2, 3.3, 4.1, 4.2, 4.3, 9.1, 9.2_

  - [x] 2.3 实现 `DedupService.DecrRef` 方法
    - 接受 `pgx.Tx` 参数（调用方控制事务边界）
    - 调用 `GetFileStoreForUpdate` 加锁行（防止并发删除+上传竞态）
    - 调用 `DecrFileStoreRefCount` 递减
    - 若 `ref_count == 0`：调用 `DeleteFileStore`，返回 `storagePath`、`hlsOutputDir`、`deleted=true`
    - 若 `ref_count > 0`：返回 `deleted=false`
    - _Requirements: 1.4, 1.5, 5.1, 5.2, 9.3_

  - [x] 2.4 实现 `DedupService.FindByUserHash` 方法
    - 调用 `GetTaskByUserHash`（仅查当前用户 tasks，不查 file_store）
    - 返回 `(sqlc.GetTaskByUserHashRow, bool, error)`
    - _Requirements: 8.1, 8.4, 10.4_

  - [ ]* 2.5 为 DedupService 编写属性测试（Property 1：SHA-256 哈希计算正确性）
    - **Property 1: SHA-256 哈希计算正确性**
    - 使用 `pgregory.net/rapid`，对任意字节序列验证 `streamToFileWithHash` 计算结果等于 `sha256.Sum256`
    - **Validates: Requirements 2.1, 2.2, 10.1, 10.2**

  - [ ]* 2.6 为 DedupService 编写属性测试（Property 2：引用计数并发安全）
    - **Property 2: 引用计数单调递增（并发安全）**
    - 使用 `pgregory.net/rapid` + testcontainers-go，N 个并发 upsert 相同 hash 后验证 `ref_count == N`
    - **Validates: Requirements 1.3, 9.1**

  - [ ]* 2.7 为 DedupService 编写属性测试（Property 3：引用计数递减原子性）
    - **Property 3: 引用计数递减与任务删除原子性**
    - 创建 `ref_count=N` 的 file_store 行，删除 N 个 task 后验证 file_store 行不存在
    - **Validates: Requirements 1.4, 5.1, 5.2, 5.5**

  - [ ]* 2.8 为 DedupService 编写属性测试（Property 7：Per-User 哈希查询数据隔离）
    - **Property 7: Per-User 哈希查询数据隔离**
    - 为用户 B 创建 hash H 的 task，查询用户 A 的 per-user hash 验证结果为空
    - **Validates: Requirements 8.1, 8.4, 10.4**

- [x] 3. Checkpoint — 确保 DedupService 核心逻辑通过测试
  - 确保所有测试通过，如有疑问请向用户确认。

- [x] 4. 修改 UploadService
  - [x] 4.1 在 `backend/internal/service/upload.go` 中添加 `streamToFileWithHash` 函数
    - 使用 `io.MultiWriter(dst, h)` 单次 pass 同时写文件和计算 SHA-256
    - 返回 `(hash string, n int64, err error)`
    - 处理超出 `maxBytes` 的情况（删除临时文件，返回 `fileutil.ErrFileTooLarge`）
    - _Requirements: 2.1, 2.3, 2.4_

  - [x] 4.2 修改 `UploadService` 结构体，注入 `*pgxpool.Pool` 和 `*DedupService`
    - 更新 `NewUploadService` 签名：新增 `pg *pgxpool.Pool` 和 `dedup *DedupService` 参数
    - 更新 `UploadResult` 结构体：新增 `Status string` 字段（供 handler 内部使用）
    - _Requirements: 3.1, 3.4_

  - [x] 4.3 重写 `UploadService.Upload` 方法，集成去重逻辑
    - 将 `streamToFile` 替换为 `streamToFileWithHash`
    - 调用 `s.dedup.FindOrCreate` 替换原有 `s.queries.CreateTask`
    - 若 `res.Deduped == true`：删除刚写入的临时文件（`cleanup()`）
    - 若 `res.AlreadyDone == false`：提交 Worker 任务（`input_path` 使用 `res.FileStore.StoragePath`）
    - 若 `res.AlreadyDone == true`：不提交 Worker 任务
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 7.1, 7.3, 7.4_

  - [ ]* 4.4 为 UploadService 编写属性测试（Property 4：去重任务继承已完成状态）
    - **Property 4: 去重任务继承已完成状态**
    - 对 `hls_output_dir` 非空的 file_store 行，验证新 task 的 status 为 `completed` 且字段一致
    - **Validates: Requirements 3.2, 4.2**

  - [ ]* 4.5 为 UploadService 编写属性测试（Property 5：去重任务继承进行中状态）
    - **Property 5: 去重任务继承进行中状态**
    - 对 `hls_output_dir` 为空的 file_store 行，验证新 task 的 status 为 `pending` 且 `input_path == file_store.storage_path`
    - **Validates: Requirements 3.3, 4.3**

- [x] 5. 修改 UploadSessionService
  - [x] 5.1 修改 `UploadSessionService` 结构体，注入 `*pgxpool.Pool` 和 `*DedupService`
    - 更新 `NewUploadSessionService` 签名：新增 `pg *pgxpool.Pool` 和 `dedup *DedupService` 参数
    - _Requirements: 4.1, 4.4_

  - [x] 5.2 重写 `UploadSessionService.Complete` 方法，集成去重逻辑
    - 在 `os.Rename` 之前计算已组装文件的 SHA-256（使用 `streamToFileWithHash` 或独立哈希计算函数）
    - 调用 `s.dedup.FindOrCreate` 替换原有 `s.queries.CreateTask`
    - 若 `res.Deduped == true`：删除临时文件（`os.RemoveAll(sessionDir)`）
    - 若 `res.Deduped == false`：将临时文件 rename 到 canonical path（`file_store.StoragePath`）
    - 若 `res.AlreadyDone == false`：提交 Worker 任务
    - 若 `res.AlreadyDone == true`：不提交 Worker 任务
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 2.2, 2.5_

  - [ ]* 5.3 为 UploadSessionService.Complete 编写单元测试
    - 覆盖：新哈希（rename + submit job）、已完成去重（删除临时文件 + 不 submit）、进行中去重（删除临时文件 + submit）
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [x] 6. 修改 TaskService
  - [x] 6.1 修改 `TaskService` 结构体，注入 `*pgxpool.Pool`、`*DedupService` 和 `zerolog.Logger`
    - 更新 `NewTaskService` 签名：新增 `pg *pgxpool.Pool`、`dedup *DedupService`、`logger zerolog.Logger` 参数
    - _Requirements: 5.1, 5.5_

  - [x] 6.2 重写 `TaskService.DeleteForUser` 方法，实现事务性删除与引用计数递减
    - 开启事务，在事务内调用 `DeleteTaskBySlugForUser`
    - 若 `t.FileStoreID.Valid`：在同一事务内调用 `s.dedup.DecrRef(ctx, tx, t.FileStoreID.Int64)`
    - 提交事务后执行磁盘清理（best-effort）：若 `physicalDeleted`，删除 `storagePath` 和 `hlsOutputDir`
    - 若 `!t.FileStoreID.Valid`（legacy）：走原有路径删除 `input_path` 和 `output_dir`
    - 磁盘删除失败时记录 WARN 日志，返回成功
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 9.3_

  - [ ]* 6.3 为 TaskService.DeleteForUser 编写单元测试
    - 覆盖：有 `file_store_id`（ref_count > 1）、有 `file_store_id`（ref_count = 1，触发物理删除）、无 `file_store_id`（legacy 路径）
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [x] 7. Checkpoint — 确保服务层修改通过测试
  - 确保所有测试通过，如有疑问请向用户确认。

- [x] 8. 修改 FFmpegProcessor（Worker）
  - [x] 8.1 修改 `FFmpegProcessor` 结构体，注入 `*pgxpool.Pool`
    - 更新 `NewFFmpegProcessor` 签名：新增 `pg *pgxpool.Pool` 参数
    - 同步更新 `worker/pool.go` 中 `NewPool` 对 `NewFFmpegProcessor` 的调用，传入 `deps.PG`
    - _Requirements: 6.1, 6.2_

  - [x] 8.2 在 `FFmpegProcessor` 中实现 `completeWithDedup` 方法
    - 开启事务
    - 调用 `UpdateTaskCompleted` 标记原始 task 完成
    - 调用 `GetFileStoreByTaskID` 获取关联的 file_store 行（失败时记录 ERROR 日志，继续提交）
    - 调用 `UpdateFileStoreHLSDir` 更新 `hls_output_dir`（失败时记录 ERROR 日志，继续提交）
    - 若 `fs.RefCount > 1`：调用 `UpdatePendingTasksCompleted` 批量更新 pending tasks
    - 提交事务
    - _Requirements: 6.1, 6.2, 6.3_

  - [x] 8.3 在 `FFmpegProcessor.Process` 中将原有 `UpdateTaskCompleted` 调用替换为 `completeWithDedup`
    - 确保 Redis progress 更新（设为 100）仍在 `completeWithDedup` 成功后执行
    - _Requirements: 6.1, 6.2, 6.3_

  - [ ]* 8.4 为 FFmpegProcessor 编写属性测试（Property 8：Worker 完成后批量更新 pending tasks）
    - **Property 8: Worker 完成后批量更新 pending tasks**
    - 共享同一 `file_store_id` 的 N 个 pending tasks，Worker 完成后验证所有 tasks 状态为 `completed` 且字段一致
    - **Validates: Requirements 6.1, 6.2**

- [x] 9. 修改 UploadHandler，支持 X-File-SHA256 快速跳过
  - [x] 9.1 在 `UploadService` 中添加 `CheckUserHash` 方法
    - 调用 `s.dedup.FindByUserHash(ctx, userID, hash)`
    - 返回 `(UploadResult, bool, error)`，`UploadResult.Status` 填充 task 的当前状态
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

  - [x] 9.2 修改 `UploadHandler.Upload`，在读取文件体之前处理 `X-File-SHA256` 请求头
    - 读取 `c.Request().Header.Get("X-File-SHA256")`
    - 若 header 非空，调用 `h.svc.CheckUserHash`
    - 若命中（`found=true`）：返回 HTTP 200 `{"task_id": ..., "status": ...}`，不处理文件体
    - 若未命中：继续正常上传流程
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

  - [ ]* 9.3 为 UploadHandler 编写属性测试（Property 6：API 响应结构一致性）
    - **Property 6: API 响应结构一致性（防 Hash Confirmation Attack）**
    - 对去重和非去重场景，验证 HTTP 响应结构相同，且不含 `file_store_id`、`sha256_hash`、`deduplicated`、`cached`、`existing_task` 字段
    - **Validates: Requirements 3.5, 4.5, 7.1, 7.2, 7.3, 7.4**

- [x] 10. 更新依赖注入（app.go）
  - [x] 10.1 修改 `backend/internal/app/app.go` 中的 `buildHandlers` 函数
    - 新增 `pg *pgxpool.Pool` 参数
    - 实例化 `dedupSvc := service.NewDedupService(pg, logger)`
    - 更新 `NewUploadService`、`NewUploadSessionService`、`NewTaskService` 调用，传入 `pg` 和 `dedupSvc`
    - 更新 `buildHandlers` 的调用处（`New` 函数），传入 `pg`
    - _Requirements: 3.1, 4.1, 5.1, 6.1_

  - [ ]* 10.2 编写集成测试，验证并发上传相同哈希的正确性
    - 使用 testcontainers-go 启动真实 PostgreSQL，并发上传相同 hash，验证 `ON CONFLICT DO UPDATE` 正确性
    - 验证迁移脚本正确性（表结构、索引、约束、`ref_count CHECK >= 0`）
    - _Requirements: 9.1, 9.2, 1.1, 1.2_

- [x] 11. Final Checkpoint — 确保所有测试通过
  - 确保所有测试通过，如有疑问请向用户确认。

## Notes

- 标有 `*` 的子任务为可选项，可跳过以加快 MVP 交付
- 每个任务均引用了具体的需求条款以保证可追溯性
- 属性测试使用 `pgregory.net/rapid`（与项目 Go 技术栈一致）
- 集成测试使用 `testcontainers-go` 启动真实 PostgreSQL 实例
- 历史 task（`file_store_id IS NULL`）在删除时走原有路径，保持向后兼容
- `DedupService.FindOrCreate` 自行管理事务边界；`DedupService.DecrRef` 接受外部事务（由 `TaskService` 控制）

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2"] },
    { "id": 1, "tasks": ["1.3", "1.4"] },
    { "id": 2, "tasks": ["1.5"] },
    { "id": 3, "tasks": ["2.1"] },
    { "id": 4, "tasks": ["2.2", "2.3", "2.4"] },
    { "id": 5, "tasks": ["2.5", "2.6", "2.7", "2.8", "4.1", "4.2"] },
    { "id": 6, "tasks": ["4.3", "5.1", "6.1", "8.1"] },
    { "id": 7, "tasks": ["4.4", "4.5", "5.2", "6.2", "8.2", "9.1"] },
    { "id": 8, "tasks": ["5.3", "6.3", "8.3", "9.2"] },
    { "id": 9, "tasks": ["8.4", "9.3", "10.1"] },
    { "id": 10, "tasks": ["10.2"] }
  ]
}
```
