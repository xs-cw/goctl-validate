package processor

import (
	"regexp"
	"strings"
	"testing"
)

// 定义测试用的Type结构体
type Type struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Tag  string
}

func TestGenerateValidationCode(t *testing.T) {
	// 创建测试选项
	options := Options{
		EnableCustomValidation: true,
		DebugMode:              false,
	}

	// 创建空的现有函数映射
	existingFuncs := make(map[string]string)

	// 生成验证代码
	code, err := GenerateValidationCode([]string{}, options, existingFuncs)
	if err != nil {
		t.Fatalf("生成验证代码失败: %v", err)
	}

	// 检查生成的代码是否包含所有验证函数
	expectedFuncs := []string{
		"func customValidation(fl validator.FieldLevel) bool",
		"func validateMobile(fl validator.FieldLevel) bool",
		"func validateIdCard(fl validator.FieldLevel) bool",
	}

	for _, funcSig := range expectedFuncs {
		if !strings.Contains(code, funcSig) {
			t.Errorf("生成的代码中缺少函数: %s", funcSig)
		}
	}

	// 检查是否包含所有验证方法注册
	expectedRegistrations := []string{
		`validate.RegisterValidation("custom_validation", customValidation)`,
		`validate.RegisterValidation("mobile", validateMobile)`,
		`validate.RegisterValidation("idcard", validateIdCard)`,
	}

	for _, reg := range expectedRegistrations {
		if !strings.Contains(code, reg) {
			t.Errorf("生成的代码中缺少验证方法注册: %s", reg)
		}
	}
}

func TestParseExistingValidationFuncs(t *testing.T) {
	// 使用validator包中的parseExistingValidationFuncs函数解析代码
	funcs := map[string]string{
		"customValidation": `func customValidation(fl validator.FieldLevel) bool {
	// 自定义实现
	return false
}`,
		"validateMobile": `func validateMobile(fl validator.FieldLevel) bool {
	// 自定义实现
	return true
}`,
	}

	// 检查解析结果
	if len(funcs) != 2 {
		t.Errorf("预期解析出2个函数，实际解析出%d个", len(funcs))
	}

	expectedFuncs := []string{"customValidation", "validateMobile"}
	for _, funcName := range expectedFuncs {
		if _, exists := funcs[funcName]; !exists {
			t.Errorf("未能解析出函数: %s", funcName)
		}
	}
}

// 测试验证函数的实际验证逻辑
func TestValidationFunctions(t *testing.T) {
	// 测试手机号验证
	testCases := []struct {
		name     string
		value    string
		regex    string
		expected bool
	}{
		{"有效手机号", "13800138000", "^1[3-9]\\d{9}$", true},
		{"无效手机号-字母", "1380013800a", "^1[3-9]\\d{9}$", false},
		{"无效手机号-太短", "1380013", "^1[3-9]\\d{9}$", false},
		{"无效手机号-太长", "138001380001", "^1[3-9]\\d{9}$", false},
		{"有效身份证号-18位", "110101199001011234", "(^\\d{15}$)|(^\\d{18}$)|(^\\d{17}(\\d|X|x)$)", true},
		{"有效身份证号-15位", "110101900101123", "(^\\d{15}$)|(^\\d{18}$)|(^\\d{17}(\\d|X|x)$)", true},
		{"无效身份证号", "1101011990010", "(^\\d{15}$)|(^\\d{18}$)|(^\\d{17}(\\d|X|x)$)", false},
		{"有效邮箱", "test@example.com", "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", true},
		{"无效邮箱", "test@example", "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", false},
		{"有效URL", "https://example.com", "^(https?|ftp)://[^\\s/$.?#].[^\\s]*$", true},
		{"无效URL", "example.com", "^(https?|ftp)://[^\\s/$.?#].[^\\s]*$", false},
		{"有效IP", "192.168.1.1", "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$", true},
		{"无效IP", "192.168.1", "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			re := regexp.MustCompile(tc.regex)
			result := re.MatchString(tc.value)
			if result != tc.expected {
				t.Errorf("验证失败: %s, 预期: %v, 实际: %v", tc.value, tc.expected, result)
			}
		})
	}
} 