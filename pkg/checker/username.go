package checker

import (
	"fmt"
	"regexp"
)

var (
	// 检查昵称字符范围，包括字母、数字、下划线和表情符号
	// 使用\p{So} 支持表情
	// 使用\p{Han} 支持中文字符
	usernameReg = regexp.MustCompile(`^[a-zA-Z0-9_\-_\s'·.()（）\p{So}\p{Han}]+$`)
)

func IsInvalidUsername(username string) bool {
	fmt.Println("处理: ", username, " ", len(username))
	if len(username) < 2 || len(username) > 30 {
		return true
	}
	return !usernameReg.MatchString(username)
}
