package main

import (
	"fmt"
	"github.com/linabellbiu/goctl-validate/internal/processor"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: test <types.go文件路径>")
		os.Exit(1)
	}

	// 获取命令行参数
	typesFile := os.Args[1]

	// 处理types.go文件
	options := processor.Options{
		EnableCustomValidation: true, // 启用自定义验证
		DebugMode:              true,
	}

	fmt.Printf("处理文件: %s\n", typesFile)
	if err := processor.ProcessTypesFile(typesFile, options); err != nil {
		fmt.Printf("处理文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("处理成功！")
}
