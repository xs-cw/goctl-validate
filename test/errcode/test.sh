#!/bin/bash
set -e

# 切换到项目根目录
cd ../..

# 清理旧的测试文件
rm -f test/errcode/validation.go

# 重新构建插件
echo "构建插件..."
go build -o goctl-validate

# 使用错误类型模式处理types.go文件
echo "使用错误类型模式处理types.go文件..."
./goctl-validate --custom --with-errcode --debug test/errcode/types.go

# 检查生成的文件
echo "生成的validation.go文件内容:"
cat test/errcode/validation.go

# 检查是否包含错误类型定义
echo "检查是否包含错误类型定义:"
grep -A 20 "type ValidationError struct" test/errcode/validation.go

# 检查Validate方法
echo "检查生成的Validate方法:"
grep -A 15 "func (r \*UserInfo) Validate" test/errcode/types.go

# 创建一个简单的测试程序
cat > test/errcode/run_test.go << 'EOF'
package main

import (
	"fmt"
	"test/errcode/types"
)

func main() {
	// 创建一个有错误的用户信息
	user := &types.UserInfo{
		Name:   "",             // 必填字段缺失
		Mobile: "123",          // 手机号格式错误
		IdCard: "12345",        // 身份证号格式错误
		Email:  "invalid-email", // 邮箱格式错误
		Age:    0,              // 必填字段缺失
	}

	// 验证结构体
	err := user.Validate()
	
	// 使用错误类型断言
	if err != nil {
		// 断言错误类型为ValidationErrors
		if validationErrs, ok := err.(*types.ValidationErrors); ok {
			// 获取第一个错误的错误码
			errCode := validationErrs.FirstCode()
			fmt.Printf("错误码: %d\n", errCode)
			
			// 获取错误信息
			fmt.Printf("错误信息: %s\n", validationErrs.Error())
			
			// 获取第一个错误
			firstErr := validationErrs.First()
			if firstErr != nil {
				fmt.Printf("字段: %s, 标签: %s, 错误码: %d\n", 
					firstErr.Field, firstErr.Tag, firstErr.Code)
			}
		} else {
			// 处理其他类型的错误
			fmt.Printf("其他错误: %v\n", err)
		}
	}
}
EOF

# 初始化Go模块
echo "初始化测试模块..."
cd test/errcode
go mod init test/errcode
go mod tidy

# 运行测试
echo "运行测试..."
go run run_test.go 