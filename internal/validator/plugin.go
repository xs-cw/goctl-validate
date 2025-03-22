package validator

import (
	"fmt"
	"github.com/linabellbiu/goctl-validate/internal/processor"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/plugin"
)

// ProcessPlugin 处理插件逻辑
func ProcessPlugin(p *plugin.Plugin, options processor.Options) error {
	// 生成验证文件
	validationFile, err := GenerateValidationFile(p.ApiFilePath, options)
	if err != nil {
		return err
	}
	
	// 输出结果
	fmt.Printf("成功生成验证文件: %s\n", validationFile)

	// 查找并处理types.go文件
	err = filepath.Walk(p.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "types.go") {
			if err := processor.ProcessTypesFile(path, options); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
