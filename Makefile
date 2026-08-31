.PHONY: all frontend backend clean run-sqlite run-postgres

# 构建所有
all: frontend backend

# 仅构建前端
frontend:
	cd frontend && pnpm install --frozen-lockfile && pnpm run build
	rm -rf internal/static/dist/*
	cp -r frontend/dist/* internal/static/dist/

# 仅构建后端
backend:
	go mod tidy
	go build -o bin/xboard-go ./cmd/server/

# 使用 SQLite 配置运行（单文件部署）
run-sqlite: all
	./bin/xboard-go -config config.sqlite.yaml

# 使用 PostgreSQL 配置运行
run-postgres: all
	./bin/xboard-go -config config.yaml

# 清理构建产物
clean:
	rm -rf frontend/dist
	rm -f bin/xboard-go bin/xboard-go.exe
	# 恢复 embed 占位文件
	mkdir -p internal/static/dist
	touch internal/static/dist/.gitkeep

# 交叉编译 Linux
build-linux:
	cd frontend && pnpm install --frozen-lockfile && pnpm run build
	cp -r frontend/dist/* internal/static/dist/
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/xboard-go-linux-amd64 ./cmd/server/

# 交叉编译 Windows
build-windows:
	cd frontend && pnpm install --frozen-lockfile && pnpm run build
	cp -r frontend/dist/* internal/static/dist/
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/xboard-go-windows-amd64.exe ./cmd/server/
