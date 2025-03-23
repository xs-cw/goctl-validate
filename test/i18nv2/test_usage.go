package types

import (
	"fmt"
)

// 演示如何使用带翻译器的验证
func ExampleUseValidateWithTranslator() {
	// 创建一个有错误的用户信息
	user := &UserInfo{
		Name:   "",                 // 缺少必填字段
		Mobile: "123",              // 不符合手机号格式
		IdCard: "12345",            // 不符合身份证格式
		Email:  "invalid-email",    // 不符合邮箱格式
		Age:    0,                  // 缺少必填字段
		Weight: 0,                  // 缺少必填字段
	}

	// 验证结构体
	err := user.Validate()
	
	// 使用辅助函数获取错误消息
	if err != nil {
		errMsg := GetValidateErrorMsg(err)
		fmt.Println("错误消息:", errMsg)
		
		// 您也可以直接使用原始错误进行错误类型判断和处理
		// 例如检查是否是验证错误
		if _, ok := err.(validator.ValidationErrors); ok {
			fmt.Println("这是一个验证错误")
		}
	}
}

// 演示如何使用中文翻译器错误消息
func ExampleHandleValidationErrors() {
	// 创建一个有错误的用户信息
	user := &UserInfo{
		Name:   "测试用户",
		Mobile: "123",           // 不符合手机号格式
		IdCard: "123456789012345678",  // 不符合身份证格式
		Email:  "user@example.com",
		Age:    20,
		Weight: 60,
	}

	// 验证结构体
	err := user.Validate()
	
	if err != nil {
		// 直接获取错误消息
		chineseMsg := GetValidateErrorMsg(err)
		fmt.Println("中文错误消息:", chineseMsg)
		
		// 或者手动处理ValidationErrors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// 输出所有的验证错误
			fmt.Println("所有验证错误:")
			for i, e := range validationErrors {
				fmt.Printf("%d. %s\n", i+1, e.Translate(translator))
			}
			
			// 检查是否是特定类型的错误
			for _, e := range validationErrors {
				if e.Tag() == "mobile" {
					fmt.Println("手机号验证失败:", e.Field())
				}
			}
		}
	}
} 