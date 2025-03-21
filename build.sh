#!/bin/bash

# 确保目录存在
mkdir -p bin

# 编译主程序
echo "编译 goctl-validate..."
go build -o bin/goctl-validate ./cmd/goctl-validate

# 检查编译结果
if [ $? -eq 0 ]; then
    echo "编译成功，输出文件: $(pwd)/bin/goctl-validate"
    
    # 显示帮助信息
    echo ""
    echo "使用方法示例:"
    echo "bin/goctl-validate validate --help"
    echo "goctl api plugin -p \"$(pwd)/bin/goctl-validate=\\\"validate --custom\\\"\" --api your_api.api --dir ."
else
    echo "编译失败"
    exit 1
fi 