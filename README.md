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
- 内置自定义验证方法（如手机号、身份证验证等）
- 内置验证错误信息的翻译功能（默认为中文）
- 支持选择翻译语言（通过`--lang`标志设置，支持`zh`、`zh_tw`和`en`）
- 智能处理生成的types.go文件，保持正确的包声明位置
- **将`validate`变量定义在`types.go`文件中，避免变量重复定义的问题**


## 安装

```bash
go install github.com/linabellbiu/goctl-validate@latest
```

## 使用方法

1. 首先使用goctl生成API代码：

```bash
goctl api go -api your_api.api -dir .
```

2. 然后使用goctl-validate插件来添加验证逻辑：

```bash
# 基本用法（默认启用所有功能）
goctl api plugin -p goctl-validate="validate" --api your_api.api --dir .

# 启用调试模式（用于排查问题）
goctl api plugin -p goctl-validate="validate --debug" --api your_api.api --dir .

# 指定翻译语言
goctl api plugin -p goctl-validate="validate --lang zh_tw" --api your_api.api --dir .

# 同时使用多个选项
goctl api plugin -p goctl-validate="validate --debug --lang en" --api your_api.api --dir .
```

## 目录结构

插件生成的代码将会按照以下结构组织：

- `types.go` - 包含API请求和响应结构体定义、**`validate`变量定义**和结构体验证方法
- `validation.go` - 包含验证逻辑、验证器注册和内置验证方法
- `translations.go` - 包含验证错误信息的翻译逻辑

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
    Mobile      string  `json:"mobile" validate:"mobile"` // 使用内置的手机号验证
}

service example-api {
    @handler GetStatus
    get /status (StatusReq) returns (CommonResp)
    
    @handler CreateItem
    post /items (CreateItemReq) returns (CommonResp)
}
```

运行goctl-validate插件后，会生成以下文件和代码：

1. 在`types.go`文件中：

```go
package types

import (
    "github.com/go-playground/validator/v10"
)

// validator实例定义在types.go中，避免重复定义问题
var validate = validator.New()

type StatusReq struct {
    Id   int64  `json:"id" validate:"required,gt=0"`
    Name string `json:"name" validate:"required"`
}

type CreateItemReq struct {
    Name        string  `json:"name" validate:"required,min=2,max=50"`
    Description string  `json:"description" validate:"omitempty,max=200"`
    Price       float64 `json:"price" validate:"required,gt=0"`
    Mobile      string  `json:"mobile" validate:"mobile"`
}

// 自动生成的验证方法
func (r *StatusReq) Validate() error {
    return validate.Struct(r)
}

func (r *CreateItemReq) Validate() error {
    return validate.Struct(r)
}
```

2. 在`validation.go`文件中会包含验证逻辑和内置的验证方法：

```go
package types

import (
    "regexp"
    "github.com/go-playground/validator/v10"
    // 其他导入...
)

// 验证方法映射
var registerValidation = map[string]validator.Func{
    "mobile": validateMobile, // 手机号验证
    "idcard": validateIdCard, // 身份证号验证
    // 其他自定义验证...
}

func init() {
    // 注册所有验证方法
    for tag, handler := range registerValidation {
        _ = validate.RegisterValidation(tag, handler)
    }
    
    // 初始化翻译器...
}

// 验证手机号
func validateMobile(fl validator.FieldLevel) bool {
    mobile := fl.Field().String()
    match, _ := regexp.MatchString("^1[3-9]\\d{9}$", mobile)
    return match
}

// 验证身份证号
func validateIdCard(fl validator.FieldLevel) bool {
    idCard := fl.Field().String()
    match, _ := regexp.MatchString("(^\\d{15}$)|(^\\d{18}$)|(^\\d{17}(\\d|X|x)$)", idCard)
    return match
}
```

3. 在`translations.go`文件中会包含验证错误信息的翻译逻辑。

## 验证标签

您可以在API结构体字段定义中使用validator的标签来定义验证规则：

```api
type CreateUserReq {
    Username string `json:"username" validate:"required,min=4,max=20"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"gte=18,lte=120"`
    Mobile   string `json:"mobile" validate:"mobile"` // 使用内置的手机号验证
    IdCard   string `json:"idCard" validate:"idcard"` // 使用内置的身份证验证
}
```

除了validator内置的验证标签外，本插件还提供了以下内置验证：

| 标签 | 描述 |
|-----|------|
| mobile | 中国大陆手机号验证 |
| idcard | 中国身份证号验证 |

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

有关可用验证标签的完整列表，请参阅[validator文档](https://pkg.go.dev/github.com/go-playground/validator/v10)。
