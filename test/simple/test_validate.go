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
