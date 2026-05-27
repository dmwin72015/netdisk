---
name: project-infra
description: Backend and frontend service ports, build commands, and infrastructure layout for videoFlowConvert project
metadata:
  type: project
---

Backend: Go server built with `make build` (produces `bin/server`), listens on port 8080.
PostgreSQL on localhost:15432, Redis on localhost:16379.
Frontend: Vite + pnpm dev server, default port 5173 (falls back to 5174 if occupied).
Database migrations are managed separately (Prisma or similar); confirmed up-to-date as of 2026-05-27.

**Why:** These are the standard ports and build commands used in local development.
**How to apply:** When restarting services, check for port conflicts on 8080 and 5173 first.
