package translation_test

import (
	"fmt"

	"github.com/linabellbiu/goctl-validate/test/manual_test/internal/types"
)

func main() {
	// 创建一个故意不符合验证规则的请求
	req := &types.TestTranslationReq{
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
		translatedErr := types.TranslateError(err)
		fmt.Println(translatedErr)
	}
}
