package types

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

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
