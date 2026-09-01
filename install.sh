#!/bin/bash

# ═══════════════════════════════════════════════════════════════
# Xboard-Go 一键部署/更新脚本
# ═══════════════════════════════════════════════════════════════

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

REPO_OWNER="1712872354"
DOCKER_IMAGE="ghcr.io/${REPO_OWNER}/xboard-go:latest"
RELEASE_URL="https://github.com/${REPO_OWNER}/Xboard-Go/releases/latest/download"
DEFAULT_PORT=8080
DEFAULT_GRPC_PORT=50051
INSTALL_DIR="/opt/xboard-go"
DATA_DIR="${INSTALL_DIR}/data"
CONFIG_PATH="${DATA_DIR}/config.yaml"

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

generate_random_string() {
    openssl rand -base64 32 2>/dev/null | tr -dc 'a-zA-Z0-9' | head -c 32 || \
    cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 32
}

# 检测当前部署方式
detect_deployment() {
    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "xboard-go"; then
        CURRENT_DEPLOY="docker"
    elif systemctl is-active --quiet xboard-go 2>/dev/null; then
        CURRENT_DEPLOY="systemd"
    elif [[ -f "${INSTALL_DIR}/xboard-go" ]]; then
        CURRENT_DEPLOY="binary"
    else
        CURRENT_DEPLOY="none"
    fi
}

# 获取当前版本
get_current_version() {
    if [[ "$CURRENT_DEPLOY" == "docker" ]]; then
        docker inspect --format='{{.Config.Image}}' xboard-go 2>/dev/null | grep -oP ':\K.*' || echo "unknown"
    elif [[ -f "${INSTALL_DIR}/xboard-go" ]]; then
        ${INSTALL_DIR}/xboard-go --version 2>/dev/null || echo "unknown"
    else
        echo "unknown"
    fi
}

# ═══════════════════════════════════════════════════════════════
# 配置迁移：检查并补充缺失的配置项
# ═══════════════════════════════════════════════════════════════

migrate_config() {
    local cfg="${CONFIG_PATH}"
    local changed=false

    if [[ ! -f "$cfg" ]]; then
        return
    fi

    log_step "检查配置文件完整性..."

    # 检查 jwt 段
    if ! grep -q "^jwt:" "$cfg"; then
        log_warn "缺少 jwt 配置段，正在添加..."
        local jwt_secret=$(grep -oP 'key: \K.*' "$cfg" 2>/dev/null | head -1)
        if [[ -z "$jwt_secret" ]]; then
            jwt_secret=$(generate_random_string)
        fi
        cat >> "$cfg" << EOF

jwt:
  secret: "${jwt_secret}"
  access_token_ttl: 7200
  refresh_token_ttl: 604800
EOF
        changed=true
        log_info "已添加 jwt 配置 (access_token_ttl=7200, refresh_token_ttl=604800)"
    else
        # 检查 access_token_ttl 是否为 0 或缺失
        local ttl=$(grep -oP 'access_token_ttl: \K\d+' "$cfg" 2>/dev/null)
        if [[ -z "$ttl" || "$ttl" == "0" ]]; then
            log_warn "jwt.access_token_ttl 为 $ttl，修正为 7200..."
            sed -i 's/access_token_ttl:.*/access_token_ttl: 7200/' "$cfg"
            changed=true
        fi
        # 检查 secret 是否为空
        local secret=$(grep -oP 'secret: \K.*' "$cfg" 2>/dev/null | head -1 | tr -d '"' | tr -d "'")
        if [[ -z "$secret" ]]; then
            local new_secret=$(generate_random_string)
            sed -i "s/secret:.*/secret: \"${new_secret}\"/" "$cfg"
            changed=true
            log_info "已生成新的 jwt.secret"
        fi
    fi

    # 检查 log 段（修正 logger → log）
    if grep -q "^logger:" "$cfg"; then
        log_warn "发现 'logger:' 配置段，修正为 'log:'..."
        sed -i 's/^logger:/log:/' "$cfg"
        changed=true
    fi
    if ! grep -q "^log:" "$cfg"; then
        log_warn "缺少 log 配置段，正在添加..."
        cat >> "$cfg" << EOF

log:
  level: info
  format: json
  output: stdout
EOF
        changed=true
    fi

    # 检查 redis 段
    if ! grep -q "^redis:" "$cfg"; then
        log_warn "缺少 redis 配置段，正在添加..."
        cat >> "$cfg" << EOF

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
  pool_size: 100
EOF
        changed=true
    fi

    # 检查 rate_limit 段
    if ! grep -q "^rate_limit:" "$cfg"; then
        log_warn "缺少 rate_limit 配置段，正在添加..."
        cat >> "$cfg" << EOF

rate_limit:
  enabled: true
  ip_limit: 100
  user_limit: 300
  ip_whitelist:
    - "127.0.0.1"
    - "::1"
  path_whitelist:
    - "/healthz"
EOF
        changed=true
    fi

    if $changed; then
        log_info "配置迁移完成"
    else
        log_info "配置文件完整，无需迁移"
    fi
}

# ═══════════════════════════════════════════════════════════════
# 更新功能
# ═══════════════════════════════════════════════════════════════

update_docker() {
    log_step "更新 Docker 部署..."

    log_info "备份当前镜像..."
    docker tag ${DOCKER_IMAGE} ${DOCKER_IMAGE}:backup 2>/dev/null || true

    log_info "拉取最新镜像..."
    docker pull ${DOCKER_IMAGE}

    log_info "停止旧容器..."
    docker stop xboard-go 2>/dev/null || true
    docker rm xboard-go 2>/dev/null || true

    local http_port=$(grep -oP 'port: \K\d+' ${CONFIG_PATH} 2>/dev/null | head -1 || echo "8080")
    local grpc_port=$(grep -oP 'grpc:.*port: \K\d+' ${CONFIG_PATH} 2>/dev/null || echo "50051")

    log_info "启动新容器..."
    docker run -d \
        --name xboard-go \
        --restart always \
        -p ${http_port}:8080 \
        -p ${grpc_port}:50051 \
        -v ${DATA_DIR}:/data \
        ${DOCKER_IMAGE}

    log_info "Docker 更新完成"
}

update_binary() {
    log_step "更新二进制部署..."

    # 迁移配置文件
    migrate_config

    # 备份当前版本
    if [[ -f "${INSTALL_DIR}/xboard-go" ]]; then
        log_info "备份当前版本..."
        cp "${INSTALL_DIR}/xboard-go" "${INSTALL_DIR}/xboard-go.bak"
    fi

    # 停止服务
    log_info "停止当前服务..."
    systemctl stop xboard-go 2>/dev/null || true
    sleep 1
    killall xboard-go 2>/dev/null || true
    pkill -9 -f "xboard-go" 2>/dev/null || true
    sleep 2

    # 检测架构
    local arch=$(uname -m)
    case $arch in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) log_error "不支持的架构: $arch"; exit 1 ;;
    esac

    # 下载到临时文件
    local binary_url="${RELEASE_URL}/xboard-go-linux-${arch}"
    local tmp_file="${INSTALL_DIR}/xboard-go.new"
    log_info "下载: ${binary_url}"

    if command -v curl &> /dev/null; then
        curl -L -o "${tmp_file}" "${binary_url}"
    elif command -v wget &> /dev/null; then
        wget -O "${tmp_file}" "${binary_url}"
    else
        log_error "需要 curl 或 wget"
        exit 1
    fi

    # 替换二进制
    mv "${tmp_file}" "${INSTALL_DIR}/xboard-go"
    chmod +x "${INSTALL_DIR}/xboard-go"

    # 启动服务
    log_info "启动服务..."
    systemctl start xboard-go 2>/dev/null || {
        log_warn "systemd 启动失败，使用 nohup..."
        nohup ${INSTALL_DIR}/xboard-go -config ${CONFIG_PATH} > ${DATA_DIR}/xboard-go.log 2>&1 &
        echo $! > ${DATA_DIR}/xboard-go.pid
    }

    # 健康检查
    health_check

    log_info "二进制更新完成"
}

health_check() {
    local http_port=$(grep -oP 'port: \K\d+' ${CONFIG_PATH} 2>/dev/null | head -1 || echo "8080")
    log_info "健康检查..."
    for i in $(seq 1 15); do
        if curl -s "http://localhost:${http_port}/healthz" > /dev/null 2>&1; then
            log_info "服务运行正常"
            return 0
        fi
        sleep 1
    done
    log_error "健康检查失败，请检查日志: journalctl -u xboard-go -f"
    return 1
}

# 回滚
rollback() {
    log_step "回滚到上一版本..."

    if [[ "$CURRENT_DEPLOY" == "docker" ]]; then
        docker stop xboard-go 2>/dev/null || true
        docker rm xboard-go 2>/dev/null || true
        docker tag ${DOCKER_IMAGE}:backup ${DOCKER_IMAGE}
        update_docker
    elif [[ -f "${INSTALL_DIR}/xboard-go.bak" ]]; then
        systemctl stop xboard-go 2>/dev/null || true
        cp "${INSTALL_DIR}/xboard-go.bak" "${INSTALL_DIR}/xboard-go"
        chmod +x "${INSTALL_DIR}/xboard-go"
        systemctl start xboard-go 2>/dev/null || true
        health_check
        log_info "回滚完成"
    else
        log_error "没有找到备份版本"
    fi
}

# ═══════════════════════════════════════════════════════════════
# 安装功能
# ═══════════════════════════════════════════════════════════════

install_docker() {
    log_step "安装 Docker..."
    curl -fsSL https://get.docker.com | bash
    systemctl enable docker && systemctl start docker
    log_info "Docker 安装完成"
}

deploy_docker() {
    log_step "Docker 部署..."

    if ! command -v docker &> /dev/null; then
        install_docker
    fi

    docker stop xboard-go 2>/dev/null || true
    docker rm xboard-go 2>/dev/null || true

    log_info "拉取镜像..."
    docker pull ${DOCKER_IMAGE}

    log_info "启动容器..."
    docker run -d \
        --name xboard-go \
        --restart always \
        -p ${HTTP_PORT}:8080 \
        -p ${GRPC_PORT}:50051 \
        -v ${DATA_DIR}:/data \
        ${DOCKER_IMAGE}

    log_info "Docker 容器已启动"
}

deploy_binary() {
    log_step "二进制部署..."
    mkdir -p "${INSTALL_DIR}" "${DATA_DIR}"

    local arch=$(uname -m)
    case $arch in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) log_error "不支持的架构: $arch"; exit 1 ;;
    esac

    local binary_url="${RELEASE_URL}/xboard-go-linux-${arch}"
    log_info "下载: ${binary_url}"

    curl -L -o "${INSTALL_DIR}/xboard-go" "${binary_url}" || \
    wget -O "${INSTALL_DIR}/xboard-go" "${binary_url}"
    chmod +x "${INSTALL_DIR}/xboard-go"

    cat > /etc/systemd/system/xboard-go.service << EOF
[Unit]
Description=Xboard-Go
After=network.target

[Service]
Type=simple
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/xboard-go -config ${CONFIG_PATH}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable xboard-go
    systemctl start xboard-go
    log_info "服务已启动"
}

generate_config() {
    log_step "生成配置文件..."
    mkdir -p "${DATA_DIR}"

    # 构建数据库配置段
    local db_section=""
    case $DB_DRIVER in
        sqlite)
            db_section="  driver: sqlite
  dbname: ${DATA_DIR}/xboard.db"
            ;;
        mysql)
            db_section="  driver: mysql
  host: ${DB_HOST}
  port: ${DB_PORT}
  user: ${DB_USER}
  password: \"${DB_PASS}\"
  dbname: ${DB_NAME}
  max_idle_conns: 10
  max_open_conns: 100"
            ;;
        postgres)
            db_section="  driver: postgres
  host: ${DB_HOST}
  port: ${DB_PORT}
  user: ${DB_USER}
  password: \"${DB_PASS}\"
  dbname: ${DB_NAME}
  sslmode: disable
  max_idle_conns: 10
  max_open_conns: 100"
            ;;
    esac

    cat > "${CONFIG_PATH}" << EOF
server:
  host: "0.0.0.0"
  port: ${HTTP_PORT}
  mode: release

database:
${db_section}

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
  pool_size: 100

grpc:
  enabled: true
  port: ${GRPC_PORT}

app:
  name: ${SITE_NAME}
  node_api_key: ${NODE_API_KEY}
  default_user_role: user

jwt:
  secret: "${APP_KEY}"
  access_token_ttl: 7200
  refresh_token_ttl: 604800

log:
  level: info
  format: json
  output: stdout

rate_limit:
  enabled: true
  ip_limit: 100
  user_limit: 300
  ip_whitelist:
    - "127.0.0.1"
    - "::1"
  path_whitelist:
    - "/healthz"
EOF

    log_info "配置已生成: ${CONFIG_PATH}"
}

init_admin() {
    log_step "初始化管理员..."
    log_info "等待服务启动..."

    for i in $(seq 1 30); do
        if curl -s "http://localhost:${HTTP_PORT}/healthz" > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done

    # 注册管理员账户
    log_info "创建管理员账户..."
    curl -s -X POST "http://localhost:${HTTP_PORT}/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASS}\"}" > /dev/null 2>&1 || true

    # 自动设置管理员角色
    log_info "设置管理员角色..."
    sleep 1
    case $DB_DRIVER in
        sqlite)
            if command -v sqlite3 &> /dev/null; then
                sqlite3 "${DATA_DIR}/xboard.db" "UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';" 2>/dev/null
                log_info "管理员角色已设置 (SQLite)"
            else
                log_warn "sqlite3 未安装，请手动执行:"
                echo "  sqlite3 ${DATA_DIR}/xboard.db \"UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';\""
            fi
            ;;
        mysql)
            mysql -u "${DB_USER}" -p"${DB_PASS}" "${DB_NAME}" \
                -e "UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';" 2>/dev/null
            log_info "管理员角色已设置 (MySQL)"
            ;;
        postgres)
            PGPASSWORD="${DB_PASS}" psql -U "${DB_USER}" -d "${DB_NAME}" \
                -c "UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';" 2>/dev/null
            log_info "管理员角色已设置 (PostgreSQL)"
            ;;
    esac
}

# ═══════════════════════════════════════════════════════════════
# 状态查看
# ═══════════════════════════════════════════════════════════════

show_status() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}  Xboard-Go 状态${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo ""

    detect_deployment
    echo -e "  部署方式: ${CURRENT_DEPLOY}"

    if [[ "$CURRENT_DEPLOY" == "docker" ]]; then
        echo -e "  容器状态: $(docker ps --filter name=xboard-go --format '{{.Status}}')"
        echo -e "  镜像版本: $(docker inspect --format='{{.Config.Image}}' xboard-go)"
        echo ""
        echo "  日志查看: docker logs -f xboard-go"
        echo "  重启服务: docker restart xboard-go"
        echo "  停止服务: docker stop xboard-go"
    elif [[ "$CURRENT_DEPLOY" == "systemd" ]] || [[ "$CURRENT_DEPLOY" == "binary" ]]; then
        echo -e "  服务状态: $(systemctl is-active xboard-go 2>/dev/null || echo 'unknown')"
        if [[ -f "${INSTALL_DIR}/xboard-go" ]]; then
            echo -e "  版本: $(${INSTALL_DIR}/xboard-go --version 2>/dev/null || echo 'unknown')"
        fi
        echo ""
        echo "  查看状态: systemctl status xboard-go"
        echo "  查看日志: journalctl -u xboard-go -f"
        echo "  重启服务: systemctl restart xboard-go"
        echo "  停止服务: systemctl stop xboard-go"
    else
        echo -e "  状态: 未安装"
    fi
    echo ""
}

# ═══════════════════════════════════════════════════════════════
# 卸载
# ═══════════════════════════════════════════════════════════════

uninstall() {
    log_step "卸载 Xboard-Go..."

    read -p "确定要卸载吗? 数据将保留 [y/N]: " confirm < /dev/tty
    if [[ "${confirm,,}" != "y" ]]; then
        log_info "取消卸载"
        return
    fi

    detect_deployment

    if [[ "$CURRENT_DEPLOY" == "docker" ]]; then
        docker stop xboard-go 2>/dev/null || true
        docker rm xboard-go 2>/dev/null || true
        docker rmi ${DOCKER_IMAGE} 2>/dev/null || true
    elif [[ "$CURRENT_DEPLOY" == "systemd" ]]; then
        systemctl stop xboard-go 2>/dev/null || true
        systemctl disable xboard-go 2>/dev/null || true
        rm -f /etc/systemd/system/xboard-go.service
        systemctl daemon-reload
    fi

    rm -f "${INSTALL_DIR}/xboard-go"
    rm -f "${INSTALL_DIR}/xboard-go.bak"

    log_info "卸载完成 (数据保留在 ${DATA_DIR})"
}

# ═══════════════════════════════════════════════════════════════
# 主菜单
# ═══════════════════════════════════════════════════════════════

main_menu() {
    echo -e "${CYAN}"
    echo "  ╔═══════════════════════════════════════════════════╗"
    echo "  ║       Xboard-Go 管理脚本                          ║"
    echo "  ╚═══════════════════════════════════════════════════╝"
    echo -e "${NC}"

    detect_deployment

    echo "  1) 全新安装"
    echo "  2) 更新版本"
    echo "  3) 查看状态"
    echo "  4) 回滚版本"
    echo "  5) 卸载"
    echo "  6) 退出"
    echo ""

    read -p "请选择操作 [1-6]: " choice < /dev/tty

    case $choice in
        1)
            collect_install_config
            if [[ "$DEPLOY_METHOD" == "1" ]]; then
                deploy_docker
            else
                deploy_binary
            fi
            init_admin
            show_install_result
            ;;
        2)
            if [[ "$CURRENT_DEPLOY" == "none" ]]; then
                log_error "未检测到安装，请先安装"
                return
            fi
            if [[ "$CURRENT_DEPLOY" == "docker" ]]; then
                update_docker
            else
                update_binary
            fi
            show_status
            ;;
        3)
            show_status
            ;;
        4)
            rollback
            ;;
        5)
            uninstall
            ;;
        6)
            exit 0
            ;;
        *)
            log_error "无效选项"
            ;;
    esac
}

collect_install_config() {
    echo ""
    echo -e "${YELLOW}请选择部署方式:${NC}"
    echo "  1) Docker 部署 (推荐)"
    echo "  2) 二进制部署"
    echo ""
    read -p "请输入选项 [1]: " DEPLOY_METHOD < /dev/tty
    DEPLOY_METHOD=${DEPLOY_METHOD:-1}

    echo ""
    read -p "HTTP 端口 [${DEFAULT_PORT}]: " HTTP_PORT < /dev/tty
    HTTP_PORT=${HTTP_PORT:-$DEFAULT_PORT}

    read -p "gRPC 端口 [${DEFAULT_GRPC_PORT}]: " GRPC_PORT < /dev/tty
    GRPC_PORT=${GRPC_PORT:-$DEFAULT_GRPC_PORT}

    echo ""
    echo -e "${YELLOW}数据库:${NC}"
    echo "  1) SQLite (推荐)"
    echo "  2) MySQL"
    echo "  3) PostgreSQL"
    echo ""
    read -p "请选择 [1]: " DB_TYPE < /dev/tty
    DB_TYPE=${DB_TYPE:-1}

    case $DB_TYPE in
        1)
            DB_DRIVER="sqlite"
            ;;
        2)
            DB_DRIVER="mysql"
            read -p "MySQL 主机 [localhost]: " DB_HOST < /dev/tty; DB_HOST=${DB_HOST:-localhost}
            read -p "MySQL 端口 [3306]: " DB_PORT < /dev/tty; DB_PORT=${DB_PORT:-3306}
            read -p "MySQL 数据库 [xboard]: " DB_NAME < /dev/tty; DB_NAME=${DB_NAME:-xboard}
            read -p "MySQL 用户 [root]: " DB_USER < /dev/tty; DB_USER=${DB_USER:-root}
            read -sp "MySQL 密码: " DB_PASS < /dev/tty; echo ""
            ;;
        3)
            DB_DRIVER="postgres"
            read -p "PostgreSQL 主机 [localhost]: " DB_HOST < /dev/tty; DB_HOST=${DB_HOST:-localhost}
            read -p "PostgreSQL 端口 [5432]: " DB_PORT < /dev/tty; DB_PORT=${DB_PORT:-5432}
            read -p "PostgreSQL 数据库 [xboard]: " DB_NAME < /dev/tty; DB_NAME=${DB_NAME:-xboard}
            read -p "PostgreSQL 用户 [postgres]: " DB_USER < /dev/tty; DB_USER=${DB_USER:-postgres}
            read -sp "PostgreSQL 密码: " DB_PASS < /dev/tty; echo ""
            ;;
    esac

    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}  管理员账户${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo ""

    read -p "管理员邮箱 [admin@example.com]: " ADMIN_EMAIL < /dev/tty
    ADMIN_EMAIL=${ADMIN_EMAIL:-admin@example.com}

    read -sp "管理员密码 (留空自动生成): " ADMIN_PASS < /dev/tty; echo ""
    if [[ -z "$ADMIN_PASS" ]]; then
        ADMIN_PASS=$(generate_random_string)
        log_info "生成密码: ${ADMIN_PASS}"
    fi

    echo ""
    read -p "站点名称 [Xboard-Go]: " SITE_NAME < /dev/tty; SITE_NAME=${SITE_NAME:-Xboard-Go}

    APP_KEY=$(generate_random_string)
    NODE_API_KEY=$(generate_random_string)

    generate_config
}

show_install_result() {
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  部署完成！${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
    echo ""
    echo "  访问地址:"
    echo "    用户面板: http://localhost:${HTTP_PORT}/user/login"
    echo "    管理后台: http://localhost:${HTTP_PORT}/admin/login"
    echo ""
    echo "  管理员账户:"
    echo "    邮箱: ${ADMIN_EMAIL}"
    echo "    密码: ${ADMIN_PASS}"
    echo ""
    echo "  配置文件: ${CONFIG_PATH}"
    echo "  节点密钥: ${NODE_API_KEY}"
    echo ""
}

# 运行
main_menu
