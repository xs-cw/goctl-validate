package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var validate = validator.New()

// 全局翻译器
var (
	translator ut.Translator
	uni        *ut.UniversalTranslator
)

// 请求结构体
type TestTranslationReq struct {
	Name   string `json:"name" validate:"required,min=2,max=10"`
	Age    int    `json:"age" validate:"required,gt=0,lt=150"`
	Email  string `json:"email" validate:"required,email"`
	Mobile string `json:"mobile" validate:"required,mobile"`
}

// 验证方法
func (r *TestTranslationReq) Validate() error {
	return validate.Struct(r)
}

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

	// 中文翻译器初始化
	zh := zh.New()
	uni = ut.New(zh, zh)
	translator, _ = uni.GetTranslator("zh")
	// 注册翻译器
	_ = zhTranslations.RegisterDefaultTranslations(validate, translator)

	// 注册自定义验证器的翻译
	registerCustomValidationTranslations(translator)
}

// 注册自定义验证器的翻译
func registerCustomValidationTranslations(trans ut.Translator) {
	// 注册 mobile 验证器的翻译
	_ = trans.Add("mobile", "{0}必须是有效的手机号码", true)
	_ = validate.RegisterTranslation("mobile", trans, func(ut ut.Translator) error {
		return nil
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("mobile", fe.Field())
		return t
	})

	// 注册 idcard 验证器的翻译
	_ = trans.Add("idcard", "{0}必须是有效的身份证号码", true)
	_ = validate.RegisterTranslation("idcard", trans, func(ut ut.Translator) error {
		return nil
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("idcard", fe.Field())
		return t
	})
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

// TranslateError 翻译验证错误信息
func TranslateError(err error) string {
	if err == nil {
		return ""
	}
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}

	// 翻译每个错误
	var errMessages []string
	for _, e := range validationErrors {
		errMessages = append(errMessages, e.Translate(translator))
	}

	return strings.Join(errMessages, ", ")
}

func main() {
	// 创建一个故意不符合验证规则的请求
	req := &TestTranslationReq{
		Name:   "a",             // 不符合最小长度要求
		Age:    0,               // 不符合大于0的要求
		Email:  "invalid-email", // 不符合邮箱格式
		Mobile: "123",           // 不符合手机号格式
	}

	// 验证请求
	err := req.Validate()
	if err != nil {
		// 打印原始错误信息
		fmt.Println("原始错误信息:")
		fmt.Println(err.Error())

		fmt.Println("\n翻译后的错误信息:")
		// 翻译错误信息
		translatedErr := TranslateError(err)
		fmt.Println(translatedErr)
	}
}
