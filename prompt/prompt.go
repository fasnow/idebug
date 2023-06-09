package prompt

import (
	"bufio"
	"fmt"
	"idebug/cmd"
	"idebug/config"
	"idebug/plugin"
	"idebug/utils"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Client struct {
	wxClient *plugin.WeChat
}

func (p *Client) Run() {
	wxClient, err := plugin.NewWxClient()
	if err != nil {
		utils.Error(err)
		return
	}
	p.wxClient = wxClient
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("idebug > ")
		scanner.Scan()
		input := scanner.Text()
		if p.executor(input) {
			break
		}
	}
}

func (p *Client) executor(in string) bool {
	in = strings.TrimSpace(in)
	switch in {
	case "cls", "clear":
		cmd1 := exec.Command("clear") // for Linux/MacOS
		if _, err := os.Stat("c:\\windows\\system32\\cmd.exe"); err == nil {
			cmd1 = exec.Command("cmd", "/c", "cls") // for Windows
		}
		cmd1.Stdout = os.Stdout
		err := cmd1.Run()
		if err != nil {
			utils.Error(err)
			break
		}
	case "info":
		showInfo()
		break
	case "run":
		if config.Info.CorpId == "" {
			utils.Warning("请先设置corpid")
			break
		}
		if config.Info.CorpSecret == "" {
			utils.Warning("请先设置corpsecret")
			break
		}
		token, i, _, err := p.wxClient.GetAccessToken(config.Info.CorpId, config.Info.CorpSecret)
		if err != nil {
			utils.Error(err)
			break
		}
		config.Info.AccessToken = token
		config.Info.ExpireIn = i
		// 现在需要登录才能调用且只能查询登录企业的access_token权限
		//_, err = p.wxClient.GetAccessTokenInfoByAccessToken(token)
		//if err != nil {
		//	utils.Error(err)
		//	break
		//}
		info := p.wxClient.GetClientInfo()
		config.Info.App = info.App
		config.Info.Tag = info.Tag
		config.Info.Department = info.Tag
		config.Info.ReliableDomain = info.ReliableDomain
		config.Info.Resource = info.Resource
		config.Info.Member = info.Member
		showInfo()
		break
	case "exit":
		return true
	default:
		args := strings.Fields(in)
		cmd.Root.SetArgs(args)
		err := cmd.Root.Execute()
		if err != nil {
			utils.Error(err)
			break
		}
	}
	return false
}

func (p *Client) hasAccessToken() bool {
	return p.wxClient.GetClientInfo().AccessToken != ""
}

func showInfo() {
	expireIn := strconv.Itoa(config.Info.ExpireIn)
	if expireIn == "0" {
		expireIn = ""
	}
	fmt.Printf("%s\n", strings.Repeat("=", 20))
	fmt.Println(fmt.Sprintf("%-17s: %s", "proxy", config.Proxy))
	fmt.Println(fmt.Sprintf("%-17s: %s", "corpid", config.Info.CorpId))
	fmt.Println(fmt.Sprintf("%-17s: %s", "corpsecret", config.Info.CorpSecret))
	fmt.Println(fmt.Sprintf("%-17s: %s", "access_token", config.Info.AccessToken))
	fmt.Println(fmt.Sprintf("%-17s: %s", "expire_in (s)", expireIn))
	//fmt.Println(fmt.Sprintf("%-15s: %s", "来源", config.Info.Resource))
	//fmt.Println(fmt.Sprintf("%-10s: %s", "通讯录范围 - 部门", strings.Join(config.Info.Department, "、")))
	//fmt.Println(fmt.Sprintf("%-10s: %s", "通讯录范围 - 成员", strings.Join(config.Info.Member, "、")))
	//fmt.Println(fmt.Sprintf("%-10s: %s", "通讯录范围 - 标签", strings.Join(config.Info.Tag, "、")))
	//fmt.Println(fmt.Sprintf("%-13s: %s", "应用权限", strings.Join(config.Info.App, "、")))
	//fmt.Println(fmt.Sprintf("%-13s: %s", "可信域名", strings.Join(config.Info.ReliableDomain, "、")))
	fmt.Printf("%s\n", strings.Repeat("=", 20))
}
