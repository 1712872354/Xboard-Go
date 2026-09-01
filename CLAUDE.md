# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此仓库中工作时提供指导。

## 项目概述

Xboard-Go 是一个代理面板管理系统（Xboard 的 Go 重写版）。Go 后端（Gin + GORM + gRPC）+ React 19 + TypeScript + shadcn/ui 前端。前端通过 Vite 编译，并通过 `go:embed`（`internal/static/`）嵌入到 Go 二进制文件中，因此单个二进制文件同时提供 API 和 SPA 服务。

关联文件 `AGENTS.md` 包含更详细的信息（节点协议列表、前端组件约定、"禁止事项"）。本文件未覆盖的内容请参考它。

同工作区下的两个兄弟项目与节点通信相关（但它们是独立仓库）：`../Xboard-Node-Go`（通过 gRPC 连接的 Go 节点代理）和 `../Xboard-Node`（原始参考节点代理）。

## 命令

```bash
# 完整构建：前端 → 复制 dist 到嵌入目录 → 编译 Go 二进制文件
cd frontend && pnpm install && pnpm run build && cd ..
cp -r frontend/dist/* internal/static/dist/      # (Makefile / build.ps1 自动化此步骤)
go build -o bin/xboard-go ./cmd/server/

# Windows 一键构建（PowerShell）
.\build.ps1

# 仅 Go 构建（前端 dist 已存在于 internal/static/dist/）
go build -o bin/xboard-go ./cmd/server/

# 运行（服务器参数是 -config，不是 -c；-c 属于 cmd/migrate）
./bin/xboard-go -config config.yaml

# 前端开发服务器（端口 3000，代理 /api → localhost:8080）
cd frontend && pnpm run dev

# 代码检查
golangci-lint run

# 测试
go test ./... -v
go test ./internal/service -run TestName   # 单个测试
```

前端构建脚本仅是 `vite build` — 没有 `tsc -b`，所以 TypeScript 错误**不会**阻止构建。

## 架构

后端是严格的分层设计。请求流向 `handler → service → repository → model`：

- `cmd/server/main.go` — 入口点。启动顺序：加载配置 → 初始化日志 → 初始化数据库 → `AutoMigrate` → 初始化 Redis（可选）→ `router.SetupRouter` → 启动 gRPC（如启用）→ 启动节点健康检查协程 → 运行 HTTP。
- `cmd/migrate/main.go` — 辅助 CLI，用于手动数据库迁移/状态检查（`-c` 配置参数，`-action up|down|status`）。
- `config/config.go` — 基于 Viper 的 YAML 加载；环境变量以 `XBOARD` 为前缀覆盖（`.` → `_`）。`config.Get()` 返回已加载的单例。
- `internal/model/` — GORM 模型，每个实体一个文件（约 24 个实体）。
- `internal/repository/` — 数据访问层；每个实体都有 `*_repo.go` 和 `New*Repository()` 构造函数。
- `internal/service/` — 业务逻辑；构造函数接收其仓库作为依赖（手动 DI，无框架）。
- `internal/handler/` — Gin 处理器；每个领域一个文件，与服务 1:1 匹配。
- `internal/router/router.go` — **单文件注册所有路由**。所有仓库/服务/处理器在此构造并连接。要添加端点，遵循现有模式：构造 repo → service → handler，然后注册路由。
- `internal/middleware/` — CORS、请求日志、恢复、限流（基于 Redis，回退到内存）、JWT 认证、`AdminRequired` RBAC。
- `internal/grpc/` — 用于节点通信的 gRPC 服务器。
- `internal/scheduler/tasks/` — 定时任务（佣金、订单、流量），由 `pkg/scheduler` 驱动。
- `pkg/` — 共享工具库：`database`、`jwt`、`redis`、`email`、`payment`（支付宝/微信）、`ratelimit`、`response`、`captcha`、`i18n`、`utils`。

数据库模式由启动时的 **GORM** **`AutoMigrate`** 管理（参见 `cmd/server/main.go` 中的模型列表），而不是手动运行 `migrations/*.sql`。

## 关键约定

- **不使用 protoc。** gRPC 使用手写 Go 结构体加自定义 JSON 编解码器（`internal/grpc/codec.go`）。永远不要运行 `protoc`。
- **API 响应信封。** 每个 REST 响应都是 `{ "code": 0, "message": "success", "data": <payload> }`。`code 0` = 成功。认证失败返回 HTTP 200 和 `code 401` — 前端 axios 拦截器解包此响应并处理 token 刷新。
- **前端 API 访问**通过 `frontend/src/lib/api.ts` 中的单一 axios 实例；所有 hook 导入它。它自动附加 Bearer token 并解包信封（解析 `res.data`）。
- **前端路径别名。** `@/` 映射到 `src/`（`vite.config.ts` + `tsconfig.json`）。
- **节点通信。** gRPC 在端口 50051（可通过 `grpc` 配置）；通过 `app.node_api_key` + `node_id` 元数据认证。Admin 处理器通过 `grpc.NodeBroadcaster.BroadcastConfig()` / `BroadcastUsers()` 向已连接节点推送实时更改（在 `cmd/server/main.go` 中通过 `handler.NotifyNodeConfigChange` / `handler.NotifyUserChange` 连接）。
- **节点配置**存储为 `nodes` 表上的 `server_info` JSON 字段；面板将结构化字段组装到其中并传递给节点。
- **前端技术栈**：Zustand（`src/stores/`）、TanStack Query（`src/hooks/`）、react-hook-form + zod、echarts-for-react（不是 recharts）、shadcn/ui + Radix 原语（`src/components/ui/`）。

## 节点管理页面结构（不可合并）

节点管理相关页面必须保持独立，不得合并到单个页面中：

| 菜单项   | 路由                       | 页面组件                     |
| ----- | ------------------------ | ------------------------ |
| 服务器管理 | `/admin/server-machines` | `ServerMachinesPage.tsx` |
| 节点管理  | `/admin/nodes`           | `NodesPage.tsx`          |
| 权限组管理 | `/admin/server-groups`   | `ServerGroupsPage.tsx`   |
| 路由管理  | `/admin/server-routes`   | `ServerRoutesPage.tsx`   |

每个页面独立维护，遵循原版 Xboard 的菜单结构。不要将这些功能合并到 NodesPage 的 Tab 中。

## 设计原则

设计每个页面功能时，都需要深度分析原版 Xboard 关于该功能的数据结构、API 接口、页面布局和交互逻辑，用于对齐原版面板。

### 分析源目录

根据功能类型选择对应的原版代码目录：

| 功能类型   | 分析目录                                  | 说明                                |
| ------ | ------------------------------------- | --------------------------------- |
| 管理员面板  | `E:\idea\Xboard-Go\Xboard-admin-dist` | 原版 admin 前端编译产物，分析路由、页面组件、API 调用  |
| 用户面板   | `E:\idea\Xboard-Go\Xboard-user`       | 原版用户前端，分析用户端页面、订阅、订单等功能           |
| 节点/服务端 | `E:\idea\Xboard-Go\Xboard-Node`       | 原版节点代理，分析节点通信协议、配置格式、上报逻辑         |
| 面板后端   | `E:\idea\Xboard-Go\Xboard`            | 原版 Laravel 后端，分析 API 接口、数据模型、业务逻辑 |

### 分析流程

1. 根据功能类型选择对应的分析源目录
2. 查看原版代码中对应的页面组件 / API 接口 / 数据模型
3. 提取 API 端点、请求参数、响应结构
4. 分析页面布局、表格列、表单字段、操作按钮
5. 对齐原版的菜单分组、路由路径、页面标题
6. 遵循 shadcn-ui 组件规范实现

## 禁止事项

- 不要运行 `protoc`（proto 类型是手写的）。
- 不要直接编辑 `internal/static/dist/` — 它在构建时从 `frontend/dist/` 填充。
- 不要将节点管理的独立页面合并到单个页面中。

