# 小型网盘系统 — 项目规划文档 v4

> **文档用途**：供 AI 编码助手直接读取并执行。  
> **定位**：单机小型网盘 + 可选媒体库。  
> **存储方案**：服务器本地磁盘，预留 Storage 接口，后续可替换为 S3 / MinIO / OSS。  
> **版本**：v4

---

## v4 关键调整

相对 v3，本版本做以下修正：

1. 上传视频后**不自动转码**。用户需要主动将视频加入媒体库，才创建媒体条目并触发 FFmpeg 转码。
2. 修正 `pre_sha1` 逻辑：`pre_sha1` 是前 256KB 内容的哈希，不再错误地用完整文件哈希前缀匹配。
3. `/upload/pre-check` 只做低成本预判，不创建上传任务；上传任务统一由 `/upload/init` 创建或恢复。
4. 文件指纹从 SHA-1 升级为 **SHA-256**。字段名统一为 `file_hash`，算法字段为 `hash_algo`。
5. 分片合并后必须重新计算完整 SHA-256 校验，不能作为可选项。
6. 配额更新改为数据库原子更新，避免并发上传突破配额。
7. 同一目录下增加未删除文件名唯一约束，目录和文件都不允许同名，避免重复条目造成交互混乱。
8. 合并、秒传、永久删除等高并发路径增加事务和锁要求。
9. 媒体库从“转码任务”升级为独立业务模块：`media_items` + `media_transcodes` + `media_jobs`。
10. 明确本地磁盘方案边界：v1/v4 是单机部署架构，不支持多实例共享本地文件系统。
11. 上传前增加同目录重复检测。若检测到同目录已存在同名文件、同名目录或相同内容文件，前端必须弹出确认框；用户确认后才继续走秒传或分片上传逻辑。
12. “同一个文件”的精确判断以完整 SHA-256 为准；`pre_hash` 只能用于低成本疑似判断。
13. 上传模块必须与网盘目录、媒体库、头像等业务解耦。上传只负责生成或复用 `physical_files`，返回 `physical_file_slug`；创建网盘文件条目由文件模块完成。
14. 用户主表保持精简，只放认证身份和账户状态。用户资料、存储统计、等级、交易记录等业务信息拆到独立表。
15. HLS 转码产物按物理文件去重。不同用户把同一个物理文件加入媒体库时，复用同一份 `media_transcodes` 和 HLS 文件，不重复转码。

---

## 技术栈

| 层 | 技术 |
|----|------|
| 后端语言 | Go 1.22+ |
| Web 框架 | Echo v4 |
| 数据库 | PostgreSQL 16 |
| 缓存 | Redis 7 |
| 查询生成 | sqlc |
| 文件存储 | 服务器本地磁盘 |
| 媒体处理 | FFmpeg（按需转码 + HLS 切片） |
| 前端框架 | SvelteKit + Svelte 5 + bits-ui |
| 前端构建 | Vite |
| 哈希算法 | SHA-256 |
| 认证方式 | JWT access token 15min + refresh token 7d |
| 容器化 | Docker + docker-compose |

---

## 架构边界

本方案适合：

- 单机部署。
- 家庭、个人、小团队或内部工具。
- 数据量可控的文件存储。
- 可接受通过服务器磁盘备份做灾备。

本方案暂不适合：

- 多台后端实例同时读写不同本地磁盘。
- 高可用对象存储场景。
- 海量文件和大规模并发下载。
- 跨机器调度 FFmpeg worker。

后续如需扩容，应替换 `Storage` 接口实现，将本地磁盘迁移到 S3 / MinIO / OSS。

---

## 本地文件存储规范

### 目录结构

SHA-256 共 64 位十六进制字符，取前 4 位拆成两级目录，文件名为完整 SHA-256，无扩展名。

```text
{STORAGE_ROOT}/
├── ab/
│   └── cd/
│       └── abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890
├── tmp/
├── avatars/
└── hls/
```

路径生成：

```go
func StoragePath(fileHash string) string {
    return filepath.Join(fileHash[0:2], fileHash[2:4], fileHash)
}

func AbsPath(root, fileHash string) string {
    return filepath.Join(root, StoragePath(fileHash))
}
```

所有参与路径拼接的参数必须先校验：

- `file_hash`：`^[a-f0-9]{64}$`
- `slug` / `upload_slug`：`^[A-Za-z0-9_-]{21}$`
- 分片文件名由服务端生成，禁止使用客户端传入路径。

---

## 推荐目录结构

```text
netdisk/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repo/
│   │   │   └── queries/
│   │   ├── storage/
│   │   ├── cache/
│   │   ├── media/
│   │   ├── middleware/
│   │   └── model/
│   ├── migrations/
│   ├── sqlc.yaml
│   ├── config.yaml
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/
│   │   │   ├── workers/
│   │   │   ├── stores/
│   │   │   └── components/
│   │   └── routes/
│   ├── package.json
│   └── vite.config.ts
├── docker-compose.yml
└── README.md
```

---

## 数据库 Schema

约定：所有表使用 `BIGSERIAL` 自增主键，对外暴露 `slug`，HTTP 入参和响应不暴露内部 ID。

### users

用户主表只保存登录身份、认证凭据和账户状态。不要把等级、余额、交易、配额、头像等可扩展业务字段塞进主表。

```sql
CREATE TABLE users (
    id            BIGSERIAL    PRIMARY KEY,
    slug          VARCHAR(21)  NOT NULL UNIQUE,
    username      VARCHAR(64)  NOT NULL UNIQUE,
    email         VARCHAR(256) NOT NULL UNIQUE,
    password_hash VARCHAR(256) NOT NULL,
    status        SMALLINT     NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_slug ON users (slug);
CREATE INDEX idx_users_email ON users (email);
```

### user_profiles

用户展示资料。头像、昵称等非认证字段放这里。

```sql
CREATE TABLE user_profiles (
    id           BIGSERIAL    PRIMARY KEY,
    user_id      BIGINT       NOT NULL UNIQUE REFERENCES users(id),
    display_name VARCHAR(64),
    avatar_path  TEXT,
    bio          TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_profiles_user ON user_profiles (user_id);
```

### user_storage_stats

用户存储用量和配额。所有配额增减都更新这张表，不更新 `users` 主表。

```sql
CREATE TABLE user_storage_stats (
    id            BIGSERIAL    PRIMARY KEY,
    user_id       BIGINT       NOT NULL UNIQUE REFERENCES users(id),
    storage_used  BIGINT       NOT NULL DEFAULT 0,
    storage_quota BIGINT       NOT NULL DEFAULT 10737418240,
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CHECK (storage_used >= 0),
    CHECK (storage_quota >= 0)
);

CREATE INDEX idx_user_storage_stats_user ON user_storage_stats (user_id);
```

配额原子增加 SQL：

```sql
UPDATE user_storage_stats
SET storage_used = storage_used + $1,
    updated_at = NOW()
WHERE user_id = $2
  AND storage_used + $1 <= storage_quota
RETURNING storage_used, storage_quota;
```

### user_levels

用户等级信息。MVP 可只建表不实现完整升级规则。

```sql
CREATE TABLE user_levels (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     BIGINT       NOT NULL UNIQUE REFERENCES users(id),
    level_code  VARCHAR(32)  NOT NULL DEFAULT 'free',
    level_name  VARCHAR(64)  NOT NULL DEFAULT '免费用户',
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_levels_user ON user_levels (user_id);
CREATE INDEX idx_user_levels_code ON user_levels (level_code);
```

### user_transactions

用户交易记录。用于后续会员购买、容量扩容、退款等场景。

```sql
CREATE TABLE user_transactions (
    id              BIGSERIAL    PRIMARY KEY,
    slug            VARCHAR(21)  NOT NULL UNIQUE,
    user_id         BIGINT       NOT NULL REFERENCES users(id),
    transaction_no  VARCHAR(64)  NOT NULL UNIQUE,
    type            VARCHAR(32)  NOT NULL,
    -- membership | storage_upgrade | refund | adjustment
    amount_cents    BIGINT       NOT NULL DEFAULT 0,
    currency        VARCHAR(8)   NOT NULL DEFAULT 'CNY',
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
    -- pending | paid | failed | refunded | closed
    metadata        JSONB        NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_transactions_user ON user_transactions (user_id, created_at DESC);
CREATE INDEX idx_user_transactions_status ON user_transactions (status);
```

### physical_files

```sql
CREATE TABLE physical_files (
    id           BIGSERIAL    PRIMARY KEY,
    slug         VARCHAR(21)  NOT NULL UNIQUE,
    hash_algo    VARCHAR(16)  NOT NULL DEFAULT 'sha256',
    file_hash    VARCHAR(64)  NOT NULL UNIQUE,
    pre_hash     VARCHAR(64)  NOT NULL,
    file_size    BIGINT       NOT NULL,
    mime_type    VARCHAR(128) NOT NULL DEFAULT 'application/octet-stream',
    storage_path TEXT         NOT NULL,
    status       VARCHAR(20)  NOT NULL DEFAULT 'completed',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_physical_files_hash ON physical_files (hash_algo, file_hash);
CREATE INDEX idx_physical_files_pre ON physical_files (file_size, pre_hash);
```

### user_files

```sql
CREATE TABLE user_files (
    id               BIGSERIAL    PRIMARY KEY,
    slug             VARCHAR(21)  NOT NULL UNIQUE,
    user_id          BIGINT       NOT NULL REFERENCES users(id),
    physical_file_id BIGINT       REFERENCES physical_files(id),
    parent_id        BIGINT       REFERENCES user_files(id),
    file_name        VARCHAR(512) NOT NULL,
    is_dir           BOOLEAN      NOT NULL DEFAULT FALSE,
    file_size        BIGINT       NOT NULL DEFAULT 0,
    mime_type        VARCHAR(128),
    is_starred       BOOLEAN      NOT NULL DEFAULT FALSE,
    is_trashed       BOOLEAN      NOT NULL DEFAULT FALSE,
    trashed_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CHECK (
        (is_dir = TRUE AND physical_file_id IS NULL)
        OR
        (is_dir = FALSE AND physical_file_id IS NOT NULL)
    )
);

CREATE INDEX idx_user_files_user_parent ON user_files (user_id, parent_id)
    WHERE is_trashed = FALSE;
CREATE INDEX idx_user_files_user_starred ON user_files (user_id)
    WHERE is_starred = TRUE AND is_trashed = FALSE;
CREATE INDEX idx_user_files_slug ON user_files (slug);

CREATE UNIQUE INDEX uq_user_files_name_root
ON user_files (user_id, file_name)
WHERE parent_id IS NULL AND is_trashed = FALSE;

CREATE UNIQUE INDEX uq_user_files_name_child
ON user_files (user_id, parent_id, file_name)
WHERE parent_id IS NOT NULL AND is_trashed = FALSE;
```

### upload_tasks

上传任务是基础能力，只表示“某个用户发起了一次物理文件上传会话”。它不保存目录、文件名、媒体库等业务字段。

```sql
CREATE TABLE upload_tasks (
    id           BIGSERIAL    PRIMARY KEY,
    slug         VARCHAR(21)  NOT NULL UNIQUE,
    owner_user_id BIGINT      NOT NULL REFERENCES users(id),
    hash_algo    VARCHAR(16)  NOT NULL DEFAULT 'sha256',
    file_hash    VARCHAR(64)  NOT NULL,
    pre_hash     VARCHAR(64)  NOT NULL,
    file_size    BIGINT       NOT NULL,
    mime_type    VARCHAR(128) NOT NULL DEFAULT 'application/octet-stream',
    total_chunks INT          NOT NULL,
    chunk_size   INT          NOT NULL DEFAULT 4194304,
    status       VARCHAR(20)  NOT NULL DEFAULT 'created',
    physical_file_id BIGINT   REFERENCES physical_files(id),
    error_msg    TEXT,
    expires_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW() + INTERVAL '7 days',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_upload_tasks_owner ON upload_tasks (owner_user_id, status);
CREATE INDEX idx_upload_tasks_hash ON upload_tasks (hash_algo, file_hash);
```

### refresh_tokens

```sql
CREATE TABLE refresh_tokens (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    BIGINT       NOT NULL REFERENCES users(id),
    token_hash VARCHAR(64)  NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ  NOT NULL,
    revoked    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id)
    WHERE revoked = FALSE;
```

### media_transcodes

共享转码产物。它属于物理文件层，不属于某个用户。多个用户的媒体库条目可以引用同一个转码产物。

```sql
CREATE TABLE media_transcodes (
    id               BIGSERIAL    PRIMARY KEY,
    slug             VARCHAR(21)  NOT NULL UNIQUE,
    physical_file_id BIGINT       NOT NULL REFERENCES physical_files(id),
    profile          VARCHAR(32)  NOT NULL DEFAULT 'hls_1080p',
    status           VARCHAR(20)  NOT NULL DEFAULT 'pending',
    -- pending | processing | done | failed
    hls_dir          TEXT,
    duration_sec     INT,
    error_msg        TEXT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_media_transcodes_file_profile
ON media_transcodes (physical_file_id, profile);

CREATE INDEX idx_media_transcodes_status ON media_transcodes (status)
    WHERE status IN ('pending', 'processing');
```

### media_items

媒体库条目。用户主动把网盘里的视频加入媒体库时创建。它记录用户视角的标题、封面等个性化信息，并引用共享转码产物。

```sql
CREATE TABLE media_items (
    id               BIGSERIAL    PRIMARY KEY,
    slug             VARCHAR(21)  NOT NULL UNIQUE,
    user_id          BIGINT       NOT NULL REFERENCES users(id),
    user_file_id     BIGINT       NOT NULL REFERENCES user_files(id),
    physical_file_id BIGINT       NOT NULL REFERENCES physical_files(id),
    transcode_id     BIGINT       REFERENCES media_transcodes(id),
    title            VARCHAR(512) NOT NULL,
    poster_path      TEXT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_media_items_user_file ON media_items (user_id, user_file_id);
CREATE INDEX idx_media_items_user ON media_items (user_id, created_at DESC);
CREATE INDEX idx_media_items_transcode ON media_items (transcode_id);
```

### media_jobs

```sql
CREATE TABLE media_jobs (
    id            BIGSERIAL    PRIMARY KEY,
    slug          VARCHAR(21)  NOT NULL UNIQUE,
    transcode_id  BIGINT       NOT NULL REFERENCES media_transcodes(id),
    status        VARCHAR(20)  NOT NULL DEFAULT 'pending',
    -- pending | processing | done | failed
    error_msg     TEXT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_media_jobs_transcode_active
ON media_jobs (transcode_id)
WHERE status IN ('pending', 'processing');

CREATE INDEX idx_media_jobs_status ON media_jobs (status)
    WHERE status IN ('pending', 'processing');
```

---

## Redis Key 规范

| Key 模式 | 类型 | TTL | 说明 |
|----------|------|-----|------|
| `nd:pre:{file_size}:{pre_hash}` | String | 7d | 预判缓存，值为 physical_file.slug |
| `nd:challenge:{user_id}:{file_hash}` | Hash | 2min | 拥有权挑战，字段：offset、token |
| `nd:chunks:{upload_slug}` | Set | 24h | 已完成分块索引 |
| `nd:lock:merge:{file_hash}` | String | 10min | 合并同一文件时的互斥锁 |
| `nd:rate:{ip}:{action}` | String | 1min | 限流计数器 |
| `nd:media:transcode:{transcode_slug}:progress` | String | 转码中 | 共享转码进度，0-100 |

---

## 本地存储接口

```go
type Storage interface {
    WriteChunk(uploadSlug string, chunkIndex int, data io.Reader) error
    ListChunks(uploadSlug string) ([]int, error)
    ReadAt(fileHash string, offset, length int64) ([]byte, error)
    MergeChunks(uploadSlug string, fileHash string, totalChunks int) error
    HashFile(fileHash string) (string, error)
    Exists(fileHash string) bool
    Open(fileHash string) (*os.File, error)
    Delete(fileHash string) error
    CleanupUpload(uploadSlug string) error
}
```

实现要求：

- `WriteChunk` 使用服务端生成的 `chunk_%06d` 文件名。
- `MergeChunks` 必须先写入临时目标文件，例如 `{file_hash}.merging`，校验成功后再原子 rename 到最终路径。
- 合并后必须计算 SHA-256，并与 `upload_tasks.file_hash` 比对。
- 校验失败时删除临时目标文件，任务状态改为 `failed`。
- `Delete` 只能删除哈希路径下的物理文件，禁止删除任意传入路径。

---

## API 设计

### 认证模块 `/api/v1/auth`

保持 v3 设计，补充要求：

- refresh token 每次刷新后轮换，旧 token 立即 revoked。
- logout 需要提交 refresh token，服务端吊销该 refresh token。
- access token 短期有效，默认不做黑名单。
- 登录限流：`nd:rate:{ip}:login` 每分钟最多 10 次。

---

## 用户模块 `/api/v1/user`

用户模块按业务表聚合返回，不把所有字段塞进 `users`。

### GET `/api/v1/user/me`

```json
{
  "slug": "V1StGXR8_Z5jdHi6B-myT",
  "username": "alice",
  "email": "alice@example.com",
  "status": 1,
  "profile": {
    "display_name": "Alice",
    "avatar_url": "/api/v1/user/avatar/V1StGXR8_Z5jdHi6B-myT",
    "bio": ""
  },
  "storage": {
    "storage_used": 1073741824,
    "storage_quota": 10737418240
  },
  "level": {
    "level_code": "free",
    "level_name": "免费用户",
    "expires_at": null
  },
  "created_at": "2024-01-01T00:00:00Z"
}
```

服务端查询：

- `users`：身份、邮箱、账号状态。
- `user_profiles`：展示资料。
- `user_storage_stats`：容量用量和配额。
- `user_levels`：用户等级。

### PATCH `/api/v1/user/profile`

只更新 `user_profiles`。

```json
{
  "display_name": "Alice",
  "bio": "hello"
}
```

### POST `/api/v1/user/me/password`

只更新 `users.password_hash`。

```json
{
  "old_password": "...",
  "new_password": "..."
}
```

### POST `/api/v1/user/me/avatar`

头像保存到 `{STORAGE_ROOT}/avatars/{user_slug}.jpg`，路径写入 `user_profiles.avatar_path`。

要求：

- 仅接受 jpg/png。
- 最大 5MB。
- 校验 MIME 和图片解码结果，不只看扩展名。

### GET `/api/v1/user/avatar/:user_slug`

读取 `user_profiles.avatar_path`，未设置则返回默认头像。

### GET `/api/v1/user/transactions`

分页返回当前用户交易记录。

```json
{
  "items": [
    {
      "slug": "...",
      "transaction_no": "...",
      "type": "membership",
      "amount_cents": 990,
      "currency": "CNY",
      "status": "paid",
      "created_at": "..."
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

---

## 文件模块 `/api/v1/files`

核心接口保持：

- `GET /api/v1/files`
- `POST /api/v1/files/mkdir`
- `POST /api/v1/files/check-conflict`
- `POST /api/v1/files/check-duplicate`
- `POST /api/v1/files/import`
- `DELETE /api/v1/files/:slug`
- `POST /api/v1/files/:slug/restore`
- `DELETE /api/v1/files/:slug/permanent`
- `POST /api/v1/files/:slug/rename`
- `POST /api/v1/files/:slug/move`
- `POST /api/v1/files/:slug/star`
- `GET /api/v1/files/:slug/download`
- `GET /api/v1/files/trash`
- `GET /api/v1/files/starred`

补充规则：

- 同一用户同一目录下，未删除条目的 `file_name` 不允许重复，文件和目录共用同一命名空间。
- `POST /api/v1/files/mkdir` 创建目录时，如果同一目录下已存在同名目录或同名文件，返回 409。
- `rename` 和 `move` 后如果目标目录出现同名未删除条目，返回 409。
- 下载、删除、移动、重命名必须校验 `user_files.user_id = current_user.id`。
- 永久删除时，先删除 `user_files`，再检查物理文件引用数。若引用数为 0，删除磁盘文件。
- 删除媒体库中的源文件时，不自动删除媒体库条目，应将 `media_items` 标记为不可播放或在永久删除时级联清理，二选一。MVP 推荐永久删除时清理对应 `media_items`、`media_jobs`、HLS 输出目录。
- `Content-Disposition` 文件名必须安全转义。

### POST `/api/v1/files/check-conflict`

上传前的低成本业务冲突检查。该接口属于文件模块，不属于上传模块；它只检查“把文件放入某个目录”是否可能产生冲突。

```json
{
  "file_name": "video.mp4",
  "file_size": 1073741824,
  "pre_hash": "sha256_of_first_256kb",
  "parent_slug": "目标目录slug或空"
}
```

响应：

```json
{ "status": "OK" }
```

或：

```json
{
  "status": "NAME_CONFLICT",
  "message": "当前目录已存在同名条目，是否继续上传？",
  "existing": {
    "slug": "...",
    "file_name": "video.mp4",
    "is_dir": false,
    "file_size": 1073741824
  }
}
```

或：

```json
{
  "status": "SAME_FILE_CONFLICT",
  "message": "当前目录可能已存在相同文件，是否继续检查并上传？",
  "existing": {
    "slug": "...",
    "file_name": "video-copy.mp4",
    "file_size": 1073741824
  }
}
```

### POST `/api/v1/files/check-duplicate`

完整哈希计算完成后的精确重复检查。

```json
{
  "file_hash": "full_sha256",
  "parent_slug": "目标目录slug或空"
}
```

响应：

```json
{ "status": "OK" }
```

或：

```json
{
  "status": "DUPLICATE_FILE",
  "message": "当前目录已存在相同文件，是否仍然上传？",
  "existing": {
    "slug": "...",
    "file_name": "video.mp4",
    "file_size": 1073741824
  }
}
```

### POST `/api/v1/files/import`

将上传模块产出的物理文件导入当前用户网盘目录。这个接口是文件业务的边界：配额、目录重名、`user_files` 创建都在这里完成。

```json
{
  "physical_file_slug": "...",
  "file_name": "video.mp4",
  "parent_slug": "目标目录slug或空"
}
```

响应：

```json
{
  "file_slug": "...",
  "file_name": "video.mp4"
}
```

服务端逻辑：

1. 校验 `physical_file_slug` 对应的物理文件存在且状态为 `completed`。
2. 校验 `parent_slug` 属于当前用户。
3. 校验同目录未删除条目不重名，否则返回 409。
4. 使用数据库原子更新增加 `user_storage_stats.storage_used`，超配额返回 507。
5. 事务内插入 `user_files`。
6. 该接口可被普通上传、秒传、外部导入等多个来源复用。

---

## 上传模块 `/api/v1/upload`

上传模块是基础业务，不直接创建 `user_files`，不感知目录、媒体库、头像、分享等上层业务。它的唯一目标是可靠地产生 `physical_files`，并把 `physical_file_slug` 返回给调用方。

### POST `/api/v1/upload/pre-check`

低成本预判。该接口不创建上传任务。

```json
{
  "pre_hash": "sha256_of_first_256kb",
  "file_size": 1073741824
}
```

响应：

```json
{ "status": "SUSPECT_HIT" }
```

或：

```json
{ "status": "NOT_FOUND" }
```

服务端逻辑：

1. 校验单文件大小不超过系统上传上限 `storage.max_upload_size`，否则返回 413。
2. 查 Redis `nd:pre:{file_size}:{pre_hash}`。
3. 未命中则查 `physical_files WHERE file_size = ? AND pre_hash = ? AND status = 'completed'`。
4. 命中返回 `SUSPECT_HIT`，未命中返回 `NOT_FOUND`。

### POST `/api/v1/upload/request-challenge`

疑似命中后，前端计算完整 SHA-256，再申请拥有权挑战。

```json
{
  "file_hash": "full_sha256"
}
```

响应：

```json
{
  "status": "CHALLENGE",
  "challenge_offset": 568231936,
  "challenge_token": "64位hex随机字符串"
}
```

或：

```json
{ "status": "NOT_FOUND" }
```

服务端逻辑：

1. 查 `physical_files WHERE hash_algo='sha256' AND file_hash=? AND status='completed'`。
2. 不存在返回 `NOT_FOUND`，由前端进入 `/upload/init`。
3. 存在则生成随机 offset 和 token。
4. 写 Redis `nd:challenge:{user_id}:{file_hash}`，TTL 120 秒。

### POST `/api/v1/upload/verify`

```json
{
  "file_hash": "full_sha256",
  "proof_code": "sha256(1kb文件数据 + challenge_token)"
}
```

响应：

```json
{ "status": "HIT", "physical_file_slug": "..." }
```

或：

```json
{ "status": "MISS" }
```

服务端逻辑：

1. Lua 原子取删 challenge，防重放。
2. `storage.ReadAt(file_hash, offset, 1024)`。
3. 服务端计算 `sha256(disk_bytes + token)` 并比对。
4. 一致则返回已有 `physical_file_slug`。
5. 不一致返回 `MISS`，调用方进入 `/upload/init`。
6. 调用方拿到 `physical_file_slug` 后，再由具体业务接口决定如何使用。例如网盘文件页调用 `/api/v1/files/import`。

### POST `/api/v1/upload/init`

创建或恢复断点上传任务。

```json
{
  "file_hash": "full_sha256",
  "pre_hash": "sha256_of_first_256kb",
  "file_size": 1073741824,
  "mime_type": "video/mp4"
}
```

响应：

```json
{
  "upload_slug": "...",
  "total_chunks": 256,
  "chunk_size": 4194304,
  "completed_chunks": [0, 1, 2]
}
```

服务端逻辑：

1. 校验 `file_hash`、`pre_hash`、`file_size`、`mime_type` 和系统上传大小上限。
2. 若当前用户已有未完成的相同 `file_hash` 任务，则恢复该任务。
3. 否则创建 `upload_tasks`。
4. 返回临时目录中已存在分片列表。

### POST `/api/v1/upload/chunk`

`multipart/form-data`：

- `upload_slug`
- `chunk_index`
- `chunk_data`
- 可选 `chunk_hash`

服务端逻辑：

1. 校验任务归属。
2. 校验 `chunk_index` 范围。
3. 校验分片大小。非最后一片必须等于 `chunk_size`。
4. 如提交 `chunk_hash`，服务端校验分片 SHA-256。
5. 写入临时目录，并记录 Redis Set。

### POST `/api/v1/upload/complete`

```json
{ "upload_slug": "..." }
```

响应：

```json
{ "status": "MERGING" }
```

异步合并流程：

1. 更新任务为 `merging`。
2. 获取 `nd:lock:merge:{file_hash}` 互斥锁，避免同一物理文件重复合并。
3. 校验所有分片存在。
4. 合并到临时目标文件。
5. 强制计算完整 SHA-256，比对 `upload_tasks.file_hash`。
6. 使用 `INSERT ... ON CONFLICT DO NOTHING` 创建 `physical_files`。
7. 查询最终 `physical_files.id`，更新 `upload_tasks.status='done'` 和 `physical_file_id`。
8. 回填 Redis 预判缓存。
9. 清理临时分片和 Redis chunk key。
10. 不创建 `user_files`，不更新 `user_storage_stats.storage_used`。这些属于具体业务的导入/引用阶段。

### GET `/api/v1/upload/:upload_slug/status`

```json
{
  "status": "merging | done | failed",
  "physical_file_slug": "...",
  "error": "..."
}
```

---

## 媒体库模块 `/api/v1/media`

媒体库是独立业务。上传视频后只会出现在网盘文件列表，不会自动转码。用户主动点击“加入媒体库”后，系统创建媒体库条目，并复用或创建共享转码产物。

### POST `/api/v1/media/items`

将网盘中的视频文件加入媒体库。

```json
{
  "file_slug": "...",
  "title": "可选标题，不传则使用文件名"
}
```

响应：

```json
{
  "media_slug": "...",
  "transcode_slug": "...",
  "transcode_status": "pending",
  "transcode_reused": false
}
```

服务端逻辑：

1. 校验 `file_slug` 属于当前用户。
2. 校验目标不是目录。
3. 校验 MIME 是视频类型，或通过文件头探测确认为视频。
4. 若该文件已加入媒体库，返回已有 `media_item`，不重复创建。
5. 根据 `physical_file_id + profile` 查找或创建 `media_transcodes`。
6. 若 `media_transcodes.status='done'`，直接复用已有 HLS 产物，不创建 `media_jobs`。
7. 若 `media_transcodes.status IN ('pending', 'processing')`，直接引用该 transcode，不创建重复任务，前端静默展示同一个进度。
8. 若不存在 transcode，则创建 `media_transcodes(status='pending')` 和 `media_jobs(status='pending')`。
9. 创建 `media_items`，写入 `transcode_id`。

### GET `/api/v1/media/items`

列出当前用户媒体库。

```json
{
  "items": [
    {
      "slug": "...",
      "title": "电影.mp4",
      "transcode_status": "done",
      "transcode_reused": true,
      "duration_sec": 7200,
      "poster_url": null,
      "created_at": "..."
    }
  ],
  "total": 1
}
```

### GET `/api/v1/media/items/:media_slug`

获取媒体详情。

```json
{
  "slug": "...",
  "title": "电影.mp4",
  "transcode_status": "processing",
  "transcode_reused": true,
  "progress_pct": 45,
  "play_url": null
}
```

转码完成后：

```json
{
  "slug": "...",
  "title": "电影.mp4",
  "transcode_status": "done",
  "progress_pct": 100,
  "play_url": "/api/v1/media/hls/{media_slug}/index.m3u8"
}
```

### POST `/api/v1/media/items/:media_slug/retry`

共享转码失败后重试。只有当当前 `media_item` 引用的 `media_transcodes.status='failed'` 时允许。

```json
{
  "transcode_slug": "...",
  "job_slug": "..."
}
```

### DELETE `/api/v1/media/items/:media_slug`

从媒体库移除。默认只删除当前用户的 `media_items`，不删除网盘源文件，也不立即删除共享 HLS 产物。

HLS 清理建议：

- `media_transcodes` 是共享产物，不能因为某个用户移除媒体库条目就删除。
- 可通过后台清理任务删除“没有任何 `media_items` 引用，且超过保留期”的转码产物和 HLS 目录。

### GET `/api/v1/media/hls/:media_slug/*`

鉴权后返回 HLS 文件。服务端必须校验：

- `media_slug` 属于当前用户。
- 关联的 `media_transcodes.status = 'done'`。
- `*` 路径规范化后仍位于该 transcode 的 HLS 输出目录内。

---

## FFmpeg 转码规范

### HLS 存储路径

```text
{STORAGE_ROOT}/hls/{transcode_slug}/
├── index.m3u8
├── seg0000.ts
├── seg0001.ts
└── ...
```

### 命令建议

优先使用 `-progress pipe:1` 获取结构化进度，避免解析 stderr 中不稳定的 `time=` 字段。

```go
func BuildFFmpegArgs(inputPath, outputDir string) []string {
    return []string{
        "-i", inputPath,
        "-c:v", "libx264",
        "-preset", "fast",
        "-crf", "22",
        "-c:a", "aac",
        "-b:a", "128k",
        "-vf", "scale='min(1920,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease",
        "-hls_time", "10",
        "-hls_playlist_type", "vod",
        "-hls_segment_filename", filepath.Join(outputDir, "seg%04d.ts"),
        "-progress", "pipe:1",
        filepath.Join(outputDir, "index.m3u8"),
    }
}
```

### worker 流程

1. 定时查询 pending 任务：

```sql
SELECT mj.*
FROM media_jobs mj
JOIN media_transcodes mt ON mt.id = mj.transcode_id
WHERE mj.status = 'pending'
  AND mt.status = 'pending'
ORDER BY created_at
LIMIT $1
FOR UPDATE SKIP LOCKED;
```

2. 将 `media_jobs.status` 和 `media_transcodes.status` 改为 `processing`。
3. 通过 `media_transcodes.physical_file_id` 找到 `physical_files.storage_path`，输入文件直接从本地磁盘读取。
4. 输出到 `{STORAGE_ROOT}/hls/{transcode_slug}`。
5. FFmpeg 进度写入 `nd:media:transcode:{transcode_slug}:progress`。
6. 成功后更新：
   - `media_jobs.status='done'`
   - `media_transcodes.status='done'`
   - `media_transcodes.hls_dir`
   - `media_transcodes.duration_sec`
7. 失败后更新 `media_jobs.status='failed'` 和 `media_transcodes.status='failed'`，并清理未完成输出目录。
8. 多个 `media_items` 引用同一个 `media_transcodes` 时，所有用户看到同一份静默进度。

---

## 前端规划

### 上传状态机

```text
IDLE
  -> PRE_HASHING
  -> CHECKING_CONFLICT
      OK
        -> PRE_CHECKING
      NAME_CONFLICT | SAME_FILE_CONFLICT
        -> WAIT_USER_CONFIRM
            CANCEL  -> IDLE
            CONFIRM -> PRE_CHECKING
  -> PRE_CHECKING
      SUSPECT_HIT
        -> FULL_HASHING
        -> CHECKING_DUPLICATE
            DUPLICATE_FILE
              -> WAIT_USER_CONFIRM
                  CANCEL  -> IDLE
                  CONFIRM -> CHALLENGING
            OK
              -> CHALLENGING
        -> CHALLENGING
        -> VERIFYING
            HIT  -> IMPORT_FILE
            MISS -> INIT_UPLOAD
      NOT_FOUND
        -> FULL_HASHING
        -> CHECKING_DUPLICATE
            DUPLICATE_FILE
              -> WAIT_USER_CONFIRM
                  CANCEL  -> IDLE
                  CONFIRM -> INIT_UPLOAD
            OK
              -> INIT_UPLOAD
        -> INIT_UPLOAD
  -> UPLOADING
  -> MERGING
  -> IMPORT_FILE
  -> DONE
  -> ERROR
```

上传确认框规则：

- `NAME_CONFLICT`：提示“当前目录已存在同名文件或目录，继续上传可能失败或需要改名”。
- `SAME_FILE_CONFLICT`：提示“当前目录可能已有相同文件，是否继续检查并上传？”。
- `DUPLICATE_FILE`：提示“当前目录已存在相同文件，是否仍然上传一份？”。
- 用户取消时不调用 `/upload/pre-check`，也不创建上传任务。
- 用户确认后才继续 `/upload/pre-check -> request-challenge/init`。
- “同一个文件”的最终确认必须以 `/files/check-duplicate` 的完整 SHA-256 检测结果为准。
- 上传模块完成后只拿到 `physical_file_slug`。文件页必须再调用 `/files/import`，把物理文件导入目标目录。
- 即使用户确认，`/files/import` 仍必须执行唯一约束校验；若最终插入同名条目冲突，返回 409，前端提示用户改名。

### 媒体库交互

文件列表中的视频文件提供“加入媒体库”操作：

1. 用户点击加入媒体库。
2. 前端调用 `POST /api/v1/media/items`。
3. 跳转或提示进入媒体库页面。
4. 媒体库卡片展示转码状态和进度。
5. 转码完成后可播放 HLS。

共享转码 UI 规则：

- 如果 `transcode_reused=true` 且状态为 `processing`，前端无需提示“正在为你新建转码任务”，只在媒体卡片上静默显示进度。
- 如果 `transcode_reused=true` 且状态为 `done`，加入媒体库后应立即可播放。
- 同一个物理文件被多个用户加入媒体库时，多个用户看到同一个转码进度，但 HLS 访问仍按各自 `media_slug` 做鉴权。

页面路由：

| 路由 | 页面 |
|------|------|
| `/login` | 登录 / 注册 |
| `/files` | 文件列表 |
| `/files/trash` | 回收站 |
| `/files/starred` | 收藏 |
| `/profile` | 个人中心 |
| `/media` | 媒体库列表 |
| `/media/:slug` | 媒体播放页 |

---

## 执行任务列表

### TASK-01：项目脚手架

创建 Go 后端、SvelteKit 前端、docker-compose、配置文件。

验收：

- PostgreSQL 16 和 Redis 7 能启动。
- `go build ./...` 通过。
- 前端 dev server 能启动。

### TASK-02：数据库迁移

创建全部 DDL，包括 `media_items` 和 `media_jobs`。

验收：

- 迁移可重复执行和回滚。
- 索引、唯一约束、CHECK 约束存在。

### TASK-03：sqlc 查询层

至少包含：

- users
- user_profiles
- user_storage_stats
- user_levels
- user_transactions
- physical_files
- user_files
- upload_tasks
- refresh_tokens
- media_transcodes
- media_items
- media_jobs

验收：

- `sqlc generate` 通过。
- 生成代码可编译。

### TASK-04：本地存储层

实现分片写入、合并、强制哈希校验、读取、删除。

验收：

- 单元测试覆盖 `WriteChunk -> MergeChunks -> HashFile -> ReadAt -> Delete`。
- 合并后哈希不一致时必须失败并清理临时文件。

### TASK-05：Redis 缓存层

实现：

- pre-cache
- challenge 原子取删 Lua
- chunk set
- merge lock
- rate limit
- media progress

验收：

- 并发调用 challenge 取删，只有一个请求成功。
- merge lock 同一 `file_hash` 只能一个持有者。

### TASK-06：认证模块

实现注册、登录、刷新、登出。注册成功时必须同时初始化：

- `users`
- `user_profiles`
- `user_storage_stats`
- `user_levels`

验收：

- refresh token 轮换。
- 旧 refresh token 不能再次使用。
- 新用户注册后能查询到 profile、storage、level 默认记录。

### TASK-06B：用户模块

实现个人信息、头像、密码修改、交易记录查询。

验收：

- `/user/me` 聚合返回 users、user_profiles、user_storage_stats、user_levels。
- 修改展示资料只更新 `user_profiles`。
- 修改密码只更新 `users.password_hash`。
- 上传头像只更新 `user_profiles.avatar_path`。
- `/user/transactions` 从 `user_transactions` 分页读取，不影响 `users` 主表。

### TASK-07：文件目录模块

实现文件列表、目录、上传冲突检查、物理文件导入、重命名、移动、软删除、永久删除、下载。

验收：

- 同目录创建同名目录返回 409。
- 同目录创建同名文件或文件与目录同名返回 409。
- 重命名、移动导致同目录重名时返回 409。
- `/files/check-conflict` 能检测同目录同名条目。
- `/files/check-conflict` 能检测同目录可能相同内容文件，并交给前端确认。
- `/files/check-duplicate` 能基于完整 SHA-256 检测同目录相同文件。
- `/files/import` 能把 `physical_file_slug` 导入网盘目录，并在事务内更新配额和创建 `user_files`。
- Range 下载可用。
- 永久删除最后引用时磁盘文件被清理。

### TASK-08：上传模块

按 v4 上传流程实现。

验收：

- 上传模块不读取或写入 `user_files`。
- 上传模块不接收 `parent_slug`、`file_name` 等目录业务字段。
- 首次上传成功后返回 `physical_file_slug`。
- 同文件秒传成功后返回已有 `physical_file_slug`。
- 伪造 proof 返回 MISS。
- 断点续传跳过已完成分片。
- 并发上传同一文件不会生成多个物理文件。
- 上传完成不更新 `user_storage_stats.storage_used`，配额由 `/files/import` 处理。

### TASK-09：媒体库模块

实现主动加入媒体库、列表、详情、删除、重试、HLS 鉴权访问。

验收：

- 上传视频后不会自动转码。
- 点击加入媒体库后才创建 `media_items`，并查找或创建共享 `media_transcodes`。
- 如果同一物理文件已有完成的 `media_transcodes`，直接复用 HLS，不创建新的 `media_jobs`。
- 如果同一物理文件已有 pending/processing 的 `media_transcodes`，复用同一个转码进度，不创建重复 `media_jobs`。
- 同一用户同一文件重复加入时返回已有媒体条目。
- 删除媒体库条目不删除网盘源文件，也不立即删除共享 HLS 产物。

### TASK-10：FFmpeg worker

实现后台 worker、HLS 输出、进度解析、失败清理。

验收：

- pending job 会被 worker 处理。
- 转码完成后 HLS 可播放。
- 多 worker 不会重复处理同一 job。
- 不同用户加入同一个物理文件时，只产生一份 HLS 目录。

### TASK-11：前端上传与文件页

实现 Web Worker 哈希、上传状态机、断点续传、文件列表。

验收：

- 大文件哈希不阻塞主线程。
- 检测到同目录同名条目时，先弹出确认框；用户取消则不上传。
- 检测到同目录可能已有相同文件时，先弹出确认框；用户确认后才继续上传流程。
- 上传或秒传拿到 `physical_file_slug` 后，前端调用 `/files/import` 创建网盘文件条目。
- 上传进度、合并状态、错误状态展示完整。

### TASK-12：前端媒体库

实现媒体库列表、加入媒体库、转码进度、播放页。

验收：

- 文件列表的视频可加入媒体库。
- 媒体库中展示转码进度。
- 复用其他用户正在处理的转码任务时，UI 静默展示进度，不制造重复任务提示。
- 转码完成后可播放。

### TASK-13：集成测试

| 场景 | 预期 |
|------|------|
| 首次上传物理文件 | 磁盘文件、physical_files 新增，不创建 user_files |
| 上传完成后导入网盘 | `/files/import` 创建 user_files，并更新 user_storage_stats.storage_used |
| 同目录创建同名目录 | 返回 409 |
| 同目录上传同名文件 | 前端弹出确认框，用户确认后继续，最终同名冲突由服务端返回 409 或前端要求改名 |
| 同目录上传相同内容文件 | 前端弹出确认框，用户取消则不上传，用户确认后继续 |
| 第二次上传同一文件 | 秒传成功，返回已有 physical_file_slug，磁盘无新文件 |
| 不同用户上传同一文件后导入 | 一条 physical_files，多条 user_files |
| 网络中断后恢复 | 已完成分片被跳过 |
| 伪造 proof_code | 返回 MISS |
| challenge 超时重放 | 返回 404 或 MISS |
| 并发超配额导入 | 只有满足配额的 `/files/import` 请求成功 |
| 永久删除最后引用 | 物理文件被删除 |
| 上传 MP4 | 不自动转码 |
| 加入媒体库 | 创建 media_item，并创建或复用 media_transcode |
| 不同用户加入同一物理视频 | 复用同一 media_transcode 和 HLS 目录，不重复转码 |
| 已有转码进行中时加入媒体库 | 不创建重复 job，UI 静默展示共享进度 |
| 删除媒体库条目 | 删除当前 media_item，不删除源文件，不立即删除共享 HLS |
| Range 下载 | 返回指定范围字节 |

---

## 安全红线

1. 随机数必须使用 `crypto/rand`。
2. challenge 必须原子取删。
3. challenge key 必须绑定用户 ID。
4. 文件路径参数必须严格正则校验。
5. 下载和 HLS 必须做资源归属校验。
6. 配额增加必须使用数据库原子更新。
7. 合并后必须强制校验完整 SHA-256。
8. `physical_files` 创建必须使用唯一约束和 `ON CONFLICT`。
9. 永久删除物理文件必须在确认无引用后执行，并考虑并发锁。
10. FFmpeg 输入输出路径必须由服务端生成，禁止透传用户路径。
11. HLS wildcard 路径必须规范化并限制在媒体输出目录内。
12. refresh token 只存哈希，不存明文。

---

## MVP 范围建议

第一阶段优先实现：

1. 注册登录。
2. 文件列表和目录。
3. 普通上传、断点续传、下载。
4. 秒传挑战。
5. 主动加入媒体库。
6. FFmpeg HLS 转码和播放。

第二阶段再补：

1. 右键菜单和拖拽移动。
2. 海报截帧。
3. 字幕管理。
4. 媒体元数据刮削。
5. 分享链接。
6. 对象存储适配。
