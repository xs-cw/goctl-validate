#!/bin/bash

# 测试目录
TEST_DIR="./test/trans_test"

# 清理上一次测试的目录
rm -rf ${TEST_DIR}
mkdir -p ${TEST_DIR}

# 复制API定义文件
cp ./test/trans_test.api ${TEST_DIR}/trans_test.api

# 切换到测试目录
cd ${TEST_DIR}

echo "================================================================"
echo "开始测试翻译功能..."
echo "================================================================"

# 使用goctl生成基础代码
echo "第1步：使用goctl生成基础代码..."
goctl api go -api trans_test.api -dir .
echo "基础代码生成完成！"
echo "----------------------------------------------------------------"

# 使用goctl-validate插件添加验证（默认不启用翻译）
echo "第2步：使用goctl-validate插件添加基本验证..."
../../goctl-validate

# 查看生成的types.go文件
echo "查看生成的types.go文件内容："
cat ./types/types.go
echo "----------------------------------------------------------------"

# 查看生成的validation.go文件
echo "查看生成的validation.go文件内容："
cat ./types/validation.go
echo "----------------------------------------------------------------"

# 清理目录，重新测试翻译功能
echo "第3步：清理目录，重新测试翻译功能..."
rm -rf ./*
goctl api go -api trans_test.api -dir .

# 使用goctl-validate插件添加验证，并启用翻译（默认中文）
echo "第4步：使用goctl-validate插件添加验证，并启用翻译（默认中文）..."
../../goctl-validate --trans

# 查看生成的validation.go文件
echo "查看生成的带翻译功能的validation.go文件内容："
cat ./types/validation.go
echo "----------------------------------------------------------------"

# 清理目录，测试其他翻译语言
echo "第5步：清理目录，测试英文翻译..."
rm -rf ./*
goctl api go -api trans_test.api -dir .

# 使用goctl-validate插件添加验证，并启用英文翻译
echo "第6步：使用goctl-validate插件添加验证，并启用英文翻译..."
../../goctl-validate --trans --lang en

# 查看生成的validation.go文件
echo "查看生成的带英文翻译的validation.go文件内容："
cat ./types/validation.go
echo "----------------------------------------------------------------"

echo "测试完成！"
echo "================================================================" 