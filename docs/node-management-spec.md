# 节点管理功能规范（原版 Xboard 对齐）

> 本文档基于对原版 Xboard admin 前端编译产物的深度分析，作为节点管理页面开发的权威参考。

## 路由结构

```
节点管理 (nav:nodeManagement)
├── /server/machine   → 服务器管理 (nav:machineManagement)
├── /server/manage    → 节点管理 (nav:nodeManagement)
├── /server/group     → 权限组管理 (nav:permissionGroupManagement)
└── /server/route     → 路由管理 (nav:routeManagement)
```

每个页面是**独立组件**，不合并到 Tab 中。

---

## 一、节点管理页面 (`/server/manage`)

### API 端点

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/server/manage/getNodes` | GET | 无 | 获取所有节点 |
| `/server/manage/save` | POST | 节点完整数据对象 | 创建/更新节点 |
| `/server/manage/drop` | POST | `{id}` | 删除节点 |
| `/server/manage/batchDelete` | POST | `{ids: number[]}` | 批量删除 |
| `/server/manage/batchUpdate` | POST | `{ids, show?, enabled?, machine_id?}` | 批量更新 |
| `/server/manage/copy` | POST | `{id}` | 复制节点 |
| `/server/manage/update` | POST | `{id, type, show?, enabled?, machine_id?}` | 更新节点状态 |
| `/server/manage/sort` | POST | `[{id, order}]` | 拖拽排序 |
| `/server/manage/resetTraffic` | POST | `{id}` | 重置流量 |
| `/server/manage/batchResetTraffic` | POST | `{ids: number[]}` | 批量重置流量 |
| `/server/manage/generateEchKey` | GET | `{public_name}` | 生成 ECH 密钥 |

### 表格列

| 列 ID | 字段 | 宽度 | 排序 | 可隐藏 | 说明 |
|-------|------|------|------|--------|------|
| `select` | — | 40 | ✗ | ✗ | 复选框全选 |
| `id` | `id` | 50 | ✓ | — | 节点ID，带状态指示点 |
| `show` | `show` | 50 | ✗ | — | 显示/隐藏开关 |
| `name` | `name` | 260 | ✗ | — | 节点名称 + 协议标签 + 状态指示 + 服务器信息 |
| `machine` | `machine_id` | 200 | ✗ | — | 部署位置 |
| `host` | `host` | — | ✗ | ✓ | 地址 |
| `online` | `online` | 80 | ✓ | ✓ | 在线用户数 |
| `rate` | `rate` | 80 | ✗ | ✓ | 倍率 |
| `group_ids` | `group_ids` | — | ✗ | — | 权限组标签列表 |
| `type` | `type` | 100 | ✗ | ✓ | 协议类型 Badge |
| `traffic` | — | 120 | ✗ | ✓ | 流量进度条 + Tooltip |
| `actions` | — | 50 | — | — | 操作下拉菜单 |

### 行内操作

- 编辑 — 打开编辑对话框
- 复制 — 调用 `/server/manage/copy`
- 重置流量 — 带确认弹窗
- 删除 — 带确认弹窗（destructive）

### 工具栏功能

- 添加节点按钮
- 搜索框（按 name 筛选）
- 筛选器：协议类型、服务器、权限组
- 批量操作：显示/隐藏、启用/禁用、重置流量、删除
- 排序模式：拖拽排序，保存调用 sort API

### 节点表单 Schema

```typescript
{
  id: number | null
  specific_key: string | null
  code: string                    // 可选
  show: boolean                   // 默认 false
  name: string                    // 必填
  rate: string                    // 必填，>= 0
  rate_time_enable: boolean       // 动态倍率开关
  rate_time_ranges: [{start, end, rate}]  // 动态倍率规则
  tags: string[]                  // 标签数组
  transfer_enable_gb: string      // 流量限制(GB)
  excludes: string[]
  ips: string[]
  group_ids: string[]             // 权限组 ID 数组
  host: string                    // 必填
  port: string                    // 必填
  server_port: string             // 必填
  parent_id: string               // 默认 "0"
  route_ids: string[]             // 路由 ID 数组
  custom_outbounds: object[]      // 自定义 Outbounds
  custom_routes: object[]         // 自定义 Routes
  protocol_settings: object       // 协议配置
  listen_address: string
  machine_id: number | null
  enabled: boolean | null
}
```

### 表单字段顺序

1. 协议类型选择器（带颜色圆点）
2. 名称
3. 倍率（子节点时禁用）
4. 动态倍率开关 + 规则列表
5. 流量限制 (GB)
6. 代号（可选）
7. 标签
8. 权限组（多选，带快速创建按钮）
9. 地址
10. 端口（带同步到 server_port 按钮）
11. 服务端口
12. 协议配置（根据类型动态渲染）
13. 父节点
14. 路由（多选）
15. 服务器（含状态指示点 + 启用开关）

### 特殊功能

- **URL 参数联动**: `?machine_id=xxx` 自动按服务器筛选，`?machine_id=xxx&open_create=1` 自动打开创建对话框
- **ECH 密钥生成**: 调用 generateEchKey API
- **协议配置动态表单**: 每种协议有独立配置
- **高级配置对话框**: TLS / Multiplex / Outbounds / Routes 四个 Tab
- **拖拽排序**: 排序模式下保存调用 sort API

---

## 二、服务器管理页面 (`/server/machine`)

### API 端点

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/server/machine/fetch` | GET | 无 | 获取所有服务器 |
| `/server/machine/save` | POST | `{id?, name, notes, is_active}` | 创建/更新服务器 |
| `/server/machine/getToken` | GET | `{id}` | 获取 Token |
| `/server/machine/installCommand` | GET | `{id}` | 获取安装命令 |
| `/server/machine/resetToken` | POST | `{id}` | 重置 Token |
| `/server/machine/drop` | POST | `{id}` | 删除服务器 |
| `/server/machine/nodes` | GET | `{machine_id}` | 获取关联节点 |
| `/server/machine/history` | GET | `{machine_id, limit, range_hours?}` | 获取负载历史 |

### 表格列

| 列 ID | 字段 | 宽度 | 说明 |
|-------|------|------|------|
| `name` | `name` | 360 | 名称 + 状态点 + SID + 高负载标签 + 备注 |
| `status` | `last_seen_at` | 110 | online/offline/inactive/never Badge |
| `node_state` | `servers_count` | — | 虚拟列：with_nodes/idle_nodes/high_load |
| `load` | `load_status` | 210 | CPU/MEM/DISK 进度条 |
| `relation` | `servers_count` | 220 | 节点数量 + "查看详情"按钮 |
| `last_seen` | `last_seen_at` | 140 | 最后在线时间 + 最后上报时间 |
| `actions` | — | — | 操作按钮组 |

### 概览统计卡片

- 总数 — 服务器总数
- 在线 — 在线服务器数
- 离线 — 离线服务器数
- 高负载 — 高负载服务器数
- 托管节点 — 所有服务器托管的节点总数

### 工具栏功能

- 添加服务器按钮
- 搜索框（按名称/SID/备注筛选）
- 状态筛选：online/offline/inactive/never
- 节点状态筛选：with_nodes/idle_nodes/high_load
- 在线率显示

### 服务器表单 Schema

```typescript
{
  id: number | optional
  name: string         // 必填
  notes: string        // 可选
  is_active: boolean   // 默认 true
}
```

### 服务器详情弹窗

1. **头部信息** — 名称 + SID + 状态 + CPU% + 最后在线 + 节点数 + 备注
2. **操作按钮** — 添加节点到服务器 / 打开节点管理
3. **负载趋势图** — 支持 1h/6h/12h/24h 时间范围
   - CPU / MEM / DISK / 上行速度 / 下行速度
   - 可切换显示/隐藏各指标
4. **当前负载卡片** — CPU/MEM/DISK 进度条 + 网络速度
5. **Token 管理** — 显示/隐藏/重置 Token（180秒自动隐藏）
6. **安装命令** — 显示安装命令 + 复制按钮
7. **关联节点列表** — 表格显示节点名/类型/地址/启用状态/解绑操作
   - 绑定已有节点弹窗 — 搜索 + 多选 + 批量绑定

### 负载阈值配置

```javascript
{ cpu: { warn: 70, danger: 85 }, mem: { warn: 75, danger: 90 }, disk: { warn: 80, danger: 90 } }
```

### 状态颜色配置

```
online:    bg-emerald-500
offline:   bg-red-500
inactive:  bg-slate-400
never:     bg-slate-400 / amber
```

---

## 三、权限组管理页面 (`/server/group`)

### API 端点

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/server/group/fetch` | GET | 无 | 获取所有权限组 |
| `/server/group/save` | POST | `{id?, name}` | 创建/更新权限组 |
| `/server/group/drop` | POST | `{id}` | 删除权限组 |

### 表格列

| 列 ID | 字段 | 排序 | 说明 |
|-------|------|------|------|
| `id` | `id` | ✓ | ID Badge |
| `name` | `name` | — | 名称（截断显示） |
| `users_count` | `users_count` | ✓ | 用户数（带用户图标） |
| `server_count` | `server_count` | ✓ | 服务器数（带服务器图标） |
| `actions` | — | — | 编辑/删除按钮 |

### 权限组表单 Schema

```typescript
{
  name: string  // 必填，min(1)，max(40)，正则 /^[a-zA-Z0-9\u4e00-\u9fa5_-]+$/
}
```

---

## 四、路由管理页面 (`/server/route`)

### API 端点

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/server/route/fetch` | GET | 无 | 获取所有路由 |
| `/server/route/save` | POST | `{id?, remarks, match[], action, action_value?}` | 创建/更新路由 |
| `/server/route/drop` | POST | `{id}` | 删除路由 |

### 表格列

| 列 ID | 字段 | 宽度 | 说明 |
|-------|------|------|------|
| `id` | `id` | — | ID Badge |
| `remarks` | `remarks` | — | 备注名 |
| `action_value` | `action` + `action_value` + `match` | 300 | 动作值 + 匹配规则数 |
| `action` | `action` | 9000 | 动作类型 Badge |
| `actions` | — | — | 编辑/删除按钮 |

### 动作类型样式

```
block:  ShieldX  destructive red
dns:    Globe    secondary   blue
direct: ArrowRight secondary  green
proxy:  Wifi     default     purple
```

### 路由表单 Schema

```typescript
{
  id: number | optional
  remarks: string           // 必填
  match: string[] | object  // 匹配规则（每行一条）
  action: string            // "block" | "dns" | "direct" | "proxy"
  action_value: object | null  // dns 时为 DNS 地址，proxy 时代理标签
}
```

### 表单字段

1. 备注（必填）
2. 匹配规则（多行文本区域，每行一条）
3. 动作选择：block / dns / direct / proxy
4. DNS 服务器（仅 action="dns" 时显示）
5. 代理标签（仅 action="proxy" 时显示）

---

## 五、协议类型支持

```javascript
shadowsocks: "#489851"  vmess: "#CB3180"    trojan: "#EBB749"
hysteria: "#5684e6"     vless: "#1a1a1a"    tuic: "#00C853"
socks: "#2196F3"        naive: "#9C27B0"    http: "#FF5722"
mieru: "#4CAF50"        anytls: "#7E57C2"
```

---

## 六、关键实现要点

1. **节点表单**是所有页面中最复杂的，包含协议动态表单、高级配置（TLS/Multiplex/ECH/Outbounds/Routes）
2. **服务器详情弹窗**包含多个子组件：负载图表、Token管理、安装命令、关联节点
3. **批量操作**需要选中状态管理 + 确认弹窗
4. **拖拽排序**需要独立的排序模式切换
5. **URL 参数联动**支持从服务器管理页面跳转到节点管理并自动筛选
