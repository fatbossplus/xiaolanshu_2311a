package validate

import (
	"fmt"
	"regexp"
)

var PhoneValidator validator.Func = func(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	fmt.Println("手机号：", phone)
	// 验证手机号格式
	// 验证规则
	reg := `^1[3456789]\d{9}$`
	// 验证手机号是否符合规则
	if !regexp.MustCompile(reg).MatchString(phone) {
		return false
	}
	return true
}
