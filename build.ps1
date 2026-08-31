# XBoard-Go 一键构建脚本 (Windows PowerShell)
# 用法:
#   .\build.ps1            # 构建前端 + 后端，输出到 bin\xboard-go.exe
#   .\build.ps1 -Frontend  # 仅构建前端
#   .\build.ps1 -Backend   # 仅构建后端
#   .\build.ps1 -Sqlite    # 使用 SQLite 配置构建

param(
    [switch]$Frontend,
    [switch]$Backend,
    [switch]$Sqlite,
    [string]$Output = "bin/xboard-go.exe",
    [string]$LdFlags = ""
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  XBoard-Go Build Script" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

# 如果没有指定 -Frontend 或 -Backend，则构建全部
$BuildAll = -not $Frontend -and -not $Backend

# --- Step 1: 构建前端 ---
if ($BuildAll -or $Frontend) {
    Write-Host "`n[1/3] Building frontend..." -ForegroundColor Yellow

    $frontendDir = Join-Path $ProjectRoot "frontend"
    if (-not (Test-Path $frontendDir)) {
        Write-Error "Frontend directory not found: $frontendDir"
        exit 1
    }

    # 检查 pnpm 是否可用
    $pnpmCmd = Get-Command pnpm -ErrorAction SilentlyContinue
    if (-not $pnpmCmd) {
        Write-Error "pnpm not found. Install with: npm install -g pnpm"
        exit 1
    }

    Push-Location $frontendDir
    try {
        Write-Host "  Installing dependencies..."
        pnpm install --frozen-lockfile
        if ($LASTEXITCODE -ne 0) { throw "pnpm install failed" }

        Write-Host "  Building frontend (vite build)..."
        pnpm run build
        if ($LASTEXITCODE -ne 0) { throw "vite build failed" }
    }
    finally {
        Pop-Location
    }

    Write-Host "  Frontend build complete." -ForegroundColor Green
}

# --- Step 2: 复制前端 dist 到 embed 目录 ---
if ($BuildAll -or $Frontend) {
    Write-Host "`n[2/3] Copying frontend dist to embed directory..." -ForegroundColor Yellow

    $srcDist = Join-Path $ProjectRoot "frontend\dist"
    $dstDist = Join-Path $ProjectRoot "internal\static\dist"

    if (-not (Test-Path $srcDist)) {
        Write-Error "Frontend dist not found: $srcDist. Run frontend build first."
        exit 1
    }

    # 清空旧的 dist 目录（保留 .gitkeep）
    if (Test-Path $dstDist) {
        Get-ChildItem $dstDist -Exclude ".gitkeep" | Remove-Item -Recurse -Force
    } else {
        New-Item -ItemType Directory -Path $dstDist -Force | Out-Null
    }

    # 复制所有文件
    Copy-Item -Path "$srcDist\*" -Destination $dstDist -Recurse -Force
    Write-Host "  Copied $(Get-ChildItem $dstDist -Recurse -File | Measure-Object | Select-Object -ExpandProperty Count) files." -ForegroundColor Green
}

# --- Step 3: 构建后端 ---
if ($BuildAll -or $Backend) {
    Write-Host "`n[3/3] Building Go backend..." -ForegroundColor Yellow

    Set-Location $ProjectRoot

    Write-Host "  Running go mod tidy..."
    go mod tidy
    if ($LASTEXITCODE -ne 0) { Write-Warning "go mod tidy returned errors, continuing..." }

    # 确保 bin 目录存在
    $binDir = Split-Path $Output -Parent
    if ($binDir -and -not (Test-Path $binDir)) {
        New-Item -ItemType Directory -Path $binDir -Force | Out-Null
    }

    # 构建参数
    $buildArgs = @("build", "-o", $Output)
    if ($LdFlags) {
        $buildArgs += @("-ldflags", $LdFlags)
    }
    $buildArgs += "./cmd/server/"

    Write-Host "  Compiling: go $($buildArgs -join ' ')"
    go @buildArgs
    if ($LASTEXITCODE -ne 0) { throw "go build failed" }

    Write-Host "  Backend build complete: $Output" -ForegroundColor Green
}

# --- 选择配置文件 ---
$configFile = "config.yaml"
if ($Sqlite) {
    $configFile = "config.sqlite.yaml"
}

Write-Host "`n=========================================" -ForegroundColor Cyan
Write-Host "  Build Complete!" -ForegroundColor Green
Write-Host "=========================================`n" -ForegroundColor Cyan
Write-Host "Binary: $Output"
Write-Host "Config: $configFile"
Write-Host ""
Write-Host "Run:    .\$Output -config $configFile"
Write-Host ""
