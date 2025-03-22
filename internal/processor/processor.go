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

	// 自定义验证方法部分（如果用户需要）
	CustomValidationPart = `	// 注册自定义验证方法
	validate.RegisterValidation("custom_validation", customValidation)
`

	// 初始化函数结束
	ValidateInitFuncEnd = `}
`

	// 自定义验证方法定义（如果用户需要）
	CustomValidationFunc = `
// 自定义验证方法示例
func customValidation(fl validator.FieldLevel) bool {
	// 在这里实现自定义验证逻辑
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

	// 如果启用了调试模式，打印文件内容
	if options.DebugMode {
		fmt.Println("============= 原始文件内容 =============")
		fmt.Println(string(fileContent))
		fmt.Println("=======================================")
	}

	// 检查文件内容中是否已经包含验证器代码，如果包含则直接返回
	if bytes.Contains(fileContent, []byte("var validate = validator.New()")) {
		if options.DebugMode {
			fmt.Println("文件已包含验证器代码，无需修改")
		}
		return nil
	}

	// 解析Go源代码
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, fileContent, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("解析文件失败: %w", err)
	}

	// 找到package声明的位置
	packagePos := findPackagePosition(string(fileContent))
	if packagePos == -1 {
		return fmt.Errorf("无法确定package位置")
	}

	// 寻找所有的请求结构体并生成验证方法
	var reqStructs []string
	// 定义变量，但不使用，防止编译错误
	existValidatorImport := false
	_ = existValidatorImport
	existCustomValidation := false

	// 检查imports
	for _, imp := range f.Imports {
		if imp.Path.Value == ValidateImport {
			existValidatorImport = true
			break
		}
	}

	// 检查是否已存在自定义验证方法
	if bytes.Contains(fileContent, []byte("func customValidation")) {
		existCustomValidation = true
	}

	// 收集所有请求结构体
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

			// 如果是结构体类型且名称以Req结尾
			if _, ok := typeSpec.Type.(*ast.StructType); ok && strings.HasSuffix(typeSpec.Name.Name, "Req") {
				reqStructs = append(reqStructs, typeSpec.Name.Name)
			}
		}
	}

	// 没有找到请求结构体，直接返回
	if len(reqStructs) == 0 {
		return nil
	}

	// 获取文件所在的目录路径
	dirPath := filepath.Dir(filePath)

	// 验证文件的路径（与types.go在同一目录）
	validationFilePath := filepath.Join(dirPath, "validation.go")

	// 检查验证文件是否已存在
	validationExists := false
	// 定义变量，但不使用，防止编译错误
	validationContent := ""
	_ = validationContent

	if _, err := os.Stat(validationFilePath); err == nil {
		// 验证文件已存在，读取内容
		validationBytes, err := os.ReadFile(validationFilePath)
		if err != nil {
			return fmt.Errorf("读取现有验证文件失败: %w", err)
		}
		validationContent = string(validationBytes)
		validationExists = true
	}

	// 获取包名
	packageName := f.Name.Name

	// 生成验证文件内容
	var validationFileContent strings.Builder

	// 如果文件不存在，添加包声明
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

		// 如果启用了自定义验证且不存在，添加自定义验证部分
		if options.EnableCustomValidation && !existCustomValidation {
			validationFileContent.WriteString(CustomValidationPart)
		}

		// 添加init函数结束
		validationFileContent.WriteString(ValidateInitFuncEnd + "\n")

		// 添加内置验证函数
		validationFileContent.WriteString(BuiltInValidationFunc + "\n")

		// 如果启用了自定义验证且不存在，添加自定义验证函数
		if options.EnableCustomValidation && !existCustomValidation {
			validationFileContent.WriteString(CustomValidationFunc + "\n")
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
