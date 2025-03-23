package main

import (
	"fmt"
	
	"validator/types"
)

func main() {
	// 创建一个无效的用户信息对象
	userInfo := types.UserInfo{
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
		if validateErr, ok := err.(*types.ValidateError); ok {
			// 获取错误码
			fmt.Println("错误码:", validateErr.Code())
			// 获取错误信息
			fmt.Println("错误信息:", validateErr.Error())
			// 获取字段名
			fmt.Println("字段名:", validateErr.Field())
			// 获取中文错误信息
			fmt.Println("中文错误信息:", validateErr.Msg)
		} else {
			// 如果不是简化错误类型，直接打印错误
			fmt.Println("验证错误:", err)
		}
	}

	// 创建一个有效的用户信息对象
	validUserInfo := types.UserInfo{
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