# NetDisk - 小型网盘系统

单机部署的个人/小团队网盘，支持文件管理、分片上传、秒传、断点续传、视频媒体库与 HLS 在线播放。

## 功能特性

| 功能 | 说明 |
|------|------|
| 文件管理 | 目录树、列表/网格视图、排序、搜索、重命名、移动 |
| 分片上传 | 4MB 分片，支持暂停/恢复/重试/并发队列 |
| 秒传 (SHA-256) | 预检 → 挑战 → 验证，服务端存量文件直接导入 |
| 上传任务 | 历史记录、状态追踪、失败重试 |
| 回收站 | 单文件还原、全部还原、清空（原子存储回收） |
| 收藏 | 星标文件快速访问 |
| 媒体库 | 添加视频到媒体库，自动 HLS 转码 + 在线播放 |
| 个人中心 | 头像、资料编辑、密码修改、存储用量（按类别细分） |
| 管理后台 | 用户管理、角色分配、存储配额配置 |
| 国际化 | 英文 / 中文（Paraglide） |
| 存储分类 | 文件自动归类：视频/音频/图片/文档/压缩包/其他 |

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.22+ / Echo v4 / sqlc / pgx |
| 数据库 | PostgreSQL 16 |
| 缓存 | Redis 7 |
| 前端 | SvelteKit + Svelte 5 + Tailwind v4 |
| 媒体处理 | FFmpeg (HLS 转码 + 封面图提取) |
| 认证 | JWT (access 15min + refresh 7d)，前端自动刷新 |

## 环境要求

- Go 1.22+
- Node 22 + pnpm
- PostgreSQL 16
- Redis 7
- FFmpeg / FFprobe （需在 PATH 中）
- 可选: golang-migrate、sqlc （本地二进制，否则回落到 Docker）

## 快速启动

### 1. 启动数据库和缓存

使用本地已有的 PostgreSQL （端口 15432） 和 Redis （端口 16379），或通过 Docker：

```bash
docker compose up -d
```

### 2. 运行数据库迁移

```bash
cd backend
make migrate-up
```

### 3. 启动后端

```bash
cd backend
make run          # 或: go run ./cmd/server
```

监听 `:8080`，健康检查 `GET /healthz`。

### 4. 启动前端

```bash
cd frontend
pnpm install
pnpm dev          # http://localhost:5173
```

Vite 将 `/api` 反代到 `:8080`。

## 项目结构

```
backend/
  cmd/server/              # 入口
  internal/
    app/                   # Echo 初始化、路由注册
    cache/                 # Redis 缓存 (预判、challenge、分片、锁、限流、进度)
    config/                # 配置加载
    db/
      migrations/          # SQL 迁移文件
      query/               # sqlc 查询源文件
      sqlc/                # sqlc 生成代码
    handler/               # HTTP 处理器
    logging/               # 日志
    media/                 # FFmpeg 转码 worker
    middleware/            # JWT 鉴权、限流、日志
    model/                 # 错误定义
    repo/                  # 数据仓库层
    service/               # 业务逻辑
    storage/               # 本地文件存储
    store/                 # 数据库连接
  pkg/
    fileutil/              # 文件工具
    jwtutil/               # JWT 管理
frontend/
  src/
    lib/
      api/                 # API 客户端
      components/          # UI 组件
      stores/              # 状态管理
      workers/             # Web Worker (SHA-256 哈希)
    routes/
      +page.svelte         # 仪表盘首页
      login/               # 登录
      register/            # 注册
      files/               # 文件列表
      files/starred/       # 收藏
      files/trash/         # 回收站
      drive/               # 网盘主页（旧版，待迁移）
      media/               # 媒体库
      media/[slug]/        # 媒体播放
      tasks/               # 上传任务管理
      account/             # 个人中心
      admin/               # 管理后台
      videos/              # 视频库（旧版，待迁移）
docker-compose.yml         # PostgreSQL + Redis
```

## API 接口

Base URL: `/api/v1`。除 `/auth/*` 外所有写接口需 `Authorization: Bearer <access_token>`。

### 认证 `/api/v1/auth`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/register` | 注册 |
| POST | `/auth/login` | 登录 |
| POST | `/auth/refresh` | 刷新 token |
| POST | `/auth/logout` | 登出 （吊销 refresh token） |

### 用户 `/api/v1/user`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/user/me` | 个人信息 （含 profile、storage、level） |
| PATCH | `/user/profile` | 更新展示资料 |
| POST | `/user/me/password` | 修改密码 |
| POST | `/user/me/avatar` | 上传头像 |
| GET | `/user/avatar/:slug` | 获取头像 （公开） |
| GET | `/user/storage-breakdown` | 存储用量按类别细分 |
| GET | `/user/transactions` | 交易记录 （分页） |

### 文件 `/api/v1/files`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/files` | 文件列表 （支持排序、按 mimeType/类别过滤、parentSlug、分页） |
| GET | `/files/recent` | 最近文件（跨目录） |
| GET | `/files/trash` | 回收站列表 |
| GET | `/files/starred` | 收藏列表 |
| GET | `/files/:slug/breadcrumb` | 面包屑路径 |
| GET | `/files/:slug/download` | 下载 （支持 Range） |
| POST | `/files/mkdir` | 创建目录 |
| POST | `/files/check-conflict` | 上传前冲突检测 |
| POST | `/files/check-duplicate` | 完整哈希重复检测 |
| POST | `/files/import` | 导入物理文件到网盘 |
| POST | `/files/trash/empty` | 清空回收站 |
| POST | `/files/trash/restore-all` | 恢复全部 |
| DELETE | `/files/:slug` | 移入回收站 |
| POST | `/files/:slug/restore` | 从回收站恢复 |
| DELETE | `/files/:slug/permanent` | 永久删除 |
| POST | `/files/:slug/rename` | 重命名 |
| POST | `/files/:slug/move` | 移动 |
| POST | `/files/:slug/star` | 收藏/取消收藏 |

### 上传 `/api/v1/upload`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/upload/pre-check` | 低成本预判 （pre_hash） |
| POST | `/upload/request-challenge` | 秒传挑战 |
| POST | `/upload/verify` | 验证秒传 |
| POST | `/upload/init` | 创建/恢复上传任务 |
| POST | `/upload/chunk` | 上传分片 |
| POST | `/upload/complete` | 触发合并 |
| POST | `/upload/update-hash` | 上传完成后更新哈希 |
| GET | `/upload/:slug/status` | 查询上传状态 |
| GET | `/upload/tasks` | 上传历史任务列表 （分页） |
| POST | `/upload/tasks/:slug/retry` | 重试失败任务 |

### 媒体库 `/api/v1/media`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/media/items` | 加入媒体库 |
| GET | `/media/items` | 媒体库列表 （分页） |
| GET | `/media/items/:slug` | 媒体详情 （含转码进度） |
| DELETE | `/media/items/:slug` | 从媒体库移除 |
| GET | `/media/poster/:slug` | 封面图 |
| GET | `/media/hls/:slug/*` | HLS 流 （鉴权） |

### 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/healthz` | 健康检查 |

### 前端页面路由

| 路径 | 说明 |
|------|------|
| `/` | 仪表盘：欢迎 + 存储概览 + 快捷入口 + 最近文件 |
| `/files/all` | 文件浏览器：面包屑、搜索、排序、列表/网格、上传、预览 |
| `/files/starred` | 收藏文件 |
| `/files/trash` | 回收站管理 |
| `/media` | 媒体库：封面网格、状态标识、播放入口 |
| `/media/[slug]` | HLS 视频播放器 |
| `/tasks` | 上传任务：状态追踪、失败重试 |
| `/account` | 个人中心：资料、头像、密码、存储用量 |
| `/admin` | 管理后台：用户管理、配额配置 （仅 admin） |

## 核心设计

### 上传流程

```
前端计算 pre_hash → /upload/pre-check
  ├─ SUSPECT_HIT → 计算完整 SHA-256 → /upload/request-challenge → /upload/verify
  │    ├─ HIT → 拿到 physical_file_slug → /files/import （秒传）
  │    └─ MISS → /upload/init → 分片上传 → /upload/complete → /files/import
  └─ NOT_FOUND → 计算完整 SHA-256 → /upload/init → 分片上传 → /upload/complete → /files/import
```

- 上传模块只负责产生 `physical_files`，不创建网盘文件条目
- `/files/import` 负责配额检查、目录唯一约束、创建 `user_files`

### 文件存储

SHA-256 前 4 位拆两级目录，文件名为完整哈希，无扩展名：

```
data/
├── ab/cd/abcdef1234...    # 物理文件
├── tmp/                   # 上传临时分片
├── avatars/               # 用户头像
└── hls/                   # HLS 转码产物
```

- 相同哈希的物理文件在服务端仅存一份，多用户共享
- 合并后强制 SHA-256 校验
- 分布式 Redis 锁防止并发合并

### 媒体库

- 上传视频后**不自动转码**，用户主动加入媒体库才触发
- 转码产物按物理文件哈希去重，多用户共享同一份 HLS
- FFmpeg 提取封面图（视频 10% 位置）
- 转码进度通过 Redis 实时推送，前端 3s 轮询
- 后台 worker 每 5s 轮询待处理任务

### 上传任务管理

- 每次上传任务在 `upload_tasks` 表记录完整生命周期
- 状态流转：`created` → `uploading` → `merging` → `done` / `failed`
- 失败任务可一键重试（生成新 slug 跳转上传页）
- 分页列表展示所有历史任务

### 存储分类

文件自动按 MIME type 归类为：video、audio、image、document、archive、other。
个人中心展示按类别的用量占比条形图。

## 测试

```bash
# 后端
cd backend && make test

# 前端
cd frontend && pnpm test
```

## 配置

`backend/config.yaml`:

```yaml
server:
  port: 8080
db:
  dsn: "postgres://postgres:root1234@localhost:15432/netdisk?sslmode=disable"
redis:
  addr: "localhost:16379"
storage:
  root: "./data"
  max_upload_size: 2147483648  # 2 GiB
ffmpeg:
  path: "ffmpeg"
```

## 安全

- 随机数使用 `crypto/rand`
- challenge 原子取删 （Lua 脚本）
- 文件路径严格正则校验
- 下载和 HLS 做资源归属校验
- 配额原子更新，防止并发突破
- 合并后强制 SHA-256 校验
- refresh token 只存哈希
- 头像上传限制 JPEG/PNG/WebP
