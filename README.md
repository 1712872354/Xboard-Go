# Xboard-Go

Xboard-Go 是 [Xboard](https://github.com/cedar2025/Xboard) 的 Go 语言重写版本，一个高性能的代理面板管理系统。

## 特性

- **高性能**: Go 语言编写，单二进制文件部署
- **多协议**: 支持 VMess, VLESS, Trojan, Shadowsocks, Hysteria2, TUIC, AnyTLS, Naive, SOCKS, HTTP, Mieru
- **多订阅格式**: 支持 Clash, V2Ray, Sing-box, Shadowrocket, Surge, Loon, Quantumult X, Stash, Surfboard
- **节点通讯**: gRPC + WebSocket + REST 三重通讯方式
- **嵌入式前端**: 前端编译后嵌入 Go 二进制文件，单文件部署
- **完整功能**: 用户管理、套餐管理、订单系统、工单系统、优惠券、礼品卡、邀请佣金等

## 快速开始

### 一键部署 (推荐)

```bash
# Linux / macOS
curl -fsSL https://raw.githubusercontent.com/1712872354/Xboard-Go/master/install.sh | bash

# 或者下载脚本后运行
wget https://raw.githubusercontent.com/1712872354/Xboard-Go/master/install.sh
chmod +x install.sh
sudo ./install.sh
```

```powershell
# Windows PowerShell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/1712872354/Xboard-Go/master/install.ps1" -OutFile "install.ps1"
.\install.ps1
```

部署脚本支持：
- 自动检测系统和架构
- 自由选择 Docker 或二进制部署
- 自动配置数据库 (SQLite/MySQL/PostgreSQL)
- 自动生成配置文件
- 初始化管理员账户

### Docker 部署

```bash
docker run -d --restart=always \
  --name xboard-go \
  -p 8080:8080 \
  -p 50051:50051 \
  -v $(pwd)/data:/data \
  ghcr.io/1712872354/xboard-go:latest
```

### Docker Compose

```bash
git clone https://github.com/1712872354/Xboard-Go.git
cd Xboard-Go
mkdir -p data
cp config.yaml.example data/config.yaml
docker-compose up -d
```

### 直接运行

```bash
wget https://github.com/1712872354/Xboard-Go/releases/latest/download/xboard-go-linux-amd64
chmod +x xboard-go-linux-amd64
./xboard-go-linux-amd64 -config config.yaml
```

## 配置

配置文件 `config.yaml` 示例:

```yaml
server:
  port: 8080
  mode: release

database:
  driver: sqlite
  source: data/xboard.db

app:
  name: Xboard-Go
  key: your-secret-key-here
  node_api_key: your-node-api-key-here

# 可选: Redis 配置
# redis:
#   addr: localhost:6379
#   password: ""
#   db: 0

# 可选: gRPC 配置
# grpc:
#   port: 50051
```

## 访问地址

- 用户面板: `http://your-domain:8080/user/login`
- 管理后台: `http://your-domain:8080/admin/login`

## 节点部署

参考 [Xboard-Node-Go](https://github.com/1712872354/Xboard-Node-Go) 项目部署节点。

## API 文档

### 用户 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/auth/register` | POST | 用户注册 |
| `/api/v1/auth/login` | POST | 用户登录 |
| `/api/v1/user/profile` | GET | 获取用户信息 |
| `/api/v1/plans` | GET | 获取套餐列表 |
| `/api/v1/orders` | POST | 创建订单 |
| `/api/v1/tickets` | GET/POST | 工单管理 |
| `/api/v1/client/subscribe` | GET | 客户端订阅 |

### 管理 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/users` | GET | 用户列表 |
| `/api/v1/admin/plans` | GET/POST | 套餐管理 |
| `/api/v1/admin/orders` | GET | 订单列表 |
| `/api/v1/admin/nodes` | GET/POST | 节点管理 |
| `/api/v1/admin/settings` | GET/PUT | 系统设置 |

### 节点 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v2/server/handshake` | POST | 节点握手 |
| `/api/v2/server/config` | GET | 获取节点配置 |
| `/api/v2/server/user` | GET | 获取用户列表 |
| `/api/v2/server/report` | POST | 上报数据 |
| `/api/v2/server/ws` | WebSocket | 实时推送 |

## 技术栈

- **后端**: Go + Gin + GORM + gRPC
- **前端**: React 19 + TypeScript + Vite + shadcn/ui
- **数据库**: SQLite (默认) / MySQL / PostgreSQL
- **缓存**: 可选 Redis

## 许可证

MIT License
