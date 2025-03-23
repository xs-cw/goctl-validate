#!/bin/bash
set -e

# 切换到项目根目录
cd ../..

# 清理旧的测试文件
rm -f test/i18nv2/validation.go

# 重新构建插件
echo "构建插件..."
go build -o goctl-validate

# 使用翻译器模式处理types.go文件
echo "使用翻译器模式处理types.go文件..."
./goctl-validate --custom --i18n --debug test/i18nv2/types.go

# 检查生成的文件
echo "生成的validation.go文件内容:"
cat test/i18nv2/validation.go

# 检查是否包含GetValidateErrorMsg函数
echo "检查是否包含GetValidateErrorMsg函数:"
grep -A 10 "func GetValidateErrorMsg" test/i18nv2/validation.go

# 检查Validate方法
echo "检查生成的Validate方法:"
grep -A 5 "func (r \*UserInfo) Validate" test/i18nv2/types.go 