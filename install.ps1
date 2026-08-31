# ═══════════════════════════════════════════════════════════════
# Xboard-Go Windows 一键部署脚本
# ═══════════════════════════════════════════════════════════════

param(
    [string]$Port = "8080",
    [string]$GrpcPort = "50051",
    [string]$InstallDir = "C:\xboard-go"
)

$ErrorActionPreference = "Stop"

# 颜色函数
function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Green }
function Write-Warn { Write-Host "[WARN] $args" -ForegroundColor Yellow }
function Write-Error { Write-Host "[ERROR] $args" -ForegroundColor Red }
function Write-Step { Write-Host "[STEP] $args" -ForegroundColor Cyan }

# 显示 Banner
function Show-Banner {
    Write-Host ""
    Write-Host "  ╔═══════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "  ║                                                   ║" -ForegroundColor Cyan
    Write-Host "  ║       Xboard-Go Windows 部署脚本                  ║" -ForegroundColor Cyan
    Write-Host "  ║       高性能代理面板管理系统                       ║" -ForegroundColor Cyan
    Write-Host "  ║                                                   ║" -ForegroundColor Cyan
    Write-Host "  ╚═══════════════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
}

# 生成随机字符串
function New-RandomString {
    param([int]$Length = 32)
    $chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    $result = ""
    for ($i = 0; $i -lt $Length; $i++) {
        $result += $chars[(Get-Random -Maximum $chars.Length)]
    }
    return $result
}

# 收集配置
function Get-Configuration {
    Write-Host ""
    Write-Host "═══════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host "  配置信息收集" -ForegroundColor Cyan
    Write-Host "═══════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host ""

    # 部署方式
    Write-Host "请选择部署方式:" -ForegroundColor Yellow
    Write-Host "  1) Docker 部署 (推荐)"
    Write-Host "  2) 二进制部署 (直接运行)"
    Write-Host ""
    $script:DeployMethod = Read-Host "请输入选项 [1]"
    if ([string]::IsNullOrEmpty($script:DeployMethod)) { $script:DeployMethod = "1" }

    # 端口配置
    $script:HttpPort = Read-Host "请输入 HTTP 端口 [$Port]"
    if ([string]::IsNullOrEmpty($script:HttpPort)) { $script:HttpPort = $Port }

    $script:GrpcPortValue = Read-Host "请输入 gRPC 端口 [$GrpcPort]"
    if ([string]::IsNullOrEmpty($script:GrpcPortValue)) { $script:GrpcPortValue = $GrpcPort }

    # 数据库配置
    Write-Host ""
    Write-Host "请选择数据库:" -ForegroundColor Yellow
    Write-Host "  1) SQLite (推荐，无需额外配置)"
    Write-Host "  2) MySQL"
    Write-Host "  3) PostgreSQL"
    Write-Host ""
    $script:DbType = Read-Host "请输入选项 [1]"
    if ([string]::IsNullOrEmpty($script:DbType)) { $script:DbType = "1" }

    switch ($script:DbType) {
        "1" {
            $script:DbDriver = "sqlite"
            $script:DbSource = "$InstallDir\data\xboard.db"
        }
        "2" {
            $script:DbDriver = "mysql"
            $script:DbHost = Read-Host "MySQL 主机 [localhost]"
            if ([string]::IsNullOrEmpty($script:DbHost)) { $script:DbHost = "localhost" }
            $script:DbPort = Read-Host "MySQL 端口 [3306]"
            if ([string]::IsNullOrEmpty($script:DbPort)) { $script:DbPort = "3306" }
            $script:DbName = Read-Host "MySQL 数据库名 [xboard]"
            if ([string]::IsNullOrEmpty($script:DbName)) { $script:DbName = "xboard" }
            $script:DbUser = Read-Host "MySQL 用户名 [root]"
            if ([string]::IsNullOrEmpty($script:DbUser)) { $script:DbUser = "root" }
            $script:DbPass = Read-Host "MySQL 密码" -AsSecureString
            $BSTR = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($script:DbPass)
            $script:DbPassPlain = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($BSTR)
            $script:DbSource = "$($script:DbUser):$($script:DbPassPlain)@tcp($($script:DbHost):$($script:DbPort))/$($script:DbName)?charset=utf8mb4&parseTime=True&loc=Local"
        }
        "3" {
            $script:DbDriver = "postgres"
            $script:DbHost = Read-Host "PostgreSQL 主机 [localhost]"
            if ([string]::IsNullOrEmpty($script:DbHost)) { $script:DbHost = "localhost" }
            $script:DbPort = Read-Host "PostgreSQL 端口 [5432]"
            if ([string]::IsNullOrEmpty($script:DbPort)) { $script:DbPort = "5432" }
            $script:DbName = Read-Host "PostgreSQL 数据库名 [xboard]"
            if ([string]::IsNullOrEmpty($script:DbName)) { $script:DbName = "xboard" }
            $script:DbUser = Read-Host "PostgreSQL 用户名 [postgres]"
            if ([string]::IsNullOrEmpty($script:DbUser)) { $script:DbUser = "postgres" }
            $script:DbPass = Read-Host "PostgreSQL 密码" -AsSecureString
            $BSTR = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($script:DbPass)
            $script:DbPassPlain = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($BSTR)
            $script:DbSource = "host=$($script:DbHost) user=$($script:DbUser) password=$($script:DbPassPlain) dbname=$($script:DbName) port=$($script:DbPort) sslmode=disable"
        }
    }

    # Redis 配置
    Write-Host ""
    $script:UseRedis = Read-Host "是否配置 Redis? (用于缓存和限流) [y/N]"
    if ([string]::IsNullOrEmpty($script:UseRedis)) { $script:UseRedis = "N" }

    if ($script:UseRedis -eq "y" -or $script:UseRedis -eq "Y") {
        $script:RedisAddr = Read-Host "Redis 地址 [localhost:6379]"
        if ([string]::IsNullOrEmpty($script:RedisAddr)) { $script:RedisAddr = "localhost:6379" }
        $script:RedisPass = Read-Host "Redis 密码 (无密码直接回车)"
        $script:RedisDb = Read-Host "Redis 数据库 [0]"
        if ([string]::IsNullOrEmpty($script:RedisDb)) { $script:RedisDb = "0" }
    }

    # 管理员配置
    Write-Host ""
    Write-Host "═══════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host "  管理员账户配置" -ForegroundColor Cyan
    Write-Host "═══════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host ""

    $script:AdminEmail = Read-Host "管理员邮箱 [admin@example.com]"
    if ([string]::IsNullOrEmpty($script:AdminEmail)) { $script:AdminEmail = "admin@example.com" }

    $script:AdminPass = Read-Host "管理员密码"
    if ([string]::IsNullOrEmpty($script:AdminPass)) {
        $script:AdminPass = New-RandomString -Length 12
        Write-Info "已生成随机密码: $($script:AdminPass)"
    }

    # 站点配置
    Write-Host ""
    $script:SiteName = Read-Host "站点名称 [Xboard-Go]"
    if ([string]::IsNullOrEmpty($script:SiteName)) { $script:SiteName = "Xboard-Go" }

    $script:SiteUrl = Read-Host "站点 URL (如 http://your-domain:$($script:HttpPort))"
    if ([string]::IsNullOrEmpty($script:SiteUrl)) { $script:SiteUrl = "http://localhost:$($script:HttpPort)" }

    # 生成密钥
    $script:AppKey = New-RandomString -Length 32
    $script:NodeApiKey = New-RandomString -Length 32
}

# 生成配置文件
function New-ConfigFile {
    Write-Step "生成配置文件..."

    $dataDir = "$InstallDir\data"
    if (-not (Test-Path $dataDir)) {
        New-Item -ItemType Directory -Path $dataDir -Force | Out-Null
    }

    $redisConfig = ""
    if ($script:UseRedis -eq "y" -or $script:UseRedis -eq "Y") {
        $redisConfig = @"
redis:
  addr: $($script:RedisAddr)
  password: "$($script:RedisPass)"
  db: $($script:RedisDb)
"@
    }

    $configContent = @"
# Xboard-Go 配置文件
# 生成时间: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

server:
  host: "0.0.0.0"
  port: $($script:HttpPort)
  mode: release

database:
  driver: $($script:DbDriver)
  source: $($script:DbSource)

$redisConfig

grpc:
  enabled: true
  port: $($script:GrpcPortValue)

app:
  name: $($script:SiteName)
  key: $($script:AppKey)
  node_api_key: $($script:NodeApiKey)
  default_user_role: user
  subscribe_token_length: 32

rate_limit:
  enabled: $(if ($script:UseRedis -eq "y" -or $script:UseRedis -eq "Y") { "true" } else { "false" })
  ip_limit: 100
  user_limit: 200

logger:
  level: info
  format: json
"@

    $configContent | Out-File -FilePath "$dataDir\config.yaml" -Encoding UTF8
    Write-Info "配置文件已生成: $dataDir\config.yaml"
}

# Docker 部署
function Install-Docker {
    Write-Step "使用 Docker 部署..."

    # 检查 Docker
    try {
        $dockerVersion = docker --version
        Write-Info "Docker 已安装: $dockerVersion"
    } catch {
        Write-Error "Docker 未安装，请先安装 Docker Desktop"
        Write-Info "下载地址: https://www.docker.com/products/docker-desktop"
        exit 1
    }

    # 停止旧容器
    docker stop xboard-go 2>$null
    docker rm xboard-go 2>$null

    # 拉取镜像
    Write-Info "拉取 Docker 镜像..."
    docker pull ghcr.io/1712872354/xboard-go:latest

    # 启动容器
    Write-Info "启动容器..."
    $dataDir = "$InstallDir\data"
    docker run -d `
        --name xboard-go `
        --restart always `
        -p "$($script:HttpPort):8080" `
        -p "$($script:GrpcPortValue):50051" `
        -v "${dataDir}:/data" `
        ghcr.io/1712872354/xboard-go:latest

    Write-Info "Docker 容器已启动"
}

# 二进制部署
function Install-Binary {
    Write-Step "使用二进制部署..."

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    # 下载二进制文件
    $binaryName = "xboard-go-windows-amd64.exe"
    $downloadUrl = "https://github.com/1712872354/Xboard-Go/releases/latest/download/$binaryName"
    Write-Info "下载二进制文件: $downloadUrl"

    Invoke-WebRequest -Uri $downloadUrl -OutFile "$InstallDir\xboard-go.exe"

    # 创建启动脚本
    $startScript = @"
@echo off
cd /d "$InstallDir"
xboard-go.exe -config data\config.yaml
pause
"@
    $startScript | Out-File -FilePath "$InstallDir\start.bat" -Encoding ASCII

    # 创建 Windows 服务
    Write-Info "创建 Windows 服务..."

    # 使用 NSSM 创建服务 (如果可用)
    try {
        $nssmPath = Get-Command nssm -ErrorAction Stop
        nssm install xboard-go "$InstallDir\xboard-go.exe" "-config" "$InstallDir\data\config.yaml"
        nssm start xboard-go
        Write-Info "Windows 服务已创建并启动"
    } catch {
        Write-Warn "NSSM 未安装，将使用计划任务启动"

        # 创建计划任务
        $action = New-ScheduledTaskAction -Execute "$InstallDir\xboard-go.exe" -Argument "-config $InstallDir\data\config.yaml"
        $trigger = New-ScheduledTaskTrigger -AtStartup
        $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable
        Register-ScheduledTask -TaskName "Xboard-Go" -Action $action -Trigger $trigger -Settings $settings -RunLevel Highest -Force

        # 立即启动
        Start-Process -FilePath "$InstallDir\xboard-go.exe" -ArgumentList "-config $InstallDir\data\config.yaml" -WindowStyle Hidden
        Write-Info "Xboard-Go 已启动"
    }
}

# 初始化管理员
function Initialize-Admin {
    Write-Step "初始化管理员账户..."

    # 等待服务启动
    Write-Info "等待服务启动..."
    Start-Sleep -Seconds 5

    # 检查服务是否启动
    $maxRetries = 30
    $retry = 0
    while ($retry -lt $maxRetries) {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:$($script:HttpPort)/healthz" -UseBasicParsing -TimeoutSec 2
            if ($response.StatusCode -eq 200) {
                break
            }
        } catch {
            # 忽略错误
        }
        $retry++
        Start-Sleep -Seconds 1
    }

    if ($retry -eq $maxRetries) {
        Write-Error "服务启动超时"
        return
    }

    Write-Info "服务已启动"

    # 注册管理员账户
    Write-Info "创建管理员账户..."
    try {
        $body = @{
            email = $script:AdminEmail
            password = $script:AdminPass
        } | ConvertTo-Json

        $response = Invoke-RestMethod -Uri "http://localhost:$($script:HttpPort)/api/v1/auth/register" -Method Post -Body $body -ContentType "application/json"
        Write-Info "管理员账户创建成功"
    } catch {
        Write-Warn "管理员账户可能已存在或创建失败"
    }

    # 提示设置管理员角色
    Write-Warn "请手动将管理员角色设置为 admin:"
    Write-Host ""
    if ($script:DbDriver -eq "sqlite") {
        Write-Host "  sqlite3 $($InstallDir)\data\xboard.db `"UPDATE users SET role='admin' WHERE email='$($script:AdminEmail)';`""
    } elseif ($script:DbDriver -eq "mysql") {
        Write-Host "  mysql -u $($script:DbUser) -p $($script:DbName) -e `"UPDATE users SET role='admin' WHERE email='$($script:AdminEmail)';`""
    } elseif ($script:DbDriver -eq "postgres") {
        Write-Host "  psql -U $($script:DbUser) -d $($script:DbName) -c `"UPDATE users SET role='admin' WHERE email='$($script:AdminEmail)';`""
    }
    Write-Host ""
}

# 显示部署结果
function Show-Result {
    Write-Host ""
    Write-Host "═══════════════════════════════════════════════════" -ForegroundColor Green
    Write-Host "  ✅ 部署完成！" -ForegroundColor Green
    Write-Host "═══════════════════════════════════════════════════" -ForegroundColor Green
    Write-Host ""
    Write-Host "  访问地址:" -ForegroundColor Cyan
    Write-Host "    用户面板: $($script:SiteUrl)/user/login"
    Write-Host "    管理后台: $($script:SiteUrl)/admin/login"
    Write-Host ""
    Write-Host "  管理员账户:" -ForegroundColor Cyan
    Write-Host "    邮箱: $($script:AdminEmail)"
    Write-Host "    密码: $($script:AdminPass)"
    Write-Host ""
    Write-Host "  配置文件:" -ForegroundColor Cyan
    Write-Host "    $InstallDir\data\config.yaml"
    Write-Host ""
    Write-Host "  节点通讯密钥:" -ForegroundColor Cyan
    Write-Host "    $($script:NodeApiKey)"
    Write-Host ""

    if ($script:DeployMethod -eq "1") {
        Write-Host "  Docker 命令:" -ForegroundColor Cyan
        Write-Host "    查看日志: docker logs -f xboard-go"
        Write-Host "    重启服务: docker restart xboard-go"
        Write-Host "    停止服务: docker stop xboard-go"
    } else {
        Write-Host "  服务管理:" -ForegroundColor Cyan
        Write-Host "    启动服务: $InstallDir\start.bat"
        Write-Host "    停止服务: taskkill /IM xboard-go.exe /F"
    }
    Write-Host ""
    Write-Host "  节点部署:" -ForegroundColor Cyan
    Write-Host "    参考: https://github.com/1712872354/Xboard-Node-Go"
    Write-Host ""
}

# 主流程
function Main {
    Show-Banner

    # 收集配置
    Get-Configuration

    # 生成配置文件
    New-ConfigFile

    # 部署
    if ($script:DeployMethod -eq "1") {
        Install-Docker
    } else {
        Install-Binary
    }

    # 初始化管理员
    Initialize-Admin

    # 显示结果
    Show-Result
}

# 运行主流程
Main
