# Plan: 统一节点服务端 — 保留 Xboard-Node，删除 Xboard-Node-Go

## 背景

当前有两个节点服务端实现：
- **Xboard-Node** (`E:\idea\Xboard-Go\Xboard-Node`)：成熟版本，支持 12 种协议(sing-box) + 5 种(xray)，ACME 证书管理，WebSocket+REST 实时推送，机器模式，热重载
- **Xboard-Node-Go** (`E:\idea\Xboard-Go\Xboard-Node-Go`)：半成品，仅 5 种协议(sing-box) + 4 种(xray)，内核是 stub 未真正实现，gRPC 通信

**关键发现**：面板 (Xboard-Go) 只有 gRPC 节点通信端点，没有 REST/WebSocket。Xboard-Node 使用 REST+WebSocket。因此需要在面板中添加兼容层。

## 方案

### 第一阶段：面板添加 REST+WebSocket 节点通信端点

在 Xboard-Go 面板中实现 Xboard-Node 需要的 REST API 和 WebSocket 端点，使其能同时支持 gRPC（现有）和 REST+WebSocket（Xboard-Node）两种节点通信方式。

#### 1. 添加节点 REST API 认证中间件
- 文件：`internal/middleware/node_auth.go`
- 实现 `NodeAPIKeyAuth` 中间件，从 query params 或 JSON body 中提取 `token` + `node_id`
- 验证 token 是否匹配 `config.App.NodeAPIKey`

#### 2. 添加节点 REST API Handler
- 文件：`internal/handler/node_server.go`
- 复用现有的 `UniProxyService` 和 gRPC handler 中的逻辑
- 端点：
  - `POST /api/v2/server/handshake` — 节点握手认证
  - `GET /api/v2/server/config` — 获取节点配置
  - `GET /api/v2/server/user` — 获取用户列表
  - `POST /api/v2/server/report` — 上报流量/状态/在线IP

#### 3. 添加 WebSocket 推送端点
- 文件：`internal/handler/node_ws.go`
- 使用 `gorilla/websocket` 库
- 端点：`GET /api/v2/server/ws` — WebSocket 连接
- 复用 `grpc.Broadcaster` 的事件推送逻辑，让 admin 操作同时通知 gRPC 和 WebSocket 连接的节点
- 推送事件：`sync.config`, `sync.users`, `sync.user.delta`

#### 4. 注册路由
- 文件：`internal/router/router.go`
- 在 `/api/v2/server/` 路由组下注册新端点
- 使用 `NodeAPIKeyAuth` 中间件保护

#### 5. 修改 Broadcaster 支持多协议推送
- 文件：`internal/grpc/broadcaster.go` 或新建 `internal/broadcast/broadcaster.go`
- 扩展现有 Broadcaster，使其同时支持 gRPC 和 WebSocket 连接的节点

### 第二阶段：清理 Xboard-Node-Go

#### 6. 删除 Xboard-Node-Go 目录
- 删除 `E:\idea\Xboard-Go\Xboard-Node-Go/` 整个目录
- 从工作区配置中移除引用（如有）

### 第三阶段：优化 Xboard-Node 的 IPv6 支持

Xboard-Node 已经双栈监听 `::`，但可以在以下方面优化：

#### 7. 增强 IPv6 流量统计区分
- 文件：`Xboard-Node/internal/kernel/singbox/conntracker.go`
- 在 alive IP 上报中区分 IPv4/IPv6 来源
- 在系统状态中添加 IPv6 连接数统计

#### 8. 确保 Windows 双栈兼容
- 检查 Windows 上 `::` 监听是否真正同时接受 IPv4 和 IPv6
- 如需要，添加 `ipv6only=0` socket 选项

## 影响范围

- **新增文件**：`internal/middleware/node_auth.go`, `internal/handler/node_server.go`, `internal/handler/node_ws.go`
- **修改文件**：`internal/router/router.go`, `internal/grpc/broadcaster.go`（或新建 broadcast 包）
- **删除**：`E:\idea\Xboard-Go\Xboard-Node-Go/` 整个目录
- **前端无影响**

## 依赖

- 需要添加 `gorilla/websocket` 依赖到 go.mod

## 验证

1. Xboard-Node 能通过 REST API 完成握手、获取配置、获取用户、上报流量
2. Xboard-Node 能通过 WebSocket 接收实时配置/用户更新
3. 现有 gRPC 通信不受影响（向后兼容）
4. IPv4/IPv6 双栈正常工作
