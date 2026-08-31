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

### Docker 部署 (推荐)

```bash
# 创建数据目录
mkdir -p data

# 复制配置文件
cp config.yaml.example data/config.yaml
# 编辑配置文件
vi data/config.yaml

# 启动服务
docker run -d --restart=always \
  --name xboard-go \
  -p 8080:8080 \
  -p 50051:50051 \
  -v $(pwd)/data:/data \
  ghcr.io/1712872354/xboard-go:latest
```

### Docker Compose

```bash
# 克隆代码
git clone https://github.com/1712872354/Xboard-Go.git
cd Xboard-Go

# 创建数据目录并复制配置
mkdir -p data
cp config.yaml.example data/config.yaml
# 编辑配置文件
vi data/config.yaml

# 启动服务
docker-compose up -d
```

### 直接运行

```bash
# 下载对应平台的二进制文件
wget https://github.com/1712872354/Xboard-Go/releases/latest/download/xboard-go-linux-amd64
chmod +x xboard-go-linux-amd64

# 复制配置文件
cp config.yaml.example config.yaml
# 编辑配置文件
vi config.yaml

# 运行
./xboard-go-linux-amd64 -config config.yaml
```

### 从源码构建

```bash
# 克隆代码
git clone https://github.com/1712872354/Xboard-Go.git
cd Xboard-Go

# 安装前端依赖并构建
cd frontend && pnpm install && pnpm run build && cd ..

# 复制前端到嵌入目录
cp -r frontend/dist/* internal/static/dist/

# 编译 Go 二进制
go build -o bin/xboard-go ./cmd/server/

# 运行
./bin/xboard-go -config config.yaml
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
