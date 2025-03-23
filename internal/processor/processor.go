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
	// 是否启用验证标签翻译
	EnableTranslation bool
	// 翻译语言，默认为中文
	TranslationLanguage string
}

// 验证器常量
const (
	ValidateImport = `"github.com/go-playground/validator/v10"`
	ValidateVar    = `var validate = validator.New()`

	// 翻译相关导入
	TranslationZhImport             = `"github.com/go-playground/locales/zh"`
	TranslationZhTwImport           = `"github.com/go-playground/locales/zh_Hant_TW"`
	TranslationEnImport             = `"github.com/go-playground/locales/en"`
	UniversalTranslatorImport       = `ut "github.com/go-playground/universal-translator"`
	ValidatorTranslationsImport     = `zhTranslations "github.com/go-playground/validator/v10/translations/zh"`
	ValidatorTranslationsZhTwImport = `zhTwTranslations "github.com/go-playground/validator/v10/translations/zh_tw"`
	ValidatorTranslationsEnImport   = `enTranslations "github.com/go-playground/validator/v10/translations/en"`

	// 翻译器变量
	TranslationVars = `// 全局翻译器
var (
	translator ut.Translator
	uni *ut.UniversalTranslator
)
`

	// 中文翻译初始化
	ZhTranslationInit = `
	// 中文翻译器初始化
	zh := zh.New()
	uni = ut.New(zh, zh)
	translator, _ = uni.GetTranslator("zh")
	// 注册翻译器
	_ = zhTranslations.RegisterDefaultTranslations(validate, translator)
	
	// 注册自定义验证器的翻译，函数位于 translations.go
	registerCustomValidationTranslations(translator)
`

	// 繁体中文翻译初始化
	ZhTwTranslationInit = `
	// 繁体中文翻译器初始化
	zhTw := zh_Hant_TW.New()
	uni = ut.New(zhTw, zhTw)
	translator, _ = uni.GetTranslator("zh_Hant_TW")
	// 注册翻译器
	_ = zhTwTranslations.RegisterDefaultTranslations(validate, translator)
	
	// 注册自定义验证器的翻译，函数位于 translations.go
	registerCustomValidationTranslations(translator)
`

	// 英文翻译初始化
	EnTranslationInit = `
	// 英文翻译器初始化
	en := en.New()
	uni = ut.New(en, en)
	translator, _ = uni.GetTranslator("en")
	// 注册翻译器
	_ = enTranslations.RegisterDefaultTranslations(validate, translator)
	
	// 注册自定义验证器的翻译，函数位于 translations.go
	registerCustomValidationTranslations(translator)
`

	// 翻译函数
	TranslateErrorFunc = `
// TranslateError 翻译验证错误信息
func TranslateError(err error) string {
	if err == nil {
		return ""
	}
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}
	
	// 翻译每个错误
	var errMessages []string
	for _, e := range validationErrors {
		errMessages = append(errMessages, e.Translate(translator))
	}
	
	return strings.Join(errMessages, ", ")
}
`

	// 自定义翻译注册函数 - 将被移到单独的translations.go文件中
	CustomTranslationsFunc = `// 注册自定义验证器的翻译
func registerCustomValidationTranslations(trans ut.Translator) {
	// 注册 mobile 验证器的翻译
	_ = trans.Add("mobile", "{0}必须是有效的手机号码", true)
	_ = validate.RegisterTranslation("mobile", trans, func(ut ut.Translator) error {
		return nil
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("mobile", fe.Field())
		return t
	})
	
	// 注册 idcard 验证器的翻译
	_ = trans.Add("idcard", "{0}必须是有效的身份证号码", true)
	_ = validate.RegisterTranslation("idcard", trans, func(ut ut.Translator) error {
		return nil
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("idcard", fe.Field())
		return t
	})
}`

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
	if options.DebugMode {
		fmt.Println("==== 开始处理文件:", filePath, " ====")
		fmt.Printf("选项: EnableCustomValidation=%v, EnableTranslation=%v, DebugMode=%v\n",
			options.EnableCustomValidation, options.EnableTranslation, options.DebugMode)
	}

	// 读取文件内容
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	if options.DebugMode {
		fmt.Println("成功读取文件，长度:", len(fileContent))
	}

	// 解析文件
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, fileContent, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("解析文件失败: %w", err)
	}

	if options.DebugMode {
		fmt.Println("成功解析文件")
	}

	// 获取文件所在目录和包名
	dirPath := filepath.Dir(filePath)
	packageName := f.Name.Name

	if options.DebugMode {
		fmt.Println("文件目录:", dirPath)
		fmt.Println("包名:", packageName)
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
								// 分析验证标签中的所有验证器
								validators := strings.Split(validateTag, ",")
								for _, v := range validators {
									// 提取验证器名称，如min=10中的min
									parts := strings.Split(v, "=")
									baseValidator := parts[0]

									// 跳过空验证器
									if baseValidator == "" {
										continue
									}

									// 跳过内置验证器
									if isBuiltInValidator(baseValidator) {
										continue
									}

									// 添加自定义验证标签 - 使用基础验证器名称
									customTags[baseValidator] = true
									if options.DebugMode {
										fmt.Printf("找到自定义验证标签: %s\n", baseValidator)
									}

									// 确认该验证器的验证函数是否已经存在
									if bytes.Contains(fileContent, []byte(fmt.Sprintf("func validate%s", strings.Title(baseValidator)))) {
										existingValidations[baseValidator] = true
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

	if options.DebugMode {
		fmt.Println("发现的请求结构体:", reqStructs)
		fmt.Println("收集到的自定义标签:")
		for tag := range customTags {
			fmt.Printf("  - %s\n", tag)
		}
	}

	// 检查验证文件是否存在
	validationFilePath := filepath.Join(dirPath, "validation.go")
	validationExists := false
	var validationContent string

	if _, err := os.Stat(validationFilePath); err == nil {
		validationExists = true
		validationBytes, err := os.ReadFile(validationFilePath)
		if err != nil {
			return fmt.Errorf("读取验证文件失败: %w", err)
		}
		validationContent = string(validationBytes)

		if options.DebugMode {
			fmt.Println("验证文件已存在，长度:", len(validationContent))
		}
	} else {
		if options.DebugMode {
			fmt.Println("验证文件不存在，将创建新文件:", validationFilePath)
		}
	}

	// 如果启用了翻译功能，生成translations.go文件
	if options.EnableTranslation {
		if err := generateTranslationsFile(dirPath, packageName, options); err != nil {
			return err
		}
	}

	// 生成验证文件内容
	var validationFileContent strings.Builder
	var newValidationContent string

	// 如果文件不存在，添加基本结构
	if !validationExists {
		validationFileContent.WriteString(fmt.Sprintf("package %s\n\n", packageName))

		// 添加导入
		validationFileContent.WriteString("import (\n")
		validationFileContent.WriteString("\t\"regexp\"\n")
		validationFileContent.WriteString("\t\"strings\"\n")
		validationFileContent.WriteString("\t" + ValidateImport + "\n")

		// 如果启用了翻译
		if options.EnableTranslation {
			// 根据选择的语言添加相应的导入
			switch options.TranslationLanguage {
			case "zh":
				validationFileContent.WriteString("\t" + TranslationZhImport + "\n")
				validationFileContent.WriteString("\t" + UniversalTranslatorImport + "\n")
				validationFileContent.WriteString("\t" + ValidatorTranslationsImport + "\n")
			case "zh_tw":
				validationFileContent.WriteString("\t" + TranslationZhTwImport + "\n")
				validationFileContent.WriteString("\t" + UniversalTranslatorImport + "\n")
				validationFileContent.WriteString("\t" + ValidatorTranslationsZhTwImport + "\n")
			case "en":
				validationFileContent.WriteString("\t" + TranslationEnImport + "\n")
				validationFileContent.WriteString("\t" + UniversalTranslatorImport + "\n")
				validationFileContent.WriteString("\t" + ValidatorTranslationsEnImport + "\n")
			default:
				// 默认使用中文
				validationFileContent.WriteString("\t" + TranslationZhImport + "\n")
				validationFileContent.WriteString("\t" + UniversalTranslatorImport + "\n")
				validationFileContent.WriteString("\t" + ValidatorTranslationsImport + "\n")
			}
		}

		validationFileContent.WriteString(")\n\n")

		// 添加验证器变量
		validationFileContent.WriteString(ValidateVar + "\n\n")

		// 如果启用了翻译，添加翻译器变量
		if options.EnableTranslation {
			validationFileContent.WriteString(TranslationVars + "\n")
		}

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
			for i, tag := range sortedTags {
				comma := ","
				if i == len(sortedTags)-1 {
					comma = ""
				}
				validationFileContent.WriteString(fmt.Sprintf("\t\"%s\": validate%s%s\n", tag, strings.Title(tag), comma))
			}
		}

		// 结束map定义
		validationFileContent.WriteString("}\n")

		// 添加init函数，并在其中初始化翻译器
		validationFileContent.WriteString("\n// 初始化并注册所有验证方法\nfunc init() {\n")
		validationFileContent.WriteString("\t// 遍历注册所有验证方法\n\tfor tag, handler := range registerValidation {\n\t\t_ = validate.RegisterValidation(tag, handler)\n\t}\n")

		// 如果启用了翻译，添加翻译器初始化代码
		if options.EnableTranslation {
			switch options.TranslationLanguage {
			case "zh":
				validationFileContent.WriteString(ZhTranslationInit)
			case "zh_tw":
				validationFileContent.WriteString(ZhTwTranslationInit)
			case "en":
				validationFileContent.WriteString(EnTranslationInit)
			default:
				// 默认使用中文
				validationFileContent.WriteString(ZhTranslationInit)
			}
		}

		validationFileContent.WriteString("}\n\n")

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

		// 如果启用了翻译，添加翻译错误信息的函数
		if options.EnableTranslation {
			validationFileContent.WriteString(TranslateErrorFunc)
		}

		newValidationContent = validationFileContent.String()
	} else {
		// 文件已存在，需要更新
		if options.DebugMode {
			fmt.Println("验证文件已存在，需要更新")
		}

		// 初始化newValidationContent为当前内容
		newValidationContent = validationContent

		// 分析现有的验证方法映射和验证函数
		mapRegex := regexp.MustCompile(`(?s)var\s+registerValidation\s*=\s*map\[string\]validator\.Func\{[\s\S]*?\}`)
		mapMatch := mapRegex.FindString(validationContent)

		if options.DebugMode {
			fmt.Println("找到的验证方法映射:", len(mapMatch) > 0)
		}

		// 提取现有的自定义验证函数
		funcRegex := regexp.MustCompile(`(?s)func\s+validate(\w+)\s*\(\s*fl\s+validator\.FieldLevel\s*\)\s*bool\s*\{`)
		funcMatches := funcRegex.FindAllStringSubmatch(validationContent, -1)

		existingValidations := make(map[string]bool)
		existingTags := make(map[string]bool)

		// 提取现有标签
		tagRegex := regexp.MustCompile(`"(\w+)":\s*validate\w+`)
		tagMatches := tagRegex.FindAllStringSubmatch(validationContent, -1)

		for _, match := range tagMatches {
			if len(match) > 1 {
				tag := match[1]
				existingTags[tag] = true
				if options.DebugMode {
					fmt.Println("找到已注册的标签:", tag)
				}
			}
		}

		for _, match := range funcMatches {
			if len(match) > 1 {
				funcName := match[1]
				funcNameLower := strings.ToLower(funcName)

				// 跳过内置验证函数
				if funcNameLower == "mobile" || funcNameLower == "idcard" {
					continue
				}

				existingValidations[funcNameLower] = true

				// 同时将验证函数名添加到customTags中，确保其被保留
				customTags[funcNameLower] = true

				if options.DebugMode {
					fmt.Println("找到已有验证函数:", funcNameLower)
				}
			}
		}

		// 同时确保已注册的标签也被添加到customTags中
		for tag := range existingTags {
			// 跳过内置验证标签
			if tag == "mobile" || tag == "idcard" {
				continue
			}
			customTags[tag] = true
		}

		// 检查自定义标签是否已添加到验证方法映射中
		var tagsToAppend []string
		for tag := range customTags {
			// 跳过内置和已经存在的标签
			if tag == "mobile" || tag == "idcard" || existingTags[tag] {
				continue
			}
			tagsToAppend = append(tagsToAppend, tag)
		}

		// 处理验证映射的更新
		if len(tagsToAppend) > 0 {
			if options.DebugMode {
				fmt.Println("需要追加的标签:", tagsToAppend)
			}
			// 在映射的结尾前追加新标签
			closeBraceIndex := strings.LastIndex(mapMatch, "}")
			if closeBraceIndex > 0 {
				mapContentBeforeBrace := mapMatch[:closeBraceIndex]

				if options.DebugMode {
					fmt.Println("在位置", closeBraceIndex, "处追加标签")
				}

				// 添加新的标签
				var newTagsContent strings.Builder
				// 检查是否需要添加逗号
				if !strings.HasSuffix(strings.TrimSpace(mapContentBeforeBrace), ",") {
					newTagsContent.WriteString(",")
				}
				newTagsContent.WriteString("\n")

				for _, tag := range tagsToAppend {
					newTagsContent.WriteString(fmt.Sprintf("\t\"%s\": validate%s,\n", tag, strings.Title(tag)))
				}

				// 重建映射内容
				updatedMapContent := mapContentBeforeBrace + newTagsContent.String() + "}"

				// 替换原有的映射
				newValidationContent = strings.Replace(validationContent, mapMatch, updatedMapContent, 1)

				if options.DebugMode {
					fmt.Println("更新后的映射内容长度:", len(updatedMapContent))
					fmt.Println("新内容中是否包含uuid2:", strings.Contains(newValidationContent, "uuid2"))
				}
			} else {
				// 如果无法找到闭合括号，可能需要完全重建映射
				if options.DebugMode {
					fmt.Println("无法找到闭合括号，使用替代方法")
				}

				// 构建一个全新的验证方法映射
				var newMapContent strings.Builder
				newMapContent.WriteString("var registerValidation = map[string]validator.Func{\n")

				// 添加内置验证
				newMapContent.WriteString("\t\"mobile\": validateMobile,\n")
				newMapContent.WriteString("\t\"idcard\": validateIdCard,\n")

				// 添加所有自定义验证
				var tags []string
				for tag := range customTags {
					if tag != "mobile" && tag != "idcard" {
						tags = append(tags, tag)
					}
				}

				// 按字母排序，保证输出稳定
				sort.Strings(tags)

				// 添加自定义标签
				for i, tag := range tags {
					comma := ","
					if i == len(tags)-1 {
						comma = ""
					}
					newMapContent.WriteString(fmt.Sprintf("\t\"%s\": validate%s%s\n", tag, strings.Title(tag), comma))
				}

				newMapContent.WriteString("}")

				// 使用正则表达式替换整个映射
				newValidationContent = mapRegex.ReplaceAllString(validationContent, newMapContent.String())

				if options.DebugMode {
					fmt.Println("完全重建映射:", newMapContent.String())
				}
			}
		}

		// 添加缺失的验证函数到文件末尾
		var missingFuncs []string
		for tag := range customTags {
			// 跳过内置验证标签和已经存在的验证函数
			if tag == "mobile" || tag == "idcard" || existingValidations[tag] {
				continue
			}

			// 为不存在的验证函数生成代码
			funcCode := fmt.Sprintf(CustomValidationFuncTemplate, tag, strings.Title(tag), tag)
			missingFuncs = append(missingFuncs, funcCode)
		}

		// 如果有缺失的验证函数，添加到文件末尾
		if len(missingFuncs) > 0 {
			for _, funcCode := range missingFuncs {
				newValidationContent += "\n" + funcCode
			}

			if options.DebugMode {
				fmt.Println("添加了", len(missingFuncs), "个缺失的验证函数")
			}
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

		// 不在types.go中添加验证器变量的声明，因为validation.go中已经有了
	}

	for _, structName := range reqStructs {
		// 检查是否已经存在该结构体的Validate方法
		if !strings.Contains(string(fileContent), "func (r *"+structName+") Validate()") {
			methodsBuilder.WriteString(fmt.Sprintf("\nfunc (r *%s) Validate() error {\n\treturn validate.Struct(r)\n}\n", structName))
		}
	}

	// 将方法添加到types.go文件末尾
	if methodsBuilder.Len() > 0 {
		// 检查文件中是否已经包含validate变量的声明
		if !strings.Contains(string(fileContent), "var validate = validator.New()") {
			// 添加validate变量声明
			// 找到import语句结束的位置
			importEndPos := strings.Index(string(fileContent), ")")
			if importEndPos > 0 {
				// 在import语句后添加validate变量声明
				importEndPos = importEndPos + 1
				modifiedContent := string(fileContent[:importEndPos]) + "\n\nvar validate = validator.New()" + string(fileContent[importEndPos:]) + methodsBuilder.String()

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
					fmt.Printf("成功添加验证方法和validate变量到 %s\n", filePath)
				}
			} else {
				// 找不到import语句，直接添加在文件末尾
				modifiedContent := string(fileContent) + "\n\nvar validate = validator.New()" + methodsBuilder.String()

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
					fmt.Printf("成功添加验证方法和validate变量到 %s\n", filePath)
				}
			}
		} else {
			// 已存在validate变量，只添加验证方法
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
	}

	// 格式化并写入文件
	formatted, err := format.Source([]byte(newValidationContent))
	if err != nil {
		if options.DebugMode {
			fmt.Println("格式化代码失败:", err)
			fmt.Println("代码内容长度:", len(newValidationContent))
			fmt.Println("代码内容前100个字符:", newValidationContent[:min(100, len(newValidationContent))])
			// 仍然尝试写入未格式化的内容
			if err := os.WriteFile(validationFilePath, []byte(newValidationContent), 0644); err != nil {
				fmt.Println("写入未格式化的验证文件失败:", err)
				return fmt.Errorf("写入未格式化的验证文件失败: %w", err)
			}
			fmt.Println("已写入未格式化的验证文件")
		}
		return fmt.Errorf("格式化更新的验证文件代码失败: %w", err)
	}

	if err := os.WriteFile(validationFilePath, formatted, 0644); err != nil {
		if options.DebugMode {
			fmt.Println("写入验证文件失败:", err)
		}
		return fmt.Errorf("写入更新的验证文件失败: %w", err)
	}

	if options.DebugMode {
		fmt.Printf("成功更新验证文件: %s (大小: %d字节)\n", validationFilePath, len(formatted))
	}

	return nil
}

// min returns the smaller of x or y.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
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

	// mobile和idcard是我们自定义的，但它们是内置在插件中的验证器
	builtInCustomValidators := map[string]bool{
		"mobile": true,
		"idcard": true,
	}

	if builtInCustomValidators[validator] {
		return true
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

// 这是新添加的函数，用于生成translations.go文件
func generateTranslationsFile(dirPath, packageName string, options Options) error {
	// 翻译文件的路径（与types.go在同一目录）
	translationsFilePath := filepath.Join(dirPath, "translations.go")

	// 检查translations.go文件是否已存在
	if _, err := os.Stat(translationsFilePath); err == nil {
		// 文件已存在，跳过生成
		return nil
	}

	// 创建translations.go文件内容
	var translationsFileContent strings.Builder

	// 添加包声明和导入
	translationsFileContent.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	translationsFileContent.WriteString("import (\n")
	translationsFileContent.WriteString("\t" + ValidateImport + "\n")
	translationsFileContent.WriteString("\t" + UniversalTranslatorImport + "\n")
	translationsFileContent.WriteString(")\n\n")

	// 添加自定义翻译注册函数
	translationsFileContent.WriteString(CustomTranslationsFunc)

	// 格式化和写入文件
	formatted, err := format.Source([]byte(translationsFileContent.String()))
	if err != nil {
		return fmt.Errorf("格式化翻译文件代码失败: %w", err)
	}

	if err := os.WriteFile(translationsFilePath, formatted, 0644); err != nil {
		return fmt.Errorf("写入翻译文件失败: %w", err)
	}

	if options.DebugMode {
		fmt.Printf("成功创建翻译文件: %s\n", translationsFilePath)
	}

	return nil
}
