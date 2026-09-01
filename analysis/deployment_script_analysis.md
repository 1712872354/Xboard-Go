# 部署脚本和配置文件缺陷分析报告

## 发现的问题

### 1. 配置文件字段不一致

| 问题 | config.yaml.example | config.go 定义 | install.sh | install.ps1 |
|------|---------------------|----------------|------------|-------------|
| 数据库配置 | `source` | 结构化字段 (host, port, dbname) | 结构化字段 ✅ | 结构化字段 ✅ |
| JWT 密钥 | `key` | `secret` | `secret` ✅ | `secret` ✅ |
| 日志配置段 | `logger` | `log` | `log` ✅ | `log` ✅ |
| 缺少配置 | 缺少 `captcha` | 有 `captcha` | 缺少 | 缺少 |
| 缺少配置 | 缺少 `subscribe_token_length` | 有此字段 | 缺少 | 缺少 |

### 2. install.sh 问题

1. **数据库配置字段名**：使用 `dbname` 但应该是 `source`（根据 config.yaml.example）
2. **缺少 subscribe_token_length**：配置文件中缺少此字段
3. **缺少 captcha 配置段**：配置文件中缺少验证码配置
4. **配置迁移不完整**：migrate_config 函数没有检查所有必要字段

### 3. install.ps1 问题

1. **SQLite 路径**：使用正斜杠 `/`，Windows 应该使用反斜杠 `\`
2. **缺少 subscribe_token_length**：配置文件中缺少此字段
3. **缺少 captcha 配置段**：配置文件中缺少验证码配置

### 4. config.yaml.example 问题

1. **数据库配置格式**：使用 `source` 字段，但代码期望结构化字段
2. **日志配置段名**：使用 `logger`，但代码期望 `log`
3. **JWT 密钥字段名**：使用 `key`，但代码期望 `secret`
4. **缺少配置段**：缺少 `captcha`、`subscribe_token_length`

---

## 修复方案

### 1. 更新 config.yaml.example

需要更新为与 config.go 定义一致的格式。

### 2. 更新 install.sh

- 修复数据库配置字段
- 添加缺失的配置项
- 完善配置迁移函数

### 3. 更新 install.ps1

- 修复 SQLite 路径
- 添加缺失的配置项

---

## 修复后的配置文件模板

```yaml
# Xboard-Go 配置文件

server:
  host: "0.0.0.0"
  port: 8080
  mode: release
  allowed_origins: []

database:
  driver: sqlite
  dbname: data/xboard.db
  # MySQL 配置示例:
  # driver: mysql
  # host: localhost
  # port: 3306
  # user: root
  # password: "your_password"
  # dbname: xboard
  # max_idle_conns: 10
  # max_open_conns: 100
  
  # PostgreSQL 配置示例:
  # driver: postgres
  # host: localhost
  # port: 5432
  # user: postgres
  # password: "your_password"
  # dbname: xboard
  # sslmode: disable
  # max_idle_conns: 10
  # max_open_conns: 100

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
  pool_size: 100

grpc:
  enabled: true
  port: 50051

app:
  name: Xboard-Go
  node_api_key: change-me-node-api-key
  default_user_role: user
  subscribe_token_length: 32
  bcrypt_cost: 10

jwt:
  secret: change-me-to-a-random-secret-key
  access_token_ttl: 7200
  refresh_token_ttl: 604800

log:
  level: info
  format: json
  output: stdout
  # file_path: /var/log/xboard-go.log

rate_limit:
  enabled: true
  ip_limit: 100
  user_limit: 300
  ip_whitelist:
    - "127.0.0.1"
    - "::1"
  path_whitelist:
    - "/healthz"

# 验证码配置 (可选)
# captcha:
#   provider: turnstile  # turnstile, recaptcha_v2, recaptcha_v3
#   site_key: your_site_key
#   secret_key: your_secret_key
#   min_score: 0.5  # reCAPTCHA v3 最小分数
```
