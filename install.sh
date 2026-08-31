#!/bin/bash

# ═══════════════════════════════════════════════════════════════
# Xboard-Go 一键部署脚本
# 支持 Linux / macOS / Windows (WSL)
# 支持 Docker / 二进制部署
# ═══════════════════════════════════════════════════════════════

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

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

# 检测操作系统
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="darwin"
    elif [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
        OS="windows"
    else
        OS="unknown"
    fi
}

# 检测架构
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l|armhf)
            ARCH="arm"
            ;;
        *)
            log_error "不支持的架构: $ARCH"
            exit 1
            ;;
    esac
}

# 检测是否有 root 权限
check_root() {
    if [[ $EUID -ne 0 ]] && [[ "$OS" != "windows" ]]; then
        log_error "请使用 root 权限运行此脚本"
        echo "  sudo bash $0"
        exit 1
    fi
}

# 检测 Docker 是否可用
check_docker() {
    if command -v docker &> /dev/null; then
        DOCKER_AVAILABLE=true
        DOCKER_VERSION=$(docker --version | grep -oP '\d+\.\d+\.\d+' | head -1)
    else
        DOCKER_AVAILABLE=false
    fi
}

# 生成随机字符串
generate_random_string() {
    local length=${1:-32}
    if command -v openssl &> /dev/null; then
        openssl rand -base64 $length | tr -dc 'a-zA-Z0-9' | head -c $length
    else
        cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c $length
    fi
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

    # 部署方式
    echo -e "${YELLOW}请选择部署方式:${NC}"
    echo "  1) Docker 部署 (推荐，简单快速)"
    echo "  2) 二进制部署 (直接运行)"
    echo ""
    read -p "请输入选项 [1]: " DEPLOY_METHOD
    DEPLOY_METHOD=${DEPLOY_METHOD:-1}

    if [[ "$DEPLOY_METHOD" == "1" ]] && [[ "$DOCKER_AVAILABLE" == "false" ]]; then
        log_warn "Docker 未安装，将自动安装 Docker"
        INSTALL_DOCKER=true
    fi

    # 端口配置
    echo ""
    read -p "请输入 HTTP 端口 [${DEFAULT_PORT}]: " HTTP_PORT
    HTTP_PORT=${HTTP_PORT:-$DEFAULT_PORT}

    read -p "请输入 gRPC 端口 [${DEFAULT_GRPC_PORT}]: " GRPC_PORT
    GRPC_PORT=${GRPC_PORT:-$DEFAULT_GRPC_PORT}

    # 数据库配置
    echo ""
    echo -e "${YELLOW}请选择数据库:${NC}"
    echo "  1) SQLite (推荐，无需额外配置)"
    echo "  2) MySQL"
    echo "  3) PostgreSQL"
    echo ""
    read -p "请输入选项 [1]: " DB_TYPE
    DB_TYPE=${DB_TYPE:-1}

    case $DB_TYPE in
        1)
            DB_DRIVER="sqlite"
            DB_SOURCE="${DATA_DIR}/xboard.db"
            ;;
        2)
            DB_DRIVER="mysql"
            read -p "MySQL 主机 [localhost]: " DB_HOST
            DB_HOST=${DB_HOST:-localhost}
            read -p "MySQL 端口 [3306]: " DB_PORT
            DB_PORT=${DB_PORT:-3306}
            read -p "MySQL 数据库名 [xboard]: " DB_NAME
            DB_NAME=${DB_NAME:-xboard}
            read -p "MySQL 用户名 [root]: " DB_USER
            DB_USER=${DB_USER:-root}
            read -sp "MySQL 密码: " DB_PASS
            echo ""
            DB_SOURCE="${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?charset=utf8mb4&parseTime=True&loc=Local"
            ;;
        3)
            DB_DRIVER="postgres"
            read -p "PostgreSQL 主机 [localhost]: " DB_HOST
            DB_HOST=${DB_HOST:-localhost}
            read -p "PostgreSQL 端口 [5432]: " DB_PORT
            DB_PORT=${DB_PORT:-5432}
            read -p "PostgreSQL 数据库名 [xboard]: " DB_NAME
            DB_NAME=${DB_NAME:-xboard}
            read -p "PostgreSQL 用户名 [postgres]: " DB_USER
            DB_USER=${DB_USER:-postgres}
            read -sp "PostgreSQL 密码: " DB_PASS
            echo ""
            DB_SOURCE="host=${DB_HOST} user=${DB_USER} password=${DB_PASS} dbname=${DB_NAME} port=${DB_PORT} sslmode=disable"
            ;;
    esac

    # Redis 配置 (可选)
    echo ""
    read -p "是否配置 Redis? (用于缓存和限流) [y/N]: " USE_REDIS
    USE_REDIS=${USE_REDIS:-N}

    if [[ "${USE_REDIS,,}" == "y" ]]; then
        read -p "Redis 地址 [localhost:6379]: " REDIS_ADDR
        REDIS_ADDR=${REDIS_ADDR:-localhost:6379}
        read -sp "Redis 密码 (无密码直接回车): " REDIS_PASS
        echo ""
        read -p "Redis 数据库 [0]: " REDIS_DB
        REDIS_DB=${REDIS_DB:-0}
    fi

    # 管理员配置
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}  管理员账户配置${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo ""

    read -p "管理员邮箱 [admin@example.com]: " ADMIN_EMAIL
    ADMIN_EMAIL=${ADMIN_EMAIL:-admin@example.com}

    read -sp "管理员密码: " ADMIN_PASS
    echo ""
    if [[ -z "$ADMIN_PASS" ]]; then
        ADMIN_PASS=$(generate_random_string 12)
        log_info "已生成随机密码: ${ADMIN_PASS}"
    fi

    # 站点配置
    echo ""
    read -p "站点名称 [Xboard-Go]: " SITE_NAME
    SITE_NAME=${SITE_NAME:-Xboard-Go}

    read -p "站点 URL (如 http://your-domain:${HTTP_PORT}): " SITE_URL
    if [[ -z "$SITE_URL" ]]; then
        SITE_URL="http://localhost:${HTTP_PORT}"
    fi

    # 生成密钥
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
# Xboard-Go 配置文件
# 生成时间: $(date '+%Y-%m-%d %H:%M:%S')

server:
  host: "0.0.0.0"
  port: ${HTTP_PORT}
  mode: release

database:
  driver: ${DB_DRIVER}
  source: ${DB_SOURCE}

$(if [[ "${USE_REDIS,,}" == "y" ]]; then
cat << REDIS
redis:
  addr: ${REDIS_ADDR}
  password: "${REDIS_PASS}"
  db: ${REDIS_DB}
REDIS
fi)

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
  enabled: $([ "${USE_REDIS,,}" == "y" ] && echo "true" || echo "false")
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
        log_info "Docker 已安装 (版本: ${DOCKER_VERSION})"
        return
    fi

    if [[ "${INSTALL_DOCKER}" != "true" ]]; then
        return
    fi

    log_step "安装 Docker..."

    if [[ "$OS" == "linux" ]]; then
        # 使用官方脚本安装 Docker
        curl -fsSL https://get.docker.com | bash
        systemctl enable docker
        systemctl start docker
    elif [[ "$OS" == "darwin" ]]; then
        log_error "请手动安装 Docker Desktop for Mac"
        log_info "下载地址: https://www.docker.com/products/docker-desktop"
        exit 1
    fi

    DOCKER_AVAILABLE=true
    log_info "Docker 安装完成"
}

# ═══════════════════════════════════════════════════════════════
# Docker 部署
# ═══════════════════════════════════════════════════════════════

deploy_docker() {
    log_step "使用 Docker 部署..."

    # 停止旧容器
    docker stop xboard-go 2>/dev/null || true
    docker rm xboard-go 2>/dev/null || true

    # 拉取镜像
    log_info "拉取 Docker 镜像..."
    docker pull ${DOCKER_IMAGE}

    # 启动容器
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

    # 下载二进制文件
    local binary_name="xboard-go-${OS}-${ARCH}"
    if [[ "$OS" == "windows" ]]; then
        binary_name="${binary_name}.exe"
    fi

    local download_url="${RELEASE_URL}/${binary_name}"
    log_info "下载二进制文件: ${download_url}"

    if command -v curl &> /dev/null; then
        curl -L -o "${INSTALL_DIR}/xboard-go" "${download_url}"
    elif command -v wget &> /dev/null; then
        wget -O "${INSTALL_DIR}/xboard-go" "${download_url}"
    else
        log_error "请安装 curl 或 wget"
        exit 1
    fi

    chmod +x "${INSTALL_DIR}/xboard-go"

    # 创建 systemd 服务
    if command -v systemctl &> /dev/null; then
        log_info "创建 systemd 服务..."
        cat > /etc/systemd/system/xboard-go.service << EOF
[Unit]
Description=Xboard-Go Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/xboard-go -config ${DATA_DIR}/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable xboard-go
        systemctl start xboard-go
        log_info "systemd 服务已启动"
    else
        # 直接运行
        log_info "启动 Xboard-Go..."
        nohup ${INSTALL_DIR}/xboard-go -config ${DATA_DIR}/config.yaml > ${DATA_DIR}/xboard-go.log 2>&1 &
        echo $! > ${DATA_DIR}/xboard-go.pid
        log_info "Xboard-Go 已启动 (PID: $(cat ${DATA_DIR}/xboard-go.pid))"
    fi
}

# ═══════════════════════════════════════════════════════════════
# 初始化管理员
# ═══════════════════════════════════════════════════════════════

init_admin() {
    log_step "初始化管理员账户..."

    # 等待服务启动
    log_info "等待服务启动..."
    sleep 5

    # 检查服务是否启动
    local max_retries=30
    local retry=0
    while [[ $retry -lt $max_retries ]]; do
        if curl -s "http://localhost:${HTTP_PORT}/healthz" > /dev/null 2>&1; then
            break
        fi
        retry=$((retry + 1))
        sleep 1
    done

    if [[ $retry -eq $max_retries ]]; then
        log_error "服务启动超时"
        return 1
    fi

    log_info "服务已启动"

    # 注册管理员账户
    log_info "创建管理员账户..."
    local response=$(curl -s -X POST "http://localhost:${HTTP_PORT}/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"${ADMIN_EMAIL}\",
            \"password\": \"${ADMIN_PASS}\"
        }")

    if echo "$response" | grep -q '"code":0'; then
        log_info "管理员账户创建成功"
    else
        log_warn "管理员账户可能已存在或创建失败"
    fi

    # 设置管理员角色
    log_info "设置管理员角色..."
    # 这里需要通过数据库直接更新角色
    # 由于不同数据库语法不同，我们提示用户手动设置
    log_warn "请手动将管理员角色设置为 admin:"
    echo ""
    if [[ "$DB_DRIVER" == "sqlite" ]]; then
        echo "  sqlite3 ${DATA_DIR}/xboard.db \"UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';\""
    elif [[ "$DB_DRIVER" == "mysql" ]]; then
        echo "  mysql -u ${DB_USER} -p ${DB_NAME} -e \"UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';\""
    elif [[ "$DB_DRIVER" == "postgres" ]]; then
        echo "  psql -U ${DB_USER} -d ${DB_NAME} -c \"UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';\""
    fi
    echo ""
}

# ═══════════════════════════════════════════════════════════════
# 显示部署结果
# ═══════════════════════════════════════════════════════════════

show_result() {
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  ✅ 部署完成！${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "  ${CYAN}访问地址:${NC}"
    echo -e "    用户面板: ${SITE_URL}/user/login"
    echo -e "    管理后台: ${SITE_URL}/admin/login"
    echo ""
    echo -e "  ${CYAN}管理员账户:${NC}"
    echo -e "    邮箱: ${ADMIN_EMAIL}"
    echo -e "    密码: ${ADMIN_PASS}"
    echo ""
    echo -e "  ${CYAN}配置文件:${NC}"
    echo -e "    ${DATA_DIR}/config.yaml"
    echo ""
    echo -e "  ${CYAN}节点通讯密钥:${NC}"
    echo -e "    ${NODE_API_KEY}"
    echo ""

    if [[ "$DEPLOY_METHOD" == "1" ]]; then
        echo -e "  ${CYAN}Docker 命令:${NC}"
        echo -e "    查看日志: docker logs -f xboard-go"
        echo -e "    重启服务: docker restart xboard-go"
        echo -e "    停止服务: docker stop xboard-go"
    else
        echo -e "  ${CYAN}服务管理:${NC}"
        if command -v systemctl &> /dev/null; then
            echo -e "    查看状态: systemctl status xboard-go"
            echo -e "    查看日志: journalctl -u xboard-go -f"
            echo -e "    重启服务: systemctl restart xboard-go"
            echo -e "    停止服务: systemctl stop xboard-go"
        else
            echo -e "    查看日志: tail -f ${DATA_DIR}/xboard-go.log"
            echo -e "    停止服务: kill \$(cat ${DATA_DIR}/xboard-go.pid)"
        fi
    fi
    echo ""
    echo -e "  ${CYAN}节点部署:${NC}"
    echo -e "    参考: https://github.com/${REPO_OWNER}/Xboard-Node-Go"
    echo ""
}

# ═══════════════════════════════════════════════════════════════
# 主流程
# ═══════════════════════════════════════════════════════════════

main() {
    print_banner

    # 检测系统
    detect_os
    detect_arch
    log_info "检测到系统: ${OS} (${ARCH})"

    # 检查 root 权限
    check_root

    # 检查 Docker
    check_docker

    # 收集配置
    collect_config

    # 安装 Docker (如果需要)
    if [[ "$DEPLOY_METHOD" == "1" ]]; then
        install_docker
    fi

    # 生成配置文件
    generate_config

    # 部署
    if [[ "$DEPLOY_METHOD" == "1" ]]; then
        deploy_docker
    else
        deploy_binary
    fi

    # 初始化管理员
    init_admin

    # 显示结果
    show_result
}

# 运行主流程
main "$@"
