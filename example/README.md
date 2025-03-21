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

5. 在处理程序中使用验证方法：

```go
func (h *StatusHandler) Status(ctx *svc.ServiceContext, req *types.StatusReq) (resp *types.CommonResp, err error) {
    // 验证请求
    if err := req.Validate(); err != nil {
        return &types.CommonResp{
            Code:    400,
            Message: err.Error(),
        }, nil
    }
    
    // 处理业务逻辑...
    // ...
}
``` 