package types

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

var validate = validator.New()

func init() {
	// 注册手机号验证方法
	validate.RegisterValidation("mobile", validateMobile)
	// 注册身份证号验证方法
	validate.RegisterValidation("idcard", validateIdCard)
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
