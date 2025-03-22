package processor

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"regexp"
)

// ParseAPIFile 解析API文件并返回结构体类型信息
func ParseAPIFile(filePath string) ([]string, error) {
	// 读取文件内容
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取API文件失败: %w", err)
	}

	// 使用正则表达式识别结构体
	re := regexp.MustCompile(`(?m)type\s+\([\s\S]+?\)`)
	matches := re.FindAllString(string(fileContent), -1)

	var types []string
	for _, match := range matches {
		types = append(types, match)
	}

	return types, nil
}

// GenerateValidationCode 生成验证代码
func GenerateValidationCode(types []string, options Options, existingFuncs map[string]string) (string, error) {
	var buf bytes.Buffer
	packageName := "types" // 默认包名为types

	// 写入包名
	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// 写入import
	buf.WriteString("import (\n")
	buf.WriteString("\t\"regexp\"\n")
	buf.WriteString(fmt.Sprintf("\t%s\n", ValidateImport))
	buf.WriteString(")\n\n")

	// 写入validator变量
	buf.WriteString(ValidateVar + "\n\n")

	// 生成初始化函数和自定义验证方法
	hasCustomValidation := false
	hasMobileValidation := false
	hasIdCardValidation := false

	// 检查现有函数
	for funcName, funcCode := range existingFuncs {
		if funcName == "customValidation" {
			hasCustomValidation = true
		} else if funcName == "validateMobile" {
			hasMobileValidation = true
		} else if funcName == "validateIdCard" {
			hasIdCardValidation = true
		}
		// 避免未使用变量警告
		_ = funcCode
	}

	// 写入init函数
	buf.WriteString("func init() {\n")
	
	// 如果开启了自定义验证且没有现有的自定义验证函数
	if options.EnableCustomValidation && !hasCustomValidation {
		buf.WriteString("\t// 注册自定义验证方法\n")
		buf.WriteString("\tvalidate.RegisterValidation(\"custom_validation\", customValidation)\n")
	}
	
	// 添加手机号验证
	if !hasMobileValidation {
		buf.WriteString("\t// 注册手机号验证方法\n")
		buf.WriteString("\tvalidate.RegisterValidation(\"mobile\", validateMobile)\n")
	}
	
	// 添加身份证验证
	if !hasIdCardValidation {
		buf.WriteString("\t// 注册身份证号验证方法\n")
		buf.WriteString("\tvalidate.RegisterValidation(\"idcard\", validateIdCard)\n")
	}
	
	buf.WriteString("}\n\n")

	// 按顺序添加现有函数
	for funcName, funcCode := range existingFuncs {
		buf.WriteString(funcCode + "\n\n")
		// 避免未使用变量警告
		_ = funcName
	}

	// 添加未定义的验证函数
	if options.EnableCustomValidation && !hasCustomValidation {
		buf.WriteString("// 自定义验证方法示例\n")
		buf.WriteString("func customValidation(fl validator.FieldLevel) bool {\n")
		buf.WriteString("\t// 在这里实现自定义验证逻辑\n")
		buf.WriteString("\treturn true\n")
		buf.WriteString("}\n\n")
	}

	if !hasMobileValidation {
		buf.WriteString("// 验证手机号\n")
		buf.WriteString("func validateMobile(fl validator.FieldLevel) bool {\n")
		buf.WriteString("\tmobile := fl.Field().String()\n")
		buf.WriteString("\t// 使用正则表达式验证中国大陆手机号(13,14,15,16,17,18,19开头的11位数字)\n")
		buf.WriteString("\tmatch, _ := regexp.MatchString(`^1[3-9]\\d{9}$`, mobile)\n")
		buf.WriteString("\treturn match\n")
		buf.WriteString("}\n\n")
	}

	if !hasIdCardValidation {
		buf.WriteString("// 验证身份证号\n")
		buf.WriteString("func validateIdCard(fl validator.FieldLevel) bool {\n")
		buf.WriteString("\tidCard := fl.Field().String()\n")
		buf.WriteString("\t// 支持15位或18位身份证号\n")
		buf.WriteString("\tmatch, _ := regexp.MatchString(`(^\\d{15}$)|(^\\d{18}$)|(^\\d{17}(\\d|X|x)$)`, idCard)\n")
		buf.WriteString("\treturn match\n")
		buf.WriteString("}\n\n")
	}

	// 格式化代码
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.String(), fmt.Errorf("格式化验证代码失败: %w", err)
	}

	return string(formatted), nil
} 