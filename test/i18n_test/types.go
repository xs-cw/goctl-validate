package types

// 测试翻译器功能
type UserForm struct {
	Name     string  `json:"name" validate:"required"`                  // 用户名
	Mobile   string  `json:"mobile" validate:"required,mobile"`         // 手机号
	IdCard   string  `json:"id_card" validate:"required,idcard"`        // 身份证号
	Email    string  `json:"email" validate:"required,email"`           // 邮箱
	Age      int     `json:"age" validate:"required,min=18,max=120"`    // 年龄
	Weight   float64 `json:"weight" validate:"required,min=30,max=200"` // 体重
	Website  string  `json:"website" validate:"required,url"`           // 网站
	IP       string  `json:"ip" validate:"required,ip"`                 // IP地址
	Password string  `json:"password" validate:"required,min=6,max=20"` // 密码
	Username string  `json:"username" validate:"required,alpha"`        // 用户名
	Nickname string  `json:"nickname" validate:"required,alphanum"`     // 昵称
}
