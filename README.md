# VideoFlow Convert

MP4 → HLS (M3U8) 转换服务。Go (Echo + sqlc + pgx + Redis + FFmpeg Worker Pool) 后端,SvelteKit + Tailwind v4 前端,JWT 认证 + SSE 进度推送。

## 环境要求

- Go 1.22+(本仓库构建于 1.25)
- Node 22 + pnpm 10
- PostgreSQL 16+(本机已用 :15432)
- Redis 7+(本机已用 :16379)
- FFmpeg / FFprobe(在 PATH 中可执行)
- 可选: golang-migrate、sqlc(若不用 docker)

## 快速启动

### 1. 准备数据库 / 缓存

```bash
# 已有外部 PG/Redis 时可跳过 docker。否则:
docker compose up -d
```

如果用现成的 PG,请确保已有 `videoconv` 数据库,并把 `backend/config.yaml` 的 `db.dsn` 改成你的连接串。Redis 端口同理。

### 2. 跑迁移

```bash
cd backend
PATH=$HOME/go/bin:$PATH make migrate-up
```

`make` 优先使用本地 `migrate` 二进制,找不到才回落到 docker 镜像。

### 3. 启动后端

```bash
cd backend
make run        # 或: go run ./cmd/server
```

监听 `:8080`,健康检查 `GET /healthz`。

### 4. 启动前端(开发模式)

```bash
cd frontend
pnpm install
pnpm dev        # http://localhost:5173
```

Vite 将 `/api` 与 `/hls` 反代到 `:8080`,因此前端无需 CORS 配置即可访问后端。

### 5. 验证连通

```bash
cd backend
./bin/dbcheck   # 同时检测 PG 和 Redis;返回非零表示有问题
```

## 主要 API

Base URL: `/api/v1`。所有写接口除 `/auth/*` 外都需 `Authorization: Bearer <access_token>`。

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/register` | 注册 |
| POST | `/auth/login` | 登录,返回 access + refresh token |
| POST | `/auth/refresh` | 用 refresh token 换取新对 |
| POST | `/auth/logout` | refresh 写入 Redis 黑名单 |
| POST | `/upload` | 上传 MP4(multipart),返回 task_id |
| GET | `/tasks` | 当前用户任务列表(分页) |
| GET | `/tasks/:id` | 任务详情 |
| GET | `/tasks/:id/progress` | SSE 实时进度推送 |
| DELETE | `/tasks/:id` | 删除任务与文件 |
| GET | `/videos` | 已完成视频列表 |
| GET | `/videos/:id` | 视频详情(含 m3u8_url) |
| GET | `/hls/:task_id/index.m3u8` | HLS 播放列表(走 JWT) |
| GET | `/hls/:task_id/*.ts` | TS 分片 |
| GET | `/admin/users` 等 | 管理员接口(需 role=admin) |

EventSource / `<video src>` 无法发送自定义 header,因此 `GET` 路由额外接受 `?access_token=<jwt>`。

## 架构

```
SvelteKit  -->  Echo HTTP  -->  PostgreSQL (users / tasks)
              |             |
              |             +-> Redis (task:progress, refresh deny-list, rate limit)
              |
              +-> Worker Pool (size = min(NumCPU, 4))
                     |
                     +-> FFmpeg / FFprobe   --> 本地磁盘 HLS 输出
```

- **Worker Pool**: 固定大小,buffered job channel,panic 隔离,Shutdown 优雅等待。
- **进度推送**: FFmpeg stderr 流式解析 `time=`,每帧写 Redis,每 5% 写一次 DB,SSE 每秒推一次。
- **认证**: HS256 JWT,access 15 分钟,refresh 7 天 + Redis 黑名单。bcrypt cost=12。
- **上传**: 流式落盘(`io.Copy` + `LimitReader`),魔数校验文件类型(不信 Content-Type),上限 2 GiB。

## 测试

```bash
cd backend
make test     # go test -race ./...
```

## 目录结构

```
backend/   # Go 后端
  cmd/server/         # 入口
  cmd/dbcheck/        # 一次性 PG + Redis 连通性检测
  internal/{config,db,handler,middleware,model,service,store,worker}
  pkg/{jwtutil,fileutil}
frontend/  # SvelteKit + Tailwind v4
docker-compose.yml    # 可选: 一键起 PG + Redis
```
