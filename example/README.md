# goctl-validate 示例

本目录包含使用 goctl-validate 插件的示例。

## 步骤

1. 首先，我们有一个API定义文件 `example.api`，其中包含请求结构体定义：

```api
type StatusReq {
    Id   int64  `json:"id" validate:"required,gt=0"`
    Name string `json:"name" validate:"required"`
}

type CreateItemReq {
    Name        string  `json:"name" validate:"required,min=2,max=50"`
    Description string  `json:"description" validate:"omitempty,max=200"`
    Price       float64 `json:"price" validate:"required,gt=0"`
}
```

2. 使用 goctl 生成 API 代码：

```bash
goctl api go -api example.api -dir .
```

3. 使用 goctl-validate 插件添加验证逻辑：

```bash
# 基本用法
goctl api plugin -p goctl-validate="validate" --api example.api --dir .

# 添加自定义验证方法
goctl api plugin -p goctl-validate="validate --custom" --api example.api --dir .

# 添加验证错误信息翻译（默认中文）
goctl api plugin -p goctl-validate="validate --trans" --api example.api --dir .
```

4. 验证生成的 `types.go` 文件，确认已添加了验证逻辑：

```go
import (
    // 其他导入...
    "github.com/go-playground/validator/v10"
)

var validate = validator.New()

func (r *StatusReq) Validate() error {
    return validate.Struct(r)
}

func (r *CreateItemReq) Validate() error {
    return validate.Struct(r)
}
```

5. 如果启用了翻译功能，将看到额外的翻译相关代码：

```go
import (
    "regexp"
    "strings"
    "github.com/go-playground/validator/v10"
    "github.com/go-playground/locales/zh"
    ut "github.com/go-playground/universal-translator"
    zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var validate = validator.New()

// 全局翻译器
var (
    translator ut.Translator
    uni *ut.UniversalTranslator
)

// 初始化并注册所有验证方法
func init() {
    // 注册验证方法
    for tag, handler := range registerValidation {
        _ = validate.RegisterValidation(tag, handler)
    }
    
    // 中文翻译器初始化
    zh := zh.New()
    uni = ut.New(zh, zh)
    translator, _ = uni.GetTranslator("zh")
    // 注册翻译器
    _ = zhTranslations.RegisterDefaultTranslations(validate, translator)
}

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
```

6. 使用翻译功能的示例：

```go
func (l *StatusLogic) Status(req *types.StatusReq) (resp *types.StatusResp, err error) {
    // 验证请求参数
    if err := req.Validate(); err != nil {
        // 翻译错误信息为中文
        return nil, errors.New(types.TranslateError(err))
    }
    
    // 业务逻辑处理
    // ...
    
    return &types.StatusResp{
        Message: "success",
    }, nil
}
``` 