package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/linabellbiu/goctl-validate/internal/processor"

	"github.com/zeromicro/go-zero/tools/goctl/plugin"
)

// ProcessPlugin 处理插件逻辑
func ProcessPlugin(p *plugin.Plugin, options processor.Options) error {
	// 根据p.Api 直接处理
	// return processor.ProcessTypesAPI(p, options)
	// 查找并处理types.go文件
	// 目录中是否已经生成过声明变量
	genFlag := false
	err := filepath.Walk(p.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.Contains(path, "internal/types/") && strings.HasSuffix(info.Name(), ".go") {
			if options.DebugMode {
				fmt.Printf("处理文件: %s\n", path)
			}
			gen, err := processor.ProcessTypesFile(genFlag, path, options)
			if err != nil {
				return err
			}
			if gen {
				genFlag = true
			}
		}
		return nil
	})
	return err
}
