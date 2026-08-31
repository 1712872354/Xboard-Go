#!/bin/bash

# ═══════════════════════════════════════════════════════════════
# Xboard-Go 一键部署脚本
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

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

generate_random_string() {
    openssl rand -base64 32 2>/dev/null | tr -dc 'a-zA-Z0-9' | head -c 32 || \
    cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 32
}

main() {
    echo -e "${CYAN}"
    echo "  ╔═══════════════════════════════════════════════════╗"
    echo "  ║       Xboard-Go 一键部署脚本                      ║"
    echo "  ╚═══════════════════════════════════════════════════╝"
    echo -e "${NC}"

    # 检测系统
    OS="linux"
    ARCH=$(uname -m)
    case $ARCH in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) log_error "不支持的架构: $ARCH"; exit 1 ;;
    esac
    log_info "系统: ${OS} (${ARCH})"

    # 检查 root
    if [[ $EUID -ne 0 ]]; then
        log_error "请使用 root 权限运行: sudo bash install.sh"
        exit 1
    fi

    # 部署方式
    echo ""
    echo -e "${YELLOW}请选择部署方式:${NC}"
    echo "  1) Docker 部署 (推荐)"
    echo "  2) 二进制部署"
    echo ""
    read -p "请输入选项 [1]: " DEPLOY_METHOD < /dev/tty
    DEPLOY_METHOD=${DEPLOY_METHOD:-1}

    # 端口
    echo ""
    read -p "HTTP 端口 [${DEFAULT_PORT}]: " HTTP_PORT < /dev/tty
    HTTP_PORT=${HTTP_PORT:-$DEFAULT_PORT}

    read -p "gRPC 端口 [${DEFAULT_GRPC_PORT}]: " GRPC_PORT < /dev/tty
    GRPC_PORT=${GRPC_PORT:-$DEFAULT_GRPC_PORT}

    # 数据库
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
            DB_CONFIG="  dbname: ${DATA_DIR}/xboard.db"
            ;;
        2)
            DB_DRIVER="mysql"
            read -p "MySQL 主机 [localhost]: " DB_HOST < /dev/tty; DB_HOST=${DB_HOST:-localhost}
            read -p "MySQL 端口 [3306]: " DB_PORT < /dev/tty; DB_PORT=${DB_PORT:-3306}
            read -p "MySQL 数据库 [xboard]: " DB_NAME < /dev/tty; DB_NAME=${DB_NAME:-xboard}
            read -p "MySQL 用户 [root]: " DB_USER < /dev/tty; DB_USER=${DB_USER:-root}
            read -sp "MySQL 密码: " DB_PASS < /dev/tty; echo ""
            DB_CONFIG="  host: ${DB_HOST}\n  port: ${DB_PORT}\n  user: ${DB_USER}\n  password: \"${DB_PASS}\"\n  dbname: ${DB_NAME}"
            ;;
        3)
            DB_DRIVER="postgres"
            read -p "PostgreSQL 主机 [localhost]: " DB_HOST < /dev/tty; DB_HOST=${DB_HOST:-localhost}
            read -p "PostgreSQL 端口 [5432]: " DB_PORT < /dev/tty; DB_PORT=${DB_PORT:-5432}
            read -p "PostgreSQL 数据库 [xboard]: " DB_NAME < /dev/tty; DB_NAME=${DB_NAME:-xboard}
            read -p "PostgreSQL 用户 [postgres]: " DB_USER < /dev/tty; DB_USER=${DB_USER:-postgres}
            read -sp "PostgreSQL 密码: " DB_PASS < /dev/tty; echo ""
            DB_CONFIG="  host: ${DB_HOST}\n  port: ${DB_PORT}\n  user: ${DB_USER}\n  password: \"${DB_PASS}\"\n  dbname: ${DB_NAME}\n  sslmode: disable"
            ;;
    esac

    # Redis
    echo ""
    read -p "是否配置 Redis? [y/N]: " USE_REDIS < /dev/tty
    USE_REDIS=${USE_REDIS:-N}

    if [[ "${USE_REDIS,,}" == "y" ]]; then
        read -p "Redis 地址 [localhost:6379]: " REDIS_ADDR < /dev/tty; REDIS_ADDR=${REDIS_ADDR:-localhost:6379}
        read -sp "Redis 密码 (无密码回车): " REDIS_PASS < /dev/tty; echo ""
        read -p "Redis 数据库 [0]: " REDIS_DB < /dev/tty; REDIS_DB=${REDIS_DB:-0}
    fi

    # 管理员
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
    read -p "站点 URL [http://localhost:${HTTP_PORT}]: " SITE_URL < /dev/tty
    SITE_URL=${SITE_URL:-"http://localhost:${HTTP_PORT}"}

    APP_KEY=$(generate_random_string)
    NODE_API_KEY=$(generate_random_string)

    # 生成配置
    log_step "生成配置文件..."
    mkdir -p "${DATA_DIR}"

    cat > "${DATA_DIR}/config.yaml" << EOF
server:
  host: "0.0.0.0"
  port: ${HTTP_PORT}
  mode: release

database:
  driver: ${DB_DRIVER}
$(echo -e "${DB_CONFIG}")
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

logger:
  level: info
  format: json
EOF

    log_info "配置已生成: ${DATA_DIR}/config.yaml"

    # 部署
    if [[ "$DEPLOY_METHOD" == "1" ]]; then
        # Docker 部署
        log_step "Docker 部署..."

        if ! command -v docker &> /dev/null; then
            log_info "安装 Docker..."
            curl -fsSL https://get.docker.com | bash
            systemctl enable docker && systemctl start docker
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
    else
        # 二进制部署
        log_step "二进制部署..."
        mkdir -p "${INSTALL_DIR}"

        BINARY_URL="${RELEASE_URL}/xboard-go-${OS}-${ARCH}"
        log_info "下载: ${BINARY_URL}"
        curl -L -o "${INSTALL_DIR}/xboard-go" "${BINARY_URL}" || \
        wget -O "${INSTALL_DIR}/xboard-go" "${BINARY_URL}"
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
    fi

    # 初始化管理员
    log_step "初始化管理员..."
    log_info "等待服务启动..."
    sleep 5

    for i in $(seq 1 30); do
        if curl -s "http://localhost:${HTTP_PORT}/healthz" > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done

    curl -s -X POST "http://localhost:${HTTP_PORT}/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASS}\"}" > /dev/null 2>&1 || true

    # 显示结果
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  ✅ 部署完成！${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
    echo ""
    echo "  访问地址:"
    echo "    用户面板: ${SITE_URL}/user/login"
    echo "    管理后台: ${SITE_URL}/admin/login"
    echo ""
    echo "  管理员账户:"
    echo "    邮箱: ${ADMIN_EMAIL}"
    echo "    密码: ${ADMIN_PASS}"
    echo ""
    echo "  配置文件: ${DATA_DIR}/config.yaml"
    echo "  节点密钥: ${NODE_API_KEY}"
    echo ""
    echo "  设置管理员角色:"
    if [[ "$DB_DRIVER" == "sqlite" ]]; then
        echo "    sqlite3 ${DATA_DIR}/xboard.db \"UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';\""
    elif [[ "$DB_DRIVER" == "mysql" ]]; then
        echo "    mysql -u ${DB_USER} -p ${DB_NAME} -e \"UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';\""
    elif [[ "$DB_DRIVER" == "postgres" ]]; then
        echo "    psql -U ${DB_USER} -d ${DB_NAME} -c \"UPDATE users SET role='admin' WHERE email='${ADMIN_EMAIL}';\""
    fi
    echo ""
}

main
