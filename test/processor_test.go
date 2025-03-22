package test

import (
	"github.com/linabellbiu/goctl-validate/internal/processor"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testTypesContent = `package types

type StatusReq struct {
	Id   int64  ` + "`" + `json:"id" validate:"required,gt=0"` + "`" + `
	Name string ` + "`" + `json:"name" validate:"required"` + "`" + `
}

type CreateItemReq struct {
	Name        string  ` + "`" + `json:"name" validate:"required,min=2,max=50"` + "`" + `
	Description string  ` + "`" + `json:"description" validate:"omitempty,max=200"` + "`" + `
	Price       float64 ` + "`" + `json:"price" validate:"required,gt=0"` + "`" + `
}
`

func TestProcessTypesFile(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "goctl-validate-test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	typesFile := filepath.Join(tempDir, "types.go")
	if err := os.WriteFile(typesFile, []byte(testTypesContent), 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	// 处理文件
	options := processor.Options{
		EnableCustomValidation: false,
		DebugMode:              false,
	}
	if err := processor.ProcessTypesFile(typesFile, options); err != nil {
		t.Fatalf("处理文件失败: %v", err)
	}

	// 读取处理后的types.go文件
	typesContent, err := os.ReadFile(typesFile)
	if err != nil {
		t.Fatalf("读取处理后的types.go文件失败: %v", err)
	}
	
	// 读取生成的validation.go文件
	validationFile := filepath.Join(tempDir, "validation.go")
	validationContent, err := os.ReadFile(validationFile)
	if err != nil {
		t.Fatalf("读取生成的validation.go文件失败: %v", err)
	}

	// 验证results
	processedTypesContent := string(typesContent)
	validationString := string(validationContent)

	// 检查types.go文件中的验证方法
	// 检查是否添加了StatusReq的Validate方法
	if !contains(processedTypesContent, "func (r *StatusReq) Validate() error {") {
		t.Error("未添加StatusReq的Validate方法")
	}

	// 检查是否添加了CreateItemReq的Validate方法
	if !contains(processedTypesContent, "func (r *CreateItemReq) Validate() error {") {
		t.Error("未添加CreateItemReq的Validate方法")
	}
	
	// 检查validation.go文件中的内容
	// 检查是否添加了validator导入
	if !contains(validationString, `"github.com/go-playground/validator/v10"`) {
		t.Error("未添加validator导入")
	}

	// 检查是否添加了validate变量
	if !contains(validationString, "var validate = validator.New()") {
		t.Error("未添加validate变量")
	}
	
	// 检查是否添加了验证函数
	if !contains(validationString, "func validateMobile(fl validator.FieldLevel) bool {") {
		t.Error("未添加手机号验证函数")
	}
	
	if !contains(validationString, "func validateIdCard(fl validator.FieldLevel) bool {") {
		t.Error("未添加身份证号验证函数")
	}
}

func TestProcessTypesFileWithCustomValidation(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "goctl-validate-test-custom")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	typesFile := filepath.Join(tempDir, "types.go")
	if err := os.WriteFile(typesFile, []byte(testTypesContent), 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	// 处理文件（启用自定义验证）
	options := processor.Options{
		EnableCustomValidation: true,
		DebugMode:              false,
	}
	if err := processor.ProcessTypesFile(typesFile, options); err != nil {
		t.Fatalf("处理文件失败: %v", err)
	}

	// 读取生成的validation.go文件
	validationFile := filepath.Join(tempDir, "validation.go")
	validationContent, err := os.ReadFile(validationFile)
	if err != nil {
		t.Fatalf("读取生成的validation.go文件失败: %v", err)
	}

	// 验证结果
	validationString := string(validationContent)

	// 检查是否添加了自定义验证方法
	if !contains(validationString, "func customValidation(fl validator.FieldLevel) bool {") {
		t.Error("未添加自定义验证方法")
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
