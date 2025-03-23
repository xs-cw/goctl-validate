#!/bin/bash

# 切换到项目根目录
cd /Users/coke/Documents/HecokeWork/github/goctl-validate/

# 清理旧的生成文件
rm -f test/simpletype/validation.go

# 重新构建插件
go build -o goctl-validate

# 运行插件处理types.go文件，启用简化错误类型和中文翻译
./goctl-validate --simplified --i18n --debug test/simpletype/types.go

echo "============= 运行独立测试示例 ============="
cd test/simpletype/
go run test_simple.go

# 返回到项目根目录
cd ../../ 