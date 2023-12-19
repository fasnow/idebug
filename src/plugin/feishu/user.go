package feishu

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fasnow/ghttp"
	"idebug/plugin"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type UserEntry struct {
	UserId        string `json:"user_id"`        // 用户的user_id，租户内用户的唯一标识，不同ID的说明参见 [用户相关的 ID 概念](https://open.feishu.cn/document/home/user-identity-introduction/introduction)
	OpenId        string `json:"open_id"`        // 用户的open_id，应用内用户的唯一标识，不同ID的说明参见 [用户相关的 ID 概念](https://open.feishu.cn/document/home/user-identity-introduction/introduction)
	Name          string `json:"name"`           // 用户名
	EnName        string `json:"en_name"`        // 英文名
	Nickname      string `json:"nickname"`       // 别名
	Email         string `json:"email"`          // 邮箱;;注意：;1. 非中国大陆手机号成员必须同时添加邮箱;2. 邮箱不可重复
	Mobile        string `json:"mobile"`         // 手机号，在本企业内不可重复；未认证企业仅支持添加中国大陆手机号，通过飞书认证的企业允许添加海外手机号，注意国际电话区号前缀中必须包含加号 +
	MobileVisible bool   `json:"mobile_visible"` // 手机号码可见性，true 为可见，false 为不可见，目前默认为 true。不可见时，组织员工将无法查看该员工的手机号码
	Gender        int    `json:"gender"`         // 性别
	Status        struct {
		IsFrozen    bool `json:"is_frozen"`    // 是否暂停
		IsResigned  bool `json:"is_resigned"`  // 是否离职
		IsActivated bool `json:"is_activated"` // 是否激活
		IsExited    bool `json:"is_exited"`    // 是否主动退出，主动退出一段时间后用户会自动转为已离职
		IsUnjoin    bool `json:"is_unjoin"`    // 是否未加入，需要用户自主确认才能加入团队
	} `json:"status"` // 用户状态，枚举类型，包括is_frozen、is_resigned、is_activated、is_exited 。;;用户状态转移参见：[用户状态图](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/contact-v3/user/field-overview#4302b5a1)
	DepartmentIds   []string `json:"department_ids"`    // 用户所属部门的ID列表，一个用户可属于多个部门。;;ID值的类型与查询参数中的department_id_type 对应。;;不同 ID 的说明与department_id的获取方式参见 [部门ID说明](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/contact-v3/department/field-overview#23857fe0)
	LeaderUserId    string   `json:"leader_user_id"`    // 用户的直接主管的用户ID，ID值与查询参数中的user_id_type 对应。;;不同 ID 的说明参见 [用户相关的 ID 概念](https://open.feishu.cn/document/home/user-identity-introduction/introduction);;获取方式参见[如何获取user_id](https://open.feishu.cn/document/home/user-identity-introduction/how-to-get)
	City            string   `json:"city"`              // 工作城市
	Country         string   `json:"country"`           // 国家或地区Code缩写，具体写入格式请参考 [国家/地区码表](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/contact-v3/user/country-code-description)
	WorkStation     string   `json:"work_station"`      // 工位
	JoinTime        int      `json:"join_time"`         // 入职时间，时间戳格式，表示从1970年1月1日开始所经过的秒数
	IsTenantManager bool     `json:"is_tenant_manager"` // 是否是租户超级管理员
	EmployeeNo      string   `json:"employee_no"`       // 工号
	EmployeeType    int      `json:"employee_type"`     // 员工类型，可选值有：;- `1`：正式员工;- `2`：实习生;- `3`：外包;- `4`：劳务;- `5`：顾问   ;同时可读取到自定义员工类型的 int 值，可通过下方接口获取到该租户的自定义员工类型的名称，参见[获取人员类型](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/contact-v3/employee_type_enum/list)

	EnterpriseEmail string `json:"enterprise_email"` // 企业邮箱，请先确保已在管理后台启用飞书邮箱服务;;创建用户时，企业邮箱的使用方式参见[用户接口相关问题](https://open.feishu.cn/document/ugTN1YjL4UTN24CO1UjN/uQzN1YjL0cTN24CN3UjN#77061525)

	JobTitle string `json:"job_title"` // 职务
	IsFrozen bool   `json:"is_frozen"` // 是否暂停用户
}

type GetUsersByDepartmentIdReqBuilder struct {
	req Req
}

type GetUsersByDepartmentIdReq struct {
	req Req
}

func NewGetUsersByDepartmentIdReqBuilder(f *Client) *GetUsersByDepartmentIdReqBuilder {
	builder := &GetUsersByDepartmentIdReqBuilder{}
	builder.req = Req{
		PathParams:  &plugin.PathParams{},
		QueryParams: &plugin.QueryParams{},
		Client:      f,
	}
	return builder
}

func (builder *GetUsersByDepartmentIdReqBuilder) DepartmentId(id string) *GetUsersByDepartmentIdReqBuilder {
	builder.req.QueryParams.Set("department_id", id)
	return builder
}

func (builder *GetUsersByDepartmentIdReqBuilder) DepartmentIdType(t string) *GetUsersByDepartmentIdReqBuilder {
	builder.req.QueryParams.Set("department_id_type", t)
	return builder
}

func (builder *GetUsersByDepartmentIdReqBuilder) UserIdType(t string) *GetUsersByDepartmentIdReqBuilder {
	builder.req.QueryParams.Set("user_id_type", t)
	return builder
}

func (builder *GetUsersByDepartmentIdReqBuilder) PageSize(size int) *GetUsersByDepartmentIdReqBuilder {
	builder.req.QueryParams.Set("page_size", strconv.Itoa(size))
	return builder
}

func (builder *GetUsersByDepartmentIdReqBuilder) Build() *GetUsersByDepartmentIdReq {
	req := &GetUsersByDepartmentIdReq{}
	req.req = builder.req
	return req
}

func (u *user) GetUsersByDepartmentId(req *GetUsersByDepartmentIdReq) ([]*UserEntry, error) {
	id := strings.TrimSpace(req.req.QueryParams.Get("department_id"))
	if id == "" {
		return nil, errors.New("部门ID不能为空")
	}
	request, err := http.NewRequest("GET", getUsersIdUrl+"?"+req.req.QueryParams.Encode(), nil)
	if err != nil {
		return nil, err
	}
	token, err := req.req.Client.autoGetTenantAccessToken()
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	if err != nil {
		return nil, err
	}
	response, err := req.req.Client.http.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, err
	}
	var tmp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			HasMore   bool         `json:"has_more"`
			PageToken string       `json:"page_token"`
			Items     []*UserEntry `json:"items"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, err
	}
	if tmp.Code != 0 {
		return nil, fmt.Errorf("from server - " + tmp.Msg)
	}
	var users []*UserEntry
	users = append(users, tmp.Data.Items...)
	if !tmp.Data.HasMore {
		return users, nil
	}
	time.Sleep(defaultInterval)
	user, err := u.moreUser(req.req.Client, req, tmp.Data.PageToken)
	if err != nil {
		return nil, err
	}
	users = append(users, user...)
	return users, nil
}

func (u *user) moreUser(f *Client, req *GetUsersByDepartmentIdReq, pageToken string) ([]*UserEntry, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s&page_token=%s", getUsersIdUrl, req.req.QueryParams.Encode(), pageToken), nil)
	if err != nil {
		return nil, err
	}
	pageToken, err = req.req.Client.autoGetTenantAccessToken()
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+pageToken)
	if err != nil {
		return nil, err
	}
	response, err := f.http.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, err
	}
	var tmp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			HasMore   bool         `json:"has_more"`
			PageToken string       `json:"page_token"`
			Items     []*UserEntry `json:"items"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, err
	}
	if tmp.Code != 0 {
		return nil, fmt.Errorf("from server - " + tmp.Msg)
	}
	var users []*UserEntry
	users = append(users, tmp.Data.Items...)
	if !tmp.Data.HasMore {
		return users, nil
	}
	time.Sleep(defaultInterval)
	moreUsers, err := u.moreUser(f, req, tmp.Data.PageToken)
	if err != nil {
		return nil, err
	}
	return append(users, moreUsers...), nil
}

type GetUserReqBuilder struct {
	req Req
}

type GetUserReq struct {
	req Req
}

func NewGetUserReqBuilder(f *Client) *GetUserReqBuilder {
	builder := &GetUserReqBuilder{}
	builder.req = Req{
		PathParams:  &plugin.PathParams{},
		QueryParams: &plugin.QueryParams{},
		Client:      f,
	}
	return builder
}
func (builder *GetUserReqBuilder) UserId(id string) *GetUserReqBuilder {
	builder.req.PathParams.Set(":user_id", id)
	return builder
}

func (builder *GetUserReqBuilder) UserIdType(t string) *GetUserReqBuilder {
	builder.req.QueryParams.Set("user_id_type", t)
	return builder
}

func (builder *GetUserReqBuilder) DepartmentIdType(t string) *GetUserReqBuilder {
	builder.req.QueryParams.Set("department_id_type", t)
	return builder
}

func (builder *GetUserReqBuilder) Build() *GetUserReq {
	req := &GetUserReq{}
	req.req = builder.req
	return req
}

func (u *user) Get(req *GetUserReq) (*UserEntry, error) {
	userEntry := &UserEntry{}
	id := strings.TrimSpace(req.req.PathParams.Get(":user_id"))
	if id == "" {
		return nil, errors.New("用户ID不能为空")
	}
	request, err := http.NewRequest("GET", strings.Replace(fmt.Sprintf("%s?%s", getUserUrl, req.req.QueryParams.Encode()), ":user_id", id, 1), nil)
	if err != nil {
		return nil, err
	}
	token, err := req.req.Client.autoGetTenantAccessToken()
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	if err != nil {
		return nil, err
	}
	response, err := req.req.Client.http.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, err
	}
	var tmp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			User UserEntry `json:"user"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, err
	}
	if tmp.Code != 0 {
		return nil, fmt.Errorf("from server - " + tmp.Msg)
	}
	userEntry = &tmp.Data.User
	return userEntry, nil
}

type UserEmailPasswordUpdateReqBuilder struct {
	req Req
}

type UserEmailPasswordChangeReq struct {
	req Req
}

func NewUserEmailPasswordChangeReqBuilder(f *Client) *UserEmailPasswordUpdateReqBuilder {
	builder := &UserEmailPasswordUpdateReqBuilder{}
	builder.req = Req{
		PathParams:  &plugin.PathParams{},
		QueryParams: &plugin.QueryParams{},
		Client:      f,
	}
	return builder
}

func (builder *UserEmailPasswordUpdateReqBuilder) PostData(userId, password string) *UserEmailPasswordUpdateReqBuilder {
	var tmp struct {
		Password struct {
			EntEmailPassword string `json:"ent_email_password"`
		} `json:"password"`
		UserID string `json:"user_id"`
	}
	tmp.Password.EntEmailPassword = password
	tmp.UserID = userId
	postData, _ := json.Marshal(tmp)
	builder.req.Body = string(postData)
	return builder
}

func (builder *UserEmailPasswordUpdateReqBuilder) UserIdType(t string) *UserEmailPasswordUpdateReqBuilder {
	builder.req.QueryParams.Set("user_id_type", t)
	return builder
}

func (builder *UserEmailPasswordUpdateReqBuilder) Build() *UserEmailPasswordChangeReq {
	req := &UserEmailPasswordChangeReq{}
	req.req = builder.req
	return req
}

func (u *user) EmailPasswordUpdate(req *UserEmailPasswordChangeReq) error {
	value := ""
	if v, ok := req.req.Body.(string); ok {
		value = v
	}
	postData := strings.NewReader(value)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s?%s", userEmailPasswordChangeUrl, req.req.QueryParams.Encode()), postData)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}
	token, err := req.req.Client.autoGetTenantAccessToken()
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	if err != nil {
		return err
	}
	response, err := req.req.Client.http.Do(request)
	if err != nil {
		return err
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return err
	}
	var tmp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
		} `json:"data"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return err
	}
	if tmp.Code != 0 {
		return fmt.Errorf("from server - " + tmp.Msg)
	}
	return nil
}
