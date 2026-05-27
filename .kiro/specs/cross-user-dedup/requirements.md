# Requirements Document

## Introduction

本功能为 videoFlowConvert 项目增加「跨用户文件去重」能力。当前系统允许不同用户重复上传相同的视频文件，导致存储空间和 FFmpeg 转码资源的浪费。

核心设计原则：
- **服务端静默复用**：相同内容的物理文件只存储一份，转码结果只生成一份，但每个用户保有独立的 task 记录，用户体验不受影响。
- **防止 Hash Confirmation Attack**：去重逻辑对用户完全透明，不通过任何 API 响应泄露其他用户的文件存在性。
- **引用计数管理**：通过 `file_store` 表追踪物理文件的引用数，最后一个引用被删除时才真正删除磁盘文件。
- **客户端快速跳过**：客户端可预计算 SHA-256 并在上传前查询，仅对当前用户的历史记录做快速跳过（不跨用户暴露信息）。

涉及的上传路径：
1. 单次上传（`POST /upload`）
2. 断点续传完成（`POST /uploads/:id/complete`）

## Glossary

- **FileStore**：管理物理文件的数据库表，每行代表一个唯一的物理文件（以 SHA-256 哈希标识），包含存储路径、引用计数和 HLS 输出目录。
- **SHA-256 Hash**：对文件全部字节计算的 SHA-256 摘要，用于跨用户去重的唯一标识。
- **Reference Count**：`file_store` 表中记录当前有多少个 `tasks` 行引用同一物理文件的计数器。
- **Hash Confirmation Attack**：攻击者通过上传已知文件并观察响应差异（如速度、状态码）来推断其他用户是否上传过相同文件的侧信道攻击。
- **Dedup Service**：负责协调文件哈希查找、引用计数增减、物理文件清理的服务层组件。
- **Upload Service**：处理单次上传（`POST /upload`）的服务层组件。
- **UploadSession Service**：处理断点续传（`POST /uploads/*`）的服务层组件。
- **Task Service**：管理 task 记录生命周期（查询、删除）的服务层组件。
- **Worker Pool**：异步执行 FFmpeg 转码任务的工作池。
- **HLS Output**：FFmpeg 将输入视频转码后生成的 HLS 播放列表（`.m3u8`）及分片（`.ts`）文件集合。
- **Canonical Input Path**：`file_store` 中记录的物理输入文件路径，所有引用该文件的 task 共享此路径。
- **Per-User Hash Check**：仅在当前用户自己的 task 历史中查找相同哈希，不跨用户查询，用于客户端快速跳过场景。

---

## Requirements

### Requirement1：file_store 表与引用计数

**User Story:** As a system administrator, I want a centralized file registry that tracks physical files and their reference counts, so that the system can safely share files across users without orphaning or prematurely deleting data.

#### Acceptance Criteria

1. THE FileStore SHALL maintain a `file_store` table with columns：`id`（BIGSERIAL 主键）、`sha256_hash`（VARCHAR(64) UNIQUE NOT NULL）、`storage_path`（TEXT NOT NULL，物理输入文件路径）、`hls_output_dir`（TEXT，转码完成后填充）、`ref_count`（INTEGER NOT NULL DEFAULT 1 CHECK (ref_count >= 0)）、`created_at`（TIMESTAMPTZ NOT NULL DEFAULT NOW()）。
2. THE FileStore SHALL enforce a UNIQUE constraint on `sha256_hash` to guarantee at most one physical file per unique content hash.
3. WHEN a new task references a file already present in `file_store`, THE FileStore SHALL increment `ref_count` by 1 using an atomic `UPDATE ... SET ref_count = ref_count + 1` operation.
4. WHEN a task is deleted and its associated `file_store` row's `ref_count` decrements to 0, THE FileStore SHALL delete the `file_store` row and the corresponding physical files from disk.
5. IF a `ref_count` decrement would result in a value below 0, THEN THE FileStore SHALL reject the operation and return an internal error.
6. THE FileStore SHALL expose `file_store_id` as a nullable foreign key column on the `tasks` table, referencing `file_store(id)` with `ON DELETE SET NULL`.

---

### Requirement2：SHA-256 哈希计算

**User Story:** As a developer, I want the system to compute SHA-256 hashes of uploaded files, so that identical files can be identified and deduplicated across users.

#### Acceptance Criteria

1. WHEN a file upload is received via `POST /upload`, THE Upload Service SHALL compute the SHA-256 hash of the complete file bytes before writing the task row.
2. WHEN a resumable upload is completed via `POST /uploads/:id/complete`, THE UploadSession Service SHALL compute the SHA-256 hash of the assembled file before writing the task row.
3. THE Upload Service SHALL compute the SHA-256 hash in a single streaming pass over the file bytes to avoid reading the file twice.
4. IF the SHA-256 hash computation fails due to an I/O error, THEN THE Upload Service SHALL return an internal error and abort the upload without creating a task row.
5. IF the SHA-256 hash computation fails due to an I/O error, THEN THE UploadSession Service SHALL return an internal error and abort the completion without creating a task row.

---

### Requirement3：服务端静默去重（单次上传）

**User Story:** As a user, I want my video uploads to complete successfully regardless of whether another user has uploaded the same file, so that I always get my own task record without being aware of deduplication.

#### Acceptance Criteria

1. WHEN a file is uploaded via `POST /upload` and the computed SHA-256 hash matches an existing `file_store` row, THE Upload Service SHALL create a new `tasks` row for the current user referencing the existing `file_store` row, without copying the physical file.
2. WHEN a file is uploaded via `POST /upload` and the computed SHA-256 hash matches an existing `file_store` row with a non-null `hls_output_dir`, THE Upload Service SHALL set the new task's `status` to `completed`, `output_dir` to the existing `hls_output_dir`, `m3u8_url` to the derived HLS URL, and `duration_sec` to the existing value, without submitting a new Worker Pool job.
3. WHEN a file is uploaded via `POST /upload` and the computed SHA-256 hash matches an existing `file_store` row with a null `hls_output_dir` (transcoding still in progress), THE Upload Service SHALL create a new `tasks` row with `status` `pending` and submit a Worker Pool job using the shared `input_path`.
4. WHEN a file is uploaded via `POST /upload` and the computed SHA-256 hash does not match any existing `file_store` row, THE Upload Service SHALL insert a new `file_store` row, write the physical file to disk, create a `tasks` row, and submit a Worker Pool job.
5. THE Upload Service SHALL return the same HTTP 202 response structure to the caller regardless of whether deduplication occurred, so that no information about other users' uploads is leaked.

---

### Requirement4：服务端静默去重（断点续传）

**User Story:** As a user, I want my resumable uploads to complete successfully and benefit from deduplication, so that I get my own task record even when the same file was previously uploaded by another user.

#### Acceptance Criteria

1. WHEN a resumable upload is completed via `POST /uploads/:id/complete` and the computed SHA-256 hash matches an existing `file_store` row, THE UploadSession Service SHALL create a new `tasks` row for the current user referencing the existing `file_store` row, and SHALL delete the assembled temporary file from disk.
2. WHEN a resumable upload is completed via `POST /uploads/:id/complete` and the computed SHA-256 hash matches an existing `file_store` row with a non-null `hls_output_dir`, THE UploadSession Service SHALL set the new task's `status` to `completed`, `output_dir`, `m3u8_url`, and `duration_sec` from the existing `file_store` row, without submitting a new Worker Pool job.
3. WHEN a resumable upload is completed via `POST /uploads/:id/complete` and the computed SHA-256 hash matches an existing `file_store` row with a null `hls_output_dir`, THE UploadSession Service SHALL create a new `tasks` row with `status` `pending` and submit a Worker Pool job using the shared `input_path`.
4. WHEN a resumable upload is completed via `POST /uploads/:id/complete` and the computed SHA-256 hash does not match any existing `file_store` row, THE UploadSession Service SHALL insert a new `file_store` row, rename the assembled file to the canonical input path, create a `tasks` row, and submit a Worker Pool job.
5. THE UploadSession Service SHALL return the same HTTP 200 response structure to the caller regardless of whether deduplication occurred.

---

### Requirement5：任务删除与引用计数递减

**User Story:** As a user, I want to delete my task records, so that I can manage my video library, and the system should only delete the physical file when no other user references it.

#### Acceptance Criteria

1. WHEN a task is deleted via `DELETE /tasks/:id` and the task has a non-null `file_store_id`, THE Task Service SHALL decrement the corresponding `file_store` row's `ref_count` by 1 within the same database transaction as the task deletion.
2. WHEN the `ref_count` of a `file_store` row reaches 0 after decrement, THE Task Service SHALL delete the physical input file at `storage_path` and the HLS output directory at `hls_output_dir` from disk after the transaction commits.
3. WHEN a task is deleted and the task has a null `file_store_id` (legacy task without dedup), THE Task Service SHALL delete the physical files referenced by `input_path` and `output_dir` on the task row directly, preserving backward compatibility.
4. IF the physical file deletion fails after a successful database transaction, THEN THE Task Service SHALL log the error at WARN level and return success to the caller, treating orphaned files as a recoverable operational issue.
5. THE Task Service SHALL perform the `ref_count` decrement and task deletion atomically within a single PostgreSQL transaction to prevent reference count inconsistencies under concurrent deletions.

---

### Requirement6：Worker 完成时回写 HLS 输出目录

**User Story:** As a system, I want the transcoding worker to update the file_store record when transcoding completes, so that subsequent deduplicated tasks can immediately reference the completed HLS output.

#### Acceptance Criteria

1. WHEN the FFmpeg Worker completes transcoding successfully, THE Worker Pool SHALL update the corresponding `file_store` row's `hls_output_dir` to the output directory path within the same database operation as `UpdateTaskCompleted`.
2. WHEN the FFmpeg Worker completes transcoding successfully and the `file_store` row has `ref_count` greater than 1, THE Worker Pool SHALL update all `tasks` rows that share the same `file_store_id` and have `status` `pending` to `status` `completed`, setting `output_dir`, `m3u8_url`, and `duration_sec` from the completed job.
3. IF the `file_store` row update fails, THEN THE Worker Pool SHALL log the error at ERROR level and continue to mark the originating task as completed, treating the `file_store` update as a best-effort operation.

---

### Requirement7：防止 Hash Confirmation Attack

**User Story:** As a security engineer, I want the deduplication mechanism to be invisible to API consumers, so that users cannot infer whether other users have uploaded the same file.

#### Acceptance Criteria

1. THE Upload Service SHALL return identical HTTP status codes, response body structures, and response timing characteristics for deduplicated and non-deduplicated uploads.
2. THE UploadSession Service SHALL return identical HTTP status codes and response body structures for deduplicated and non-deduplicated resumable upload completions.
3. THE Upload Service SHALL NOT include any field in the API response that indicates whether deduplication occurred (e.g., no `deduplicated`, `cached`, or `existing_task` fields).
4. THE Upload Service SHALL NOT expose the `file_store_id` or `sha256_hash` values in any user-facing API response.
5. WHILE a deduplication lookup is in progress, THE Upload Service SHALL NOT return a response faster than the minimum time required to write a new file to disk, to prevent timing-based side-channel attacks.

---

### Requirement8：客户端 Per-User 快速跳过（可选功能）

**User Story:** As a client developer, I want to query whether the current user has already uploaded a file with a given SHA-256 hash, so that the client can skip re-uploading a file the same user has already uploaded.

#### Acceptance Criteria

1. WHERE the client provides a `X-File-SHA256` header in `POST /upload`, THE Upload Service SHALL check whether the current authenticated user has an existing task with the matching SHA-256 hash.
2. WHEN the current user has an existing task matching the provided SHA-256 hash, THE Upload Service SHALL return HTTP 200 with the existing task's slug and status, without processing the uploaded file bytes.
3. WHEN the current user does not have an existing task matching the provided SHA-256 hash, THE Upload Service SHALL proceed with the normal upload flow regardless of whether other users have uploaded the same file.
4. THE Upload Service SHALL NOT query `file_store` by hash when processing the `X-File-SHA256` header; THE Upload Service SHALL only query the current user's own `tasks` rows to prevent cross-user information leakage.
5. WHERE the client provides a `X-File-SHA256` header in `POST /uploads` (init resumable upload), THE UploadSession Service SHALL check whether the current authenticated user has an existing task with the matching SHA-256 hash and return HTTP 200 with the existing task if found.

---

### Requirement9：数据库事务一致性

**User Story:** As a system architect, I want all deduplication-related database operations to be atomic, so that the system remains consistent under concurrent uploads of the same file.

#### Acceptance Criteria

1. WHEN two concurrent uploads of the same file arrive simultaneously, THE Dedup Service SHALL use a PostgreSQL `INSERT INTO file_store ... ON CONFLICT (sha256_hash) DO UPDATE SET ref_count = file_store.ref_count + 1 RETURNING *` statement to atomically insert or increment, preventing duplicate `file_store` rows.
2. THE Dedup Service SHALL wrap the `file_store` upsert and `tasks` insert in a single PostgreSQL transaction so that a partial failure leaves no orphaned `file_store` rows with incorrect `ref_count`.
3. WHEN a task deletion and `ref_count` decrement occur concurrently with a new upload referencing the same hash, THE Task Service SHALL use `SELECT ... FOR UPDATE` on the `file_store` row within the deletion transaction to serialize access and prevent the `ref_count` from reaching 0 while a new reference is being added.
4. IF a transaction involving `file_store` upsert or `ref_count` decrement fails due to a serialization error, THEN THE Dedup Service SHALL return an internal error to the caller without retrying automatically.

---

### Requirement10：SHA-256 哈希存储在 tasks 表

**User Story:** As a developer, I want the SHA-256 hash to be stored on the tasks row, so that per-user hash lookups are efficient and do not require joining to file_store.

#### Acceptance Criteria

1. THE Upload Service SHALL store the computed SHA-256 hash in a `sha256_hash` column on the `tasks` table at task creation time.
2. THE UploadSession Service SHALL store the computed SHA-256 hash in the `sha256_hash` column on the `tasks` table at task creation time.
3. THE Task Service SHALL use an index on `(user_id, sha256_hash)` to efficiently serve per-user hash lookup queries without full table scans.
4. WHEN querying for a per-user hash match, THE Upload Service SHALL query `SELECT slug, status FROM tasks WHERE user_id = $1 AND sha256_hash = $2 LIMIT 1` to avoid exposing cross-user data.
