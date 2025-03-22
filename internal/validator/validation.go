package validator

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/linabellbiu/goctl-validate/internal/processor"
)

// GenerateValidationFile 生成验证文件
func GenerateValidationFile(apiFile string, options processor.Options) (string, error) {
	// 解析API文件
	types, err := parseAPIFile(apiFile)
	if err != nil {
		return "", fmt.Errorf("解析API文件失败: %w", err)
	}
	
	// 检查是否已存在验证文件
	validationFilePath := filepath.Join(filepath.Dir(apiFile), "validation.go")
	existingCode := ""
	
	if fileExists(validationFilePath) {
		// 读取现有文件内容
		content, err := ioutil.ReadFile(validationFilePath)
		if err != nil {
			return "", fmt.Errorf("读取现有验证文件失败: %w", err)
		}
		existingCode = string(content)
	}
	
	// 解析现有验证函数，确保不覆盖
	existingFuncs := parseExistingValidationFuncs(existingCode)
	
	// 生成新的验证代码，保留现有函数
	code, err := generateValidationCode(types, options, existingFuncs)
	if err != nil {
		return "", fmt.Errorf("生成验证代码失败: %w", err)
	}
	
	// 写入文件
	err = ioutil.WriteFile(validationFilePath, []byte(code), 0644)
	if err != nil {
		return "", fmt.Errorf("写入验证文件失败: %w", err)
	}
	
	return validationFilePath, nil
}

// parseAPIFile 解析API文件并返回结构体类型信息
func parseAPIFile(filePath string) ([]string, error) {
	// 读取文件内容
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取API文件失败: %w", err)
	}

	// 提取包名
	packageName := "types" // 默认包名
	packageRegex := regexp.MustCompile(`(?m)^package\s+(\w+)`)
	if matches := packageRegex.FindStringSubmatch(string(fileContent)); len(matches) > 1 {
		packageName = matches[1]
	}
	// 避免未使用变量警告
	_ = packageName

	// 使用正则表达式识别结构体
	re := regexp.MustCompile(`(?m)type\s+\([\s\S]+?\)`)
	matches := re.FindAllString(string(fileContent), -1)

	var types []string
	for _, match := range matches {
		types = append(types, match)
	}

	return types, nil
}

// generateValidationCode 生成验证代码
func generateValidationCode(types []string, options processor.Options, existingFuncs map[string]string) (string, error) {
	var buf strings.Builder
	packageName := "types" // 默认包名为types

	// 写入包名
	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// 写入import
	buf.WriteString("import (\n")
	buf.WriteString("\t\"regexp\"\n")
	buf.WriteString("\t\"github.com/go-playground/validator/v10\"\n")
	buf.WriteString(")\n\n")

	// 写入validator变量
	buf.WriteString("var validate = validator.New()\n\n")

	// 生成初始化函数和自定义验证方法
	hasCustomValidation := false
	hasMobileValidation := false
	hasIdCardValidation := false

	// 检查现有函数
	for funcName := range existingFuncs {
		if funcName == "customValidation" {
			hasCustomValidation = true
		} else if funcName == "validateMobile" {
			hasMobileValidation = true
		} else if funcName == "validateIdCard" {
			hasIdCardValidation = true
		}
	}

	// 写入init函数
	buf.WriteString("func init() {\n")
	
	// 如果开启了自定义验证且没有现有的自定义验证函数
	if options.EnableCustomValidation && !hasCustomValidation {
		buf.WriteString("\t// 注册自定义验证方法\n")
		buf.WriteString("\tvalidate.RegisterValidation(\"custom_validation\", customValidation)\n")
	}
	
	// 添加手机号验证
	buf.WriteString("\t// 注册手机号验证方法\n")
	buf.WriteString("\tvalidate.RegisterValidation(\"mobile\", validateMobile)\n")
	
	// 添加身份证验证
	buf.WriteString("\t// 注册身份证号验证方法\n")
	buf.WriteString("\tvalidate.RegisterValidation(\"idcard\", validateIdCard)\n")
	
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

	return buf.String(), nil
}

// parseExistingValidationFuncs 解析现有验证函数
func parseExistingValidationFuncs(code string) map[string]string {
	funcs := make(map[string]string)
	
	// 使用正则表达式匹配函数定义
	funcRegex := regexp.MustCompile(`(?m)^func\s+(\w+)\s*\([^)]*\)\s*bool\s*{`)
	matches := funcRegex.FindAllStringSubmatchIndex(code, -1)
	
	for i, match := range matches {
		if len(match) < 4 {
			continue
		}
		
		funcName := code[match[2]:match[3]]
		startPos := match[0]
		endPos := len(code)
		
		// 如果有下一个函数，找到当前函数的结束位置
		if i < len(matches)-1 {
			endPos = matches[i+1][0]
		}
		
		// 查找函数体的结束大括号
		braceCount := 1
		for j := match[1]; j < endPos; j++ {
			if code[j] == '{' {
				braceCount++
			} else if code[j] == '}' {
				braceCount--
				if braceCount == 0 {
					endPos = j + 1
					break
				}
			}
		}
		
		funcs[funcName] = code[startPos:endPos]
	}
	
	return funcs
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
} 