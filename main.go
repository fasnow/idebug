package main

import (
	"fmt"
	"idebug/cmd"
	"idebug/logger"
	"idebug/prompt"
)

func main() {
	client := prompt.Client{}
	cmd.Banner()
	version, releaseUrl, publishTime, content := cmd.CheckUpdate()
	if version != "" {
		s := fmt.Sprintf("最新版本: %s %s", version, publishTime)
		s += "\n    下载地址: " + releaseUrl
		s += "\n    更新内容: " + content
		logger.Notice(s)
	}
	client.Run()
}
