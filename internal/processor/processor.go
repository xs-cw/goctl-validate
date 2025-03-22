package processor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Options 定义处理器的选项
type Options struct {
	// 是否添加自定义验证方法
	EnableCustomValidation bool
	// 是否启用调试模式
	DebugMode bool
}

// 验证器常量
const (
	ValidateImport = `"github.com/go-playground/validator/v10"`
	ValidateVar    = `var validate = validator.New()`

	// 基本验证初始化
	ValidateInitFuncStart = `
func init() {
	// 注册手机号验证方法
	validate.RegisterValidation("mobile", validateMobile)
	// 注册身份证号验证方法
	validate.RegisterValidation("idcard", validateIdCard)
`

	// 自定义验证方法注册模板
	CustomValidationRegisterTemplate = `	// 注册自定义验证方法: %s
	validate.RegisterValidation("%s", validate%s)
`

	// 初始化函数结束
	ValidateInitFuncEnd = `}
`

	// 自定义验证方法定义模板
	CustomValidationFuncTemplate = `
// 自定义验证方法: %s
func validate%s(fl validator.FieldLevel) bool {
	// 在这里实现 %s 的验证逻辑
	return true
}
`

	// 内置验证方法
	BuiltInValidationFunc = `
// 验证手机号
func validateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	// 使用正则表达式验证中国大陆手机号(13,14,15,16,17,18,19开头的11位数字)
	match, _ := regexp.MatchString("^1[3-9]\\d{9}$", mobile)
	return match
}

// 验证身份证号
func validateIdCard(fl validator.FieldLevel) bool {
	idCard := fl.Field().String()
	// 支持15位或18位身份证号
	match, _ := regexp.MatchString("(^\\d{15}$)|(^\\d{18}$)|(^\\d{17}(\\d|X|x)$)", idCard)
	return match
}
`
)

// ProcessTypesFile 处理types.go文件，添加验证逻辑
func ProcessTypesFile(filePath string, options Options) error {
	// 读取文件内容
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	if options.DebugMode {
		fmt.Println("============= 原始文件内容 =============")
		fmt.Println(string(fileContent))
		fmt.Println("=======================================")
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, fileContent, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("解析文件失败: %w", err)
	}

	// 寻找所有的请求结构体并生成验证方法
	var reqStructs []string
	// 定义变量，但不使用，防止编译错误
	existingValidations := make(map[string]bool)

	// 检查imports
	for _, imp := range f.Imports {
		if imp.Path.Value == ValidateImport {
			break
		}
	}

	// 提取自定义验证标签
	customTags := make(map[string]bool)

	// 收集所有请求结构体和自定义验证标签
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			// 如果是结构体类型
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// 不再仅限于以Req结尾的结构体
			// 检查所有结构体是否包含validate标签
			hasValidateTag := false
			for _, field := range structType.Fields.List {
				if field.Tag != nil {
					tag := field.Tag.Value
					validateTag := extractValidateTag(tag)
					if validateTag != "" {
						hasValidateTag = true
						break
					}
				}
			}

			// 如果结构体包含验证标签或是以Req结尾，则处理
			if hasValidateTag || strings.HasSuffix(typeSpec.Name.Name, "Req") {
				reqStructs = append(reqStructs, typeSpec.Name.Name)

				// 分析结构体字段的验证标签
				if options.EnableCustomValidation {
					for _, field := range structType.Fields.List {
						if field.Tag != nil {
							tag := field.Tag.Value

							// 提取验证标签
							validateTag := extractValidateTag(tag)
							if validateTag != "" {
								// 分析验证标签中的自定义验证器
								validators := strings.Split(validateTag, ",")
								for _, v := range validators {
									// 跳过内置验证器和空验证器
									if v == "" || isBuiltInValidator(v) {
										continue
									}

									// 添加自定义验证标签
									customTags[v] = true

									// 确认该验证器的验证函数是否已经存在
									if bytes.Contains(fileContent, []byte(fmt.Sprintf("func validate%s", strings.Title(v)))) {
										existingValidations[v] = true
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// 没有找到请求结构体，直接返回
	if len(reqStructs) == 0 && len(customTags) == 0 {
		return nil
	}

	// 获取文件所在的目录路径
	dirPath := filepath.Dir(filePath)

	// 验证文件的路径（与types.go在同一目录）
	validationFilePath := filepath.Join(dirPath, "validation.go")

	// 检查验证文件是否已存在
	validationExists := false
	validationContent := ""

	if _, err := os.Stat(validationFilePath); err == nil {
		// 验证文件已存在，读取内容
		validationBytes, err := os.ReadFile(validationFilePath)
		if err != nil {
			return fmt.Errorf("读取现有验证文件失败: %w", err)
		}
		validationContent = string(validationBytes)
		validationExists = true

		// 检查现有验证文件中的验证函数
		for tag := range customTags {
			if bytes.Contains(validationBytes, []byte(fmt.Sprintf("func validate%s", strings.Title(tag)))) {
				existingValidations[tag] = true
			}
		}
	}

	// 获取包名
	packageName := f.Name.Name

	// 生成验证文件内容
	var validationFileContent strings.Builder

	// 如果文件不存在，添加基本结构
	if !validationExists {
		validationFileContent.WriteString(fmt.Sprintf("package %s\n\n", packageName))

		// 添加导入
		validationFileContent.WriteString("import (\n")
		validationFileContent.WriteString("\t\"regexp\"\n")
		validationFileContent.WriteString("\t" + ValidateImport + "\n")
		validationFileContent.WriteString(")\n\n")

		// 添加验证器变量
		validationFileContent.WriteString(ValidateVar + "\n\n")

		// 添加init函数
		validationFileContent.WriteString(ValidateInitFuncStart)

		// 如果启用了自定义验证，添加自定义验证标签的注册
		if options.EnableCustomValidation && len(customTags) > 0 {
			for tag := range customTags {
				validationFileContent.WriteString(fmt.Sprintf(CustomValidationRegisterTemplate, tag, tag, strings.Title(tag)))
			}
		}

		// 添加init函数结束
		validationFileContent.WriteString(ValidateInitFuncEnd + "\n")

		// 添加内置验证函数
		validationFileContent.WriteString(BuiltInValidationFunc + "\n")

		// 如果启用了自定义验证，添加自定义验证函数
		if options.EnableCustomValidation && len(customTags) > 0 {
			for tag := range customTags {
				if !existingValidations[tag] {
					validationFileContent.WriteString(fmt.Sprintf(CustomValidationFuncTemplate, tag, strings.Title(tag), tag))
				}
			}
		}
	} else {
		// 文件已存在，需要更新
		// 1. 提取现有的验证函数和注册
		existingFuncs := make(map[string]bool)
		existingRegs := make(map[string]bool)

		// 提取文件中所有的验证函数和注册信息
		funcRegex := regexp.MustCompile(`func validate(\w+)\(fl validator\.FieldLevel\) bool`)
		regRegex := regexp.MustCompile(`validate\.RegisterValidation\("(\w+)"`)

		// 查找所有的验证函数
		funcMatches := funcRegex.FindAllStringSubmatch(validationContent, -1)
		for _, match := range funcMatches {
			if len(match) > 1 {
				// 提取函数名，如AgeRange，变为小写作为tag
				funcName := match[1]
				if funcName != "Mobile" && funcName != "IdCard" { // 跳过内置函数
					tag := strings.ToLower(funcName[0:1]) + funcName[1:]
					existingFuncs[tag] = true
				}
			}
		}

		// 查找所有的注册
		regMatches := regRegex.FindAllStringSubmatch(validationContent, -1)
		for _, match := range regMatches {
			if len(match) > 1 {
				tag := match[1]
				if tag != "mobile" && tag != "idcard" { // 跳过内置标签
					existingRegs[tag] = true
				}
			}
		}

		// 2. 生成新的init函数
		var newInitContent strings.Builder
		newInitContent.WriteString("func init() {\n")
		newInitContent.WriteString("\t// 注册手机号验证方法\n")
		newInitContent.WriteString("\tvalidate.RegisterValidation(\"mobile\", validateMobile)\n")
		newInitContent.WriteString("\t// 注册身份证号验证方法\n")
		newInitContent.WriteString("\tvalidate.RegisterValidation(\"idcard\", validateIdCard)\n")

		// 收集处理过的标签，避免重复
		processedTags := make(map[string]bool)

		// 添加当前types.go文件中的自定义验证标签注册
		for tag := range customTags {
			processedTags[tag] = true
			if !existingRegs[tag] {
				newInitContent.WriteString(fmt.Sprintf(CustomValidationRegisterTemplate, tag, tag, strings.Title(tag)))
			} else {
				// 保留已有的注册代码
				regPattern := fmt.Sprintf(`\t// 注册自定义验证方法: %s\n\tvalidate.RegisterValidation\("%s", validate%s\)`, tag, tag, strings.Title(tag))
				regRegex := regexp.MustCompile(regPattern)
				if regRegex.MatchString(validationContent) {
					matches := regRegex.FindStringSubmatch(validationContent)
					if len(matches) > 0 {
						newInitContent.WriteString(matches[0] + "\n")
					}
				}
			}
		}

		// 添加所有已存在的验证函数注册，即使不在当前types.go中
		for tag := range existingRegs {
			if !processedTags[tag] && tag != "mobile" && tag != "idcard" {
				processedTags[tag] = true
				regPattern := fmt.Sprintf(`\t// 注册自定义验证方法: %s\n\tvalidate.RegisterValidation\("%s", validate%s\)`, tag, tag, strings.Title(tag))
				regRegex := regexp.MustCompile(regPattern)
				if regRegex.MatchString(validationContent) {
					matches := regRegex.FindStringSubmatch(validationContent)
					if len(matches) > 0 {
						newInitContent.WriteString(matches[0] + "\n")
					} else {
						// 使用标准格式添加
						newInitContent.WriteString(fmt.Sprintf(CustomValidationRegisterTemplate, tag, tag, strings.Title(tag)))
					}
				}
			}
		}

		newInitContent.WriteString("}\n")

		// 3. 替换原有的init函数
		initRegex := regexp.MustCompile(`(?s)func\s+init\(\)\s*\{.*?\}`)
		newValidationContent := initRegex.ReplaceAllString(validationContent, newInitContent.String())

		// 4. 添加缺失的验证函数
		// 首先添加当前types.go中缺少的验证函数
		for tag := range customTags {
			if !existingFuncs[tag] {
				if !strings.HasSuffix(newValidationContent, "\n\n") {
					if strings.HasSuffix(newValidationContent, "\n") {
						newValidationContent += "\n"
					} else {
						newValidationContent += "\n\n"
					}
				}

				newValidationContent += fmt.Sprintf(CustomValidationFuncTemplate, tag, strings.Title(tag), tag)
			}
		}

		// 5. 格式化并写入文件
		formatted, err := format.Source([]byte(newValidationContent))
		if err != nil {
			return fmt.Errorf("格式化更新的验证文件代码失败: %w", err)
		}

		if err := os.WriteFile(validationFilePath, formatted, 0644); err != nil {
			return fmt.Errorf("写入更新的验证文件失败: %w", err)
		}

		if options.DebugMode {
			fmt.Printf("成功更新验证文件: %s\n", validationFilePath)
		}
	}

	// 为所有请求结构体生成验证方法
	var methodsBuilder strings.Builder
	for _, structName := range reqStructs {
		// 检查是否已经存在该结构体的Validate方法
		if !strings.Contains(string(fileContent), "func (r *"+structName+") Validate()") {
			methodsBuilder.WriteString(fmt.Sprintf("\nfunc (r *%s) Validate() error {\n\treturn validate.Struct(r)\n}\n", structName))
		}
	}

	// 将方法添加到types.go文件末尾
	if methodsBuilder.Len() > 0 {
		modifiedContent := string(fileContent) + methodsBuilder.String()

		// 格式化代码
		formatted, err := format.Source([]byte(modifiedContent))
		if err != nil {
			return fmt.Errorf("格式化代码失败: %w", err)
		}

		// 写回文件
		if err := os.WriteFile(filePath, formatted, 0644); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
	}

	// 如果需要创建或更新验证文件
	if !validationExists {
		// 格式化验证文件内容
		formatted, err := format.Source([]byte(validationFileContent.String()))
		if err != nil {
			return fmt.Errorf("格式化验证文件代码失败: %w", err)
		}

		// 写入验证文件
		if err := os.WriteFile(validationFilePath, formatted, 0644); err != nil {
			return fmt.Errorf("写入验证文件失败: %w", err)
		}

		if options.DebugMode {
			fmt.Printf("成功创建验证文件: %s\n", validationFilePath)
		}
	}

	return nil
}

// 从结构体标签中提取validate标签内容
func extractValidateTag(tag string) string {
	re := regexp.MustCompile(`validate:"([^"]*)"`)
	matches := re.FindStringSubmatch(tag)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// 判断是否是内置验证器
func isBuiltInValidator(validator string) bool {
	builtInValidators := map[string]bool{
		"required": true,
		"mobile":   true,
		"idcard":   true,
		"email":    true,
		"url":      true,
		"ip":       true,
		"len":      true,
		"min":      true,
		"max":      true,
		"eq":       true,
		"ne":       true,
		"lt":       true,
		"lte":      true,
		"gt":       true,
		"gte":      true,
		"oneof":    true,
		"numeric":  true,
		"alpha":    true,
		"alphanum": true,
	}

	// 检查是否是带参数的内置验证器，如min=10
	parts := strings.Split(validator, "=")
	if len(parts) > 1 {
		return builtInValidators[parts[0]]
	}

	return builtInValidators[validator]
}

// findPackagePosition 查找package关键字在文件中的位置
func findPackagePosition(content string) int {
	// 使用正则表达式查找package关键字
	re := regexp.MustCompile(`(?m)^package\s+\w+`)
	loc := re.FindStringIndex(content)
	if loc == nil {
		return -1
	}
	return loc[0]
}
