package types

import (
	"fmt"
)

// 演示如何使用自定义错误类型和错误码
func ExampleUseErrorCodes() {
	// 创建一个有错误的用户信息
	user := &UserInfo{
		Name:   "",             // 必填字段缺失
		Mobile: "123",          // 手机号格式错误
		IdCard: "12345",        // 身份证号格式错误
		Email:  "invalid-email", // 邮箱格式错误
		Age:    0,              // 必填字段缺失
	}

	// 验证结构体
	err := user.Validate()
	
	// 使用错误类型断言
	if err != nil {
		// 断言错误类型为ValidationErrors
		if validationErrs, ok := err.(*ValidationErrors); ok {
			// 获取第一个错误的错误码
			errCode := validationErrs.FirstCode()
			fmt.Printf("错误码: %d\n", errCode)
			
			// 获取错误信息
			fmt.Printf("错误信息: %s\n", validationErrs.Error())
			
			// 获取第一个错误
			firstErr := validationErrs.First()
			if firstErr != nil {
				fmt.Printf("字段: %s, 标签: %s, 错误码: %d\n", 
					firstErr.Field, firstErr.Tag, firstErr.Code)
			}
			
			// 遍历所有错误
			fmt.Println("所有错误:")
			for i, err := range validationErrs.Errors {
				fmt.Printf("%d. 字段: %s, 错误码: %d, 消息: %s\n", 
					i+1, err.Field, err.Code, err.Message)
			}
			
			// 根据错误码进行不同处理
			switch errCode {
			case ErrCodeRequired:
				fmt.Println("处理必填字段缺失的情况")
			case ErrCodeMobile:
				fmt.Println("处理手机号格式错误的情况")
			case ErrCodeEmail:
				fmt.Println("处理邮箱格式错误的情况")
			default:
				fmt.Println("处理其他错误情况")
			}
		} else {
			// 处理其他类型的错误
			fmt.Printf("其他错误: %v\n", err)
		}
	}
}

// 演示如何在API处理函数中使用错误类型
func ExampleUseInAPIHandler() {
	// 假设这是API处理函数
	handler := func(user *UserInfo) (int, string) {
		// 验证请求参数
		err := user.Validate()
		if err == nil {
			// 验证通过，处理业务逻辑
			return 200, "处理成功"
		}
		
		// 验证失败，返回错误码和错误信息
		if validationErrs, ok := err.(*ValidationErrors); ok {
			// 获取第一个错误的错误码
			errCode := validationErrs.FirstCode()
			
			// 映射错误码到HTTP状态码和错误消息
			switch errCode {
			case ErrCodeRequired:
				return 400, "缺少必填参数"
			case ErrCodeMobile:
				return 400, "手机号格式错误"
			case ErrCodeEmail:
				return 400, "邮箱格式错误"
			case ErrCodeIDCard:
				return 400, "身份证号格式错误"
			default:
				return 400, "参数格式错误"
			}
		}
		
		// 其他错误
		return 500, "服务器错误"
	}
	
	// 测试不同的参数
	cases := []*UserInfo{
		// 有效参数
		{
			Name:   "张三",
			Mobile: "13812345678",
			IdCard: "110101199001011234",
			Email:  "test@example.com",
			Age:    30,
		},
		// 缺少必填参数
		{
			Name:   "",
			Mobile: "13812345678",
			IdCard: "110101199001011234",
			Email:  "test@example.com",
			Age:    30,
		},
		// 手机号格式错误
		{
			Name:   "张三",
			Mobile: "123",
			IdCard: "110101199001011234",
			Email:  "test@example.com",
			Age:    30,
		},
	}
	
	// 测试不同参数的处理结果
	for i, c := range cases {
		code, msg := handler(c)
		fmt.Printf("Case %d: 状态码=%d, 消息=%s\n", i+1, code, msg)
	}
} 