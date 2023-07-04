package cmd

import (
	"context"
	"fmt"
	"github.com/fasnow/ghttp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"idebug/logger"
	fs "idebug/plugin/feishu"
	"idebug/plugin/wechat"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Module string

const (
	WxModule     Module = "wechat"
	FeiShuModule Module = "feishu"
	NoModule     Module = ""
)

var (
	ModuleChangedCh       *chan bool
	ctx, cancel           = context.WithCancel(context.Background())
	CurrentModule         *Module
	WxClient              *wechat.Client
	FeiShuClient          *fs.Client
	FeiShuDefaultInterval = 100 * time.Millisecond
	departmentIdTypeMap   = map[string]string{"id": "department_id", "openid": "open_department_id"} //飞书部门ID类型
	userIdTypeMap         = map[string]string{"id": "user_id", "openid": "open_id"}                  //飞书用户ID类型
	retry                 = 3
)

var (
	Proxy                 *string
	userId                string
	departmentId          string
	userIdType            string //飞书用户ID类型
	departmentIdType      string //飞书部门ID类型缓存
	userIdTypeCache       string //飞书用户ID类型
	departmentIdTypeCache string //飞书部门ID类型缓存
	password              string //飞书企业邮箱密码更改
	recurse               bool   //递归获取
	verbose               int    //打印过程的数量
)

const mainUsage = `Global Commands:
    clear,cls         清屏
    exit              退出
    info              查看当前设置
    use               切换模块,可选值:wechat、feishu
    update            检测更新
    -h,--help,help    查看帮助
    set proxy <proxy> 设置代理,支持socks5,http
`

type mainCli struct {
	Root   *cobra.Command
	set    *cobra.Command
	proxy  *cobra.Command
	clear  *cobra.Command
	update *cobra.Command
	info   *cobra.Command
	use    *cobra.Command
	exit   *cobra.Command
}

func NewMainCli() *mainCli {
	cli := &mainCli{}
	cli.Root = cli.newRoot()
	cli.set = cli.newSet()
	cli.proxy = newProxy()
	cli.clear = cli.newClear()
	cli.update = cli.newUpdate()
	cli.info = cli.newInfo()
	cli.use = cli.newUse()
	cli.exit = cli.newExit()
	cli.init()
	return cli
}

func (cli *mainCli) init() {
	cli.set.AddCommand(cli.proxy)
	cli.Root.AddCommand(cli.set, cli.clear, cli.update, cli.info, cli.use, cli.exit)
	//cli.setHelpV1(cli.Root, "")
	//cli.setHelpV1(cli.proxy, "")
	//cli.setHelpV1(cli.set, "")
	//
	//cli.setHelpV1(cli.clear, "")
	//cli.setHelpV1(cli.update, "")
	//cli.setHelpV1(cli.info, "")
	//cli.setHelpV1(cli.use, "")
	//cli.setHelpV1(cli.exit, "")

	cli.setHelpV2(cli.Root, cli.proxy, cli.set, cli.clear, cli.update, cli.info, cli.use, cli.exit)
}

func (cli *mainCli) newExit() *cobra.Command {
	return &cobra.Command{
		Use:   "exit",
		Short: `退出模块或者程序`,
		Run: func(cmd *cobra.Command, args []string) {
			if *CurrentModule != NoModule {
				*CurrentModule = NoModule
			} else {
				if runtime.GOOS != "windows" {
					cmd := exec.Command("reset")
					err := cmd.Run()
					if err != nil {
						logger.Error(logger.FormatError(err))
					}
				}
				os.Exit(0)
			}
		},
	}
}

func (cli *mainCli) newRoot() *cobra.Command {
	return &cobra.Command{
		Use: "idebug",
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			reset()
		},
	}
}

func (cli *mainCli) newSet() *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: `设置参数`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
}

func (cli *mainCli) newClear() *cobra.Command {
	return &cobra.Command{
		Use:     `clear`,
		Short:   `清屏`,
		Aliases: []string{"cls"},
		Run: func(cmd *cobra.Command, args []string) {
			cmd1 := exec.Command("clear") // for Linux/MacOS
			if _, err := os.Stat("c:\\windows\\system32\\cmd.exe"); err == nil {
				cmd1 = exec.Command("cmd", "/c", "cls") // for Windows
			}
			cmd1.Stdout = os.Stdout
			err := cmd1.Run()
			if err != nil {
				logger.Error(logger.FormatError(err))
			}
		},
	}
}

func (cli *mainCli) newUpdate() *cobra.Command {
	return &cobra.Command{
		Use:   `update`,
		Short: `检查更新`,
		Run: func(cmd *cobra.Command, args []string) {
			version, releaseUrl, content := CheckUpdate()
			if version != "" {
				s := "最新版本: " + version
				s += "\n    下载地址: " + releaseUrl
				s += "\n    更新内容: " + content
				logger.Notice(s)
				return
			}
			logger.Success("当前已是最新版本")
		},
	}
}

func (cli *mainCli) newInfo() *cobra.Command {
	return &cobra.Command{
		Use:   `info`,
		Short: `查看环境变量`,
		Run: func(cmd *cobra.Command, args []string) {
			if Proxy == nil {
				fmt.Println(fmt.Sprintf("%-17s: %s", "proxy", ""))
			} else {
				fmt.Println(fmt.Sprintf("%-17s: %s", "proxy", *Proxy))
			}
		},
	}
}

func (cli *mainCli) newUse() *cobra.Command {
	return &cobra.Command{
		Use:   `use`,
		Short: `选择模块,可用模块:wechat、feishu `,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				logger.Warning("请选择一个有效模块: wechat、feishu")
				return
			}
			switch Module(args[0]) {
			case FeiShuModule:
				*CurrentModule = FeiShuModule
				if FeiShuClient == nil {
					FeiShuClient = fs.NewClient()
					FeiShuClient.Context(&ctx)
				}
				break
			case WxModule:
				*CurrentModule = WxModule
				if WxClient == nil {
					WxClient = wechat.NewWxClient()
					WxClient.Context(&ctx)
				}
				break
			default:
				logger.Error(fmt.Errorf("未知模块:" + args[0]))
			}
		},
	}
}

func newProxy() *cobra.Command {
	return &cobra.Command{
		Use:   `proxy`,
		Short: `设置代理`,
		Run: func(cmd *cobra.Command, args []string) {
			setProxy(args)
		},
	}
}

func commandMsiSet(cmd *cobra.Command) {
	//// 打印自定义用法
	//cmd.SetUsageFunc(func(cmd *cobra.Command) error {
	//	if name == "root" {
	//		ShowAllUsage()
	//	} else {
	//		ShowUsageByName(name)
	//	}
	//	return nil
	//})

}

func (cli *mainCli) setHelpV1(cmd *cobra.Command, extraUsage string) {
	// 不自己打印会多一个空白行
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		//返回error会自动打印用法
		logger.Error(err)
		return nil
	})

	// 禁止打印错误信息
	cmd.SilenceErrors = true
	//
	// 禁用帮助信息
	cmd.DisableSuggestions = true

	// 禁用用法信息
	cmd.SilenceUsage = true

	cmd.Flags().SortFlags = false

	var cmdCount int
	var cmdOutput string
	// 遍历子命令，生成命令列表
	for _, child := range cmd.Commands() {
		if child.IsAvailableCommand() || child.Name() == "help" {
			cmdCount++
			if cmdCount == 1 {
				cmdOutput = "Available Commands:\n"
			}
			cmdOutput += fmt.Sprintf("  %-15s %s\n", child.Name(), child.Short)
		}
	}
	flags := cmd.Flags()
	var flagCount int
	var flagOutput string
	flags.VisitAll(func(flag *pflag.Flag) {
		flagCount++
		if flagCount == 1 {
			flagOutput += "Flags:\n"
		}
		if flag.Shorthand != "" {
			flagOutput += fmt.Sprintf("  %-15s %s\n", fmt.Sprintf("-%s,--%s", flag.Shorthand, flag.Name), flag.Usage)
		} else {
			flagOutput += fmt.Sprintf("  %-15s %s\n", flag.Name, flag.Usage)
		}
	})
	var commandPath string
	paths := strings.Split(cmd.CommandPath(), " ")
	if len(paths) > 1 {
		commandPath = strings.Join(paths[1:], " ")
	}
	var usage = "Usage:\n"
	if extraUsage != "" {
		usage += fmt.Sprintf("  %s %s\n", commandPath, extraUsage)
	}
	if flagCount > 0 {
		if commandPath != "" {
			usage += fmt.Sprintf("  %s [flags]\n", commandPath)
		} else {
			usage += "  [flags]\n"
		}

	}
	if cmdCount > 0 {
		if commandPath != "" {
			usage += fmt.Sprintf("  %s [command]\n", commandPath)
		} else {
			usage += "  [command]\n"
		}
	}
	if cmdCount > 0 {
		usage += cmdOutput
	}
	if flagCount > 0 {
		usage += flagOutput
	}

	//设置自定义的使用帮助函数
	cmd.SetHelpFunc(func(cmd *cobra.Command, s []string) {
		fmt.Print(usage)
	})
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Print(usage)
		return nil
	})
}

func (cli *mainCli) setHelpV2(cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		// 不自己打印会多一个空白行
		cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
			logger.Error(err)
			return nil //返回error会自动打印用法
		})

		cmd.SilenceErrors = true      // 禁止打印错误信息
		cmd.DisableSuggestions = true // 禁用帮助信息
		cmd.SilenceUsage = true       // 禁用用法信息
		cmd.Flags().SortFlags = false

		// 设置自定义的使用帮助函数
		cmd.SetHelpFunc(func(cmd *cobra.Command, s []string) {
			fmt.Print(mainUsage)
		})
		cmd.SetUsageFunc(func(cmd *cobra.Command) error {
			fmt.Print(mainUsage)
			return nil
		})
	}
}

func setProxy(args []string) {
	if len(args) == 0 {
		*Proxy = ""
	} else {
		*Proxy = args[0]
	}
	ghttp.SetGlobalProxy(*Proxy)
	logger.Success("proxy => " + *Proxy)
}

func reset() {
	userId = ""
	departmentId = ""
	userIdType = ""
	departmentIdType = ""
	password = ""
	recurse = false
	verbose = -1
}
