package types

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	zhTrans "github.com/go-playground/validator/v10/translations/zh"
)

var (
	uni      *ut.UniversalTranslator
	trans    ut.Translator
	validate *validator.Validate
)

// 初始化翻译器
func init() {
	// 初始化翻译器
	enLocale := en.New()
	zhLocale := zh.New()
	uni = ut.New(enLocale, zhLocale)

	trans, _ = uni.GetTranslator("zh")
	validate = validator.New()

	// 注册默认翻译
	_ = zhTrans.RegisterDefaultTranslations(validate, trans)

	// 注册自定义翻译
	registerCustomTranslations(validate, trans)
}

// Translate 翻译验证错误
func Translate(err error) error {
	if err == nil {
		return nil
	}

	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	var errMsgs []string
	for _, e := range errs {
		translatedErr := e.Translate(trans)
		errMsgs = append(errMsgs, translatedErr)
	}
	// TODO 可以自定义错误类型
	return errors.New(strings.Join(errMsgs, ", "))
}

// 注册自定义翻译
func registerCustomTranslations(validate *validator.Validate, trans ut.Translator) {
	// 内置自定义验证器的翻译
	_ = trans.Add("mobile", "{0}手机号码格式不正确", false)
	_ = validate.RegisterTranslation("mobile", trans, func(ut ut.Translator) error {
		return nil
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("mobile", fe.Field())
		return t
	})

	_ = trans.Add("idcard", "{0}身份证号码格式不正确", false)
	_ = validate.RegisterTranslation("idcard", trans, func(ut ut.Translator) error {
		return nil
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("idcard", fe.Field())
		return t
	})

	_ = trans.Add("new_tag1", "{0}格式不符合要求", false)
	_ = validate.RegisterTranslation("new_tag1", trans, func(ut ut.Translator) error {
		return nil
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("new_tag1", fe.Field())
		return t
	})

	_ = trans.Add("new_tag2", "{0}格式不符合要求", false)
	_ = validate.RegisterTranslation("new_tag2", trans, func(ut ut.Translator) error {
		return nil
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("new_tag2", fe.Field())
		return t
	})
}
