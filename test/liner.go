package main

import (
	"fmt"
	"github.com/chzyer/readline"
	"idebug/logger"
	"os"
	"strings"
)

func main() {
	line, _ := readline.NewEx(&readline.Config{})
	line.Config.Stdout = logger.Writer
	line.HistoryEnable()
	for {
		line.SetPrompt(logger.ModuleSelectedV2("feishu"))
		input, err := line.Readline()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to read input:", err)
			return
		}

		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}

		if input != "" {
			fmt.Println("You entered:", input)
		}
	}
}
