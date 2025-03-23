package types

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

// registerValidation 存储所有的验证方法
// key: 验证标签名称，value: 对应的验证函数
var registerValidation = map[string]validator.Func{
	"mobile":   validateMobile,   // 手机号验证
	"idcard":   validateIdCard,   // 身份证号验证
	"new_tag1": validateNew_tag1, // new_tag1
	"new_tag2": validateNew_tag2, // new_tag2
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

// 自定义验证方法: new_tag1
func validateNew_tag1(fl validator.FieldLevel) bool {
	// 在这里实现 new_tag1 的验证逻辑
	return true
}

// 自定义验证方法: new_tag2
func validateNew_tag2(fl validator.FieldLevel) bool {
	// 在这里实现 new_tag2 的验证逻辑
	return true
}
