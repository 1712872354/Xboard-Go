# AGENTS.md — Xboard-Go

## 项目概述

代理面板管理系统。Go 后端（Gin + GORM + gRPC）+ React 前端（Vite + shadcn/ui）。前端在构建时通过 `go:embed` 嵌入到 Go 二进制文件中。

同工作区下的两个关联项目：

- **Xboard-Node-Go** (`../Xboard-Node-Go`)：通过 gRPC 连接的节点代理
- **Xboard-Node** (`../Xboard-Node`)：原始参考节点代理（用于协议/功能参考）

## 构建命令

```bash
# 完整构建（前端 → 复制到嵌入目录 → 编译 Go）
cd frontend && pnpm install && pnpm run build && cd ..
Remove-Item -Recurse -Force internal\static\dist\* -ErrorAction SilentlyContinue
Copy-Item -Path frontend\dist\* -Destination internal\static\dist\ -Recurse -Force
go build -o bin\xboard-go.exe .\cmd\server\

# Windows 一键构建（PowerShell）
.\build.ps1

# 仅 Go 构建（前端 dist 已存在于 internal/static/dist/）
go build -o bin\xboard-go.exe .\cmd\server\

# 代码检查
golangci-lint run

# 测试
go test ./... -v
```

## 架构

```
cmd/server/main.go          — 入口：配置 → 日志 → 数据库 → Redis → 路由 + gRPC
config/config.go             — 基于 Viper 的 YAML 配置加载
internal/
  model/                     — GORM 模型（24 个文件，每个实体一个）
  repository/                — 数据访问层（GORM 查询）
  service/                   — 业务逻辑层
  handler/                   — Gin HTTP 处理器（REST API）
  router/router.go           — 所有路由注册（单文件，约 500 行）
  middleware/                — CORS、JWT 认证、限流、RBAC
  grpc/                      — gRPC 服务器，用于节点通信
    server.go                — 服务器生命周期，全局 NodeBroadcaster
    handler.go               — 握手、Stream、GetConfig、GetUsers RPC
    proto.go                 — 手写 protobuf 类型（无 protoc，JSON 编解码器）
    broadcaster.go           — 每节点事件通道，用于配置/用户推送
    metrics.go               — 内存节点指标缓存
    codec.go                 — 自定义 JSON 编解码器
    auth.go                  — gRPC 认证拦截器（apikey + node_id 元数据）
  static/                    — go:embed 前端 dist
pkg/                         — 共享工具库（database、jwt、redis、email 等）
```

## 关键约定

- **不使用 protoc**：gRPC 使用手写 Go 结构体 + 自定义 JSON 编解码器（`internal/grpc/codec.go`）。永远不要运行 `protoc`。
- **API 响应信封**：所有 REST 响应使用 `{ "code": 0, "message": "success", "data": <payload> }`。前端自动解包。
- **前端 API**：单一 axios 实例位于 `frontend/src/lib/api.ts`。所有 hook 使用它。认证 token 自动附加。
- **前端路径别名**：`@/` 映射到 `src/`（在 vite.config.ts + tsconfig.json 中配置）。
- **节点通信**：gRPC 在端口 50051（可配置）。通过 config.yaml 中的 `node_api_key` 认证。节点在 gRPC 元数据中发送 `authorization` + `node_id`。
- **广播模式**：Admin REST 处理器调用 `grpc.NodeBroadcaster.BroadcastConfig()` / `BroadcastUsers()` 通过 gRPC Stream 向已连接节点推送更改。

## 前端技术栈

- **UI**：shadcn/ui v2/v3（Radix 原语）。组件位于 `src/components/ui/`。使用 `asChild` 模式。
- **表单**：react-hook-form + zod + @hookform/resolvers
- **数据**：TanStack Query（useQuery/useMutation）。所有 hook 位于 `src/hooks/`。
- **状态**：Zustand stores 位于 `src/stores/`（auth、theme、locale）
- **图表**：echarts-for-react（ECharts 封装）
- **国际化**：已准备但未完全接入。语言 JSON 文件位于 `src/locales/`。

## 协议支持

12 种协议：vmess、vless、trojan、shadowsocks、hysteria、hysteria2、tuic、anytls、naive、socks、http、mieru。

节点配置存储为 `nodes` 表上的 `server_info` JSON 字段。前端将结构化字段组装到此 JSON 中。面板通过 gRPC `NodeConfig.ServerInfo` 将其传递给节点。

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

- 不要运行 `protoc` — 所有 proto 类型都是手写的
- 不要直接修改 `internal/static/dist/` — 它在构建时从 `frontend/dist/` 填充
- 不要添加 `framer-motion` — 不在依赖中
- 不要使用 `recharts` — 使用 `echarts-for-react` 作为图表库
- `build` npm 脚本仅是 `vite build`（没有 `tsc -b`）— TypeScript 错误不会阻止构建
- 不要将节点管理的独立页面合并到单个页面中

