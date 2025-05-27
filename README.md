# goctl-validate
一个go-zero API插件，使用go-playground/validator无缝接入go-zreo。

# 原理解析
go-zreo参数验证的时候调用了validatioin/Validator 接口,只要实现了就会自动调用，这就是为啥我们只要加了 func (r *Request) Validate() error，就会自动验证 Request 对象了.超级方便!!!
```go
// Parse parses the request.
func Parse(r *http.Request, v any) error {
	...
	if valid, ok := v.(validation.Validator); ok {
		return valid.Validate()
	} else if val := getValidator(); val != nil {
		return val.Validate(r, v)
	}
	return nil
}
```
## 功能

- 自动为API中定义的请求结构体（名称以`Req`结尾）添加`Validate()`方法
- 添加`go-playground/validator/v10`依赖及初始化代码
- 支持多个请求结构体
- 支持自定义验证方法（通过`--custom`标志启用）
- 支持调试模式（通过`--debug`标志启用）
- 支持验证错误翻译器（通过`--translator`标志启用，默认为中文）
- 智能处理生成的types.go文件，保持正确的包声明位置


## 安装

```bash
go install github.com/xs-cw/goctl-validate@latest
```

## 使用方法

1. 首先使用goctl生成API代码：

```bash
goctl api go -api your_api.api -dir .
```

2. 然后使用goctl-validate插件来添加验证逻辑：

```bash
# 基本用法
goctl api plugin -p goctl-validate="validate" --api your_api.api --dir .

# 启用自定义验证方法
goctl api plugin -p goctl-validate="validate --custom" --api your_api.api --dir .

# 启用验证错误翻译器（默认中文）
goctl api plugin -p goctl-validate="validate --translator" --api your_api.api --dir .

# 启用调试模式（用于排查问题）
goctl api plugin -p goctl-validate="validate --debug" --api your_api.api --dir .

# 同时启用多个选项
goctl api plugin -p goctl-validate="validate --custom --translator --debug" --api your_api.api --dir .
```


## 示例

假设您有以下API定义：

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

service example-api {
    @handler GetStatus
    get /status (StatusReq) returns (CommonResp)
    
    @handler CreateItem
    post /items (CreateItemReq) returns (CommonResp)
}
```

运行goctl-validate插件后，会在生成的`types.go`文件中添加以下代码：

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

如果使用了`--custom`标志，还会生成一个`validation.go`文件，添加自定义验证方法：

```go
// validation.go
package types

import (
    "regexp"
    "github.com/go-playground/validator/v10"
)

// registerValidation 存储所有的验证方法
// key: 验证标签名称，value: 对应的验证函数
var registerValidation = map[string]validator.Func{
    "mobile": validateMobile, // 手机号验证
    "idcard": validateIdCard, // 身份证号验证
}

// 初始化并注册所有验证方法
func init() {
    // 遍历注册所有验证方法
    for tag, handler := range registerValidation {
        _ = validate.RegisterValidation(tag, handler)
    }
}

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
```

如果使用了`--translator`标志，还会生成一个`translator.go`文件，添加验证错误翻译功能：

```go
// translator.go
package types

import (
    "errors"
    "fmt"
    "github.com/go-playground/validator/v10"
    "github.com/go-playground/locales/en"
    "github.com/go-playground/locales/zh"
    ut "github.com/go-playground/universal-translator"
    zhTrans "github.com/go-playground/validator/v10/translations/zh"
)

var (
	uni   *ut.UniversalTranslator
	trans ut.Translator
)

// 初始化翻译器
func init() {
    // 创建中文翻译器
    zhLoc := zh.New()
    enLoc := en.New()
    
    // 创建通用翻译器并设置为中文
    uni := ut.New(enLoc, zhLoc)
    trans, _ = uni.GetTranslator("zh")
    
    // 注册默认的中文翻译
    _ = zhTrans.RegisterDefaultTranslations(validate, trans)
    
    // 注册自定义验证标签的翻译
    registerCustomTranslations(validate, trans)
}

// Translate 翻译验证错误
func Translate(err error) error {
	if err == nil {
		return nil
	}

	var errs validator.ValidationErrors
	if ok := errors.As(err, &errs); !ok {
		return err
	}

	var errMsgs []string
	for _, e := range errs {
		translatedErr := e.Translate(trans)
		errMsgs = append(errMsgs, translatedErr)
	}
	// TODO 可以自定义错误类型
	return errors.New(strings.Join(errMsgs, ", "))
}


// registerCustomTranslations 注册自定义验证标签的翻译
func registerCustomTranslations(validate *validator.Validate, trans ut.Translator) {
    // 注册手机号验证的翻译
    _ = validate.RegisterTranslation("mobile", trans, func(ut ut.Translator) error {
        return ut.Add("mobile", "{0}必须是有效的手机号码", true)
    }, func(ut ut.Translator, fe validator.FieldError) string {
        t, _ := ut.T("mobile", fe.Field())
        return t
    })

    // 注册身份证验证的翻译
    _ = validate.RegisterTranslation("idcard", trans, func(ut ut.Translator) error {
        return ut.Add("idcard", "{0}必须是有效的身份证号码", true)
    }, func(ut ut.Translator, fe validator.FieldError) string {
        t, _ := ut.T("idcard", fe.Field())
        return t
    })
}
```

常用验证标签：

| 标签 | 描述 | 示例 |
|-----|------|------|
| required | 必填字段 | `validate:"required"` |
| email | 必须是有效的电子邮件 | `validate:"email"` |
| min | 最小长度/值 | `validate:"min=4"` |
| max | 最大长度/值 | `validate:"max=100"` |
| len | 精确长度 | `validate:"len=8"` |
| eq | 等于 | `validate:"eq=10"` |
| ne | 不等于 | `validate:"ne=0"` |
| gt | 大于 | `validate:"gt=0"` |
| gte | 大于或等于 | `validate:"gte=18"` |
| lt | 小于 | `validate:"lt=100"` |
| lte | 小于或等于 | `validate:"lte=60"` |
| oneof | 枚举值 | `validate:"oneof=male female"` |
| numeric | 数字（整数或小数） | `validate:"numeric"` |
| alpha | 字母字符 | `validate:"alpha"` |
| alphanum | 字母数字字符 | `validate:"alphanum"` |
| mobile | 手机号验证（自定义） | `validate:"mobile"` |
| idcard | 身份证号验证（自定义） | `validate:"idcard"` |

有关可用验证标签的完整列表，请参阅[validator文档](https://pkg.go.dev/github.com/go-playground/validator/v10)。