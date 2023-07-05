package prompt

import (
	"fmt"
	"github.com/fasnow/go-prompt"
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
	module     *cmd.Module
	proxy      *string
	ppt        *prompt.Prompt
	cancelChan chan bool
	doneChan   chan bool
}

func (client *Client) Run() {
	client.proxy = new(string)
	client.module = (*cmd.Module)(new(string))
	*client.proxy = ""
	cmd.Proxy = client.proxy
	*client.module = cmd.NoModule
	cmd.CurrentModule = client.module
	client.cancelChan = make(chan bool, 1)
	client.doneChan = make(chan bool, 1)

	//client.ppt = prompt.New(
	//	client.executor,
	//	completer,
	//	prompt.OptionPrefix(""),
	//	prompt.OptionAddKeyBind(prompt.KeyBind{
	//		Key: prompt.ControlC,
	//		Fn: func(buf *prompt.Buffer) {
	//			if runtime.GOOS != "windows" {
	//				cmd := exec.Command("reset")
	//				err := cmd.Run()
	//				if err != nil {
	//					logger.Error(logger.FormatError(err))
	//				}
	//			}
	//			os.Exit(0)
	//		},
	//	}),
	//)
	line, err := readline.NewEx(&readline.Config{})
	if err != nil {
		logger.Error(logger.FormatError(err))
		return
	}
	line.Config.Stdout = logger.Writer
	line.HistoryEnable()
	line.Config.InterruptPrompt = ""
	go client.listener()
	for {
		line.SetPrompt(logger.ModuleSelectedV2(string(*client.module)))
		input, err := line.Readline()
		if err != nil && err.Error() != "Interrupt" {
			logger.Error(logger.FormatError(err))
			continue
		}
		client.exec(input)
	}
}

func (client *Client) executor(in string) {
	go func() {
		client.exec(in)
		client.doneChan <- true
	}()
	select {
	case <-client.cancelChan:
		return
	case <-client.doneChan:
		return
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

func (client *Client) livePrefix() (string, bool) {
	return logger.ModuleSelectedV3(string(*client.module)), true
}

func (client *Client) listener() {
	// 创建一个通道来接收中断信号
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// 启动一个 goroutine 来监听中断信号
	go func() {
		for {
			<-interrupt  // 接收http中断信号
			cmd.Cancel() // 中断http请求
			client.cancelChan <- true
		}
	}()
	// 等待程序终止信号
	select {}
}
