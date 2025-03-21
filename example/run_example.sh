#!/bin/bash

# 显示执行的命令
set -x

# 生成API代码
goctl api go -api example.api -dir .

# 使用goctl-validate插件添加验证逻辑
goctl api plugin -p goctl-validate="validate" --api example.api --dir .

# 如果想使用自定义验证功能，取消下面注释
# goctl api plugin -p goctl-validate="validate --custom" --api example.api --dir .

# 如果遇到问题，可以使用调试模式
# goctl api plugin -p goctl-validate="validate --debug" --api example.api --dir .

echo "示例运行完毕，请检查生成的types.go文件中是否添加了验证逻辑" 