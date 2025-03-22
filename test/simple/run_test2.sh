#!/bin/bash
set -e

# 切换到项目根目录
cd ../..

# 重新构建插件
echo "构建插件..."
go build -o goctl-validate

# 创建并运行测试程序
echo "创建测试程序..."
cat > test/simple/test_validate.go << 'EOF'
package main

import (
	"fmt"
	"github.com/linabellbiu/goctl-validate/internal/processor"
	"os"
	"path/filepath"
)

func main() {
	// 获取当前目录
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		os.Exit(1)
	}
	
	// 处理types.go文件
	typesFile := filepath.Join(dir, "types.go")
	options := processor.Options{
		EnableCustomValidation: false,
		DebugMode:              true,
	}
	
	fmt.Printf("处理文件: %s\n", typesFile)
	if err := processor.ProcessTypesFile(typesFile, options); err != nil {
		fmt.Printf("处理文件失败: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("处理成功！")
}
EOF

echo "运行测试程序..."
cd test/simple
go mod init test
go mod edit -replace github.com/linabellbiu/goctl-validate=../..
go mod tidy
go run test_validate.go

# 检查生成的文件
echo "检查生成的文件:"
ls -la

if [ -f validation.go ]; then
    echo "---------- validation.go ----------"
    cat validation.go
    echo "---------- types.go ----------"
    cat types.go
else
    echo "错误: validation.go 文件未生成!"
fi

echo "测试完成！" 