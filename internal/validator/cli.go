package validator

import (
	"os"
	"strings"
)

// ParsePluginArgs 解析插件参数
func ParsePluginArgs() {
	if len(os.Args) <= 1 {
		return
	}

	// 检查第一个参数是否包含等号（插件形式的调用）
	arg := os.Args[1]
	if !strings.Contains(arg, "=") {
		return
	}

	// 分割参数
	parts := strings.SplitN(arg, "=", 2)
	if len(parts) != 2 {
		return
	}

	// 解析引号内的命令
	cmd := parts[1]
	// 去掉可能存在的引号
	cmd = strings.Trim(cmd, "\"'")

	// 分割命令和参数
	cmdParts := strings.Fields(cmd)
	if len(cmdParts) == 0 {
		return
	}

	// 重建参数数组
	newArgs := []string{os.Args[0]}

	// 子命令作为第一个参数
	if cmdParts[0] != "validate" && cmdParts[0] != "goctl-validate" {
		// 如果没有指定子命令，保留默认命令
		newArgs = append(newArgs, "validate")
		// 所有部分都是标志
		newArgs = append(newArgs, cmdParts...)
	} else {
		// 子命令已指定，直接使用
		if len(cmdParts) > 1 {
			newArgs = append(newArgs, cmdParts[0])
			newArgs = append(newArgs, cmdParts[1:]...)
		} else {
			newArgs = append(newArgs, cmdParts[0])
		}
	}

	// 添加剩余原始参数（跳过第一个已处理的参数）
	if len(os.Args) > 2 {
		newArgs = append(newArgs, os.Args[2:]...)
	}

	// 更新参数
	os.Args = newArgs
}
