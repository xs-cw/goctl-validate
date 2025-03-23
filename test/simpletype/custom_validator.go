package types

import (
	"github.com/go-playground/validator/v10"
)

// 自定义验证方法示例 - 年龄范围验证
func validateAgeRange(fl validator.FieldLevel) bool {
	age := fl.Field().Int()
	// 验证年龄是否在18-120岁之间
	return age >= 18 && age <= 120
} 