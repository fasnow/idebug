package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"idebug/config"
	"idebug/http"
	"idebug/plugin"
	"idebug/utils"
	"strconv"
	"strings"
)

var (
	postData  string
	userId    string
	deptId    string
	recurse   bool
	dump      bool
	printNum  int
	printTree bool
)

var Root = &cobra.Command{
	Use: "idebug",
	Run: func(cmd *cobra.Command, args []string) {
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		postData = ""
		userId = ""
		deptId = ""
		recurse = false
		dump = false
		printNum = 5
		printTree = false
	},
}

var Help = &cobra.Command{Run: func(cmd *cobra.Command, args []string) {
	ShowAllUsage()
}}

var Proxy = &cobra.Command{
	Use: "proxy",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			err := http.SetGlobalProxy("")
			if err != nil {
				utils.Error(err)
				return
			}
			config.Proxy = ""
		} else {
			err := http.SetGlobalProxy(args[0])
			if err != nil {
				utils.Error(err)
				return
			}
			config.Proxy = args[0]
		}
		utils.Success("proxy => " + config.Proxy)
	},
}

var CorpId = &cobra.Command{
	Use: "corpid",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			utils.Warning("请提供一个值")
			return
		}
		config.Info.CorpId = args[0]
		utils.Success("corpid => " + config.Info.CorpId)
	},
}

var CorpSecret = &cobra.Command{
	Use: "corpsecret",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			utils.Warning("请提供一个值")
			return nil
		}
		config.Info.CorpSecret = args[0]
		utils.Success("corpsecret => " + config.Info.CorpSecret)
		return nil
	},
}

var Set = &cobra.Command{
	Use: "set",
}

var DpAdd = &cobra.Command{
	Use: "add",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			utils.Warning("请先提供添加部门信息")
			return
		}

	},
}

var DpUpdate = &cobra.Command{
	Use: "update",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			utils.Warning("请先提供更新部门信息")
			return
		}
	}}

var DpDelete = &cobra.Command{
	Use: "delete",
	Run: func(cmd *cobra.Command, args []string) {
		if !hasAccessToken() {
			utils.Warning("请先执行run获取access_token")
			return
		}
		if len(args) < 1 {
			utils.Warning("请提供部门ID")
			return
		}
	}}

var DpList = &cobra.Command{
	Use: "list",
	Run: func(cmd *cobra.Command, args []string) {
		var departmentTreeResource []*utils.Department
		var departmentList []plugin.WxDepartment
		if !hasAccessToken() {
			utils.Warning("请先执行run获取access_token")
			return
		}
		if len(args) < 1 {
			res, _, err := plugin.GetSubDepartmentIdList(config.Info.AccessToken)
			if err != nil {
				utils.Error(err)
				return
			}
			departmentList = res
		} else {
			departmentId, _ := strconv.Atoi(args[0])
			//递归获取指定部门所有ID
			res, _, err := plugin.GetSubDepartmentIdListById(config.Info.AccessToken, departmentId)
			if err != nil {
				utils.Error(err)
				return
			}
			departmentList = res
		}
		if len(departmentList) == 0 {
			return
		}
		for _, v := range departmentList {
			d := utils.Department{
				WxDepartment: plugin.WxDepartment{
					WxSimpleDepartment: plugin.WxSimpleDepartment{
						ID:       v.ID,
						ParentId: v.ParentId,
						Order:    v.Order,
					},
				},
			}
			departmentTreeResource = append(departmentTreeResource, &d)
		}
		tree := utils.BuildDepartmentTree(departmentTreeResource)
		if printTree {
			utils.PrintDepartmentIdsTree(tree, 0)
		}
		utils.Info("正在保存至HTML文件...")
		msg, err := utils.SaveDepartmentIdsTreeToHTML(tree, "dept_tree_ids.html")
		if err != nil {
			utils.Error(err)
			return
		}
		utils.Success(msg)
	}}

var DpDetail = &cobra.Command{
	Use: "detail",
	Run: func(cmd *cobra.Command, args []string) {
		if !hasAccessToken() {
			utils.Warning("请先执行run获取access_token")
			return
		}
		if len(args) < 1 {
			utils.Warning("请提供部门ID")
			return
		}
		departmentId, _ := strconv.Atoi(args[0])
		detail, _, err := plugin.GetDepartmentDetail(config.Info.AccessToken, departmentId)
		if err != nil {
			utils.Error(err)
			return
		}
		fmt.Printf("%s\n", strings.Repeat("=", 20))
		fmt.Printf("%-10s: %d\n", "部门ID", detail.ID)
		fmt.Printf("%-8s: %d\n", "上级部门ID", detail.ParentId)
		fmt.Printf("%-6s: %s\n", "部门中文名称", detail.Name)
		fmt.Printf("%-6s: %s\n", "部门英文名称", detail.NameEn)
		fmt.Printf("%-8s: %s\n", "部门领导", strings.Join(detail.DepartmentLeader, "、"))
		fmt.Printf("%-12s: %d\n", "Order", detail.Order)
		fmt.Printf("%s\n", strings.Repeat("=", 20))
	}}

var DpTree = &cobra.Command{
	Use: "tree",
	Run: func(cmd *cobra.Command, args []string) {
		var departmentTreeResource []*utils.Department
		var departmentList []plugin.WxDepartment
		if !hasAccessToken() {
			utils.Warning("请先执行run获取access_token")
			return
		}
		departmentList, _, err := plugin.GetDepartment(config.Info.AccessToken)
		if err != nil {
			utils.Error(err)
			return
		}
		if len(departmentList) == 0 {
			utils.Warning("无可用部门信息")
			return
		}
		for _, v := range departmentList {
			d := utils.Department{
				WxDepartment: plugin.WxDepartment{
					WxSimpleDepartment: plugin.WxSimpleDepartment{
						ID:       v.ID,
						ParentId: v.ParentId,
						Order:    v.Order,
					},
					Name:             v.Name,
					NameEn:           v.NameEn,
					DepartmentLeader: v.DepartmentLeader,
				},
			}
			departmentTreeResource = append(departmentTreeResource, &d)
		}
		tree := utils.BuildDepartmentTree(departmentTreeResource)
		if printTree {
			utils.PrintDepartmentTree(tree, 0)
		}
		utils.Info("正在保存至HTML文件...")
		msg, err := utils.SaveDepartmentTreeToHTML(tree, "dept_tree.html")
		if err != nil {
			utils.Error(err)
			return
		}
		utils.Success(msg)
	}}

var Dp = &cobra.Command{
	Use: "dp",
}

var User = &cobra.Command{
	Use: "user",
	Run: func(cmd *cobra.Command, args []string) {
		flagNum := 0
		if userId != "" && deptId != "" {
			flagNum++
		}
		if deptId != "" {
			flagNum++
		}
		if dump {
			flagNum++
		}
		if flagNum > 1 {
			utils.Warning("用户ID和部门ID不能同时设置")
			return
		}
		if !hasAccessToken() {
			utils.Warning("请先执行run获取access_token")
			return
		}
		if userId != "" {
			user, _, err := plugin.GetDepartmentSpecifiedUser(config.Info.AccessToken, userId)
			if err != nil {
				utils.Error(err)
				return
			}
			fmt.Printf("%s\n", strings.Repeat("=", 20))
			fmt.Printf("%-10s: %s\n", "ID", user.UserId)
			fmt.Printf("%-8s: %s\n", "姓名", user.Name)
			var s []string
			for _, num := range user.Department {
				s = append(s, strconv.Itoa(num))
			}
			fmt.Printf("%-5s: %s\n", "所属部门ID", strings.Join(s, "、"))
			fmt.Printf("%-8s: %s\n", "职位", user.Position)
			fmt.Printf("%-8s: %s\n", "手机", user.Mobile)
			fmt.Printf("%-8s: %s\n", "邮箱", user.Email)
			fmt.Printf("%-5s: %s\n", "微信二维码", user.QrCode)
			fmt.Printf("%s\n", strings.Repeat("=", 20))
			return
		}
		did, _ := strconv.Atoi(deptId)
		if deptId != "" {
			user, _, err := plugin.GetDepartmentUser(config.Info.AccessToken, did, recurse)
			if err != nil {
				utils.Error(err)
				return
			}
			if len(user) == 0 {
				utils.Warning("无可用用户信息")
				return
			}
			tmpNum := printNum
			if printNum > 0 {
				fmt.Printf("%s\n", strings.Repeat("=", 20))
			}
			for i := 0; i < len(user); i++ {
				if printNum < 1 {
					break
				}
				fmt.Printf("%-10s: %s\n", "ID", user[i].UserId)
				fmt.Printf("%-8s: %s\n", "姓名", user[i].Name)
				var s []string
				for _, num := range user[i].Department {
					s = append(s, strconv.Itoa(num))
				}
				fmt.Printf("%-5s: %s\n", "所属部门ID", strings.Join(s, "、"))
				fmt.Printf("%-8s: %s\n", "职位", user[i].Position)
				fmt.Printf("%-8s: %s\n", "手机", user[i].Mobile)
				fmt.Printf("%-8s: %s\n", "邮箱", user[i].Email)
				fmt.Printf("%-5s: %s\n", "微信二维码", user[i].QrCode)
				printNum--
				if i != len(user)-1 && printNum > 0 {
					fmt.Printf("%s\n", strings.Repeat("-", 20))
				}
			}
			if tmpNum > 0 {
				fmt.Printf("%s\n", strings.Repeat("=", 20))
				utils.Info("更多数据请查看文件")
			}
			utils.Info("正在保存至XLSX文件...")
			msg, err := utils.SaveUserToExcel(user, "dept_users.xlsx")
			if err != nil {
				utils.Error(err)
				return
			}
			utils.Success(msg)
			return
		}
		if dump {
			var departmentTreeResource []*utils.Department
			var departmentList []plugin.WxDepartment
			utils.Info("正在获取部门树...")
			departmentList, _, err := plugin.GetDepartment(config.Info.AccessToken)
			if err != nil {
				utils.Error(err)
				return
			}
			if len(departmentList) == 0 {
				utils.Warning("无可用部门信息")
				return
			}
			for _, v := range departmentList {
				d := utils.Department{
					WxDepartment: plugin.WxDepartment{
						WxSimpleDepartment: plugin.WxSimpleDepartment{
							ID:       v.ID,
							ParentId: v.ParentId,
							Order:    v.Order,
						},
						Name:             v.Name,
						NameEn:           v.NameEn,
						DepartmentLeader: v.DepartmentLeader,
					},
				}
				departmentTreeResource = append(departmentTreeResource, &d)
			}
			departmentTree := utils.BuildDepartmentTree(departmentTreeResource)
			utils.Info("正在获取用户...")
			users, _, err := plugin.GetDepartmentUser(config.Info.AccessToken, departmentTreeResource[0].ID, true)
			if err != nil {
				utils.Error(err)
				return
			}

			// 将用户插入到部门树中
			for _, wxUser := range users {
				departmentIDs := wxUser.Department
				for _, departmentID := range departmentIDs {
					insertUserToDepartmentTree(&wxUser, departmentID, departmentTree)
				}
			}

			//保存文件
			utils.Info("正在保存至HTML文件...")
			msg, err := utils.SaveDepartmentTreeWithUsersToHTML(departmentTree, "dept_tree_users.html")
			if err != nil {
				utils.Error(err)
				return
			}
			utils.Success(msg)
			utils.Info("正在保存至XLSX文件...")
			msg, err = utils.SaveUserTreeToExcel(departmentTree, "dept_tree_users.xlsx")
			if err != nil {
				utils.Error(err)
				return
			}
			utils.Success(msg)
		}
	},
}

var UserAdd = &cobra.Command{Use: "add"}

var UserUpdate = &cobra.Command{Use: "update"}

var UserDelete = &cobra.Command{Use: "delete"}

var UserList = &cobra.Command{Use: "list"}

func init() {
	User.Flags().StringVar(&userId, "uid", "", "用户ID")
	User.Flags().StringVar(&deptId, "did", "", "部门ID")
	User.Flags().BoolVar(&recurse, "re", false, "是否递归获取 (默认 否)")
	User.Flags().BoolVar(&dump, "dump", false, "导出所有部门和用户")
	User.Flags().IntVar(&printNum, "print", 5, "输出到控制台数目的数目 (默认 5),仅对 --did 生效")

	User.AddCommand(UserAdd)
	User.AddCommand(UserUpdate)
	User.AddCommand(UserDelete)
	User.AddCommand(UserList)
	msiSet(User, "user")
	Root.AddCommand(User)

	Dp.AddCommand(DpAdd)
	Dp.AddCommand(DpUpdate)
	Dp.AddCommand(DpDelete)
	DpList.Flags().BoolVar(&printTree, "print", false, "是否打印过程信息 (默认 否)")
	Dp.AddCommand(DpList)
	Dp.AddCommand(DpDetail)
	DpTree.Flags().BoolVar(&printTree, "print", false, "是否打印过程信息 (默认 否)")
	Dp.AddCommand(DpTree)
	msiSet(Dp, "dp")
	Root.AddCommand(Dp)

	Set.AddCommand(CorpId)
	Set.AddCommand(CorpSecret)
	Set.AddCommand(Proxy)
	msiSet(Set, "set")
	Root.AddCommand(Set)

	//msiSet(Help, "")
	Root.AddCommand(Help)

	msiSet(Root, "root")
}

func msiSet(cmd *cobra.Command, name string) {
	// 打印自定义用法
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		if name == "root" {
			ShowAllUsage()
		} else {
			ShowUsageByName(name)
		}
		return nil
	})

	// 不自己打印会多一个空白行
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		//返回error会自动打印用法
		utils.Error(err)
		return nil
	})

	// 禁止打印错误信息
	cmd.SilenceErrors = true

	// 禁用帮助信息
	cmd.DisableSuggestions = true

	//// 禁用用法信息
	//cmd.SilenceUsage = true

	cmd.Flags().SortFlags = false

}

func hasAccessToken() bool {
	return config.Info.AccessToken != ""
}

func insertUserToDepartmentTree(wxUser *plugin.WxUser, departmentID int, departments []*utils.Department) {
	for _, department := range departments {
		if department.ID == departmentID {
			// 创建一个新的用户对象并复制数据
			newUser := *wxUser
			department.User = append(department.User, &newUser)
			return
		}

		insertUserToDepartmentTree(wxUser, departmentID, department.Children)
	}
}
