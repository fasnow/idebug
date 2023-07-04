package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"idebug/logger"
	"idebug/plugin/wechat"
	"idebug/utils"
	"os"
	"strconv"
	"strings"
)

const wechatUsage = mainUsage + `wechat Module:
    set corpid     <corpid>      设置corpid
    set corpsecret <corpsecret>  设置corpsecret
    run                          获取access_token
    dp             <did>         根据<did>查看部门详情  
    dp ls          <did>         根据<did>递归获取子部门id,不提供<did>则递归获取默认部门
    dp tree        <did>         根据<did>递归获取子部门信息,稍微详细一些,不提供<did>则递归获取默认部门  
    user           <uid>         根据<uid>查看用户详情
    user ls        <did> [-r]    根据<did>查看部门用户列表,-r:递归获取(默认false)
    dump           <did>         根据<did>递归导出部门用户,不提供<did>则递归获取默认部门
`

type wechatCli struct {
	Root       *cobra.Command
	info       *cobra.Command
	run        *cobra.Command
	set        *cobra.Command
	corpId     *cobra.Command
	corpSecret *cobra.Command
	dp         *cobra.Command
	dpLs       *cobra.Command
	dpTree     *cobra.Command
	user       *cobra.Command
	userLs     *cobra.Command
	dump       *cobra.Command
}

func NewWechatCli() *wechatCli {
	cli := &wechatCli{}
	cli.Root = cli.newRoot()
	cli.info = cli.newInfo()
	cli.run = cli.newRun()
	cli.set = cli.newSet()
	cli.corpId = cli.newCorpId()
	cli.corpSecret = cli.newCorpSecret()
	cli.dp = cli.newDp()
	cli.dpLs = cli.newDpLs()
	cli.dpTree = cli.newDpTree()
	cli.user = cli.newUser()
	cli.userLs = cli.newUserLs()
	cli.dump = cli.newDump()
	cli.init()
	return cli
}

func (cli *wechatCli) init() {
	//cli.dpLs.Flags().IntVarP(&verbose, "verbose", "v", -1, "控制台输出的条数,默认全部输出")
	//cli.dpTree.Flags().IntVarP(&verbose, "verbose", "v", -1, "控制台输出的条数,默认全部输出")

	//cli.userLs.Flags().IntVarP(&verbose, "verbose", "v", -1, "控制台输出的条数,默认全部输出")
	cli.userLs.Flags().BoolVarP(&recurse, "re", "r", false, "是否递归获取,默认false")

	cli.set.AddCommand(cli.corpId)
	cli.set.AddCommand(cli.corpSecret)
	cli.set.AddCommand(newProxy())
	cli.dp.AddCommand(cli.dpLs, cli.dpTree)
	cli.user.AddCommand(cli.userLs)
	cli.Root.AddCommand(cli.set, cli.run, cli.info, cli.dp, cli.user, cli.dump)

	cli.setHelpV1(cli.Root, cli.info, cli.run, cli.set, cli.corpId, cli.corpSecret, cli.dp, cli.dpLs, cli.dpTree, cli.user, cli.userLs, cli.dump)
}

func (cli *wechatCli) newRoot() *cobra.Command {
	return &cobra.Command{
		Use:   "wechat",
		Short: `微信模块`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if *CurrentModule != WxModule {
				return logger.FormatError(fmt.Errorf("请先设置模块为wechat"))
			}
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			reset()
		},
	}
}

func (cli *wechatCli) newRun() *cobra.Command {
	return &cobra.Command{
		Use:   `run`,
		Short: `根据设置获取access_token`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			conf := WxClient.GetConfig()
			if conf.CorpId == nil || *conf.CorpId == "" {
				return fmt.Errorf("请先设置corpid")
			}
			if conf.CorpSecret == nil || *conf.CorpSecret == "" {
				return fmt.Errorf("请先设置corpsecret")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			_, err := WxClient.GetAccessToken()
			if err != nil {
				logger.Error(logger.FormatError(err))
				return
			}
			cli.showClientConfig()
		},
	}
}

func (cli *wechatCli) newSet() *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: `设置参数`,
	}
}

func (cli *wechatCli) newCorpId() *cobra.Command {
	return &cobra.Command{
		Use:   "corpid",
		Short: `设置corpid`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				logger.Warning("请提供一个值")
				return
			}
			WxClient.SetCorpId(args[0])
			logger.Success("corpid => " + args[0])
		},
	}
}

func (cli *wechatCli) newCorpSecret() *cobra.Command {
	return &cobra.Command{
		Use:   "corpsecret",
		Short: `设置corpsecret`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				logger.Warning("请提供一个值")
				return nil
			}
			WxClient.SetCorpSecret(args[0])
			logger.Success("corpsecret => " + args[0])
			return nil
		},
	}
}

func (cli *wechatCli) newDp() *cobra.Command {
	return &cobra.Command{
		Use:   "dp",
		Short: `部门操作`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !cli.hasAccessToken() {
				return fmt.Errorf("请先执行run获取access_token")
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("请提供一个参数作为部门ID或者提供一个子命令")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			req := wechat.NewGetDepartmentReqBuilder(WxClient).DepartmentId(args[0]).Build()
			deptInfo, err := WxClient.Department.Get(req)
			if err != nil {
				logger.Error(logger.FormatError(err))
				return
			}
			fmt.Printf("%s\n", strings.Repeat("=", 20))
			fmt.Printf("%-10s: %d\n", "部门ID", deptInfo.ID)
			fmt.Printf("%-8s: %d\n", "上级部门ID", deptInfo.ParentId)
			fmt.Printf("%-6s: %s\n", "部门中文名称", deptInfo.Name)
			fmt.Printf("%-6s: %s\n", "部门英文名称", deptInfo.NameEn)
			fmt.Printf("%-8s: %s\n", "部门领导", strings.Join(deptInfo.DepartmentLeader, "、"))
			fmt.Printf("%-12s: %d\n", "Order", deptInfo.Order)
			fmt.Printf("%s\n", strings.Repeat("=", 20))
		},
	}
}

func (cli *wechatCli) newDpLs() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: `根据部门ID获取子部门ID`,
		Run: func(cmd *cobra.Command, args []string) {
			var departmentList []*WxDepartmentNode
			var depts []*wechat.DepartmentEntrySimplified
			for i := 0; i < retry; i++ {
				var err error
				if len(args) == 0 {
					req := wechat.NewGetDepartmentIdListReqBuilder(WxClient).Build()
					depts, err = WxClient.Department.GetIdList(req)

				} else {
					//递归获取指定部门所有ID
					req := wechat.NewGetDepartmentIdListReqBuilder(WxClient).DepartmentId(args[0]).Build()
					depts, err = WxClient.Department.GetIdList(req)
				}
				if err != nil {
					if i == retry-1 {
						logger.Error(logger.FormatError(err))
						return
					}
					continue
				}
				break
			}
			if len(depts) == 0 {
				logger.Warning("无可用部门信息")
				return
			}
			for _, dept := range depts {
				departmentList = append(departmentList, &WxDepartmentNode{
					DepartmentEntry: wechat.DepartmentEntry{
						DepartmentEntrySimplified: *dept},
				})
			}
			tree := cli.buildDepartmentIdsTree(departmentList)
			index := 0
			cli.printDepartmentIdsTree(tree, 0, &index)
			logger.Info("正在保存至HTML文件...")
			msg, err := cli.departmentIdsTreeToHTML(tree, "wechat_dept_ids.html")
			if err != nil {
				logger.Error(logger.FormatError(err))
				return
			}
			logger.Success(msg)
		},
	}
}

func (cli *wechatCli) newDpTree() *cobra.Command {
	return &cobra.Command{
		Use:   "tree",
		Short: `根据部门ID获取子部门ID,获取的信息稍微详细点`,
		Run: func(cmd *cobra.Command, args []string) {
			var departmentList []*WxDepartmentNode
			var err error
			var departments []*wechat.DepartmentEntry
			for i := 0; i < retry; i++ {
				if len(args) == 0 {
					req := wechat.NewGetDepartmentListReqBuilder(WxClient).Build()
					departments, err = WxClient.Department.GetList(req)
				} else {
					req := wechat.NewGetDepartmentListReqBuilder(WxClient).DepartmentId(args[0]).Build()
					departments, err = WxClient.Department.GetList(req)
				}
				if err != nil {
					if i < retry-1 {
						logger.Error(logger.FormatError(err))
						return
					}
					continue
				}
				break
			}
			if len(departments) == 0 {
				logger.Warning("无可用部门信息")
				return
			}
			for _, v := range departments {
				d := wechat.DepartmentEntry{
					DepartmentEntrySimplified: wechat.DepartmentEntrySimplified{
						ID:       v.ID,
						ParentId: v.ParentId,
						Order:    v.Order,
					},
					Name:             v.Name,
					NameEn:           v.NameEn,
					DepartmentLeader: v.DepartmentLeader,
				}
				departmentList = append(departmentList, &WxDepartmentNode{DepartmentEntry: d})
			}
			tree := cli.buildDepartmentIdsTree(departmentList)
			var index int
			index = 0
			cli.printDepartmentTree(tree, 0, &index)
			logger.Info("正在保存至HTML文件...")
			msg, err := cli.saveDepartmentTreeToHTML(tree, "wechat_dept.html")
			if err != nil {
				logger.Error(logger.FormatError(err))
				return
			}
			logger.Success(msg)
		}}
}

func (cli *wechatCli) newUser() *cobra.Command {
	return &cobra.Command{
		Use:   "user",
		Short: `用户操作`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !cli.hasAccessToken() {
				return fmt.Errorf("请先执行run获取access_token")
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("请提供一个参数作为用户ID或者提供一个子命令")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			req := wechat.NewGetUserReqBuilder(WxClient).UserId(args[0]).Build()
			userInfo, err := WxClient.User.Get(req)
			if err != nil {
				logger.Error(logger.FormatError(err))
				return
			}
			cli.showUserInfo(userInfo, false)
			return
		},
	}
}

func (cli *wechatCli) newUserLs() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: `根据部门ID获取用户`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("请提供一个参数作为部门ID")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			req := wechat.NewGetUsersByDepartmentIdReqBuilder(WxClient).DepartmentId(args[0]).Fetch(recurse).Build()
			userList, err := WxClient.User.GetUsersByDepartmentId(req)
			if err != nil {
				return
			}
			if len(userList) == 0 {
				logger.Warning("无可用用户信息")
				return
			}
			for i := 0; i < len(userList); i++ {
				//if i >= verbose && verbose > -1 {
				//	logger.Info("更多数据请查看文件")
				//	break
				//}
				cli.showUserInfo(userList[i], true)
			}
			logger.Info("正在保存至XLSX文件...")
			msg, err := cli.saveUserToExcel(userList, "wechat_users.xlsx")
			if err != nil {
				logger.Error(logger.FormatError(err))
				return
			}
			logger.Success(msg)
		},
	}
}

func (cli *wechatCli) newDump() *cobra.Command {
	return &cobra.Command{
		Use:   "dump",
		Short: `根据部门ID导出用户,不提供部门ID则导出通讯录授权范围内所有部门用户`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !cli.hasAccessToken() {
				return fmt.Errorf("请先执行run获取access_token")

			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var departmentTreeResource []*wechat.DepartmentEntry
			var deptList []*wechat.DepartmentEntry
			var err error
			logger.Info("正在获取部门树...")
			for i := 0; i < retry; i++ {
				if len(args) == 0 {
					req := wechat.NewGetDepartmentListReqBuilder(WxClient).Build()
					deptList, err = WxClient.Department.GetList(req)
				} else {
					req := wechat.NewGetDepartmentListReqBuilder(WxClient).DepartmentId(args[0]).Build()
					deptList, err = WxClient.Department.GetList(req)
				}
				if err != nil {
					if i == retry-1 {
						logger.Error(logger.FormatError(err))
						return
					}
					continue
				}
				break
			}
			if len(deptList) == 0 {
				logger.Warning("无可用部门信息")
				return
			}
			for _, v := range deptList {
				d := wechat.DepartmentEntry{
					DepartmentEntrySimplified: wechat.DepartmentEntrySimplified{
						ID:       v.ID,
						ParentId: v.ParentId,
						Order:    v.Order,
					},
					Name:             v.Name,
					NameEn:           v.NameEn,
					DepartmentLeader: v.DepartmentLeader,
				}
				departmentTreeResource = append(departmentTreeResource, &d)
			}
			departmentTree := cli.buildDepartmentTree(departmentTreeResource)
			logger.Info("正在获取用户...")
			var userList []*wechat.UserEntry
			for i := 0; i < retry; i++ {
				req := wechat.NewGetUsersByDepartmentIdReqBuilder(WxClient).DepartmentId(strconv.Itoa(departmentTreeResource[0].ID)).Fetch(true).Build()
				userList, err = WxClient.User.GetUsersByDepartmentId(req)
				if err != nil {
					if i == retry-1 {
						logger.Error(logger.FormatError(err))
						return
					}
					continue
				}
				break
			}

			// 将用户插入到部门树中
			for _, wxUser := range userList {
				departmentIDs := wxUser.Department
				for _, departmentID := range departmentIDs {
					cli.insertUserToDepartmentTree(wxUser, departmentID, departmentTree)
				}
			}

			//保存文件
			logger.Info("正在保存至html文件...")
			msg, err := cli.saveDepartmentTreeWithUsersToHTML(departmentTree, "wechat_dump.html")
			if err != nil {
				logger.Error(logger.FormatError(err))
				logger.Info("保存文件失败")
			} else {
				logger.Success(msg)
			}

			logger.Info("正在保存至xlsx文件...")
			msg, err = cli.saveUserTreeToExcel(departmentTree, "wechat_dump.xlsx")
			if err != nil {
				logger.Error(logger.FormatError(err))
				logger.Info("保存文件失败")
			} else {
				logger.Success(msg)
			}
		},
	}
}

func (cli *wechatCli) newInfo() *cobra.Command {
	return &cobra.Command{
		Use:   `info`,
		Short: `查看当前设置`,
		Run: func(cmd *cobra.Command, args []string) {
			cli.showClientConfig()
		},
	}
}

func (cli *wechatCli) hasAccessToken() bool {
	return WxClient.GetAccessTokenFromCache() != ""
}

func (cli *wechatCli) showUserInfo(userInfo *wechat.UserEntry, inLine bool) {
	var depts []string
	for _, deptId := range userInfo.Department {
		for i := 0; i < retry; i++ {
			req := wechat.NewGetDepartmentReqBuilder(WxClient).DepartmentId(strconv.Itoa(deptId)).Build()
			deptInfo, err := WxClient.Department.Get(req)
			if err != nil {
				logger.Error(logger.FormatError(err))
				continue
			}
			depts = append(depts, fmt.Sprintf("%s (ID:%d)", deptInfo.Name, deptId))
			break
		}
	}
	if inLine {
		s := fmt.Sprintf("  -ID[%s] 姓名[%s] 所属部门ID[%s] 职位[%s] 手机[%s] 邮箱[%s] 微信二维码[%s]", userInfo.UserId, userInfo.Name, strings.Join(depts, "、"), userInfo.Position, userInfo.Mobile, userInfo.Email, userInfo.QrCode)
		fmt.Println(s)
		return
	}
	fmt.Printf("%s\n", strings.Repeat("=", 20))
	fmt.Printf("%-10s: %s\n", "ID", userInfo.UserId)
	fmt.Printf("%-8s: %s\n", "姓名", userInfo.Name)
	fmt.Printf("%-6s: %s\n", "所属部门", strings.Join(depts, "、"))
	fmt.Printf("%-8s: %s\n", "职位", userInfo.Position)
	fmt.Printf("%-8s: %s\n", "手机", userInfo.Mobile)
	fmt.Printf("%-8s: %s\n", "邮箱", userInfo.Email)
	fmt.Printf("%-5s: %s\n", "微信二维码", userInfo.QrCode)
	fmt.Printf("%s\n", strings.Repeat("=", 20))
}

func (cli *wechatCli) showClientConfig() {
	wxClientConfig := WxClient.GetConfig()
	fmt.Printf("%s\n", strings.Repeat("=", 20))
	if Proxy == nil {
		fmt.Println(fmt.Sprintf("%-17s: %s", "proxy", ""))
	} else {
		fmt.Println(fmt.Sprintf("%-17s: %s", "proxy", *Proxy))
	}
	if wxClientConfig.CorpId == nil {
		fmt.Println(fmt.Sprintf("%-17s: %s", "corpid", ""))
	} else {
		fmt.Println(fmt.Sprintf("%-17s: %s", "corpid", *wxClientConfig.CorpId))
	}
	if wxClientConfig.CorpSecret == nil {
		fmt.Println(fmt.Sprintf("%-17s: %s", "corpsecret", ""))
	} else {
		fmt.Println(fmt.Sprintf("%-17s: %s", "corpsecret", *wxClientConfig.CorpSecret))
	}
	if wxClientConfig.AccessToken == nil {
		fmt.Println(fmt.Sprintf("%-17s: %s", "access_token", ""))
	} else {
		fmt.Println(fmt.Sprintf("%-17s: %s", "access_token", *wxClientConfig.AccessToken))
	}
	fmt.Printf("%s\n", strings.Repeat("=", 20))
}

func (cli *wechatCli) setHelpV1(cmds ...*cobra.Command) {
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
			fmt.Print(wechatUsage)
		})
		cmd.SetUsageFunc(func(cmd *cobra.Command) error {
			fmt.Print(wechatUsage)
			return nil
		})
	}
}

type WxDepartmentNode struct {
	wechat.DepartmentEntry
	User     []*wechat.UserEntry
	Children []*WxDepartmentNode
}

func (cli *wechatCli) insertUserToDepartmentTree(wxUser *wechat.UserEntry, departmentID int, departments []*WxDepartmentNode) {
	for _, department := range departments {
		if department.ID == departmentID {
			// 创建一个新的用户对象并复制数据
			newUser := *wxUser
			department.User = append(department.User, &newUser)
			return
		}

		cli.insertUserToDepartmentTree(wxUser, departmentID, department.Children)
	}
}

func (cli *wechatCli) buildDepartmentIdsTree(nodes []*WxDepartmentNode) []*WxDepartmentNode {
	var rootDepartments []*WxDepartmentNode
	departmentMap := make(map[int]*WxDepartmentNode)
	// 创建部门映射表
	for _, dept := range nodes {
		departmentMap[dept.ID] = dept
	}

	// 构建部门树
	for _, dept := range nodes {
		departmentNode := departmentMap[dept.ID]
		parentDept, ok := departmentMap[dept.ParentId]
		if ok {
			parentDept.Children = append(parentDept.Children, departmentNode)
		} else {
			rootDepartments = append(rootDepartments, departmentNode)
		}
	}
	return rootDepartments
}

func (cli *wechatCli) buildDepartmentTree(depts []*wechat.DepartmentEntry) []*WxDepartmentNode {
	var rootDepartments []*WxDepartmentNode
	departmentMap := make(map[int]*WxDepartmentNode)
	// 创建部门映射表
	for _, dept := range depts {
		departmentMap[dept.ID] = &WxDepartmentNode{
			DepartmentEntry: *dept,
			User:            []*wechat.UserEntry{},
			Children:        []*WxDepartmentNode{},
		}
	}

	// 构建部门树
	for _, dept := range depts {
		departmentNode := departmentMap[dept.ID]
		parentDept, ok := departmentMap[dept.ParentId]
		if ok {
			parentDept.Children = append(parentDept.Children, departmentNode)
		} else {
			rootDepartments = append(rootDepartments, departmentNode)
		}
	}
	return rootDepartments
}

// printDepartmentTree 打印部门树
func (cli *wechatCli) printDepartmentTree(nodes []*WxDepartmentNode, level int, index *int) {
	for _, dept := range nodes {
		//if *index >= verbose && verbose > -1 {
		//	break
		//}
		var nameEn string
		var leaderId string
		if dept.NameEn != "" {
			nameEn = fmt.Sprintf("  英文名称[%s]", dept.NameEn)
		}
		if len(dept.DepartmentLeader) != 0 {
			leaderId = fmt.Sprintf("  领导ID[%s]", strings.Join(dept.DepartmentLeader, "、"))
		}
		fmt.Printf("%s  -ID[%d]  名称[%s]%s%s\n", strings.Repeat(" ", 3*level), dept.ID, dept.Name, nameEn, leaderId)
		*index++
		cli.printDepartmentTree(dept.Children, level+1, index)
	}
}

// printDepartmentIdsTree 打印部门ID树
func (cli *wechatCli) printDepartmentIdsTree(nodes []*WxDepartmentNode, level int, index *int) {
	for _, dept := range nodes {
		if *index >= verbose && verbose > -1 {
			break
		}
		fmt.Printf("%s  -ID:%d\n", strings.Repeat(" ", 3*level), dept.ID)
		*index++
		cli.printDepartmentIdsTree(dept.Children, level+1, index)
	}
}

// generateDepartmentTreeHTML 生成包含部门信息的部门树HTML代码
func (cli *wechatCli) generateDepartmentTreeHTML(nodes []*WxDepartmentNode, level int) string {
	html := ""
	for _, dept := range nodes {
		html += fmt.Sprintf("<div style=\"margin-left:%dem;\">", level)

		// 添加折叠/展开按钮
		if len(dept.Children) > 0 {
			html += fmt.Sprintf("<span class=\"toggle\" onclick=\"toggleDepartment(this)\">-</span>")
		} else {
			html += "<span class=\"empty-toggle\"></span>"
		}

		s := fmt.Sprintf("ID:%d&nbsp;&nbsp;%s", dept.ID, dept.Name)
		if dept.NameEn != "" {
			s += fmt.Sprintf("&nbsp;&nbsp;英文名称:%s", dept.NameEn)
		}
		if len(dept.DepartmentLeader) != 0 {
			s += fmt.Sprintf("&nbsp;&nbsp;领导ID:%s\n", strings.Join(dept.DepartmentLeader, "、"))
		}

		// 添加部门名称
		html += fmt.Sprintf("<span class=\"department\">%s</span>", s)

		// 递归生成子部门树
		if len(dept.Children) > 0 {
			html += cli.generateDepartmentTreeHTML(dept.Children, level+1)
		}

		html += "</div>"
	}
	return html
}

// saveDepartmentTreeToHTML 生成含部门信息的部门树并将其输出为HTML文档
func (cli *wechatCli) saveDepartmentTreeToHTML(nodes []*WxDepartmentNode, filename string) (string, error) {
	index := strings.LastIndex(filename, ".html")
	if index == -1 {
		filename = filename + ".html"
	}
	tmp := filename
	// 判断文件是否存在
	if utils.IsFileExists(filename) {
		newFilename := generateNewFilename(filename)
		filename = newFilename
	}
	file, err := os.Create(filename)
	if err != nil {
		return "", errors.New("创建HTML文件失败: " + err.Error())
	}
	defer file.Close()
	departmentTreeHTML := cli.generateDepartmentTreeHTML(nodes, 0)
	htmlDocument := cli.generateTreeHTMLDocument(departmentTreeHTML)
	_, err = file.WriteString(htmlDocument)
	if err != nil {
		return "", errors.New("无法写入HTML内容到文件: " + err.Error())
	}
	if filename == tmp {
		return fmt.Sprintf("文件已保存至 %s", filename), nil
	}
	return fmt.Sprintf("文件 %s 已存在,已保存至 %s", tmp, filename), nil
}

// generateDepartmentIdsTreeHTML 生成只包含ID的部门树HTML代码
func (cli *wechatCli) generateDepartmentIdsTreeHTML(nodes []*WxDepartmentNode, level int) string {
	html := ""
	for _, dept := range nodes {
		html += fmt.Sprintf("<div style=\"margin-left:%dem;\">", level)

		// 添加折叠/展开按钮
		if len(dept.Children) > 0 {
			html += fmt.Sprintf("<span class=\"toggle\" onclick=\"toggleDepartment(this)\">-</span>")
		} else {
			html += "<span class=\"empty-toggle\"></span>"
		}

		s := fmt.Sprintf("ID:%d", dept.ID)

		// 添加部门ID
		html += fmt.Sprintf("<span class=\"department\">%s</span>", s)

		// 递归生成子部门树
		if len(dept.Children) > 0 {
			html += cli.generateDepartmentTreeHTML(dept.Children, level+1)
		}

		html += "</div>"
	}
	return html
}

// departmentIdsTreeToHTML 生成只包含ID的部门树并将其输出为HTML文档
func (cli *wechatCli) departmentIdsTreeToHTML(nodes []*WxDepartmentNode, filename string) (string, error) {
	index := strings.LastIndex(filename, ".html")
	if index == -1 {
		filename = filename + ".html"
	}
	tmp := filename
	// 判断文件是否存在
	if utils.IsFileExists(filename) {
		newFilename := generateNewFilename(filename)
		filename = newFilename
	}
	file, err := os.Create(filename)
	if err != nil {
		return "", errors.New("创建HTML文件失败: " + err.Error())
	}
	defer file.Close()
	departmentTreeHTML := cli.generateDepartmentIdsTreeHTML(nodes, 0)
	htmlDocument := cli.generateTreeHTMLDocument(departmentTreeHTML)
	_, err = file.WriteString(htmlDocument)
	if err != nil {
		return "", errors.New("无法写入HTML内容到文件: " + err.Error())
	}
	if filename == tmp {
		return fmt.Sprintf("文件已保存至 %s", filename), nil
	}
	return fmt.Sprintf("文件 %s 已存在,已保存至 %s", tmp, filename), nil
}

// generateDepartmentTreeWithUsersHTML 生成包含部门信息和用户信息的部门树HTML代码
func (cli *wechatCli) generateDepartmentTreeWithUsersHTML(nodes []*WxDepartmentNode, level int) string {
	html := ""
	for _, dept := range nodes {
		html += fmt.Sprintf("<div style=\"margin-left:%dem;\">", level)

		s := fmt.Sprintf("ID:%d&nbsp;&nbsp;%s", dept.ID, dept.Name)
		if dept.NameEn != "" {
			s += fmt.Sprintf("&nbsp;&nbsp;英文名称:%s", dept.NameEn)
		}
		if len(dept.DepartmentLeader) != 0 {
			s += fmt.Sprintf("&nbsp;&nbsp;领导ID:%s\n", strings.Join(dept.DepartmentLeader, "、"))
		}

		// 添加折叠/展开按钮
		if len(dept.Children) > 0 {
			html += fmt.Sprintf("<span class=\"toggle\" onclick=\"toggleDepartment(this)\">-</span>")
		} else {
			html += "<span class=\"empty-toggle\"></span>"
		}
		html += fmt.Sprintf("<span class=\"department\">%s</span>", s)

		// 添加部门名称和用户列表的父级容器
		html += fmt.Sprintf("<div class=\"department-container\">")
		a := ""
		for i := 0; i < len(dept.User); i++ {
			user := dept.User[i]
			m := fmt.Sprintf("%s&nbsp;&nbsp;%s", user.UserId, user.Name)
			if user.Position != "" {
				m += fmt.Sprintf("&nbsp;&nbsp;%s", user.Position)
			}
			m += fmt.Sprintf("&nbsp;&nbsp;%s&nbsp;&nbsp;%s&nbsp;&nbsp;%s", user.Mobile, user.Email, user.QrCode)
			a += fmt.Sprintf("<li>%s</li>", m)
		}
		html += fmt.Sprintf("<ul class=\"user-list\">%s</ul>", a)

		// 递归生成子部门树
		if len(dept.Children) > 0 {
			html += cli.generateDepartmentTreeWithUsersHTML(dept.Children, level+1)
		}

		html += "</div></div>"
	}
	return html
}

// saveDepartmentTreeWithUsersToHTML 生成包含部门信息和用户信息的部门树并将其输出为HTML文档
func (cli *wechatCli) saveDepartmentTreeWithUsersToHTML(nodes []*WxDepartmentNode, filename string) (string, error) {
	index := strings.LastIndex(filename, ".html")
	if index == -1 {
		filename = filename + ".html"
	}
	tmp := filename
	// 判断文件是否存在
	if utils.IsFileExists(filename) {
		newFilename := generateNewFilename(filename)
		filename = newFilename
	}
	file, err := os.Create(filename)
	if err != nil {
		return "", errors.New("创建HTML文件失败: " + err.Error())
	}
	defer file.Close()
	departmentTreeHTML := cli.generateDepartmentTreeWithUsersHTML(nodes, 0)
	htmlDocument := cli.generateTreeHTMLDocument(departmentTreeHTML)
	_, err = file.WriteString(htmlDocument)
	if err != nil {
		return "", errors.New("无法写入HTML内容到文件: " + err.Error())
	}
	if filename == tmp {
		return fmt.Sprintf("文件已保存至 %s", filename), nil
	}
	return fmt.Sprintf("%s 已存在,已另存为 %s", tmp, filename), nil
}

// generateTreeHTMLDocument 生成添加折叠功能的完整的HTML文档
func (cli *wechatCli) generateTreeHTMLDocument(content string) string {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				.toggle {
					margin-right: 5px;
					cursor: pointer;
				}
				.empty-toggle {
					width: 14px;
					display: inline-block;
				}
				.department {
					font-weight: bold;
				}
				.collapse {
					display: none;
				}
				ul{
					margin:0px;
				}
			</style>
			<script>
				function toggleDepartment(button) {
					var div = button.parentNode;
					var subDepartments = div.getElementsByTagName("div");
					for (var i = 0; i < subDepartments.length; i++) {
						subDepartments[i].classList.toggle("collapse");
					}
					button.textContent = button.textContent === "+" ? "-" : "+";
				}

				function expandAll() {
					var departments = document.getElementsByClassName("toggle");
					for (var i = 0; i < departments.length; i++) {
						var button = departments[i];
						var div = button.parentNode;
						var subDepartments = div.getElementsByTagName("div");
						for (var j = 0; j < subDepartments.length; j++) {
							subDepartments[j].classList.remove("collapse");
						}
						button.textContent = "-";
					}
				}

				function collapseAll() {
					var departments = document.getElementsByClassName("toggle");
					for (var i = 0; i < departments.length; i++) {
						var button = departments[i];
						var div = button.parentNode;
						var subDepartments = div.getElementsByTagName("div");
						for (var j = 0; j < subDepartments.length; j++) {
							subDepartments[j].classList.add("collapse");
						}
						button.textContent = "+";
					}
				}
			</script>
		</head>
		<body>
			<div id="departmentTree">
				<div>
					<button onclick="expandAll()">全部展开</button>
					<button onclick="collapseAll()">全部折叠</button>
				</div>
				%s
			</div>
		</body>
		</html>
	`
	return fmt.Sprintf(html, content)
}

// saveUserTreeToExcel 生成包含部门名称的用户信息的XLSX文档
func (cli *wechatCli) saveUserTreeToExcel(nodes []*WxDepartmentNode, filename string) (string, error) {
	if !strings.HasSuffix(filename, ".xlsx") {
		filename = filename + ".xlsx"
	}
	tmp := filename
	// 判断文件是否存在
	if utils.IsFileExists(filename) {
		newFilename := generateNewFilename(filename)
		filename = newFilename
	}

	// 设置表头
	headers := []any{"id", "部门名称", "部门英文名称", "部门ID", "部门领导", "上级部门ID", "用户ID", "姓名", "性别", "电话号码", "邮箱", "职位", "微信二维码"}

	var data [][]any
	var index = 0
	for _, d := range nodes {
		for _, user := range d.User {
			index++
			gender := ""
			if user.Gender == "1" {
				gender = "男"
			} else if user.Gender == "0" {
				gender = "女"
			}
			s := []any{index, d.Name, d.NameEn, d.ID, strings.Join(d.DepartmentLeader, "、"), d.ParentId, user.UserId, user.Name, gender, user.Mobile, user.Email, user.Position, user.QrCode}
			data = append(data, s)
		}
		cli.fetchColItem(&index, &data, d.Children)
	}

	// 保存文件
	err := saveToExcel(headers, data, filename)
	if err != nil {
		return "", errors.New("保存 Excel 文件失败: " + err.Error())
	}
	if tmp != filename {
		return fmt.Sprintf("%s 已存在,已另存为 %s", tmp, filename), nil
	}
	return fmt.Sprintf("文件已保存至 %s", filename), nil
}

// saveUserToExcel 生成包含所属部门ID的用户信息的XLSX文档
func (cli *wechatCli) saveUserToExcel(users []*wechat.UserEntry, filename string) (string, error) {
	index := strings.LastIndex(filename, ".xlsx")
	if index == -1 {
		filename = filename + ".xlsx"
	}
	tmp := filename
	// 判断文件是否存在
	if utils.IsFileExists(filename) {
		newFilename := generateNewFilename(filename)
		filename = newFilename
	}

	// 设置表头
	headers := []any{"id", "所属部门ID", "用户ID", "姓名", "性别", "电话号码", "邮箱", "职位", "微信二维码"}

	var data [][]any
	// 写入内容
	for i, user := range users {
		var s []string
		for _, deptId := range user.Department {
			s = append(s, strconv.Itoa(deptId))
		}
		var gender string
		if user.Gender == "1" {
			gender = "男"
		} else if user.Gender == "0" {
			gender = "女"
		}
		d := []any{i, strings.Join(s, " "), user.UserId, user.Name, gender, user.Mobile, user.Email, user.Position, user.QrCode}
		data = append(data, d)
	}
	// 保存文件
	err := saveToExcel(headers, data, filename)
	if err != nil {
		return "", errors.New("保存 Excel 文件失败: " + err.Error())
	}
	if tmp != filename {
		return fmt.Sprintf("%s 已存在,已另存为 %s", tmp, filename), nil
	}
	return fmt.Sprintf("文件已保存至 %s", filename), nil
}

func (cli *wechatCli) fetchColItem(index *int, data *[][]any, dept []*WxDepartmentNode) {
	for _, d := range dept {
		for _, user := range d.User {
			*index++
			gender := ""
			if user.Gender == "1" {
				gender = "男"
			} else if user.Gender == "0" {
				gender = "女"
			}
			s := []any{*index, d.Name, d.NameEn, d.ID, strings.Join(d.DepartmentLeader, "、"), d.ParentId, user.UserId, user.Name, gender, user.Mobile, user.Email, user.Position, user.QrCode}
			*data = append(*data, s)
		}
		cli.fetchColItem(index, data, d.Children)
	}
}
