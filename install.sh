#!/bin/bash

# ═══════════════════════════════════════════════════════════════
# Xboard-Go 一键部署脚本
# 支持 Linux / macOS / Windows (WSL)
# 支持 Docker / 二进制部署
# ═══════════════════════════════════════════════════════════════

# 如果是通过管道运行的，先下载到临时文件再执行
if [[ ! -t 0 ]] && [[ -z "$XBOARD_REEXEC" ]]; then
    echo "检测到管道运行模式，正在下载脚本..."
    TMP_SCRIPT="/tmp/xboard-install-$$.sh"
    if command -v curl &> /dev/null; then
        curl -fsSL "https://raw.githubusercontent.com/1712872354/Xboard-Go/master/install.sh" -o "$TMP_SCRIPT"
    elif command -v wget &> /dev/null; then
        wget -q "https://raw.githubusercontent.com/1712872354/Xboard-Go/master/install.sh" -O "$TMP_SCRIPT"
    else
        echo "错误: 需要 curl 或 wget"
        exit 1
    fi
    chmod +x "$TMP_SCRIPT"
    export XBOARD_REEXEC=1
    exec bash "$TMP_SCRIPT" "$@"
fi

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 项目信息
REPO_OWNER="1712872354"
REPO_NAME="Xboard-Go"
REPO_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}"
RELEASE_URL="${REPO_URL}/releases/latest/download"
DOCKER_IMAGE="ghcr.io/${REPO_OWNER}/xboard-go:latest"
DEFAULT_PORT=8080
DEFAULT_GRPC_PORT=50051
INSTALL_DIR="/opt/xboard-go"
DATA_DIR="${INSTALL_DIR}/data"

# ═══════════════════════════════════════════════════════════════
# 工具函数
# ═══════════════════════════════════════════════════════════════

print_banner() {
    echo -e "${CYAN}"
    echo "  ╔═══════════════════════════════════════════════════╗"
    echo "  ║                                                   ║"
    echo "  ║       Xboard-Go 一键部署脚本                      ║"
    echo "  ║       高性能代理面板管理系统                       ║"
    echo "  ║                                                   ║"
    echo "  ╚═══════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="darwin"
    else
        OS="unknown"
    fi
}

detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) log_error "不支持的架构: $ARCH"; exit 1 ;;
    esac
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "请使用 root 权限运行此脚本"
        echo "  sudo bash $0"
        exit 1
    fi
}

check_docker() {
    if command -v docker &> /dev/null; then
        DOCKER_AVAILABLE=true
    else
        DOCKER_AVAILABLE=false
    fi
}

generate_random_string() {
    local length=${1:-32}
    openssl rand -base64 $length 2>/dev/null | tr -dc 'a-zA-Z0-9' | head -c $length || \
    cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c $length
}

# ═══════════════════════════════════════════════════════════════
# 配置收集
# ═══════════════════════════════════════════════════════════════

collect_config() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}  配置信息收集${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo ""

    echo -e "${YELLOW}请选择部署方式:${NC}"
    echo "  1) Docker 部署 (推荐，简单快速)"
    echo "  2) 二进制部署 (直接运行)"
    echo ""
    read -p "请输入选项 [1]: " DEPLOY_METHOD < /dev/tty
    DEPLOY_METHOD=${DEPLOY_METHOD:-1}

    echo ""
    read -p "请输入 HTTP 端口 [${DEFAULT_PORT}]: " HTTP_PORT < /dev/tty
    HTTP_PORT=${HTTP_PORT:-$DEFAULT_PORT}

    read -p "请输入 gRPC 端口 [${DEFAULT_GRPC_PORT}]: " GRPC_PORT < /dev/tty
    GRPC_PORT=${GRPC_PORT:-$DEFAULT_GRPC_PORT}

    echo ""
    echo -e "${YELLOW}请选择数据库:${NC}"
    echo "  1) SQLite (推荐)"
    echo "  2) MySQL"
    echo "  3) PostgreSQL"
    echo ""
    read -p "请输入选项 [1]: " DB_TYPE < /dev/tty
    DB_TYPE=${DB_TYPE:-1}

    case $DB_TYPE in
        1)
            DB_DRIVER="sqlite"
            DB_SOURCE="${DATA_DIR}/xboard.db"
            ;;
        2)
            DB_DRIVER="mysql"
            read -p "MySQL 主机 [localhost]: " DB_HOST < /dev/tty
            DB_HOST=${DB_HOST:-localhost}
            read -p "MySQL 端口 [3306]: " DB_PORT < /dev/tty
            DB_PORT=${DB_PORT:-3306}
            read -p "MySQL 数据库名 [xboard]: " DB_NAME < /dev/tty
            DB_NAME=${DB_NAME:-xboard}
            read -p "MySQL 用户名 [root]: " DB_USER < /dev/tty
            DB_USER=${DB_USER:-root}
            read -sp "MySQL 密码: " DB_PASS < /dev/tty
            echo ""
            DB_SOURCE="${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?charset=utf8mb4&parseTime=True&loc=Local"
            ;;
        3)
            DB_DRIVER="postgres"
            read -p "PostgreSQL 主机 [localhost]: " DB_HOST < /dev/tty
            DB_HOST=${DB_HOST:-localhost}
            read -p "PostgreSQL 端口 [5432]: " DB_PORT < /dev/tty
            DB_PORT=${DB_PORT:-5432}
            read -p "PostgreSQL 数据库名 [xboard]: " DB_NAME < /dev/tty
            DB_NAME=${DB_NAME:-xboard}
            read -p "PostgreSQL 用户名 [postgres]: " DB_USER < /dev/tty
            DB_USER=${DB_USER:-postgres}
            read -sp "PostgreSQL 密码: " DB_PASS < /dev/tty
            echo ""
            DB_SOURCE="host=${DB_HOST} user=${DB_USER} password=${DB_PASS} dbname=${DB_NAME} port=${DB_PORT} sslmode=disable"
            ;;
    esac

    echo ""
    read -p "是否配置 Redis? [y/N]: " USE_REDIS < /dev/tty
    USE_REDIS=${USE_REDIS:-N}

    if [[ "${USE_REDIS,,}" == "y" ]]; then
        read -p "Redis 地址 [localhost:6379]: " REDIS_ADDR < /dev/tty
        REDIS_ADDR=${REDIS_ADDR:-localhost:6379}
        read -sp "Redis 密码 (无密码回车): " REDIS_PASS < /dev/tty
        echo ""
        read -p "Redis 数据库 [0]: " REDIS_DB < /dev/tty
        REDIS_DB=${REDIS_DB:-0}
    fi

    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}  管理员账户配置${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo ""

    read -p "管理员邮箱 [admin@example.com]: " ADMIN_EMAIL < /dev/tty
    ADMIN_EMAIL=${ADMIN_EMAIL:-admin@example.com}

    read -sp "管理员密码: " ADMIN_PASS < /dev/tty
    echo ""
    if [[ -z "$ADMIN_PASS" ]]; then
        ADMIN_PASS=$(generate_random_string 12)
        log_info "已生成随机密码: ${ADMIN_PASS}"
    fi

    echo ""
    read -p "站点名称 [Xboard-Go]: " SITE_NAME < /dev/tty
    SITE_NAME=${SITE_NAME:-Xboard-Go}

    read -p "站点 URL [http://localhost:${HTTP_PORT}]: " SITE_URL < /dev/tty
    SITE_URL=${SITE_URL:-"http://localhost:${HTTP_PORT}"}

    APP_KEY=$(generate_random_string 32)
    NODE_API_KEY=$(generate_random_string 32)
}

# ═══════════════════════════════════════════════════════════════
# 生成配置文件
# ═══════════════════════════════════════════════════════════════

generate_config() {
    log_step "生成配置文件..."
    mkdir -p "${DATA_DIR}"

    cat > "${DATA_DIR}/config.yaml" << EOF
server:
  host: "0.0.0.0"
  port: ${HTTP_PORT}
  mode: release

database:
  driver: ${DB_DRIVER}
  source: ${DB_SOURCE}
EOF

    if [[ "${USE_REDIS,,}" == "y" ]]; then
        cat >> "${DATA_DIR}/config.yaml" << EOF

redis:
  addr: ${REDIS_ADDR}
  password: "${REDIS_PASS}"
  db: ${REDIS_DB}
EOF
    fi

    cat >> "${DATA_DIR}/config.yaml" << EOF

grpc:
  enabled: true
  port: ${GRPC_PORT}

app:
  name: ${SITE_NAME}
  key: ${APP_KEY}
  node_api_key: ${NODE_API_KEY}
  default_user_role: user
  subscribe_token_length: 32

rate_limit:
  enabled: false
  ip_limit: 100
  user_limit: 200

logger:
  level: info
  format: json
EOF

    log_info "配置文件已生成: ${DATA_DIR}/config.yaml"
}

# ═══════════════════════════════════════════════════════════════
# 安装 Docker
# ═══════════════════════════════════════════════════════════════

install_docker() {
    if [[ "$DOCKER_AVAILABLE" == "true" ]]; then
        log_info "Docker 已安装"
        return
    fi

    log_step "安装 Docker..."
    curl -fsSL https://get.docker.com | bash
    systemctl enable docker
    systemctl start docker
    log_info "Docker 安装完成"
}

# ═══════════════════════════════════════════════════════════════
# Docker 部署
# ═══════════════════════════════════════════════════════════════

deploy_docker() {
    log_step "使用 Docker 部署..."

    docker stop xboard-go 2>/dev/null || true
    docker rm xboard-go 2>/dev/null || true

    log_info "拉取 Docker 镜像..."
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

# ═══════════════════════════════════════════════════════════════
# 二进制部署
# ═══════════════════════════════════════════════════════════════

deploy_binary() {
    log_step "使用二进制部署..."
    mkdir -p "${INSTALL_DIR}"

    local binary_name="xboard-go-${OS}-${ARCH}"
    local download_url="${RELEASE_URL}/${binary_name}"
    log_info "下载: ${download_url}"

    curl -L -o "${INSTALL_DIR}/xboard-go" "${download_url}" || \
    wget -O "${INSTALL_DIR}/xboard-go" "${download_url}"

    chmod +x "${INSTALL_DIR}/xboard-go"

    cat > /etc/systemd/system/xboard-go.service << EOF
[Unit]
Description=Xboard-Go
After=network.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/xboard-go -config ${DATA_DIR}/config.yaml
Restart=always

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable xboard-go
    systemctl start xboard-go
    log_info "服务已启动"
}

# ═══════════════════════════════════════════════════════════════
# 初始化管理员
# ═══════════════════════════════════════════════════════════════

init_admin() {
    log_step "初始化管理员..."
    log_info "等待服务启动..."

    for i in $(seq 1 30); do
        if curl -s "http://localhost:${HTTP_PORT}/healthz" > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done

    log_info "创建管理员账户..."
    curl -s -X POST "http://localhost:${HTTP_PORT}/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASS}\"}" > /dev/null 2>&1 || true

    log_warn "请手动设置管理员角色:"
    if [[ "$DB_DRIVER" == "sqlite" ]]; then
        echo "  sqlite3 ${DATA_DIR}/xboard.db \"UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';\""
    elif [[ "$DB_DRIVER" == "mysql" ]]; then
        echo "  mysql -u ${DB_USER} -p ${DB_NAME} -e \"UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';\""
    elif [[ "$DB_DRIVER" == "postgres" ]]; then
        echo "  psql -U ${DB_USER} -d ${DB_NAME} -c \"UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';\""
    fi
}

# ═══════════════════════════════════════════════════════════════
# 显示结果
# ═══════════════════════════════════════════════════════════════

show_result() {
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  ✅ 部署完成！${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "  访问地址:"
    echo -e "    用户面板: ${SITE_URL}/user/login"
    echo -e "    管理后台: ${SITE_URL}/admin/login"
    echo ""
    echo -e "  管理员账户:"
    echo -e "    邮箱: ${ADMIN_EMAIL}"
    echo -e "    密码: ${ADMIN_PASS}"
    echo ""
    echo -e "  配置文件: ${DATA_DIR}/config.yaml"
    echo -e "  节点密钥: ${NODE_API_KEY}"
    echo ""
}

# ═══════════════════════════════════════════════════════════════
# 主流程
# ═══════════════════════════════════════════════════════════════

main() {
    print_banner
    detect_os
    detect_arch
    log_info "系统: ${OS} (${ARCH})"
    check_root
    check_docker
    collect_config

    if [[ "$DEPLOY_METHOD" == "1" ]]; then
        install_docker
        generate_config
        deploy_docker
    else
        generate_config
        deploy_binary
    fi

    init_admin
    show_result
}

main
