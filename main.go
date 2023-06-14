package main

import (
	"idebug/cmd"
	"idebug/prompt"
	"idebug/utils"
)

func main() {
	client := prompt.Client{}
	utils.Banner()
	cmd.ShowAllUsage()
	newVersion, releaseUrl, content := utils.CheckUpdate()
	if newVersion != "" {
		s := "最新版本: " + newVersion
		s += "\n    下载地址: " + releaseUrl
		s += "\n    更新内容: " + content
		utils.Notice(s)
	}
	client.Run()
}
