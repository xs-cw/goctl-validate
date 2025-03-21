.PHONY: build clean test install

# 变量定义
BINDIR=bin
BINARY=goctl-validate
SRC=./cmd/goctl-validate

# 默认目标：构建
all: build

# 构建项目
build:
	@echo "正在构建 $(BINARY)..."
	@mkdir -p $(BINDIR)
	@go build -o $(BINDIR)/$(BINARY) $(SRC)
	@echo "构建完成：$(BINDIR)/$(BINARY)"

# 测试
test:
	@echo "正在运行测试..."
	@go test -v ./test/...
	@echo "测试完成"

# 安装到GOPATH
install:
	@echo "正在安装 $(BINARY)..."
	@go install $(SRC)
	@echo "安装完成"

# 清理构建产物
clean:
	@echo "正在清理..."
	@rm -rf $(BINDIR)
	@echo "清理完成"

# 运行示例
example:
	@echo "正在运行示例..."
	@cd example && ./run_example.sh
	@echo "示例运行完成"
