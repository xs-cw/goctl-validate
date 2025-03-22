package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"regexp"
)

var validate = validator.New()

// 自定义验证方法示例
func customValidation(fl validator.FieldLevel) bool {
	// 在这里实现自定义验证逻辑
	return true
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

type LoginReq struct {
	Mobile   string `json:"mobile" validate:"required,mobile"`
	Password string `json:"password" validate:"required,min=6,max=20"`
}

type UserInfoReq struct {
	IDCard string `json:"id_card" validate:"required,idcard"`
	Name   string `json:"name" validate:"required"`
}

func init() {
	// 注册自定义验证方法
	validate.RegisterValidation("custom_validation", customValidation)
	// 注册手机号验证方法
	validate.RegisterValidation("mobile", validateMobile)
	// 注册身份证号验证方法
	validate.RegisterValidation("idcard", validateIdCard)
}

func (r *LoginReq) Validate() error {
	return validate.Struct(r)
}

func (r *UserInfoReq) Validate() error {
	return validate.Struct(r)
}

func main() {
	// 测试手机号验证
	loginReq := &LoginReq{
		Mobile:   "13800138000", // 有效手机号
		Password: "123456",
	}
	err := loginReq.Validate()
	fmt.Println("验证有效手机号:", err)

	// 测试无效手机号
	loginReq.Mobile = "1234567890" // 无效手机号
	err = loginReq.Validate()
	fmt.Println("验证无效手机号:", err)

	// 测试身份证号验证
	userInfo := &UserInfoReq{
		IDCard: "11010119900307123X", // 有效身份证号
		Name:   "测试用户",
	}
	err = userInfo.Validate()
	fmt.Println("验证有效身份证号:", err)

	// 测试无效身份证号
	userInfo.IDCard = "1234567890" // 无效身份证号
	err = userInfo.Validate()
	fmt.Println("验证无效身份证号:", err)
} 