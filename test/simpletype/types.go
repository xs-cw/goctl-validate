package types

// 测试简化错误类型
type UserInfo struct {
	Name        string `json:"name" validate:"required"`                // 用户名
	Mobile      string `json:"mobile" validate:"required,mobile"`       // 手机号
	IdCard      string `json:"id_card" validate:"required,idcard"`      // 身份证号
	Email       string `json:"email" validate:"required,email"`         // 邮箱
	Age         int    `json:"age" validate:"required,ageRange"`        // 年龄范围验证
} 