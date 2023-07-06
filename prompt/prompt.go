package prompt

import (
	"fmt"
	"github.com/fasnow/readline"
	"github.com/spf13/cobra"
	"idebug/cmd"
	"idebug/logger"
	"idebug/utils"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var globalCmd = []string{"clear", "cls", "use", "update", "exit"}

type Client struct {
	module *cmd.Module
	proxy  *string
	//ppt        *prompt.Prompt
}

func (client *Client) Run() {
	client.proxy = new(string)
	client.module = (*cmd.Module)(new(string))
	*client.proxy = ""
	cmd.Proxy = client.proxy
	*client.module = cmd.NoModule
	cmd.CurrentModule = client.module
	line, err := readline.NewEx(&readline.Config{})
	if err != nil {
		logger.Error(logger.FormatError(err))
		return
	}
	line.Config.Stdout = logger.Writer
	line.HistoryEnable()
	go client.ctrlCListener()
	for {
		line.SetPrompt(logger.ModuleSelectedV2(string(*client.module)))
		input, err := line.Readline()
		if err != nil {
			if err.Error() != "Interrupt" {
				logger.Error(logger.FormatError(err))
			}
			continue
		}
		client.executor(input)
	}
}

func (client *Client) executor(in string) {
	cmd.SetContext()
	go func() {
		client.exec(in)
		cmd.Cancel()
	}()
	for {
		select {
		case <-cmd.Context.Done():
			return
		}
	}
}

func (client *Client) exec(in string) {
	var moduleMap = map[cmd.Module]*cobra.Command{
		cmd.WxModule:     cmd.NewWechatCli().Root,
		cmd.FeiShuModule: cmd.NewFeiShuCli().Root,
		cmd.NoModule:     cmd.NewMainCli().Root,
	}
	in = strings.TrimSpace(in)
	args := strings.Fields(in)
	if len(args) == 0 {
		return
	}
	var cmdFunc *cobra.Command
	var ok bool
	if utils.StringInList(args[0], globalCmd) {
		cmdFunc = moduleMap[cmd.NoModule]
	} else {
		if cmdFunc, ok = moduleMap[*client.module]; !ok {
			logger.Error(logger.FormatError(fmt.Errorf("模块错误")))
			return
		}
	}
	cmdFunc.SetArgs(args)
	if err := cmdFunc.Execute(); err != nil {
		logger.Error(err)
	}
	return
}

func (client *Client) ctrlCListener() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		for {
			<-interrupt  // 接收ctrl+c中断信号
			cmd.Cancel() // 取消http上下文
			if !cmd.HttpCanceled {
				cmd.HttpCanceled = true
			}
		}
	}()
	select {}
}
