package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"idebug/logger"
	fs "idebug/plugin/feishu"
	"idebug/utils"
	"os"
	"strings"
	"time"
)

const feishuUsage = mainUsage + `feishu Module:
    关于--dt和--ut说明,飞书部门和用户一般都有两种类型id,用户为user_id和open_id,部门为department_id和open_department_id,下面统一简化为了id和openid,互不影响。注意--dt或者--ut要和提供的<did>或者<uid>实际类型对应,如没有与之对应的<did>或者<uid>参数则表示设置的是返回的部门id或者用户id的类型,例如dp <did> --dt <type> --ut <type>表示:根据did查看部门详情,如果提供的did实际类型为id,则--dt的值应该为id,那么此时的--ut则表示返回的用户类型,按需赋值即可。如果某个id类型获取不到数据请更换类型。
    set appid     <appid>                         设置appid
    set appsecret <appsecret>                     设置appsecret
    run     --dt <type> --ut <type>               获取tenant_access_token
    dp      <did> --dt <type> --ut <type>         根据<did>查看部门详情
    dp ls   <did> --dt <type> --ut <type> [-r]    根据<did>查看子部门列表,-r:递归获取(默认false)
    user    <uid> --dt <type> --ut <type>         根据<uid>查看用户详情
    user ls <did> --dt <type> --ut <type>         根据<did>查看部门直属用户列表,暂不提供递归获取,可使用dump命令代替
    email update --uid <uid> --pass --ut <type>   根据<uid>更新[企业邮箱]密码
    dump         <did> --dt <type> --ut <type>    根据<did>递归导出部门用户,如果能确定授权范围为所有部门请手动赋值为0
`

type feiShuCli struct {
	Root                *cobra.Command
	info                *cobra.Command
	set                 *cobra.Command
	appId               *cobra.Command
	appSecret           *cobra.Command
	run                 *cobra.Command
	dp                  *cobra.Command
	dpLs                *cobra.Command
	user                *cobra.Command
	userLs              *cobra.Command
	email               *cobra.Command
	emailPasswordUpdate *cobra.Command
	dump                *cobra.Command
}

func NewFeiShuCli() *feiShuCli {
	cli := &feiShuCli{}
	cli.Root = cli.newRoot()
	cli.info = cli.newInfo()
	cli.set = cli.newSet()
	cli.appId = cli.newAppId()
	cli.appSecret = cli.newAppSecret()
	cli.run = cli.newRun()
	cli.dp = cli.newDp()
	cli.dpLs = cli.newDpLs()
	cli.user = cli.newUser()
	cli.userLs = cli.newUserLs()
	cli.email = cli.newEmail()
	cli.emailPasswordUpdate = cli.newEmailPasswordUpdate()
	cli.dump = cli.newDump()
	cli.init()
	return cli
}

func (cli *feiShuCli) init() {
	cli.dp.PersistentFlags().StringVar(&departmentIdType, "dt", "", "用户ID类型,可选值: id、openid")
	cli.dp.PersistentFlags().StringVar(&userIdType, "ut", "", "用户ID类型,可选值: id、openid")
	cli.dpLs.Flags().BoolVarP(&recurse, "re", "r", false, "是否递归获取,默认false")

	cli.user.PersistentFlags().StringVar(&departmentIdType, "dt", "", "用户IID类型,可选值: id、openid")
	cli.user.PersistentFlags().StringVar(&userIdType, "ut", "", "用户ID类型,可选值: id、openid")
	//TODO: cli.userLs.Flags().BoolVarP(&recurse, "re", "r", false, "是否递归获取,默认false")

	cli.emailPasswordUpdate.Flags().StringVar(&userIdType, "ut", "", "用户ID类型,可选值: id、openid")
	cli.emailPasswordUpdate.Flags().StringVar(&password, "pass", "", "企业邮箱新密码")
	cli.emailPasswordUpdate.Flags().StringVar(&userId, "uid", "", "用户ID")
	cli.emailPasswordUpdate.MarkFlagRequired("pass")
	cli.emailPasswordUpdate.MarkFlagRequired("uid")
	cli.emailPasswordUpdate.MarkFlagRequired("ut")

	cli.run.Flags().StringVar(&userIdType, "ut", "", "用户ID类型,可选值: id、openid")
	cli.run.Flags().StringVar(&departmentIdType, "dt", "", "部门ID类型,可选值: id、openid")

	cli.dump.Flags().StringVar(&userIdType, "ut", "", "用户ID类型,可选值: id、openid")
	cli.dump.Flags().StringVar(&departmentIdType, "dt", "", "部门ID类型,可选值: id、openid")
	cli.dump.MarkFlagRequired("uid")
	cli.dump.MarkFlagRequired("ut")

	cli.set.AddCommand(cli.appId, cli.appSecret, newProxy())
	cli.dp.AddCommand(cli.dpLs)
	cli.user.AddCommand(cli.userLs)
	cli.email.AddCommand(cli.emailPasswordUpdate)
	cli.Root.AddCommand(cli.set, cli.run, cli.info, cli.dp, cli.user, cli.email, cli.dump)

	cli.setHelpV1(cli.Root, cli.info, cli.set, cli.appId, cli.appSecret, cli.run, cli.dp, cli.dpLs, cli.user, cli.userLs, cli.email, cli.emailPasswordUpdate, cli.dump)
}

func (cli *feiShuCli) newRoot() *cobra.Command {
	return &cobra.Command{
		Use:   "feishu",
		Short: `飞书模块`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if *CurrentModule != FeiShuModule {
				return logger.FormatError(fmt.Errorf("请先设置模块为feishu"))
			}
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			reset()
		},
	}
}

func (cli *feiShuCli) newInfo() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: `查看设置`,
		Run: func(cmd *cobra.Command, args []string) {
			cli.showClientConfig()
		},
	}
}

func (cli *feiShuCli) newSet() *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: `设置参数`,
	}
}

func (cli *feiShuCli) newAppId() *cobra.Command {
	return &cobra.Command{
		Use:   "appid",
		Short: `设置app_id`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				logger.Warning("请提供一个值")
				return
			}
			FeiShuClient.SetAppId(args[0])
			logger.Success("appid => " + args[0])
		},
	}
}

func (cli *feiShuCli) newAppSecret() *cobra.Command {
	return &cobra.Command{
		Use:   "appsecret",
		Short: `设置app_secret`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				logger.Warning("请提供一个值")
				return
			}
			FeiShuClient.SetAppSecret(args[0])
			logger.Success("appsecret => " + args[0])
		},
	}

}

func (cli *feiShuCli) newRun() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: `根据设置获取tenant_access_token`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			conf := FeiShuClient.GetAuthScopeFromCache()
			if conf.AppId == nil || *conf.AppId == "" {
				return fmt.Errorf("请先设置appid")

			}
			if conf.AppSecret == nil || *conf.AppSecret == "" {
				return fmt.Errorf("请先设置appsecret")
			}
			if err := cli.checkIdType(); err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("获取信息中,请稍等...")
			req := fs.NewGetAuthScopeReqBuilder(FeiShuClient).
				UserIdType(userIdTypeMap[userIdType]).
				DepartmentIdType(departmentIdTypeMap[departmentIdType]).
				Build()
			_, err := FeiShuClient.GetNewAuthScope(req)
			if err != nil {
				logger.Warning("获取tenant_access_token通信录授权范围失败")
				logger.Error(logger.FormatError(err))
				return
			}
			departmentIdTypeCache = departmentIdType
			userIdTypeCache = userIdType
			cli.showClientConfig()
		},
	}
}

func (cli *feiShuCli) newDp() *cobra.Command {
	return &cobra.Command{
		Use:   "dp",
		Short: `部门操作`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !cli.hasTenantAccessToken() {
				return fmt.Errorf("请先执行run获取tenant_access_token")
			}
			if err := cli.checkIdType(); err != nil {
				return err
			}
			if len(args) == 0 {
				return fmt.Errorf("请提供一个参数作为部门ID或者提供一个子命令")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var userIdTypeValue = strings.ToUpper(userIdTypeMap[userIdTypeCache])
			var deptIdTypeValue = strings.ToUpper(departmentIdTypeMap[departmentIdTypeCache])
			var uidType = userIdTypeMap[userIdType]
			var didType = departmentIdTypeMap[departmentIdType]
			var (
				status             string
				name               string
				nameEnUs           string
				nameJaJp           string
				did                string
				oid                string
				parentName         string
				parentId           string
				leaderUserId       string
				leaderUserName     string
				leaders            []string
				memberCount        int
				primaryMemberCount int
				hrbps              []string
			)
			req := fs.NewGetDepartmentReqBuilder(FeiShuClient).
				DepartmentId(args[0]).
				DepartmentIdType(didType).
				UserIdType(uidType).
				Build()
			deptInfo, err := FeiShuClient.Department.Get(req)
			if err != nil {
				logger.Error(logger.FormatError(err))
				return
			}
			if deptInfo.Status.IsDeleted {
				status = "已删除"
			} else {
				status = "正常"
			}
			did = deptInfo.DepartmentID
			oid = deptInfo.OpenDepartmentID
			name = deptInfo.Name
			nameEnUs = deptInfo.I18NName.EnUs
			nameJaJp = deptInfo.I18NName.JaJp
			leaderUserId = deptInfo.LeaderUserID
			parentId = deptInfo.ParentDepartmentID

			//获取上级部门信息
			req = fs.NewGetDepartmentReqBuilder(FeiShuClient).
				DepartmentId(deptInfo.ParentDepartmentID).
				DepartmentIdType(departmentIdTypeMap["id"]).
				UserIdType(uidType).
				Build()
			parentDeptInfo, err := FeiShuClient.Department.Get(req)
			if err == nil {
				parentName = parentDeptInfo.Name
			}
			if leaderUserId != "" {
				req := fs.NewGetUserReqBuilder(FeiShuClient).
					UserId(leaderUserId).
					UserIdType(uidType).
					Build()
				leaderUserInfo, err := FeiShuClient.User.Get(req)
				if err == nil {
					leaderUserName = leaderUserInfo.Name
				}
			}
			for _, leader := range deptInfo.Leaders {
				req := fs.NewGetUserReqBuilder(FeiShuClient).
					UserId(leader.LeaderID).
					UserIdType(uidType).
					Build()
				leaderInfo, err := FeiShuClient.User.Get(req)
				if err != nil {
					leaders = append(leaders, fmt.Sprintf("%s: %s", userIdTypeValue, leader.LeaderID))
					continue
				}
				leaders = append(leaders, fmt.Sprintf("%s(%s: %s)", leaderInfo.Name, userIdTypeValue, leader.LeaderID))
			}
			memberCount = deptInfo.MemberCount
			primaryMemberCount = deptInfo.PrimaryMemberCount
			if deptInfo.DepartmentHrbps != nil {
				for _, hrbp := range deptInfo.DepartmentHrbps {
					hrbps = append(hrbps, *hrbp)
				}

			}
			fmt.Printf("%s\n", strings.Repeat("=", 20))
			fmt.Println(fmt.Sprintf("%-12s: %s", "状态", status))
			fmt.Println(fmt.Sprintf("%-12s: %s", "名称", name))
			if nameEnUs != "" {
				fmt.Println(fmt.Sprintf("%-11s: %s", "英文名称", nameEnUs))
			}
			if nameJaJp != "" {
				fmt.Println(fmt.Sprintf("%-11s: %s", "日文名称", nameJaJp))
			}
			fmt.Println(fmt.Sprintf("%-14s: %s", "ID", did))
			fmt.Println(fmt.Sprintf("%-14s: %s", "OPEN_ID", oid))
			fmt.Println(fmt.Sprintf("%-10s: %s", "上级部门", fmt.Sprintf("%s(%s: %s)", parentName,
				deptIdTypeValue, parentId)))
			if leaderUserId != "" {
				fmt.Println(fmt.Sprintf("%-10s: %s", "主管用户", fmt.Sprintf("%s(%s: %s)", leaderUserName, userIdTypeValue, leaderUserId)))
			} else {
				fmt.Println(fmt.Sprintf("%-10s: %s", "主管用户", ""))
			}
			if len(leaders) != 0 {
				fmt.Println(fmt.Sprintf("%-11s: %s", "负责人", strings.Join(leaders, "、")))
			} else {
				fmt.Println(fmt.Sprintf("%-11s: %s", "负责人", ""))
			}
			fmt.Println(fmt.Sprintf("%-10s: %d", "用户个数", memberCount))
			fmt.Println(fmt.Sprintf("%-8s: %d", "主属用户个数", primaryMemberCount))
			if len(hrbps) > 0 {
				fmt.Println(fmt.Sprintf("%-12s: %s", "部门HRBP", strings.Join(hrbps, "、")))
			} else {
				fmt.Println(fmt.Sprintf("%-12s: %s", "部门HRBP", ""))
			}
			fmt.Printf("%s\n", strings.Repeat("=", 20))
		},
	}
}

func (cli *feiShuCli) newDpLs() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: `根据部门ID获取子部门列表`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("请提供一个参数作为部门ID")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var index = 0
			var depts []*fs.DepartmentEntry
			err := cli.recursePrintDept(depts, args[0], departmentIdTypeMap[departmentIdType], userIdTypeMap[userIdType], 0, &index)
			if err != nil {
				logger.Error(logger.FormatError(err))
				return
			}
		},
	}
}

func (cli *feiShuCli) newUser() *cobra.Command {
	return &cobra.Command{
		Use:   "user",
		Short: `用户操作`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !cli.hasTenantAccessToken() {
				return fmt.Errorf("请先执行run获取tenant_access_token")
			}
			if err := cli.checkIdType(); err != nil {
				return err
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("请提供一个参数作为用户ID或者提供一个子命令")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			req := fs.NewGetUserReqBuilder(FeiShuClient).UserId(args[0]).
				UserIdType(userIdTypeMap[userIdType]).
				DepartmentIdType(departmentIdTypeMap[departmentIdType]).Build()
			userInfo, err := FeiShuClient.User.Get(req)
			if err != nil {
				logger.Error(logger.FormatError(err))
				return
			}
			cli.showUserInfo(*userInfo, false)
		}}
}

func (cli *feiShuCli) newUserLs() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: `根据部门ID获取用户列表`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !cli.hasTenantAccessToken() {
				return fmt.Errorf("请先执行run获取tenant_access_token")
			}
			if err := cli.checkIdType(); err != nil {
				return err
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("请提供一个参数作为部门ID")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			req := fs.NewGetUsersByDepartmentIdReqBuilder(FeiShuClient).
				DepartmentId(args[0]).
				DepartmentIdType(departmentIdTypeMap[departmentIdType]).
				UserIdType(userIdTypeMap[userIdType]).
				PageSize(50).
				Build()
			userList, err := FeiShuClient.User.GetUsersByDepartmentId(req)
			if err != nil {
				logger.Error(logger.FormatError(err))
				return
			}
			length := len(userList)
			if length == 0 {
				logger.Info("无可用数据")
				return
			}
			for _, userInfo := range userList {
				//if (verbose > i && verbose >= 0) || (verbose < 0) {
				//	cli.showUserInfo(*userInfo, true)
				//}
				cli.showUserInfo(*userInfo, true)
			}
		},
	}
}

func (cli *feiShuCli) newEmail() *cobra.Command {
	return &cobra.Command{
		Use:   "email",
		Short: `企业邮箱操作`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !cli.hasTenantAccessToken() {
				return fmt.Errorf("请先执行run获取tenant_access_token")
			}
			var userIdTypeList []string
			for idType := range userIdTypeMap {
				userIdTypeList = append(userIdTypeList, idType)
			}
			if !utils.StringInList(userIdType, userIdTypeList) {
				return errors.New("--ut :未设置或者是错误的用户ID类型,可选值：" + strings.Join(userIdTypeList, "、"))
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
}

func (cli *feiShuCli) newEmailPasswordUpdate() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: `更新密码`,
		Run: func(cmd *cobra.Command, args []string) {
			req := fs.NewUserEmailPasswordChangeReqBuilder(FeiShuClient).
				UserIdType(userIdTypeMap[userIdType]).
				PostData(userId, password).
				Build()
			err := FeiShuClient.User.EmailPasswordUpdate(req)
			if err != nil {
				logger.Error(logger.FormatError(err))
				return
			}
			logger.Success("修改成功")
		},
	}
}

func (cli *feiShuCli) newDump() *cobra.Command {
	return &cobra.Command{
		Use:   "dump",
		Short: `根据部门ID导出用户,不提供部门ID则导出通讯录授权范围内所有部门用户`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !cli.hasTenantAccessToken() {
				return fmt.Errorf("请先执行run获取tenant_access_token")
			}
			if err := cli.checkIdType(); err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var deptNodeList []*FeiShuDepartmentNode
			var uidType = userIdTypeMap[userIdType]
			var didType = departmentIdTypeMap[departmentIdType]
			conf := FeiShuClient.GetAuthScopeFromCache()
			//根据指定部门ID先获取子部门列表
			var isTerminal bool
			var deptIds []string
			if len(args) == 0 {
				for deptId := range conf.DepartmentScope {
					deptIds = append(deptIds, deptId)
				}
			} else {
				deptIds = append(deptIds, args[0])
			}
			for _, deptId := range deptIds {
				deptNode := &FeiShuDepartmentNode{
					UnitIds:         []*string{},
					DepartmentHrbps: []*string{},
					User:            []*fs.UserEntry{},
					Children:        []*FeiShuDepartmentNode{},
				}

				var deptInfo fs.DepartmentEntry
				var err error
				for i := 0; i < retry; i++ {
					if i == 0 {
						logger.Info(fmt.Sprintf("正在获取部门[%s]的信息....", deptId))
					}
					//获取部门信息
					req := fs.NewGetDepartmentReqBuilder(FeiShuClient).
						DepartmentId(deptId).
						DepartmentIdType(didType).
						UserIdType(uidType).
						Build()
					deptInfo, err = FeiShuClient.Department.Get(req)
					if err != nil {
						if i == retry-1 {
							logger.Error(logger.FormatError(err))
							logger.Info(fmt.Sprintf("部门[%s]信息获取失败,终止获取", deptId))
							return
						}
						logger.Error(logger.FormatError(err))
						logger.Info(fmt.Sprintf("部门[%s]信息获取失败,正在重试...", deptId))
						time.Sleep(FeiShuDefaultInterval)
						continue
					}
					break
				}
				deptNode.Name = deptInfo.Name
				deptNode.ZhCnName = deptInfo.I18NName.ZhCn
				deptNode.JaJpName = deptInfo.I18NName.JaJp
				deptNode.EnUsName = deptInfo.I18NName.EnUs
				deptNode.DepartmentID = deptInfo.DepartmentID
				deptNode.OpenDepartmentID = deptInfo.OpenDepartmentID
				deptNode.ParentDepartmentID = deptInfo.ParentDepartmentID
				//deptNode.ParentDepartmentName = deptInfo
				//deptNode.Status = deptInfo
				deptNode.LeaderUserID = deptInfo.LeaderUserID
				//deptNode.LeaderUserName = deptInfo
				//deptNode.ChatID = deptInfo

				if deptInfo.Status.IsDeleted {
					deptNode.Status = "已删除"
				} else {
					deptNode.Status = "正常"
				}
				logger.Info(fmt.Sprintf("正在获取部门[%s]主管用户信息...", deptId))
				//获取部门主管领导信息
				if deptInfo.LeaderUserID != "" {
					for i := 0; i < retry; i++ {
						req := fs.NewGetUserReqBuilder(FeiShuClient).
							UserId(deptInfo.LeaderUserID).
							UserIdType(uidType).
							DepartmentIdType(didType).
							Build()
						userInfo, err := FeiShuClient.User.Get(req)
						if err != nil {
							if i == retry-1 {
								logger.Error(logger.FormatError(err))
								logger.Info(fmt.Sprintf("部门[%s]主管用户信息获取失败,将会继续执行...", deptId))
								break
							}
							logger.Error(logger.FormatError(err))
							logger.Info(fmt.Sprintf("部门[%s]主管用户信息获取失败,正在重试...", deptId))
							time.Sleep(FeiShuDefaultInterval)
							continue
						}
						deptNode.LeaderUserName = userInfo.Name
						break
					}
				}

				deptNode.ChatID = deptInfo.ChatID

				//获取部门用户
				logger.Info(fmt.Sprintf("正在获取部门[%s]直属用户列表...", deptId))
				for i := 0; i < retry; i++ {
					req1 := fs.NewGetUsersByDepartmentIdReqBuilder(FeiShuClient).
						DepartmentId(deptId).
						DepartmentIdType(didType).UserIdType(uidType).Build()
					users, err := FeiShuClient.User.GetUsersByDepartmentId(req1)
					if err != nil {
						if i == retry-1 {
							logger.Error(logger.FormatError(err))
							logger.Info(fmt.Sprintf("部门[%s]直属用户列表获取失败,终止获取", deptId))
							return
						}
						logger.Error(logger.FormatError(err))
						logger.Info(fmt.Sprintf("部门[%s]直属用户列表获取失败,正在重试...", deptId))
						time.Sleep(FeiShuDefaultInterval)
						continue
					}
					deptNodeList = append(deptNodeList, deptNode)
					deptNode.User = append(deptNode.User, users...)
					err = cli.fetchDepartment(deptNode, deptId, departmentIdTypeMap[departmentIdType], userIdTypeMap[userIdType])
					if err != nil {
						isTerminal = true
						logger.Error(logger.FormatError(err))
					}
					break
				}
				if isTerminal {
					break
				}
			}
			if isTerminal {
				logger.Warning("终止获取,会保存已获取数据")
			}
			logger.Info("正在保存至html文件...")

			msg, err := cli.saveDepartmentTreeWithUsersToHTML(deptNodeList, "feishu_dump.html")
			if err != nil {
				logger.Error(logger.FormatError(err))
				logger.Info("保存文件失败")
			} else {
				logger.Success(msg)
			}
			logger.Info("正在保存至xlsx文件...")
			msg, err = cli.saveDepartmentWithUserToExcel(deptNodeList, "feishu_dump.xlsx")
			if err != nil {
				logger.Error(logger.FormatError(err))
				logger.Info("保存文件失败")
			} else {
				logger.Success(msg)
			}

		},
	}
}

func (cli *feiShuCli) checkIdType() error {
	var deptIdTypeList []string
	var userIdTypeList []string
	for idType := range departmentIdTypeMap {
		deptIdTypeList = append(deptIdTypeList, idType)
	}
	for idType := range userIdTypeMap {
		userIdTypeList = append(userIdTypeList, idType)
	}
	if !utils.StringInList(departmentIdType, deptIdTypeList) {
		return errors.New("--dt :未设置或者是错误的部门ID类型,可选值：" + strings.Join(deptIdTypeList, "、"))
	}
	if !utils.StringInList(userIdType, userIdTypeList) {
		return errors.New("--ut :未设置或者是错误的用户ID类型,可选值：" + strings.Join(userIdTypeList, "、"))
	}
	return nil
}

func (cli *feiShuCli) hasTenantAccessToken() bool {
	return FeiShuClient.GetTenantAccessTokenFromCache() != ""
}

func (cli *feiShuCli) showUserInfo(userInfo fs.UserEntry, inLine bool) {
	var (
		name            string
		uid             string
		oid             string
		gender          string
		employeeNo      string
		status          []string
		phone           string
		email           string
		enterpriseEmail string
		isTenantManager string
		workLocation    string
		depts           []string
	)
	name = userInfo.Name
	uid = userInfo.UserId
	oid = userInfo.OpenId
	if userInfo.Gender == 0 {
		gender = "保密"
	} else if userInfo.Gender == 1 {
		gender = "男"
	} else if userInfo.Gender == 2 {
		gender = "女"
	}
	employeeNo = userInfo.EmployeeNo
	email = userInfo.Email
	enterpriseEmail = userInfo.EnterpriseEmail
	if userInfo.Status.IsUnjoin {
		status = append(status, "未加入")
	}
	if userInfo.Status.IsActivated {
		status = append(status, "已激活")
	}
	if userInfo.Status.IsResigned {
		status = append(status, "已离职")
	}
	if userInfo.Status.IsFrozen {
		status = append(status, "已冻结")
	}
	if userInfo.Status.IsExited {
		status = append(status, "已退出")
	}
	phone = userInfo.Mobile
	if userInfo.IsTenantManager {
		isTenantManager = "是"
	} else if userInfo.IsTenantManager {
		isTenantManager = "否"
	}
	for _, deptId := range userInfo.DepartmentIds {
		req := fs.NewGetDepartmentReqBuilder(FeiShuClient).DepartmentId(deptId).DepartmentIdType(departmentIdTypeMap[departmentIdType]).Build()
		deptInfo, err := FeiShuClient.Department.Get(req)
		if err != nil {
			time.Sleep(FeiShuDefaultInterval)
			continue
		}
		var info string
		if deptInfo.Name != "" {
			info = fmt.Sprintf("%s(%s: %s)", deptInfo.Name,
				strings.ToUpper(departmentIdTypeMap["id"]), deptInfo.DepartmentID)
		} else if deptInfo.Name == "" {
			info = fmt.Sprintf("%s: %s",
				strings.ToUpper(departmentIdTypeMap["id"]), deptInfo.DepartmentID)
		}
		depts = append(depts, info)
		time.Sleep(FeiShuDefaultInterval)
	}
	if userInfo.Country != "" {
		workLocation += userInfo.Country + " "
	}
	if userInfo.City != "" {
		workLocation += userInfo.City
	}
	if inLine {
		s := fmt.Sprintf("  -ID[%s] OPEN_ID[%s] 姓名[%s] 性别[%s] 工号[%s] 手机号码[%s] 邮箱[%s] 企业邮箱[%s] 状态[%s] 是否企业管理员[%s] 所属部门[%s] 工作地点[%s]", uid, oid, name, gender, employeeNo, phone, email, enterpriseEmail, strings.Join(status, "、"), isTenantManager, strings.Join(depts, "、"), workLocation)
		fmt.Println(s)
		return
	}
	fmt.Printf("%s\n", strings.Repeat("=", 20))
	fmt.Printf("%-14s: %s\n", "USER_ID", uid)
	fmt.Printf("%-14s: %s\n", "OPEN_ID", oid)
	fmt.Printf("%-12s: %s\n", "姓名", name)
	fmt.Printf("%-12s: %s\n", "性别", gender)
	fmt.Printf("%-12s: %s\n", "工号", employeeNo)
	fmt.Printf("%-10s: %s\n", "手机号码", phone)
	fmt.Printf("%-12s: %s\n", "邮箱", email)
	fmt.Printf("%-10s: %s\n", "企业邮箱", enterpriseEmail)
	fmt.Printf("%-12s: %s\n", "状态", strings.Join(status, "、"))
	fmt.Printf("%-9s: %s\n", "是否管理员", isTenantManager)
	fmt.Printf("%-10s: %s\n", "所属部门", strings.Join(depts, "、"))
	fmt.Printf("%-10s: %s\n", "工作地点", workLocation)
	fmt.Printf("%s\n", strings.Repeat("=", 20))
}

func (cli *feiShuCli) showClientConfig() {
	fsClientConfig := FeiShuClient.GetAuthScopeFromCache()
	fmt.Printf("%s\n", strings.Repeat("=", 40))
	if Proxy == nil {
		fmt.Println(fmt.Sprintf("%-17s: %s", "proxy", ""))
	} else {
		fmt.Println(fmt.Sprintf("%-17s: %s", "proxy", *Proxy))
	}
	if fsClientConfig.AppId == nil {
		fmt.Println(fmt.Sprintf("%-17s: %s", "app_id", ""))
	} else {
		fmt.Println(fmt.Sprintf("%-17s: %s", "app_id", *fsClientConfig.AppId))
	}
	if fsClientConfig.AppSecret == nil {
		fmt.Println(fmt.Sprintf("%-17s: %s", "app_secret", ""))
	} else {
		fmt.Println(fmt.Sprintf("%-17s: %s", "app_secret", *fsClientConfig.AppSecret))
	}
	if fsClientConfig.TenantAccessToken == nil {
		fmt.Println(fmt.Sprintf("%-17s: %s", "tenant_access_token", ""))
	} else {
		fmt.Println(fmt.Sprintf("%-17s: %s", "tenant_access_token", *fsClientConfig.TenantAccessToken))
	}
	var deptScope []string
	for deptId, deptName := range fsClientConfig.DepartmentScope {
		var deptIdType = "ID"
		if departmentIdTypeCache == "open_department_id" {
			deptIdType = "OPEN_ID"
		}
		if deptName != "" {
			deptScope = append(deptScope, fmt.Sprintf("%s(%s:%s)", deptName, deptIdType, deptId))
		} else {
			deptScope = append(deptScope, fmt.Sprintf("%s:%s", deptIdType, deptId))
		}

	}
	fmt.Println("-----------获取通讯录授权范围-----------")
	logger.Notice("该接口用于获取应用被授权可访问的通讯录范围，包括可访问的部门列表、用户列表和用户组列表。授权范围为全员时，" +
		"返回的部门列表为该企业所有的一级部门；否则返回的部门为管理员在设置授权范围时勾选的部门（不包含勾选部门的子部门）")
	fmt.Println(fmt.Sprintf("%-13s: %s", "部门范围", strings.Join(deptScope, "、")))
	var groupScope []string
	for groupId, groupName := range fsClientConfig.GroupScope {
		if groupName != "" {
			groupScope = append(groupScope, fmt.Sprintf("%s(ID:%s)", groupName, groupId))
		} else {
			groupScope = append(groupScope, fmt.Sprintf("ID:%s", groupId))
		}
	}
	fmt.Println(fmt.Sprintf("%-12s: %s", "用户组范围", strings.Join(groupScope, "、")))
	var userScope []string
	for id, userName := range fsClientConfig.UserScope {
		var userIdType = "ID"
		if userIdTypeCache == "open_id" {
			userIdType = "OPEN_ID"
		}
		if userName != "" {
			userScope = append(userScope, fmt.Sprintf("%s(%s:%s)", userName, userIdType, id))
		} else {
			userScope = append(userScope, fmt.Sprintf("%s:%s", userIdType, id))
		}

	}
	fmt.Println(fmt.Sprintf("%-13s: %s", "用户范围", strings.Join(userScope, "、")))
	fmt.Printf("%s\n", strings.Repeat("=", 40))
}

func (cli *feiShuCli) fetchDepartment(node *FeiShuDepartmentNode, deptId, deptIdType, userIdType string) error {
	var deptChildren []*fs.DepartmentEntry
	var err error
	logger.Info(fmt.Sprintf("正在获取部门[%s]的子部门...", deptId))
	for i := 0; i < retry; i++ {
		req := fs.NewGetDepartmentChildrenReqBuilder(FeiShuClient).
			DepartmentId(deptId).
			DepartmentIdType(deptIdType).
			UserIdType(userIdType).
			Fetch(false).
			PageSize(50).
			Build()
		deptChildren, err = FeiShuClient.Department.Children(req)
		if err != nil {
			if i == retry-1 {
				logger.Error(logger.FormatError(err))
				logger.Info(fmt.Sprintf("部门[%s]的子部门获取失败,终止获取", deptId))
				return errors.New("")
			}
			logger.Error(logger.FormatError(err))
			logger.Info(fmt.Sprintf("部门[%s]的子部门获取失败,正在重试...", deptId))
			time.Sleep(FeiShuDefaultInterval)
			continue
		}
		break
	}

	for _, child := range deptChildren {
		var id string
		if deptIdType == "department_id" {
			id = child.DepartmentID
		} else {
			id = child.OpenDepartmentID
		}
		req := fs.NewGetDepartmentReqBuilder(FeiShuClient).
			DepartmentId(id).
			DepartmentIdType(deptIdType).
			UserIdType(userIdType).
			Build()
		deptInfo, err := FeiShuClient.Department.Get(req)
		if err != nil {
			logger.Error(logger.FormatError(err))
			return errors.New("")
		}
		deptNode := &FeiShuDepartmentNode{
			Name:                 deptInfo.Name,
			ZhCnName:             deptInfo.I18NName.ZhCn,
			JaJpName:             deptInfo.I18NName.JaJp,
			EnUsName:             deptInfo.I18NName.EnUs,
			DepartmentID:         deptInfo.DepartmentID,
			OpenDepartmentID:     deptInfo.OpenDepartmentID,
			ParentDepartmentID:   deptInfo.ParentDepartmentID,
			ParentDepartmentName: node.Name,
			//Status:               "",
			LeaderUserID: deptInfo.LeaderUserID,
			//LeaderUserName:       "",
			ChatID:          deptInfo.ChatID,
			UnitIds:         []*string{},
			DepartmentHrbps: []*string{},
			User:            []*fs.UserEntry{},
			Children:        []*FeiShuDepartmentNode{},
		}
		if deptInfo.Status.IsDeleted {
			deptNode.Status = "已删除"
		} else {
			deptNode.Status = "正常"
		}

		//获取部门主管领导信息
		logger.Info(fmt.Sprintf("正在获取部门[%s]主管用户信息...", id))
		if deptInfo.LeaderUserID != "" {
			for i := 0; i < retry; i++ {
				req := fs.NewGetUserReqBuilder(FeiShuClient).
					UserId(deptInfo.LeaderUserID).
					UserIdType(userIdType).
					DepartmentIdType(deptIdType).
					Build()
				userInfo, err := FeiShuClient.User.Get(req)
				if err != nil {
					if i == retry-1 {
						logger.Error(logger.FormatError(err))
						logger.Info(fmt.Sprintf("部门[%s]主管用户信息获取失败,将会继续执行...", id))
						break
					}
					logger.Error(logger.FormatError(err))
					logger.Info(fmt.Sprintf("部门[%s]主管用户信息获取失败,正在重试...", id))
					time.Sleep(FeiShuDefaultInterval)
					continue
				}
				deptNode.LeaderUserName = userInfo.Name
				break
			}
		}

		//获取部门用户
		logger.Info(fmt.Sprintf("正在获取部门[%s]直属用户列表...", id))
		for i := 0; i < retry; i++ {
			req1 := fs.NewGetUsersByDepartmentIdReqBuilder(FeiShuClient).
				DepartmentId(id).
				DepartmentIdType(deptIdType).UserIdType(userIdType).Build()
			users, err := FeiShuClient.User.GetUsersByDepartmentId(req1)
			if err != nil {
				if i == retry-1 {
					logger.Error(logger.FormatError(err))
					logger.Info(fmt.Sprintf("部门[%s]直属用户列表获取失败,终止获取", id))
					return errors.New("")
				}
				logger.Error(logger.FormatError(err))
				logger.Info(fmt.Sprintf("部门[%s]直属用户列表获取失败,正在重试...", id))
				time.Sleep(FeiShuDefaultInterval)
				continue
			}
			node.Children = append(node.Children, deptNode)
			deptNode.User = append(deptNode.User, users...)
			err = cli.fetchDepartment(deptNode, id, deptIdType, userIdType)
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (cli *feiShuCli) setHelpV1(cmds ...*cobra.Command) {
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
			fmt.Print(feishuUsage)
		})
		cmd.SetUsageFunc(func(cmd *cobra.Command) error {
			fmt.Print(feishuUsage)
			return nil
		})
	}
}

type FeiShuDepartmentNode struct {
	Name                 string // 部门名称
	ZhCnName             string // 部门的中文名
	JaJpName             string // 部门的日文名
	EnUsName             string // 部门的英文名
	DepartmentID         string // 部门ID
	OpenDepartmentID     string // 部门open_department_id
	ParentDepartmentID   string // 上级部门ID
	ParentDepartmentName string // 上级部门名称
	Status               string // 部门状态
	LeaderUserID         string // 主管领导ID
	LeaderUserName       string // 主管领导姓名
	ChatID               string // 部门群ID
	UnitIds              []*string
	DepartmentHrbps      []*string
	User                 []*fs.UserEntry
	Children             []*FeiShuDepartmentNode
}

type colItem struct {
	Name                 string // 部门名称
	ZhCnName             string // 部门的中文名
	JaJpName             string // 部门的日文名
	EnUsName             string // 部门的英文名
	DepartmentID         string // 部门ID
	OpenDepartmentID     string // 部门open_department_id
	ParentDepartmentID   string // 上级部门ID
	ParentDepartmentName string // 上级部门名称
	Status               string // 部门状态
	LeaderUserID         string // 主管领导ID
	LeaderUserName       string // 主管领导姓名
	DepartmentHrbps      []*string
	User                 fs.UserEntry
}

func (cli *feiShuCli) fetchColItem(tree []*FeiShuDepartmentNode) []*colItem {
	var items []*colItem
	for _, node := range tree {
		for _, user := range node.User {
			var item = &colItem{
				Name:                 node.Name,
				ZhCnName:             node.ZhCnName,
				JaJpName:             node.JaJpName,
				EnUsName:             node.EnUsName,
				DepartmentID:         node.DepartmentID,
				OpenDepartmentID:     node.OpenDepartmentID,
				ParentDepartmentID:   node.ParentDepartmentID,
				ParentDepartmentName: node.ParentDepartmentName,
				Status:               node.Status,
				LeaderUserID:         node.LeaderUserID,
				LeaderUserName:       node.LeaderUserName,
				DepartmentHrbps:      node.DepartmentHrbps,
				User:                 *user,
			}
			items = append(items, item)
		}
		items = append(items, cli.fetchColItem(node.Children)...)
	}
	return items
}

// saveDepartmentWithUserToExcel 生成包含所属部门ID的用户信息的XLSX文档,保存文件为空则自动生成
func (cli *feiShuCli) saveDepartmentWithUserToExcel(tree []*FeiShuDepartmentNode, filename string) (string, error) {
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
	var items []*colItem
	items = append(items, cli.fetchColItem(tree)...)
	// 设置表头
	headers := []any{"id", "部门名称", "部门中文名称", "部门日文名称", "部门英文名称", "部门ID", "部门OPEN_ID", "上级部门ID", "上级部门名称", "部门状态", "部门主管ID", "部门主管姓名", "Hrbps", "用户ID", "用户OPEN_ID",
		"姓名", "英文姓名", "昵称", "性别", "电话号码", "邮箱", "企业邮箱", "用户状态", "用户所属部门", "工作地点", "加入时间", "工号", "是否企业管理员", "职称"}

	var data [][]any
	// 写入内容
	for i, item := range items {
		var gender string
		user := item.User
		if user.Gender == 0 {
			gender = "保密"
		} else if user.Gender == 1 {
			gender = "男"
		} else if user.Gender == 2 {
			gender = "女"
		}
		var hrbps []string
		for _, hrbp := range item.DepartmentHrbps {
			hrbps = append(hrbps, *hrbp)
		}
		var userStat []string
		if user.Status.IsUnjoin {
			userStat = append(userStat, "未加入")
		}
		if user.Status.IsResigned {
			userStat = append(userStat, "已离职")
		}
		if user.Status.IsActivated {
			userStat = append(userStat, "已激活")
		}
		if user.Status.IsExited {
			userStat = append(userStat, "已退出")
		}
		if user.Status.IsFrozen {
			userStat = append(userStat, "已冻结")
		}
		var workLocation string
		if user.Country != "" {
			workLocation += user.Country
		}
		if user.City != "" {
			workLocation += user.City
		}
		var isAdmin string
		if user.IsTenantManager {
			isAdmin = "是"
		}
		if !user.IsTenantManager {
			isAdmin = "否"
		}
		timestamp := int64(user.JoinTime)   // 已知的时间戳（以秒为单位）
		joinTime := time.Unix(timestamp, 0) // 使用time.Unix()函数将时间戳转换为本地时间
		var (
			deptName             = item.Name                 // 部门名称
			zhCnName             = item.ZhCnName             // 部门的中文名
			jaJpName             = item.JaJpName             // 部门的日文名
			enUsName             = item.EnUsName             // 部门的英文名
			departmentID         = item.DepartmentID         // 部门ID
			openDepartmentID     = item.OpenDepartmentID     // 部门open_department_id
			parentDepartmentID   = item.ParentDepartmentID   // 上级部门ID
			parentDepartmentName = item.ParentDepartmentName // 上级部门名称
			status               = item.Status               // 部门状态
			leaderUserID         = item.LeaderUserID         // 主管领导ID
			leaderUserName       = item.LeaderUserName       // 主管领导姓名
			departmentHrbps      = strings.Join(hrbps, "、")
			uid                  = user.UserId
			userOpenId           = user.OpenId
			username             = user.Name
			usernameEn           = user.EnName
			userNikname          = user.Nickname
			userGender           = gender
			userPhone            = user.Mobile
			userEmail            = user.Email
			userEnterpriseEmail  = user.EnterpriseEmail
			userStatus           = strings.Join(userStat, "、")
			userDepts            = strings.Join(user.DepartmentIds, "、")
			userWorkLocation     = workLocation
			userJoinTime         = joinTime.String()
			userEmployeeNo       = user.EmployeeNo
			userIsAdmin          = isAdmin
			userJobTitle         = user.JobTitle
		)
		d := []any{
			i + 1,
			deptName,
			zhCnName,
			jaJpName,
			enUsName,
			departmentID,
			openDepartmentID,
			parentDepartmentID,
			parentDepartmentName,
			status,
			leaderUserID,
			leaderUserName,
			departmentHrbps,
			uid,
			userOpenId,
			username,
			usernameEn,
			userNikname,
			userGender,
			userPhone,
			userEmail,
			userEnterpriseEmail,
			userStatus,
			userDepts,
			userWorkLocation,
			userJoinTime,
			userEmployeeNo,
			userIsAdmin,
			userJobTitle,
		}
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

func (cli *feiShuCli) recursePrintDept(depts []*fs.DepartmentEntry, did, didType, uidType string, level int, index *int) error {
	var deptChildren []*fs.DepartmentEntry
	var err error
	for i := 0; i < retry; i++ {
		req := fs.NewGetDepartmentChildrenReqBuilder(FeiShuClient).
			DepartmentId(did).
			DepartmentIdType(didType).
			UserIdType(uidType).
			PageSize(50).
			Build()
		deptChildren, err = FeiShuClient.Department.Children(req)
		if err != nil {
			if i < retry {
				time.Sleep(FeiShuDefaultInterval)
				continue
			}
			return err
		}
		break
	}
	var length = len(deptChildren)
	if length == 0 && level == 0 {
		logger.Info("无可用数据")
		return nil
	}
	for i := 0; i < length; i++ {
		//if verbose <= *index && verbose >= 0 {
		//	break
		//}
		*index++
		var (
			status             string
			name               string
			did                string
			oid                string
			leaderUserId       string
			memberCount        int
			primaryMemberCount int
		)
		if deptChildren[i].Status.IsDeleted {
			status = "已删除"
		} else {
			status = "正常"
		}
		name = deptChildren[i].Name
		did = deptChildren[i].DepartmentID
		oid = deptChildren[i].OpenDepartmentID
		memberCount = deptChildren[i].MemberCount
		leaderUserId = deptChildren[i].LeaderUserID
		primaryMemberCount = deptChildren[i].PrimaryMemberCount
		s := fmt.Sprintf("%s  -名称[%s] 状态[%s] 部门ID[%s] OPEN_ID[%s] 主管用户ID[%s] 用户个数[%d] 直属用户个数[%d]",
			strings.Repeat(" ", 3*level), name, status, did, oid, leaderUserId, memberCount, primaryMemberCount)
		fmt.Println(s)
		depts = append(depts, deptChildren[i])
		if recurse {
			var id string
			if didType == "department_id" {
				id = did
			} else {
				id = oid
			}
			err := cli.recursePrintDept(depts, id, didType, uidType, level+1, index)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// generateDepartmentTreeWithUsersHTML 生成包含部门信息和用户信息的部门树HTML代码
func (cli *feiShuCli) generateDepartmentTreeWithUsersHTML(nodes []*FeiShuDepartmentNode, level int) string {
	html := ""
	for _, dept := range nodes {
		html += fmt.Sprintf("<div style=\"margin-left:%dem;\">", level)

		s := fmt.Sprintf("ID:%s&nbsp;&nbsp;OPEN_ID:%s&nbsp;&nbsp;名称:%s", dept.DepartmentID, dept.OpenDepartmentID, dept.Name)
		if dept.EnUsName != "" {
			s += fmt.Sprintf("&nbsp;&nbsp;英文名称:%s", dept.EnUsName)
		}
		if dept.JaJpName != "" {
			s += fmt.Sprintf("&nbsp;&nbsp;日文名称:%s", dept.JaJpName)
		}
		if dept.LeaderUserName != "" {
			s += fmt.Sprintf("&nbsp;&nbsp;领导:%s(ID:%s)\n", dept.LeaderUserName, dept.LeaderUserID)
		} else if dept.LeaderUserID != "" {
			s += fmt.Sprintf("&nbsp;&nbsp;领导:ID:%s\n", dept.LeaderUserID)
		}
		if len(dept.DepartmentHrbps) > 0 {
			var tmp []string
			for _, hrbp := range dept.DepartmentHrbps {
				tmp = append(tmp, *hrbp)
			}
			s += fmt.Sprintf("&nbsp;&nbsp;Hrbp:%s\n", strings.Join(tmp, "、"))
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
			var gender string
			if user.Gender == 0 {
				gender = "保密"
			} else if user.Gender == 1 {
				gender = "男"
			} else if user.Gender == 2 {
				gender = "女"
			}
			var hrbps []string
			for _, hrbp := range dept.DepartmentHrbps {
				hrbps = append(hrbps, *hrbp)
			}
			var userStat []string
			if user.Status.IsUnjoin {
				userStat = append(userStat, "未加入")
			}
			if user.Status.IsResigned {
				userStat = append(userStat, "已离职")
			}
			if user.Status.IsActivated {
				userStat = append(userStat, "已激活")
			}
			if user.Status.IsExited {
				userStat = append(userStat, "已退出")
			}
			if user.Status.IsFrozen {
				userStat = append(userStat, "已冻结")
			}
			var workLocation string
			if user.Country != "" {
				workLocation += user.Country
			}
			if user.City != "" {
				workLocation += user.City
			}
			var isAdmin string
			if user.IsTenantManager {
				isAdmin = "是"
			}
			if !user.IsTenantManager {
				isAdmin = "否"
			}
			m := fmt.Sprintf("ID:%s&nbsp;&nbsp;OPEN_ID:%s&nbsp;&nbsp;姓名:%s&nbsp;&nbsp;性别:%s&nbsp;&nbsp;电话号码:%s&nbsp;&nbsp;邮箱:%s&nbsp;&nbsp;企业邮箱:%s&nbsp;&nbsp;状态:%s&nbsp;&nbsp;工号:%s&nbsp;&nbsp;是否企业管理员:%s", user.UserId, user.OpenId, user.Name, gender, user.Mobile, user.Email, user.EnterpriseEmail, userStat, user.EmployeeNo, isAdmin)
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
func (cli *feiShuCli) saveDepartmentTreeWithUsersToHTML(nodes []*FeiShuDepartmentNode, filename string) (string, error) {
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
func (cli *feiShuCli) generateTreeHTMLDocument(content string) string {
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
