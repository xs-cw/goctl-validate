#!/bin/bash
set -e

# 切换到项目根目录
cd ../..

# 重新构建插件
echo "构建插件..."
go build -o goctl-validate

# 切换回测试目录
cd test/simple

# 运行插件处理types.go文件
echo "运行插件..."
../../goctl-validate

# 检查生成的文件
echo "检查生成的文件:"
echo "---------- validation.go ----------"
cat validation.go
echo "---------- types.go ----------"
cat types.go

echo "测试完成！" 