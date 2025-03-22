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
- 智能处理生成的types.go文件，保持正确的包声明位置


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
# 基本用法
goctl api plugin -p goctl-validate="validate" --api your_api.api --dir .

# 启用自定义验证方法
goctl api plugin -p goctl-validate="validate --custom" --api your_api.api --dir .

# 启用调试模式（用于排查问题）
goctl api plugin -p goctl-validate="validate --debug" --api your_api.api --dir .

# 同时启用自定义验证和调试模式
goctl api plugin -p goctl-validate="validate --custom --debug" --api your_api.api --dir .
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

如果使用了`--custom`标志，还会添加以下代码：

```go
func init() {
    // 注册自定义验证方法
    validate.RegisterValidation("custom_validation", customValidation)
}

// 自定义验证方法示例
func customValidation(fl validator.FieldLevel) bool {
    // 在这里实现自定义验证逻辑
    return true
}
```

## 验证标签

您可以在API结构体字段定义中使用validator的标签来定义验证规则：

```api
type CreateUserReq {
    Username string `json:"username" validate:"required,min=4,max=20"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"gte=18,lte=120"`
    Profile  string `json:"profile" validate:"custom_validation"` // 使用自定义验证
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

有关可用验证标签的完整列表，请参阅[validator文档](https://pkg.go.dev/github.com/go-playground/validator/v10)。
