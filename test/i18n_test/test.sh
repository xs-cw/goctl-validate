#!/bin/bash
set -e

# 获取项目根目录的绝对路径
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

# 确保在项目根目录
if [ ! -f "go.mod" ]; then
    echo "错误：未找到 go.mod 文件，请确保在项目根目录"
    exit 1
fi

# 清理旧的测试文件
rm -f test/i18n_test/validation.go

# 重新构建插件
echo "构建插件..."
go build -o goctl-validate

# 使用翻译器模式处理types.go文件
echo "使用翻译器模式处理types.go文件..."
./goctl-validate --custom --i18n --debug test/i18n_test/types.go

# 检查生成的文件
echo "生成的validation.go文件内容:"
cat test/i18n_test/validation.go

# 检查是否包含翻译器相关代码
echo "检查是否包含翻译器相关代码:"
grep -A 10 "translator" test/i18n_test/validation.go

# 创建一个测试程序
cat > test/i18n_test/run_test.go << 'EOF'
package main

import (
	"fmt"
	"test/i18n_test/types"
)

func main() {
	// 创建一个无效的用户表单
	user := &types.UserForm{
		Name:     "",              // 必填字段缺失
		Mobile:   "123",           // 手机号格式错误
		IdCard:   "12345",         // 身份证号格式错误
		Email:    "invalid-email", // 邮箱格式错误
		Age:      10,              // 年龄小于最小值
		Weight:   20,              // 体重小于最小值
		Website:  "invalid-url",   // URL格式错误
		IP:       "invalid-ip",    // IP格式错误
		Password: "123",           // 密码太短
		Username: "123",           // 用户名包含数字
		Nickname: "test@123",      // 昵称包含特殊字符
	}

	// 验证结构体
	err := user.Validate()
	if err != nil {
		// 断言错误类型为ValidateError
		if validateErr, ok := err.(*types.ValidateError); ok {
			fmt.Printf("字段: %s\n", validateErr.Field)
			fmt.Printf("标签: %s\n", validateErr.Tag)
			fmt.Printf("错误信息: %s\n", validateErr.Message)
		} else {
			fmt.Printf("其他错误: %v\n", err)
		}
	}
}
EOF

# 初始化Go模块
echo "初始化测试模块..."
cd test/i18n_test
go mod init test/i18n_test
go mod tidy

# 运行测试
echo "运行测试..."
go run run_test.go 