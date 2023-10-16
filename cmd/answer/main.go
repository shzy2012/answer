package main

import (
	answercmd "github.com/answerdev/answer/cmd"
	_ "github.com/answerdev/plugins/connector/github" // 新增代码
)

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	answercmd.Main()
}
