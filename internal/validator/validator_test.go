package validator

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/linabellbiu/goctl-validate/internal/processor"
	"github.com/zeromicro/go-zero/tools/goctl/plugin"
)

func TestProcessPlugin(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "goctl-validate-test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试API文件
	apiContent := `
syntax = "v1"

type UserReq {
	Name   string ` + "`" + `json:"name" validate:"required"` + "`" + `
	Mobile string ` + "`" + `json:"mobile" validate:"mobile"` + "`" + `
	Email  string ` + "`" + `json:"email" validate:"email"` + "`" + `
	IDCard string ` + "`" + `json:"id_card" validate:"idcard"` + "`" + `
}

service user-api {
	@handler GetUser
	get /users/:id (UserReq) returns (UserReq)
}
`
	apiFile := filepath.Join(tempDir, "user.api")
	if err := ioutil.WriteFile(apiFile, []byte(apiContent), 0644); err != nil {
		t.Fatalf("写入API文件失败: %v", err)
	}

	// 创建types.go文件目录
	typesDir := filepath.Join(tempDir, "types")
	if err := os.MkdirAll(typesDir, 0755); err != nil {
		t.Fatalf("创建types目录失败: %v", err)
	}

	// 创建测试types.go文件
	typesContent := `
package types

type UserReq struct {
	Name   string ` + "`" + `json:"name" validate:"required"` + "`" + `
	Mobile string ` + "`" + `json:"mobile" validate:"mobile"` + "`" + `
	Email  string ` + "`" + `json:"email" validate:"email"` + "`" + `
	IDCard string ` + "`" + `json:"id_card" validate:"idcard"` + "`" + `
}
`
	typesFile := filepath.Join(typesDir, "types.go")
	if err := ioutil.WriteFile(typesFile, []byte(typesContent), 0644); err != nil {
		t.Fatalf("写入types.go文件失败: %v", err)
	}

	// 创建插件实例
	p := &plugin.Plugin{
		ApiFilePath: apiFile,
		Dir:         tempDir,
	}

	// 设置选项
	options := processor.Options{
		EnableCustomValidation: true,
		DebugMode:              true,
	}

	// 执行插件处理
	err = ProcessPlugin(p, options)
	if err != nil {
		t.Fatalf("处理插件失败: %v", err)
	}

	// 检查是否生成了验证文件
	validationFile := filepath.Join(tempDir, "validation.go")
	if _, err := os.Stat(validationFile); os.IsNotExist(err) {
		t.Fatalf("未生成验证文件")
	}

	// 读取生成的验证文件内容
	content, err := ioutil.ReadFile(validationFile)
	if err != nil {
		t.Fatalf("读取验证文件失败: %v", err)
	}

	// 检查文件内容是否包含所有验证函数
	expectedFuncs := []string{
		"func validateMobile",
		"func validateIdCard",
	}

	for _, funcName := range expectedFuncs {
		if !strings.Contains(string(content), funcName) {
			t.Errorf("生成的验证文件中缺少函数: %s", funcName)
		}
	}
} 