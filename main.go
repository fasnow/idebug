package main

import (
	"idebug/cmd"
	"idebug/logger"
	"idebug/prompt"
)

func main() {
	client := prompt.Client{}
	cmd.Banner()
	newVersion, releaseUrl, content := cmd.CheckUpdate()
	if newVersion != "" {
		s := "最新版本: " + newVersion
		s += "\n    下载地址: " + releaseUrl
		s += "\n    更新内容: " + content
		logger.Notice(s)
	}
	client.Run()
}
