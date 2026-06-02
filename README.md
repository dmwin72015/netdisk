# NetDisk - 个人/小团队网盘

NetDisk 是一个单机部署的网盘系统，面向个人和小团队使用。当前版本提供文件管理、分片上传、秒传/断点续传、上传任务管理、回收站、收藏、媒体库与 HLS 在线播放等能力。

> 本文档按当前工作区代码整理。历史遗留的 `drive`、`tasks`、`videos`、`admin` 前端 API/页面仍保留在代码中，但当前后端未注册对应路由，已在下文单独标注。

## 当前功能

| 模块 | 当前能力 |
|------|----------|
| 认证 | 注册、登录、刷新 token、登出；JWT access token + refresh token 哈希存储 |
| 文件管理 | 根目录/子目录浏览、面包屑、创建目录、重命名、移动、下载、列表/网格视图、排序、按名称前端搜索 |
| 文件过滤 | 支持按父目录、MIME 前缀、文件分类、仅目录、是否显示系统目录过滤 |
| 上传 | 4 MiB 分片上传、并发上传队列、暂停/恢复/重试、断点续传、上传速度和进度展示 |
| 秒传/去重 | `pre_hash` 预检、完整 SHA-256 挑战验证、物理文件去重、导入到当前目录 |
| 上传任务 | 上传历史分页、状态过滤、日期过滤、失败任务重试、单条/批量删除 |
| 冲突处理 | 上传前同名/疑似重复/完整重复检测，支持用户确认后继续上传或导入 |
| 回收站 | 移入回收站、恢复、永久删除、清空回收站、全部恢复；后台定期清理过期项目 |
| 收藏 | 文件星标/取消星标，收藏页快速访问 |
| 媒体库 | 将视频文件加入媒体库、创建系统上传目录、后台 FFmpeg HLS 转码、封面生成、在线播放 |
| 个人中心 | 资料编辑、头像上传、密码修改、总用量/配额展示、分类用量统计、交易记录 |
| 国际化 | 中文/英文，基于 Paraglide，语言偏好通过 cookie 管理 |
| 客户端配置 | 前端从 `/api/v1/config` 获取分片大小、最大上传大小、头像大小限制 |
| 日志/限流 | zerolog 请求日志、可选滚动文件日志、Redis API/认证限流 |

## 后端部署（非 Docker）

### 1. 编译

```bash
cd backend
make build
# 产物：bin/netdisk-server
```

### 2. 配置

```bash
cp config.example.yaml /etc/netdisk/config.yaml
# 按需修改端口、数据库、JWT secret、存储路径等
```

### 3. 数据库迁移

```bash
make migrate-up
```

### 4. 运行

```bash
./bin/netdisk-server --config /etc/netdisk/config.yaml
```

后端支持以下命令行参数：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--config` | `./config.yaml` | 配置文件路径 |
| `--port` | 配置文件中的 `server.port` | 监听端口，优先级高于配置文件 |

环境变量（优先级高于配置文件）：

- `NETDISK_SERVER_PORT` — 监听端口
- `NETDISK_DB_DSN` — PostgreSQL 连接串
- `NETDISK_REDIS_ADDR` — Redis 地址
- `NETDISK_JWT_SECRET` — JWT 签名密钥
- `NETDISK_STORAGE_ROOT` — 文件存储根目录

### 5. Systemd 服务（可选）

```ini
[Unit]
Description=NetDisk Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=netdisk
ExecStart=/usr/local/bin/netdisk-server --config /etc/netdisk/config.yaml
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

### 6. 生产注意事项

- **JWT secret**：部署前务必修改 `jwt.secret` 为足够长的随机字符串
- **数据库密码**：修改 `db.dsn` 中的密码，并使用强密码
- **日志**：生产环境建议 `log.output: file`，配合日志轮转（如 logrotate）
- **存储路径**：确保 `storage.root` 指向有足够磁盘空间的目录
- **反向代理**：生产环境建议在前置 nginx 或 Caddy 中配置 HTTPS，后端仅监听内网端口
- **健康检查**：`GET /healthz` 返回 `{"status":"ok"}`，可用于容器/负载均衡探活

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.25+、Echo v4、pgx、sqlc、Viper、zerolog |
| 数据库 | PostgreSQL 16 |
| 缓存 | Redis 7 |
| 前端 | SvelteKit 2、Svelte 5、TypeScript、Tailwind CSS v4、Bits UI |
| 上传 | Web Worker SHA-256、分片上传、Redis 分片状态/锁 |
| 媒体 | FFmpeg / FFprobe、HLS、hls.js |
| 测试 | Go test、Vitest、Svelte Check |

## 环境要求

- Go 1.25+（以 `backend/go.mod` 为准）
- Node 22+ 与 pnpm
- PostgreSQL 16（默认映射到 `localhost:15432`）
- Redis 7（默认映射到 `localhost:16379`）
- FFmpeg / FFprobe（媒体转码需要，需在 `PATH` 中）
- 可选：`migrate`、`sqlc`；未安装时 Makefile 部分命令会回退到 Docker 镜像

## 快速启动

系统提供两种部署方案，可按需选择。

### 方案一：三容器部署（推荐，nginx 反代）

前后端分离，各司其职，易于独立扩缩。

```
┌──────────┐   /api /hls /avatars   ┌──────────┐
│  nginx:80 │ ─────────────────────→ │ backend  │
│          │                         │ :8070    │
│          │   其他路由               ├──────────┤
│          │ ─────────────────────→ │ frontend │
└──────────┘                         │ :8090    │
                              (adapter-node, SSR)
```

```bash
docker compose up -d --build
```

访问 `http://localhost:8080`。

| 服务 | 容器 | 镜像 | 说明 |
|------|------|------|------|
| nginx | `netdisk_nginx` | `nginx:1.27-alpine` | 统一入口，反代 API/HLS 到后端，其余到前端 |
| 前端 | `netdisk_frontend` | `netdisk-frontend:local` | SvelteKit Node 服务（adapter-node），SSR |
| 后端 | `netdisk_backend` | `netdisk-backend:local` | Go Echo API 服务 |

启动文件：
- `docker-compose.yml` — 编排定义
- `nginx/default.conf` — nginx 反代规则
- `frontend/Dockerfile` — 前端容器构建（node:22-alpine）
- `backend/Dockerfile` — 后端容器构建（golang:1.25-alpine → alpine:3.22）

### 方案二：单容器部署（前端嵌入 Go 二进制）

前端使用 `adapter-static` 构建为 SPA，通过 `go:embed` 编译进后端，简化运维。

```
┌───────────────────────────┐
│       single container    │
│  ┌─────────────────────┐  │
│  │   Go 二进制          │  │
│  │   ├─ API handler     │  │
│  │   └─ embedded SPA    │  │
│  └─────────────────────┘  │
└───────────────────────────┘
```

```bash
docker compose -f docker-compose.single.yml up -d --build
```

访问 `http://localhost:8080`。

| 服务 | 容器 | 说明 |
|------|------|------|
| 后端 | `netdisk_server` | 单体二进制，内嵌前端静态文件，同时处理 API 和 SPA 路由 |

启动文件：
- `docker-compose.single.yml` — 编排定义
- `backend/Dockerfile.single` — 多阶段构建（先 pnpm build → 再 go build）
- `frontend/svelte.config.static.js` — adapter-static 配置（SPA fallback）
- `backend/internal/web/embed.go` — `go:embed` 将前端产物嵌入二进制

### 方案对比

| 维度 | 方案一（nginx） | 方案二（单容器） |
|------|----------------|-----------------|
| 组件数 | 3 容器 | 1 容器 |
| 前端渲染 | SSR（Node） | CSR（SPA fallback） |
| 部署复杂度 | 中等 | 低 |
| 独立升级前后端 | ✓ | 需整体构建 |
| 静态文件性能 | nginx 高效处理 | Go http.FileServer |
| HTTPS 终止 | nginx 直接支持 | 需额外反向代理 |

### 本地开发

#### 1. 启动基础设施

```bash
docker compose up -d postgres redis
```

默认地址：
- PostgreSQL：`postgres://postgres:root1234@localhost:15432/netdisk?sslmode=disable`
- Redis：`localhost:16379`

#### 2. 配置后端

```bash
cp backend/config.example.yaml backend/config.yaml
```

按需修改数据库、Redis、JWT secret、存储根目录等。

#### 3. 运行数据库迁移

```bash
cd backend && make migrate-up
```

#### 4. 启动后端

```bash
cd backend && make run
```

健康检查：

```bash
curl http://localhost:8080/healthz
# {"status":"ok"}
```

#### 5. 启动前端

```bash
cd frontend && pnpm install && pnpm dev
```

前端开发服务运行在 `http://localhost:5173`，Vite 将 `/api` 反代到 `http://localhost:8080`。

## 项目结构

```text
backend/
  cmd/server/                 # 后端入口
  internal/
    app/                      # 应用初始化、路由、中间件注册
    cache/                    # Redis 封装：challenge、分片、锁、预检、转码进度
    config/                   # YAML 配置加载
    db/
      migrations/             # 数据库迁移
      query/                  # sqlc 查询源文件
      sqlc/                   # sqlc 生成代码
    handler/                  # HTTP handler
    logging/                  # 控制台/文件日志与滚动写入
    media/                    # FFmpeg 转码 worker
    middleware/               # JWT、请求日志、Redis 限流
    model/                    # 统一错误与客户端配置模型
    service/                  # 认证、文件、上传、媒体、用户、回收站 worker
    storage/                  # 本地物理文件/分片存储
    store/                    # PostgreSQL 连接
    web/                      # go:embed 前端静态文件（方案二）
  pkg/
    fileutil/                 # MIME/分类等文件工具
    jwtutil/                  # JWT 生成与校验
frontend/
  svelte.config.js             # adapter-node 配置（方案一）
  svelte.config.static.js      # adapter-static 配置（方案二）
  messages/                    # Paraglide 源消息
  src/
    lib/
      api/                    # 当前/遗留 API 客户端
      components/             # Navbar、文件、媒体、账号等业务组件
      stores/                 # auth/config store
      ui/                     # Dialog、Drawer、Dropdown、Popover 等 UI 基础组件
      workers/                # SHA-256 Web Worker
      upload-*.ts             # 上传队列、策略、并发、断点续传工具
    routes/
      (public)/               # 登录、注册
      (protected)/            # 登录后页面：主页、文件、媒体、任务、账号等
docker-compose.yml             # 方案一：nginx + frontend + backend + PostgreSQL + Redis
docker-compose.single.yml      # 方案二：单容器（前端嵌入 Go 二进制）+ PostgreSQL + Redis
backend/Dockerfile              # 后端容器构建（独立）
backend/Dockerfile.single       # 单容器构建（前端 + 后端）
backend/config.docker.yaml      # 后端容器配置
nginx/default.conf              # nginx 反代配置（方案一，HTTP 80）
data/                           # 默认本地存储根目录
```

## 前端路由

| 路径 | 状态 | 说明 |
|------|------|------|
| `/login` | 可用 | 登录页 |
| `/register` | 可用 | 注册页，注册成功后自动登录 |
| `/` | 可用 | 仪表盘：欢迎卡片、存储概览、快捷入口、最近文件 |
| `/files/all` | 可用 | 文件浏览器根目录 |
| `/files/all/[slug]` | 可用 | 文件浏览器子目录，由同一个布局统一渲染 |
| `/files/starred` | 可用 | 收藏文件列表 |
| `/files/trash` | 可用 | 回收站列表、恢复/删除/清空 |
| `/media` | 可用 | 媒体库列表、添加视频到媒体库 |
| `/media/[slug]` | 可用 | HLS 播放页与转码状态展示 |
| `/tasks` | 可用 | 上传任务历史、筛选、重试、删除 |
| `/account` | 可用 | 个人资料、头像、密码、存储统计 |
| `/admin` | 前端存在，后端未接通 | 管理用户/角色/配额 UI 存在，但当前后端未注册 `/api/v1/admin/*` |
| `/videos`、`/videos/[id]` | 遗留 | 依赖旧 `/api/v1/videos/*`，当前后端未注册 |

## API 概览

Base URL：`/api/v1`。除注册、登录、刷新 token、健康检查和静态头像文件外，请求均需要 `Authorization: Bearer <access_token>`。

响应统一包裹在 `data` 字段中；错误响应由 `code`、`message` 等字段描述。

### 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/healthz` | 服务健康检查 |

### 认证 `/api/v1/auth`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/register` | 注册用户 |
| POST | `/auth/login` | 登录，返回用户信息与 access/refresh token |
| POST | `/auth/refresh` | 使用 refresh token 刷新 token |
| POST | `/auth/logout` | 登出并吊销 refresh token，需要 JWT |

### 用户 `/api/v1/user` 与头像

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/user/me` | 当前用户信息、profile、storage、level |
| PATCH | `/user/profile` | 更新 `displayName`、`bio`、`avatarUrl` |
| POST | `/user/me/password` | 修改密码 |
| POST | `/user/me/avatar` | 上传头像，支持 JPEG/PNG/WebP |
| GET | `/user/storage-breakdown` | 按文件分类统计存储占用 |
| GET | `/user/transactions` | 用户交易/配额记录，支持分页 |
| GET | `/avatars/*` | 静态头像文件服务 |

### 文件 `/api/v1/files`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/files` | 文件列表；支持 `page`、`pageSize`、`parentSlug`、`mimeType`、`fileCategory`、`sortBy`、`sortDir`、`onlyDirs`、`includeSystem` |
| GET | `/files/recent` | 最近文件，支持 `limit` |
| GET | `/files/trash` | 回收站列表，支持分页 |
| GET | `/files/starred` | 收藏文件列表，支持分页 |
| GET | `/files/:slug/breadcrumb` | 文件/目录面包屑 |
| GET | `/files/:slug/download` | 下载文件，支持 Range |
| POST | `/files/mkdir` | 创建目录 |
| POST | `/files/check-conflict` | 上传前同名/疑似重复检测 |
| POST | `/files/check-duplicate` | 完整 SHA-256 重复检测 |
| POST | `/files/import` | 将物理文件导入为用户文件 |
| DELETE | `/files/:slug` | 移入回收站 |
| POST | `/files/:slug/restore` | 从回收站恢复 |
| DELETE | `/files/:slug/permanent` | 永久删除 |
| POST | `/files/:slug/rename` | 重命名 |
| POST | `/files/:slug/move` | 移动到目标目录 |
| POST | `/files/:slug/star` | 收藏/取消收藏 |
| POST | `/files/trash/empty` | 清空回收站 |
| POST | `/files/trash/restore-all` | 恢复全部回收站文件 |

### 上传 `/api/v1/upload`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/upload/pre-check` | 使用 `preHash` 和文件大小做低成本预检 |
| POST | `/upload/request-challenge` | 为完整 SHA-256 申请秒传挑战 |
| POST | `/upload/verify` | 验证挑战并返回可秒传的物理文件引用 |
| POST | `/upload/init` | 创建/恢复上传任务 |
| POST | `/upload/chunk` | 上传分片，使用 multipart/form-data |
| POST | `/upload/complete` | 合并分片、校验 SHA-256、生成物理文件 |
| POST | `/upload/update-hash` | 为上传任务补充/更新完整哈希 |
| GET | `/upload/:upload_slug/status` | 查询上传任务状态 |
| GET | `/upload/tasks` | 上传任务历史；支持 `limit`、`offset`、`startDate`、`endDate`、`status` |
| POST | `/upload/tasks/:slug/retry` | 重试失败上传任务 |
| DELETE | `/upload/tasks/:slug` | 删除单条上传任务 |
| DELETE | `/upload/tasks` | 批量删除上传任务 |

### 客户端配置 `/api/v1/config`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/config` | 获取 `upload.chunkSize`、`upload.maxUploadSize`、`avatar.maxSize` |

### 媒体库 `/api/v1/media`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/media/upload-dir` | 确保并返回“媒体上传”系统目录 |
| POST | `/media/items` | 将视频文件加入媒体库并触发/复用转码任务 |
| GET | `/media/items` | 媒体库列表，支持分页 |
| GET | `/media/items/:media_slug` | 媒体详情、转码状态与进度 |
| DELETE | `/media/items/:media_slug` | 从媒体库移除 |
| GET | `/media/poster/:media_slug` | 获取封面图 |
| GET | `/media/hls/:media_slug/*` | 获取 HLS playlist/segment，需鉴权 |

### 当前未注册的历史 API

以下前端客户端或页面仍在仓库中，但当前 `backend/internal/app/router.go` 没有注册对应后端路由：

- `/api/v1/admin/*`
- `/api/v1/drive/*`
- `/api/v1/tasks/*`
- `/api/v1/videos/*`
- 旧的 `POST /api/v1/upload` 单接口上传

如需恢复这些模块，需要补齐后端 handler/service/router，或删除/迁移对应前端遗留代码。

## 上传与秒传流程

```text
选择文件
  ↓
计算 pre_hash → POST /upload/pre-check
  ├─ SUSPECT_HIT
  │   ↓
  │ 计算完整 SHA-256 → request-challenge → verify
  │   ├─ HIT  → POST /files/import                    # 秒传导入
  │   └─ MISS → init → chunk* → complete → import      # 正常分片上传
  └─ NOT_FOUND
      ↓
    计算完整 SHA-256 → init → chunk* → complete → import
```

关键点：

- 上传服务只负责生成/复用 `physical_files`，不会直接创建用户目录条目。
- `/files/import` 负责配额检查、目录唯一性检查，并创建 `user_files`。
- 前端上传队列支持文件级并发、分片级上传、暂停/恢复、失败重试和中断续传。
- Redis 用于保存分片状态、challenge、pre-check 缓存和合并锁。

## 文件存储

默认存储根目录为 `data/`，当前布局：

```text
data/
├── files/ab/cd/<sha256>       # 物理文件，按 SHA-256 前 2 字节拆分目录
├── tmp/                       # 上传临时分片
├── avatars/                   # 用户头像
└── hls/                       # HLS 转码产物和封面
```

- 相同 SHA-256 的物理文件只保存一份，可被多个用户文件引用。
- 合并完成后会强制校验 SHA-256。
- 启动时会自动创建存储目录，并兼容迁移旧布局 `data/ab/cd/<hash>` 到 `data/files/ab/cd/<hash>`。
- 文件分类基于 MIME type 归类为 `video`、`audio`、`image`、`document`、`archive`、`other`。

## 媒体库流程

- 用户上传视频后，需要主动在前端加入媒体库才会触发转码。
- 媒体库会为用户创建系统目录（可在文件设置中显示/隐藏）。
- 转码产物按物理文件和 profile 去重，多个用户可复用同一份 HLS。
- FFmpeg worker 后台轮询待处理任务，生成 HLS 与 poster。
- 转码进度写入 Redis，媒体详情页轮询状态并在完成后播放。

## 后台任务

- `media worker`：按 `media.poll_interval` 和 `media.batch_size` 处理转码队列。
- `trash worker`：按 `trash.poll_interval` 扫描过期回收站项目，超过 `trash.retention_days` 后清理。

## 配置

主要配置在 `backend/config.yaml`，可参考 `backend/config.example.yaml`：

```yaml
server:
  port: 8080
  cors_origins:
    - http://localhost:5173

db:
  dsn: "postgres://postgres:root1234@localhost:15432/netdisk?sslmode=disable"

redis:
  addr: "localhost:16379"

jwt:
  secret: "haliluya"
  access_ttl_min: 60
  refresh_ttl_hour: 168

storage:
  root: "../data"
  max_upload_size: 4294967296  # 4 GiB
  tmp_dir: "tmp"
  files_dir: "files"
  avatars_dir: "avatars"
  hls_dir: "hls"

ffmpeg:
  path: "ffmpeg"
  ffprobe_path: "ffprobe"

log:
  level: "trace"
  output: "console"           # console 或 file
  file_path: "./logs/server.log"
  max_size_mb: 5

limits:
  default_storage_quota: 536870912000  # 500 GB
  bcrypt_cost: 12
  avatar_max_size: 2097152             # 2 MiB

upload:
  chunk_size: 4194304                  # 4 MiB
  task_expiry_days: 30
  merge_lock_ttl: "10m"

rate_limit:
  api_requests_per_min: 600
  auth_requests_per_min: 20
```

## 常用命令

### 后端

```bash
cd backend
make run                  # 启动服务
make build                # 构建 bin/server
make migrate-up           # 执行迁移
make migrate-down         # 回滚一次迁移
make sqlc-gen             # 重新生成 sqlc 代码
make build-frontend-static # 构建前端静态文件到 internal/web/build（方案二用）
make build-single         # 构建单容器镜像
make test                 # Go 测试
make tidy                 # go mod tidy
```

### 前端

```bash
cd frontend
pnpm dev          # 开发服务
pnpm build        # 构建
pnpm preview      # 预览构建产物
pnpm check        # Svelte/TypeScript 检查
pnpm test         # Vitest 测试
```

## 数据库迁移概览

当前迁移包含：

- 初始用户、资料、配额、文件、上传任务、refresh token、媒体转码和媒体库表。
- 移除部分外键约束，便于异步清理/引用管理。
- 添加文件分类、父目录 slug、上传任务文件名、上传任务父目录 slug。
- 默认配额从 10 GiB 调整为 500 GiB，默认等级调整为 `VIP1`。
- 添加系统目录标记与系统目录唯一约束。

## 安全与一致性

- JWT access token + refresh token，refresh token 仅保存哈希并支持吊销。
- 上传 challenge 使用 Redis Lua 原子取删，减少重放风险。
- 路径和存储 key 使用受控生成，下载、HLS、头像均做归属或类型校验。
- 头像上传限制 MIME 与大小，支持 JPEG/PNG/WebP。
- 上传合并后强制 SHA-256 校验，Redis 锁防止并发合并。
- 配额更新通过数据库事务处理，避免并发突破。
- API 与认证接口均有 Redis 限流。

## 已知待办

- 后端尚未注册 `/api/v1/admin/*`，但前端 `/admin` 页面和 `frontend/src/lib/api/admin.ts` 已存在。
- `frontend/src/lib/api/drive.ts`、`tasks.ts`、`videos.ts` 以及 `/videos` 页面属于旧接口遗留，当前不可直接使用。
- 根目录 `.gitignore` 或开发环境中可能会出现本地运行产物（如 `backend/server.pid`、Go build cache、日志文件），提交前应按团队约定清理。
