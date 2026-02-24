.PHONY: run build tidy download clean

# 运行服务
run:
	go run main.go

# 编译二进制
build:
	go build -o bin/server main.go

# 整理依赖
tidy:
	go mod tidy

# 下载依赖
download:
	go mod download

# 清理编译产物
clean:
	rm -rf bin/ logs/

# 生成 swagger 文档（需要安装 swag: go install github.com/swaggo/swag/cmd/swag@latest）
swagger:
	swag init -g main.go -o docs/

# 格式化代码
fmt:
	go fmt ./...
	goimports -l -w .

# 代码检查
lint:
	golangci-lint run ./...
