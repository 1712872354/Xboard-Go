# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

Xboard-Go is a proxy panel management system (a Go rewrite of Xboard). Go backend (Gin + GORM + gRPC) with a React 19 + TypeScript + shadcn/ui frontend. The frontend is compiled with Vite and embedded into the Go binary at build time via `go:embed` (`internal/static/`), so a single binary serves both the API and the SPA.

A companion file, `AGENTS.md`, contains deeper detail (node protocol list, frontend component conventions, "what NOT to do"). Read it for anything this file does not cover.

Two sibling projects in the same workspace are relevant to node communication (but are separate repos): `../Xboard-Node-Go` (Go node agent that connects via gRPC) and `../Xboard-Node` (original reference node agent).

## Commands

```bash
# Full build: frontend → copy dist into embed dir → compile Go binary
cd frontend && pnpm install && pnpm run build && cd ..
cp -r frontend/dist/* internal/static/dist/      # (Makefile / build.ps1 automate this)
go build -o bin/xboard-go ./cmd/server/

# Windows one-liner (PowerShell)
.\build.ps1

# Go-only build (frontend dist already present in internal/static/dist/)
go build -o bin/xboard-go ./cmd/server/

# Run (server flag is -config, NOT -c; -c belongs to cmd/migrate)
./bin/xboard-go -config config.yaml

# Frontend dev server (port 3000, proxies /api → localhost:8080)
cd frontend && pnpm run dev

# Lint
golangci-lint run

# Tests
go test ./... -v
go test ./internal/service -run TestName   # single test
```

The frontend build script is `vite build` only — there is no `tsc -b`, so TypeScript errors do **not** block builds.

## Architecture

The backend is a strict layered design. Requests flow `handler → service → repository → model`:

- `cmd/server/main.go` — entrypoint. Startup order: load config → init logger → init DB → `AutoMigrate` → init Redis (optional) → `router.SetupRouter` → start gRPC (if enabled) → start node health-check goroutine → run HTTP.
- `cmd/migrate/main.go` — secondary CLI for manual DB migration/status checks (`-c` config flag, `-action up|down|status`).
- `config/config.go` — Viper-based YAML loading; env vars override with prefix `XBOARD` (`.` → `_`). `config.Get()` returns the loaded singleton.
- `internal/model/` — GORM models, one file per entity (~24 entities).
- `internal/repository/` — data-access layer; every entity has a `*_repo.go` with a `New*Repository()` constructor.
- `internal/service/` — business logic; constructors take their repos as dependencies (manual DI, no framework).
- `internal/handler/` — Gin handlers; one file per domain, matching services 1:1.
- `internal/router/router.go` — **single file registering every route**. All repos/services/handlers are constructed here and wired together. To add an endpoint, follow the existing pattern: construct repo → service → handler, then register the route.
- `internal/middleware/` — CORS, request logger, recovery, rate limiting (Redis-backed, falls back to in-memory), JWT auth, `AdminRequired` RBAC.
- `internal/grpc/` — gRPC server for node communication.
- `internal/scheduler/tasks/` — cron jobs (commission, order, traffic) driven by `pkg/scheduler`.
- `pkg/` — shared utilities: `database`, `jwt`, `redis`, `email`, `payment` (alipay/wechat), `ratelimit`, `response`, `captcha`, `i18n`, `utils`.

Database schema is managed by **GORM `AutoMigrate`** at startup (see the model list in `cmd/server/main.go`), not by running `migrations/*.sql` manually.

## Key conventions

- **No protoc.** gRPC uses hand-written Go structs plus a custom JSON codec (`internal/grpc/codec.go`). Never run `protoc`.
- **API response envelope.** Every REST response is `{ "code": 0, "message": "success", "data": <payload> }`. `code 0` = success. Auth failures return HTTP 200 with `code 401` — the frontend axios interceptor unwraps this and handles token refresh.
- **Frontend API access** goes through the single axios instance at `frontend/src/lib/api.ts`; all hooks import it. It auto-attaches the Bearer token and unwraps the envelope (resolves `res.data`).
- **Frontend path alias.** `@/` maps to `src/` (`vite.config.ts` + `tsconfig.json`).
- **Node communication.** gRPC on port 50051 (configurable via `grpc` config); auth via `app.node_api_key` + `node_id` metadata. Admin handlers push realtime changes to connected nodes through `grpc.NodeBroadcaster.BroadcastConfig()` / `BroadcastUsers()` (wired up in `cmd/server/main.go` via `handler.NotifyNodeConfigChange` / `handler.NotifyUserChange`).
- **Node config** is stored as a `server_info` JSON blob on the `nodes` table; the panel assembles structured fields into it and passes it through to nodes.
- **Frontend stack**: Zustand (`src/stores/`), TanStack Query (`src/hooks/`), react-hook-form + zod, echarts-for-react (not recharts), shadcn/ui with Radix primitives (`src/components/ui/`).

## What NOT to do

- Don't run `protoc` (proto types are hand-written).
- Don't edit `internal/static/dist/` directly — it is populated from `frontend/dist/` at build time.
- Don't use `recharts` (use `echarts-for-react`), `@base-ui/react`, or `framer-motion` — none are in dependencies.
