# AGENTS.md — Xboard-Go

## Project overview

Proxy panel management system. Go backend (Gin + GORM + gRPC) with React frontend (Vite + shadcn/ui). Frontend is embedded into the Go binary at build time via `go:embed`.

Two companion projects in the same workspace:
- **Xboard-Node-Go** (`../Xboard-Node-Go`): Node agent that connects via gRPC
- **Xboard-Node** (`../Xboard-Node`): Original reference node agent (read for protocol/feature reference)

## Build commands

```bash
# Full build (frontend → copy to embed → compile Go)
cd frontend && pnpm install && pnpm run build && cd ..
Remove-Item -Recurse -Force internal\static\dist\* -ErrorAction SilentlyContinue
Copy-Item -Path frontend\dist\* -Destination internal\static\dist\ -Recurse -Force
go build -o bin\xboard-go.exe .\cmd\server\

# Windows one-liner (PowerShell)
.\build.ps1

# Go only (if frontend dist already in internal/static/dist/)
go build -o bin\xboard-go.exe .\cmd\server\

# Lint
golangci-lint run

# Test
go test ./... -v
```

## Architecture

```
cmd/server/main.go          — Entry point: config → logger → DB → Redis → router + gRPC
config/config.go             — Viper-based YAML config loading
internal/
  model/                     — GORM models (24 files, one per entity)
  repository/                — Data access layer (GORM queries)
  service/                   — Business logic layer
  handler/                   — Gin HTTP handlers (REST API)
  router/router.go           — All route registration (single file, ~500 lines)
  middleware/                — CORS, JWT auth, rate limiting, RBAC
  grpc/                      — gRPC server for node communication
    server.go                — Server lifecycle, global NodeBroadcaster
    handler.go               — Handshake, Stream, GetConfig, GetUsers RPCs
    proto.go                 — Manual protobuf types (no protoc, JSON codec)
    broadcaster.go           — Per-node event channels for config/user push
    metrics.go               — In-memory node metrics cache
    codec.go                 — Custom JSON codec for gRPC
    auth.go                  — gRPC auth interceptor (apikey + node_id metadata)
  static/                    — go:embed frontend dist
pkg/                         — Shared utilities (database, jwt, redis, email, etc.)
```

## Key conventions

- **No protoc**: gRPC uses hand-written Go structs + custom JSON codec (`internal/grpc/codec.go`). Never run `protoc`.
- **API response envelope**: All REST responses use `{ "code": 0, "message": "success", "data": <payload> }`. Frontend unwraps automatically.
- **Frontend API**: Single axios instance at `frontend/src/lib/api.ts`. All hooks use it. Auth token auto-attached.
- **Frontend path alias**: `@/` maps to `src/` (configured in vite.config.ts + tsconfig.json).
- **Node communication**: gRPC on port 50051 (configurable). Auth via `node_api_key` in config.yaml. Node sends `authorization` + `node_id` in gRPC metadata.
- **Broadcast pattern**: Admin REST handlers call `grpc.NodeBroadcaster.BroadcastConfig()` / `BroadcastUsers()` to push changes to connected nodes via gRPC Stream.

## Frontend stack specifics

- **UI**: shadcn/ui v2/v3 (Radix primitives). Components in `src/components/ui/`. Uses `asChild` pattern.
- **Forms**: react-hook-form + zod + @hookform/resolvers
- **Data**: TanStack Query (useQuery/useMutation). All hooks in `src/hooks/`.
- **State**: Zustand stores in `src/stores/` (auth, theme, locale)
- **Charts**: echarts-for-react (ECharts wrapper)
- **i18n**: Prepared but not fully wired. Locale JSON files in `src/locales/`.

## Protocol support

12 protocols: vmess, vless, trojan, shadowsocks, hysteria, hysteria2, tuic, anytls, naive, socks, http, mieru.

Node config is stored as `server_info` JSON blob on the `nodes` table. Frontend assembles structured fields into this JSON. The panel passes it through to nodes via gRPC `NodeConfig.ServerInfo`.

## What NOT to do

- Don't run `protoc` — all proto types are hand-written
- Don't modify `internal/static/dist/` directly — it's populated from `frontend/dist/` during build
- Don't use `@base-ui/react` — project uses shadcn/ui v2/v3 with Radix primitives
- Don't add `framer-motion` — not in dependencies
- Don't use `recharts` — use `echarts-for-react` for charts
- The `build` npm script is `vite build` only (no `tsc -b`) — TypeScript errors don't block builds
