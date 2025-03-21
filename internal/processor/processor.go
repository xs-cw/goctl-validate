package processor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
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
	ValidateImport   = `"github.com/go-playground/validator/v10"`
	ValidateVar      = `var validate = validator.New()`
	ValidateInitFunc = `
func init() {
	// 注册自定义验证方法
	validate.RegisterValidation("custom_validation", customValidation)
}

// 自定义验证方法示例
func customValidation(fl validator.FieldLevel) bool {
	// 在这里实现自定义验证逻辑
	return true
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
	existValidatorImport := false
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

	// 生成验证方法
	var methods strings.Builder
	for _, structName := range reqStructs {
		// 检查是否已经存在该结构体的Validate方法
		if !strings.Contains(string(fileContent), "func (r *"+structName+") Validate()") {
			methods.WriteString("\nfunc (r *")
			methods.WriteString(structName)
			methods.WriteString(") Validate() error {\n\treturn validate.Struct(r)\n}\n")
		}
	}

	// 准备导入和变量声明
	var importAndVar strings.Builder
	if !existValidatorImport {
		importAndVar.WriteString("\nimport (\n\t")
		importAndVar.WriteString(ValidateImport)
		importAndVar.WriteString("\n)\n\n")
	}

	importAndVar.WriteString(ValidateVar)
	importAndVar.WriteString("\n")

	// 添加自定义验证方法
	if options.EnableCustomValidation && !existCustomValidation {
		importAndVar.WriteString(ValidateInitFunc)
	}

	// 构建新的文件内容 - 将导入和变量添加到package声明之后
	// 这里分为三部分：1. package前的内容 2. package声明 3. package后的内容
	content := string(fileContent)

	// 首先找到package行结束的位置
	packageEndPos := strings.Index(content[packagePos:], "\n")
	if packageEndPos == -1 {
		return fmt.Errorf("无法确定package声明结束位置")
	}
	packageEndPos += packagePos + 1 // +1 是为了包含换行符

	// 构建最终内容
	modifiedContent := content[:packageEndPos] +
		importAndVar.String() +
		content[packageEndPos:]

	// 如果有需要添加的验证方法，则添加到文件末尾
	if methods.Len() > 0 {
		modifiedContent += methods.String()
	}

	// 格式化代码
	formatted, err := format.Source([]byte(modifiedContent))
	if err != nil {
		if options.DebugMode {
			fmt.Println("============= 修改后文件内容（格式化前） =============")
			fmt.Println(modifiedContent)
			fmt.Println("===================================================")
			fmt.Printf("格式化错误: %v\n", err)
		}
		return fmt.Errorf("格式化代码失败: %w,%+v", err, modifiedContent)
	}

	// 在调试模式下打印格式化后的内容
	if options.DebugMode {
		fmt.Println("============= 格式化后文件内容 =============")
		fmt.Println(string(formatted))
		fmt.Println("==========================================")
	}

	// 写回文件
	if err := os.WriteFile(filePath, formatted, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
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
