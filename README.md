# Xboard-Go

Xboard-Go 是 [Xboard](https://github.com/cedar2025/Xboard) 的 Go 语言重写版本，一个高性能的代理面板管理系统。

## ✨ 特性

- **高性能**: Go 语言编写，单二进制文件部署
- **多协议**: 支持 VMess, VLESS, Trojan, Shadowsocks, Hysteria2, TUIC, AnyTLS, Naive, SOCKS, HTTP, Mieru
- **多订阅格式**: 支持 Clash, V2Ray, Sing-box, Shadowrocket, Surge, Loon, Quantumult X, Stash, Surfboard
- **节点通讯**: gRPC + WebSocket + REST 三重通讯方式
- **嵌入式前端**: 前端编译后嵌入 Go 二进制文件，单文件部署
- **完整功能**: 用户管理、套餐管理、订单系统、工单系统、优惠券、礼品卡、邀请佣金等
- **主题系统**: 支持自定义主题，可配置主题参数
- **插件系统**: 支持插件安装、卸载、配置管理
- **会话管理**: 支持多设备会话管理，可强制登出
- **快速登录**: 支持快速登录链接、Token直接登录、邮件链接登录

## 🚀 快速开始

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

## ⚙️ 配置

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

## 🔗 访问地址

- 用户面板: `http://your-domain:8080/user/login`
- 管理后台: `http://your-domain:8080/admin/login`

## 📦 节点部署

参考 [Xboard-Node-Go](https://github.com/1712872354/Xboard-Node-Go) 项目部署节点。

## 📚 API 文档

### 认证 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/auth/register` | POST | 用户注册 |
| `/api/v1/auth/login` | POST | 用户登录 |
| `/api/v1/auth/refresh` | POST | 刷新Token |
| `/api/v1/auth/forget` | POST | 忘记密码 |
| `/api/v1/auth/reset` | POST | 重置密码 |
| `/api/v1/auth/send-code` | POST | 发送验证码 |
| `/api/v1/auth/verify` | POST | 邮箱验证 |
| `/api/v1/auth/quick-login` | GET | 快速登录 |
| `/api/v1/auth/token2login` | GET | Token直接登录 |
| `/api/v1/auth/mail-link-login` | GET | 邮件链接登录 |
| `/api/v1/auth/send-mail-login-link` | POST | 发送邮件登录链接 |

### 用户 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/user/profile` | GET/PUT | 用户信息管理 |
| `/api/v1/user/password` | PUT | 修改密码 |
| `/api/v1/user/subscribe/reset` | POST | 重置订阅Token |
| `/api/v1/user/servers` | GET | 获取服务器列表 |
| `/api/v1/user/sessions` | GET | 获取活跃会话 |
| `/api/v1/user/sessions/remove` | POST | 移除会话 |
| `/api/v1/user/check-login` | GET | 检查登录状态 |
| `/api/v1/user/traffic/stats` | GET | 流量统计 |
| `/api/v1/user/traffic/history` | GET | 流量历史 |
| `/api/v1/user/traffic/daily` | GET | 每日流量 |

### 套餐 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/plans` | GET | 获取套餐列表 |
| `/api/v1/plans/{id}` | GET | 获取套餐详情 |

### 订单 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/orders` | POST/GET | 创建/查询订单 |
| `/api/v1/orders/{id}` | GET | 订单详情 |
| `/api/v1/orders/{id}/cancel` | POST | 取消订单 |
| `/api/v1/orders/checkout` | POST | 订单结账 |
| `/api/v1/orders/check` | GET | 检查订单状态 |
| `/api/v1/orders/detail` | GET | 订单详情（用户端） |
| `/api/v1/orders/payment-methods` | GET | 获取支付方式 |

### 支付 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/payment/methods` | GET | 获取支付方式 |
| `/api/v1/payment/create` | POST | 创建支付 |
| `/api/v1/payment/mock/pay` | GET | 模拟支付页面 |
| `/api/v1/payment/mock/callback` | GET | 模拟支付回调 |

### 工单 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/tickets` | POST/GET | 创建/查询工单 |
| `/api/v1/tickets/{id}` | GET | 工单详情 |
| `/api/v1/tickets/{id}/reply` | POST | 回复工单 |
| `/api/v1/tickets/{id}/close` | POST | 关闭工单 |
| `/api/v1/tickets/{id}/withdraw` | POST | 撤回工单 |
| `/api/v1/tickets/{id}/messages` | GET | 获取工单消息 |
| `/api/v1/tickets/stats` | GET | 工单统计 |

### 邀请 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/invite/code` | GET | 获取邀请码 |
| `/api/v1/invite/use` | POST | 使用邀请码 |
| `/api/v1/invite/details` | GET | 邀请详情 |
| `/api/v1/invite/commission/stats` | GET | 佣金统计 |
| `/api/v1/invite/commission/logs` | GET | 佣金日志 |
| `/api/v1/invite/commission/withdraw` | POST | 佣金提现 |

### 礼品卡 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/gift-cards/use` | POST | 使用礼品卡 |
| `/api/v1/gift-cards/history` | GET | 礼品卡历史 |
| `/api/v1/gift-cards/types` | GET | 礼品卡类型 |
| `/api/v1/gift-cards/{id}` | GET | 礼品卡详情 |

### 优惠券 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/coupons/validate` | POST | 验证优惠券 |

### 公告 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/notices` | GET | 获取公告列表 |

### 知识库 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/knowledges` | GET | 获取知识库列表 |
| `/api/v1/knowledges/categories` | GET | 获取知识库分类 |

### 主题 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/themes/current` | GET | 获取当前主题 |

### Telegram API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/telegram/bot-info` | GET | 获取Bot信息 |
| `/api/v1/telegram/webhook` | POST | Telegram Webhook |

### 订阅 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/client/subscribe` | GET | 客户端订阅 |

---

## 🛠️ 管理员 API

### 用户管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/users` | GET | 用户列表 |
| `/api/v1/admin/users/{id}` | GET/PUT | 用户详情/更新 |
| `/api/v1/admin/users/{id}/status` | PUT | 用户状态 |
| `/api/v1/admin/users/{id}/role` | PUT | 用户角色 |
| `/api/v1/admin/users/{id}` | DELETE | 删除用户 |
| `/api/v1/admin/users/generate` | POST | 批量生成用户 |
| `/api/v1/admin/users/export` | GET | 导出用户CSV |
| `/api/v1/admin/users/send-mail` | POST | 发送邮件给用户 |
| `/api/v1/admin/users/reset-secret` | POST | 重置用户密钥 |
| `/api/v1/admin/users/set-inviter` | POST | 设置邀请人 |
| `/api/v1/admin/users/transfer-balance` | POST | 转移余额 |

### 套餐管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/plans` | GET/POST | 套餐列表/创建 |
| `/api/v1/admin/plans/{id}` | PUT/DELETE | 套餐更新/删除 |
| `/api/v1/admin/plans/sort` | POST | 套餐排序 |

### 订单管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/orders` | GET | 订单列表 |
| `/api/v1/admin/orders/confirm-payment` | POST | 确认支付 |
| `/api/v1/admin/orders/assign` | POST | 分配订单 |

### 节点管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/nodes` | GET/POST | 节点列表/创建 |
| `/api/v1/admin/nodes/{id}` | GET/PUT/DELETE | 节点详情/更新/删除 |
| `/api/v1/admin/nodes/{id}/copy` | POST | 复制节点 |
| `/api/v1/admin/nodes/{id}/status` | PUT | 节点状态 |
| `/api/v1/admin/nodes/{id}/metrics` | GET | 节点指标 |
| `/api/v1/admin/nodes/{id}/reset-traffic` | POST | 重置节点流量 |
| `/api/v1/admin/nodes/batch` | POST | 批量操作 |
| `/api/v1/admin/nodes/sort` | POST | 节点排序 |
| `/api/v1/admin/nodes/batch-update` | POST | 批量更新 |

### 服务器管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/server-groups` | GET/POST | 服务器分组 |
| `/api/v1/admin/server-routes` | GET/POST | 服务器路由 |
| `/api/v1/admin/server-machines` | GET/POST | 服务器机器 |
| `/api/v1/admin/server-machines/{id}/status` | PUT | 机器状态 |
| `/api/v1/admin/server-machines/{id}/load` | PUT | 机器负载 |
| `/api/v1/admin/server-machines/{id}/reset-token` | POST | 重置Token |
| `/api/v1/admin/server-machines/{id}/install-command` | GET | 安装命令 |
| `/api/v1/admin/server-machines/{id}/nodes` | GET | 机器节点 |
| `/api/v1/admin/server-machines/{id}/load-history` | GET | 负载历史 |

### 支付管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/payment/gateways` | GET/POST | 支付网关 |
| `/api/v1/admin/payment/gateways/{id}` | GET/PUT/DELETE | 网关详情/更新/删除 |
| `/api/v1/admin/payment/gateways/{id}/status` | PUT | 网关状态 |
| `/api/v1/admin/payment/gateways/{id}/sort` | PUT | 网关排序 |
| `/api/v1/admin/payment/gateways/sort` | POST | 批量排序 |

### 工单管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/tickets` | GET | 工单列表 |
| `/api/v1/admin/tickets/{id}` | GET | 工单详情 |
| `/api/v1/admin/tickets/{id}/reply` | POST | 回复工单 |
| `/api/v1/admin/tickets/{id}/close` | POST | 关闭工单 |
| `/api/v1/admin/tickets/{id}` | DELETE | 删除工单 |
| `/api/v1/admin/tickets/{id}/messages` | GET | 工单消息 |

### 优惠券管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/coupons` | GET/POST | 优惠券列表/创建 |
| `/api/v1/admin/coupons/{id}` | GET/PUT/DELETE | 优惠券详情/更新/删除 |
| `/api/v1/admin/coupons/{id}/show` | PUT | 显示状态 |

### 礼品卡管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/gift-card-templates` | GET/POST | 模板列表/创建 |
| `/api/v1/admin/gift-card-templates/{id}` | GET/PUT/DELETE | 模板详情/更新/删除 |
| `/api/v1/admin/gift-card-codes` | GET | 卡密列表 |
| `/api/v1/admin/gift-card-codes/generate` | POST | 生成卡密 |
| `/api/v1/admin/gift-card-codes/{id}` | GET/PUT/DELETE | 卡密详情/更新/删除 |
| `/api/v1/admin/gift-card-codes/{id}/toggle` | PUT | 启用/禁用卡密 |
| `/api/v1/admin/gift-card-codes/{id}/usages` | GET | 使用记录 |
| `/api/v1/admin/gift-card-codes/export` | GET | 导出卡密 |
| `/api/v1/admin/gift-card-codes/statistics` | GET | 统计数据 |
| `/api/v1/admin/gift-card-codes/types` | GET | 卡密类型 |

### 公告管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/notices` | GET/POST | 公告列表/创建 |
| `/api/v1/admin/notices/{id}` | GET/PUT/DELETE | 公告详情/更新/删除 |
| `/api/v1/admin/notices/sort` | POST | 公告排序 |
| `/api/v1/admin/notices/{id}/show` | PUT | 显示状态 |

### 知识库管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/knowledges` | GET/POST | 知识库列表/创建 |
| `/api/v1/admin/knowledges/{id}` | GET/PUT/DELETE | 知识库详情/更新/删除 |
| `/api/v1/admin/knowledges/sort` | POST | 知识库排序 |
| `/api/v1/admin/knowledges/{id}/show` | PUT | 显示状态 |

### 邮件模板管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/mail-templates` | GET/POST | 模板列表/创建 |
| `/api/v1/admin/mail-templates/{id}` | GET/PUT/DELETE | 模板详情/更新/删除 |
| `/api/v1/admin/mail-templates/{id}/reset` | PUT | 重置模板 |
| `/api/v1/admin/mail-templates/{id}/test` | POST | 测试模板 |

### 主题管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/themes` | GET | 主题列表 |
| `/api/v1/admin/themes/{id}` | GET/DELETE | 主题详情/删除 |
| `/api/v1/admin/themes/{id}/config` | GET/PUT | 主题配置 |
| `/api/v1/admin/themes/{id}/default` | PUT | 设置默认主题 |

### 插件管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/plugins` | GET/POST | 插件列表/创建 |
| `/api/v1/admin/plugins/{id}` | GET/PUT/DELETE | 插件详情/更新/删除 |
| `/api/v1/admin/plugins/{id}/status` | PUT | 插件状态 |
| `/api/v1/admin/plugins/{id}/enable` | POST | 启用插件 |
| `/api/v1/admin/plugins/{id}/disable` | POST | 禁用插件 |
| `/api/v1/admin/plugins/upload` | POST | 上传插件 |
| `/api/v1/admin/plugins/types` | GET | 插件类型 |
| `/api/v1/admin/plugins/{id}/install` | POST | 安装插件 |
| `/api/v1/admin/plugins/{id}/uninstall` | POST | 卸载插件 |
| `/api/v1/admin/plugins/{id}/config` | GET/PUT | 插件配置 |
| `/api/v1/admin/plugins/{id}/upgrade` | POST | 升级插件 |
| `/api/v1/admin/plugins/reload` | POST | 重新加载插件 |

### 系统设置

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/settings` | GET/PUT | 系统设置 |
| `/api/v1/admin/settings/group/{group}` | GET | 分组设置 |
| `/api/v1/admin/settings/mappings` | GET | 配置映射 |
| `/api/v1/admin/settings/{key}` | PUT/DELETE | 设置更新/删除 |
| `/api/v1/admin/settings/test-email` | POST | 测试邮件 |
| `/api/v1/admin/settings/email-template` | GET | 邮件模板 |
| `/api/v1/admin/settings/theme-template` | GET | 主题模板 |

### 数据看板

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/dashboard/overview` | GET | 数据概览 |
| `/api/v1/admin/dashboard/recent-orders` | GET | 最近订单 |
| `/api/v1/admin/dashboard/recent-users` | GET | 最近用户 |
| `/api/v1/admin/dashboard/income-stats` | GET | 收入统计 |
| `/api/v1/admin/dashboard/user-growth` | GET | 用户增长 |
| `/api/v1/admin/dashboard/node-stats` | GET | 节点统计 |
| `/api/v1/admin/dashboard/node-traffic-ranking` | GET | 节点流量排行 |
| `/api/v1/admin/dashboard/user-traffic-ranking` | GET | 用户流量排行 |
| `/api/v1/admin/dashboard/invite-ranking` | GET | 邀请排行 |
| `/api/v1/admin/dashboard/commission-stats` | GET | 佣金统计 |
| `/api/v1/admin/dashboard/comprehensive-stats` | GET | 综合统计 |

### 统计报表

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/stats/server-ranking` | GET | 服务器排行 |
| `/api/v1/admin/stats/server-yesterday-ranking` | GET | 服务器昨日排行 |
| `/api/v1/admin/stats/orders` | GET | 订单统计 |
| `/api/v1/admin/stats/users` | GET | 用户统计 |
| `/api/v1/admin/stats/records` | GET | 统计记录 |

### 流量管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/traffic/sync` | POST | 同步流量 |
| `/api/v1/admin/traffic/logs` | GET | 流量日志 |
| `/api/v1/admin/traffic-reset/logs` | GET | 重置日志 |
| `/api/v1/admin/traffic-reset/stats` | GET | 重置统计 |
| `/api/v1/admin/traffic-reset/user/{id}/history` | GET | 用户重置历史 |
| `/api/v1/admin/traffic-reset/reset-user` | POST | 重置用户流量 |

### 设备管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/devices/user/{id}` | GET | 用户设备 |
| `/api/v1/admin/devices/node/{id}` | GET | 节点设备 |
| `/api/v1/admin/devices/user/{id}/count` | GET | 在线设备数 |
| `/api/v1/admin/devices/cleanup` | POST | 清理离线设备 |

### 队列管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/system/queue/stats` | GET | 队列统计 |
| `/api/v1/admin/system/queue/workload` | GET | 队列工作负载 |
| `/api/v1/admin/system/queue/failed-jobs` | GET | 失败任务列表 |
| `/api/v1/admin/system/queue/failed-jobs/{id}/retry` | POST | 重试失败任务 |
| `/api/v1/admin/system/queue/failed-jobs/{id}` | DELETE | 删除失败任务 |
| `/api/v1/admin/system/queue/failed-jobs/clear` | POST | 清空失败任务 |

### 系统管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/system/status` | GET | 系统状态 |
| `/api/v1/admin/system/info` | GET | 系统信息 |
| `/api/v1/admin/system/scheduler` | GET | 调度器状态 |
| `/api/v1/admin/system/check-update` | GET | 检查更新 |
| `/api/v1/admin/system/execute-update` | POST | 执行更新 |

### 审计日志

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/audit-logs` | GET | 审计日志列表 |
| `/api/v1/admin/audit-logs/{id}` | GET | 日志详情 |
| `/api/v1/admin/audit-logs/{id}` | DELETE | 删除日志 |

### Telegram 管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/telegram/set-webhook` | POST | 设置Webhook |

---

## 🖥️ 节点 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v2/server/handshake` | POST | 节点握手 |
| `/api/v2/server/report` | POST | 综合上报 |
| `/api/v2/server/config` | GET | 获取节点配置 |
| `/api/v2/server/user` | GET | 获取用户列表 |
| `/api/v2/server/machine/nodes` | POST | 获取机器节点 |
| `/api/v2/server/machine/status` | POST | 上报机器状态 |
| `/api/v2/server/ws` | WebSocket | 实时推送 |

### Legacy 节点 API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/server/UniProxy/config` | GET | 获取配置 |
| `/api/v1/server/UniProxy/user` | GET | 获取用户 |
| `/api/v1/server/UniProxy/push` | POST | 上报流量 |
| `/api/v1/server/UniProxy/alive` | POST | 上报在线 |
| `/api/v1/server/UniProxy/status` | POST | 上报状态 |

---

## 🛠️ 技术栈

- **后端**: Go + Gin + GORM + gRPC
- **前端**: React 19 + TypeScript + Vite + shadcn/ui
- **数据库**: SQLite (默认) / MySQL / PostgreSQL
- **缓存**: 可选 Redis

## 📊 功能覆盖率

Xboard-Go 已实现原版 Xboard 约 **98%** 的功能，包括：

- ✅ 用户认证与授权（11个API）
- ✅ 用户管理（15个API）
- ✅ 套餐管理（8个API）
- ✅ 订单管理（10个API）
- ✅ 节点管理（12个API）
- ✅ 服务器管理（10个API）
- ✅ 支付管理（8个API）
- ✅ 工单系统（8个API）
- ✅ 优惠券系统（7个API）
- ✅ 礼品卡系统（12个API）
- ✅ 邀请系统（8个API）
- ✅ 公告系统（6个API）
- ✅ 知识库系统（6个API）
- ✅ 邮件模板（5个API）
- ✅ 系统设置（8个API）
- ✅ 数据看板（11个API）
- ✅ 统计报表（5个API）
- ✅ 主题管理（7个API）
- ✅ 插件管理（10个API）
- ✅ 流量管理（8个API）
- ✅ 设备管理（4个API）
- ✅ Telegram集成（3个API）
- ✅ 系统维护（8个API）
- ✅ 队列管理（6个API）

## 📄 许可证

MIT License
