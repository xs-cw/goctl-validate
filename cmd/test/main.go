package main

import (
	"fmt"
	"github.com/linabellbiu/goctl-validate/internal/processor"
	"os"
	"path/filepath"
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
		EnableCustomValidation: false,
		DebugMode:              true,
	}
	
	fmt.Printf("处理文件: %s\n", typesFile)
	if err := processor.ProcessTypesFile(typesFile, options); err != nil {
		fmt.Printf("处理文件失败: %v\n", err)
		os.Exit(1)
	}
	
	// 检查validation.go文件是否生成
	validationFile := filepath.Join(filepath.Dir(typesFile), "validation.go")
	if _, err := os.Stat(validationFile); os.IsNotExist(err) {
		fmt.Printf("错误: validation.go 文件未生成!\n")
		os.Exit(1)
	}
	
	// 读取验证文件内容
	content, err := os.ReadFile(validationFile)
	if err != nil {
		fmt.Printf("读取验证文件失败: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("验证文件内容:")
	fmt.Println(string(content))
	
	fmt.Println("处理成功！")
} 