package prompt

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"idebug/cmd"
	"idebug/logger"
	"idebug/utils"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Client struct {
	module *cmd.Module
	proxy  *string
	ppt    *prompt.Prompt
	prefix *string
}

func (client *Client) Run() {
	client.proxy = new(string)
	client.module = (*cmd.Module)(new(string))
	*client.proxy = ""
	cmd.Proxy = client.proxy
	*client.module = cmd.NoModule
	cmd.CurrentModule = client.module
	client.ppt = prompt.New(
		client.executor,
		completer,
		prompt.OptionPrefix(""),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(buf *prompt.Buffer) {
				if runtime.GOOS != "windows" {
					cmd := exec.Command("reset")
					err := cmd.Run()
					if err != nil {
						logger.Error(logger.FormatError(err))
					}
				}
				os.Exit(0)
			},
		}),
	)

	for {
		logger.ModuleSelectedV1(string(*client.module))
		client.executor(client.ppt.Input())
	}
}

var globalCmd = []string{"clear", "cls", "use", "update", "exit"}

func completer(in prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix([]prompt.Suggest{}, in.GetWordBeforeCursor(), true)
}
func (client *Client) executor(in string) {
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
