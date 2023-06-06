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
	client.Run()
}
