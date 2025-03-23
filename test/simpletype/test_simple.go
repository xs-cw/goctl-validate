package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"regexp"
)

// ValidateError 表示简单的验证错误
type ValidateError struct {
	Msg       string      // 错误消息
	FieldName string      // 字段名
	Tag       string      // 验证标签
	Value     interface{} // 字段值
	RawError  error       // 原始错误
}

// Error 实现error接口
func (e *ValidateError) Error() string {
	return e.Msg
}

// Code 返回错误码（0表示无错误）
func (e *ValidateError) Code() int {
	return 10001
}

// Field 返回字段名
func (e *ValidateError) Field() string {
	return e.FieldName
}

// UserInfo 用户信息
type UserInfo struct {
	Name   string `json:"name" validate:"required"`
	Mobile string `json:"mobile" validate:"required,mobile"`
	IdCard string `json:"id_card" validate:"required,idcard"`
	Email  string `json:"email" validate:"required,email"`
	Age    int    `json:"age" validate:"required,gte=18,lte=120"`
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

// Validate 验证结构体字段
func (r *UserInfo) Validate() error {
	validate := validator.New()
	// 注册验证方法
	_ = validate.RegisterValidation("mobile", validateMobile)
	_ = validate.RegisterValidation("idcard", validateIdCard)
	
	err := validate.Struct(r)
	if err == nil {
		return nil
	}
	
	// 转换为ValidationErrors类型
	if fieldErrors, ok := err.(validator.ValidationErrors); ok && len(fieldErrors) > 0 {
		fieldError := fieldErrors[0]
		message := fieldError.Error()
		
		return &ValidateError{
			FieldName: fieldError.Field(),
			Tag:       fieldError.Tag(),
			Value:     fieldError.Value(),
			Msg:       message,
			RawError:  fieldError,
		}
	}
	
	return err
}

func main() {
	// 创建一个无效的用户信息对象
	userInfo := UserInfo{
		Name:   "",
		Mobile: "123",
		IdCard: "12345",
		Email:  "invalid-email",
		Age:    10,
	}

	// 调用验证方法
	err := userInfo.Validate()
	if err != nil {
		// 使用类型断言判断是否为简化的错误类型
		if validateErr, ok := err.(*ValidateError); ok {
			// 获取错误码
			fmt.Println("错误码:", validateErr.Code())
			// 获取错误信息
			fmt.Println("错误信息:", validateErr.Error())
			// 获取字段名
			fmt.Println("字段名:", validateErr.Field())
		} else {
			// 如果不是简化错误类型，直接打印错误
			fmt.Println("验证错误:", err)
		}
	}

	// 创建一个有效的用户信息对象
	validUserInfo := UserInfo{
		Name:   "张三",
		Mobile: "13800138000",
		IdCard: "110101199001011234",
		Email:  "test@example.com",
		Age:    25,
	}

	// 验证有效用户
	err = validUserInfo.Validate()
	if err != nil {
		fmt.Println("验证错误:", err)
	} else {
		fmt.Println("验证通过!")
	}
} 