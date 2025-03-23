package main

import (
	"fmt"
	"os"

	"github.com/linabellbiu/goctl-validate/internal/processor"
	"github.com/linabellbiu/goctl-validate/internal/validator"

	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/tools/goctl/plugin"
)

var (
	// 版本信息
	version = "1.1.0"
	// 是否启用调试模式
	debugMode bool
	// 翻译语言，默认为中文
	translationLanguage string
	// 是否启用验证功能（自定义验证和翻译）
	enableValidation bool

	rootCmd = &cobra.Command{
		Use:     "validate",
		Short:   "A goctl plugin to generate validation code for API types",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := plugin.NewPlugin()
			if err != nil {
				return err
			}

			// 设置处理选项
			options := processor.Options{
				EnableCustomValidation: enableValidation, // 自动启用自定义验证
				DebugMode:              debugMode,
				EnableTranslation:      enableValidation, // 自动启用翻译功能
				TranslationLanguage:    translationLanguage,
			}

			return validator.ProcessPlugin(p, options)
		},
	}
)

func init() {
	rootCmd.Flags().BoolVar(&enableValidation, "all", true, "启用完整验证（包含自定义验证和错误翻译功能）")
	rootCmd.Flags().BoolVar(&debugMode, "debug", false, "启用调试模式")
	rootCmd.Flags().StringVar(&translationLanguage, "lang", "zh", "翻译语言 (默认: zh 中文)")
}

func main() {
	// 处理 goctl 插件参数格式
	validator.ParsePluginArgs()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
