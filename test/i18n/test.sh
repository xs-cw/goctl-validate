#!/bin/bash
set -e

# 切换到项目根目录
cd ../..

# 清理旧的测试文件
rm -f test/i18n/validation.go

# 重新构建插件
echo "构建插件..."
go build -o goctl-validate

# 使用普通模式处理types.go文件
echo "使用普通模式处理types.go文件..."
./goctl-validate --custom --debug test/i18n/types.go

# 检查生成的文件
echo "普通模式生成的文件内容:"
cat test/i18n/validation.go

# 检查Validate方法
echo "普通模式下的Validate方法:"
grep -A 3 "func (r \*UserForm) Validate" test/i18n/types.go

# 清理旧的测试文件
rm -f test/i18n/validation.go test/i18n/types.go.bak
cp test/i18n/types.go test/i18n/types.go.bak
rm -f test/i18n/validation.go

# 使用翻译器模式处理types.go文件
echo "使用翻译器模式处理types.go文件..."
./goctl-validate --custom --i18n --debug test/i18n/types.go

# 检查生成的文件
echo "翻译器模式生成的文件内容:"
cat test/i18n/validation.go

# 检查Validate方法
echo "翻译器模式下的Validate方法:"
grep -A 10 "func (r \*UserForm) Validate" test/i18n/types.go

# 显示翻译器导入部分
echo "检查翻译器导入部分:"
grep -A 10 "import" test/i18n/validation.go

# 检查自定义验证器的翻译注册
echo "检查自定义验证器的翻译注册:"
grep -A 10 "注册.*的中文翻译" test/i18n/validation.go 