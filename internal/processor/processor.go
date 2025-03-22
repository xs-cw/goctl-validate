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
	"sort"
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

	// 验证方法映射注释和开始部分
	ValidationRegisterComment = `// registerValidation 存储所有的验证方法
// key: 验证标签名称，value: 对应的验证函数`

	// 验证方法映射
	ValidateRegisterMap = `var registerValidation = map[string]validator.Func{
	"mobile": validateMobile, // 手机号验证
	"idcard": validateIdCard, // 身份证号验证
`

	// 自定义验证方法映射模板
	CustomValidationMapTemplate = `	"%s": validate%s, // %s
`

	// 验证方法注册初始化
	ValidateInitFunc = `
// 初始化并注册所有验证方法
func init() {
	// 遍历注册所有验证方法
	for tag, handler := range registerValidation {
		_ = validate.RegisterValidation(tag, handler)
	}
}
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
	hasValidatorImport := false
	for _, imp := range f.Imports {
		if imp.Path.Value == ValidateImport {
			hasValidatorImport = true
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

		// 添加验证方法映射注释
		validationFileContent.WriteString(ValidationRegisterComment + "\n")

		// 添加验证方法映射开始
		validationFileContent.WriteString(ValidateRegisterMap)

		// 按字母顺序排序标签，确保生成顺序一致
		var sortedTags []string
		for tag := range customTags {
			sortedTags = append(sortedTags, tag)
		}
		sort.Strings(sortedTags)

		// 如果启用了自定义验证，添加自定义验证标签
		if options.EnableCustomValidation && len(customTags) > 0 {
			for _, tag := range sortedTags {
				validationFileContent.WriteString(fmt.Sprintf(CustomValidationMapTemplate, tag, strings.Title(tag), tag))
			}
		}

		// 结束map定义
		validationFileContent.WriteString("}\n")

		// 添加init函数
		validationFileContent.WriteString(ValidateInitFunc + "\n")

		// 添加内置验证函数
		validationFileContent.WriteString(BuiltInValidationFunc + "\n")

		// 如果启用了自定义验证，添加自定义验证函数
		if options.EnableCustomValidation && len(customTags) > 0 {
			// 按字母顺序添加验证函数
			for _, tag := range sortedTags {
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
		existingRegLines := make(map[string]string) // 存储原始的注册行，用于保持注释一致性

		// 提取文件中所有的验证函数和注册信息
		funcRegex := regexp.MustCompile(`func validate(\w+)\(fl validator\.FieldLevel\) bool`)
		regRegex := regexp.MustCompile(`\t"(\w+)":\s*validate\w+,.*`)

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

		// 查找所有的注册行和对应的tag
		regMatches := regRegex.FindAllStringSubmatchIndex(validationContent, -1)
		for _, matchIndex := range regMatches {
			if len(matchIndex) >= 4 {
				startOfLine := validationContent[matchIndex[0]:matchIndex[1]]
				tag := validationContent[matchIndex[2]:matchIndex[3]]

				if tag != "mobile" && tag != "idcard" { // 跳过内置标签
					existingRegs[tag] = true
					existingRegLines[tag] = startOfLine // 保存整行内容
				}
			}
		}

		// 2. 收集所有标签，按字母顺序排序
		var allTags []string

		// 添加内置标签(固定顺序)
		allTags = append(allTags, "mobile", "idcard")

		// 收集所有自定义标签
		for tag := range customTags {
			if tag != "mobile" && tag != "idcard" {
				allTags = append(allTags, tag)
			}
		}

		// 收集现有但不在customTags中的标签
		for tag := range existingRegs {
			if tag != "mobile" && tag != "idcard" && !customTags[tag] {
				allTags = append(allTags, tag)
			}
		}

		// 除了内置标签外，对自定义标签按字母排序
		if len(allTags) > 2 {
			sort.Strings(allTags[2:])
		}

		// 3. 生成新的验证方法映射
		var newMapContent strings.Builder
		// 添加验证映射注释
		newMapContent.WriteString(ValidationRegisterComment + "\n")
		newMapContent.WriteString("var registerValidation = map[string]validator.Func{\n")

		// 按排序后的标签顺序添加
		for _, tag := range allTags {
			if tag == "mobile" {
				newMapContent.WriteString("\t\"mobile\": validateMobile, // 手机号验证\n")
			} else if tag == "idcard" {
				newMapContent.WriteString("\t\"idcard\": validateIdCard, // 身份证号验证\n")
			} else {
				// 如果存在原始的注册行，使用它保持格式一致
				if line, exists := existingRegLines[tag]; exists {
					newMapContent.WriteString(line + "\n")
				} else {
					// 否则使用标准格式
					newMapContent.WriteString(fmt.Sprintf(CustomValidationMapTemplate, tag, strings.Title(tag), tag))
				}
			}
		}

		newMapContent.WriteString("}\n")

		// 4. 检查所有缺失的验证函数
		// 为缺失的验证函数创建内容
		var missingFuncContent strings.Builder
		var missingTags []string

		// 收集所有需要验证函数但尚未存在的标签
		for tag := range customTags {
			if !existingFuncs[tag] {
				missingTags = append(missingTags, tag)
			}
		}

		// 按字母顺序添加验证函数
		sort.Strings(missingTags)
		for _, tag := range missingTags {
			missingFuncContent.WriteString(fmt.Sprintf(CustomValidationFuncTemplate, tag, strings.Title(tag), tag))
		}

		// 5. 替换原有的验证方法映射和init函数
		// 首先替换注释和map声明部分
		commentAndMapPattern := `(?s)// registerValidation.*?var registerValidation = map\[string\]validator\.Func\{.*?\}`
		mapRegex := regexp.MustCompile(commentAndMapPattern)

		var newValidationContent string
		if mapRegex.MatchString(validationContent) {
			// 如果已经有map格式了，替换它
			newValidationContent = mapRegex.ReplaceAllString(validationContent, newMapContent.String())

			// 移除validate变量的声明(如果存在)
			validateVarPattern := `var validate = validator\.New\(\)\n*`
			validateVarRegex := regexp.MustCompile(validateVarPattern)
			newValidationContent = validateVarRegex.ReplaceAllString(newValidationContent, "")

			// 添加缺失的验证函数到文件末尾
			if missingFuncContent.Len() > 0 {
				newValidationContent = newValidationContent + "\n" + missingFuncContent.String()
			}
		} else {
			// 如果是旧格式或者格式不匹配，创建一个全新的内容
			var newFullContent strings.Builder
			newFullContent.WriteString(fmt.Sprintf("package %s\n\n", packageName))

			// 添加导入
			newFullContent.WriteString("import (\n")
			newFullContent.WriteString("\t\"regexp\"\n")
			newFullContent.WriteString("\t" + ValidateImport + "\n")
			newFullContent.WriteString(")\n\n")

			// 添加验证方法映射（不添加validator变量）
			newFullContent.WriteString(newMapContent.String() + "\n")

			// 添加init函数
			newFullContent.WriteString(ValidateInitFunc + "\n")

			// 添加内置验证函数
			newFullContent.WriteString(BuiltInValidationFunc + "\n")

			// 提取所有自定义验证函数
			customFuncPattern := `(?s)// 自定义验证方法:.*?return true\n\}`
			customFuncRegex := regexp.MustCompile(customFuncPattern)
			customFuncMatches := customFuncRegex.FindAllString(validationContent, -1)

			// 按字母顺序整理自定义验证函数
			type FuncInfo struct {
				Tag  string
				Code string
			}
			var funcInfos []FuncInfo

			// 收集所有现有的函数
			for _, funcCode := range customFuncMatches {
				funcNameRegex := regexp.MustCompile(`func validate(\w+)\(`)
				nameMatch := funcNameRegex.FindStringSubmatch(funcCode)
				if len(nameMatch) > 1 {
					funcName := nameMatch[1]
					tag := strings.ToLower(funcName[:1]) + funcName[1:]
					funcInfos = append(funcInfos, FuncInfo{Tag: tag, Code: funcCode})
				}
			}

			// 对函数按标签名排序
			sort.Slice(funcInfos, func(i, j int) bool {
				return funcInfos[i].Tag < funcInfos[j].Tag
			})

			// 添加所有排序后的函数
			for _, funcInfo := range funcInfos {
				newFullContent.WriteString(funcInfo.Code + "\n\n")
			}

			// 添加缺失的验证函数
			for _, tag := range missingTags {
				newFullContent.WriteString(fmt.Sprintf(CustomValidationFuncTemplate, tag, strings.Title(tag), tag))
			}

			newValidationContent = newFullContent.String()
		}

		// 6. 格式化并写入文件
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

	// 检查是否需要添加验证器的导入
	if !hasValidatorImport && len(reqStructs) > 0 {
		// 找到最后一个导入
		lastImportPos := -1
		for i, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if ok && genDecl.Tok == token.IMPORT {
				lastImportPos = i
			}
		}

		// 找到文件中的包声明之后的位置
		fileContentStr := string(fileContent)
		packageEndPos := findPackagePosition(fileContentStr)
		if packageEndPos > 0 {
			packageEndPos = packageEndPos + len("package "+f.Name.Name)

			// 如果已经有导入部分
			if lastImportPos >= 0 {
				// 将验证器的导入添加到现有导入部分
				// 实现比较复杂，这里简单处理为在末尾添加
			} else {
				// 在包声明之后添加导入
				importStatement := "\n\nimport (\n\t" + ValidateImport + "\n)"

				// 插入导入语句
				if options.DebugMode {
					fmt.Println("添加验证器导入")
				}

				// 将导入添加到文件内容
				fileContentStr = fileContentStr[:packageEndPos] + importStatement + fileContentStr[packageEndPos:]
				fileContent = []byte(fileContentStr)
			}
		}

		// 添加验证器变量的声明
		validateVarStatement := "\n\n" + ValidateVar
		fileContentStr = string(fileContent) + validateVarStatement
		fileContent = []byte(fileContentStr)
	}

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

		if options.DebugMode {
			fmt.Printf("成功添加验证方法到 %s\n", filePath)
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
