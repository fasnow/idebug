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

type DepartmentI18nName struct {
	ZhCn string `json:"zh_cn,omitempty"` // 部门的中文名
	JaJp string `json:"ja_jp,omitempty"` // 部门的日文名
	EnUs string `json:"en_us,omitempty"` // 部门的英文名
}

type DepartmentStatus struct {
	IsDeleted bool `json:"is_deleted,omitempty"` // 是否被删除
}

type DepartmentLeader struct {
	LeaderType int    `json:"leaderType,omitempty"` // 负责人类型
	LeaderID   string `json:"leaderID,omitempty"`   // 负责人ID
}

type DepartmentEntry struct {
	DepartmentID       string             `json:"department_id"`        // 本部门的自定义部门ID;;注意：除需要满足正则规则外，同时不能以`od-`开头
	I18NName           DepartmentI18nName `json:"i18n_name"`            // 国际化的部门名称
	MemberCount        int                `json:"member_count"`         // 部门下用户的个数
	Name               string             `json:"name"`                 // 部门名称
	OpenDepartmentID   string             `json:"open_department_id"`   // 部门的open_id，类型与通过请求的查询参数传入的department_id_type相同
	Order              string             `json:"order"`                // 部门的排序，即部门在其同级部门的展示顺序
	ParentDepartmentID string             `json:"parent_department_id"` // 父部门的ID;;* 在根部门下创建新部门，该参数值为 “0”
	PrimaryMemberCount int                `json:"primary_member_count"` // 部门下主属用户的个数
	Status             DepartmentStatus   `json:"status"`               // 部门状态

	LeaderUserID           string              `json:"leader_user_id"`            // 部门主管用户ID
	ChatID                 string              `json:"chat_id"`                   // 部门群ID
	UnitIds                []*string           `json:"unit_ids"`                  // 部门单位自定义ID列表，当前只支持一个
	Leaders                []*DepartmentLeader `json:"leaders"`                   // 部门负责人
	GroupChatEmployeeTypes []*int              `json:"group_chat_employee_types"` // 部门群雇员类型限制。[]空列表时，表示为无任何雇员类型。类型字段可包含以下值，支持多个类型值；若有多个，用英文','分隔：;1、正式员工;2、实习生;3、外包;4、劳务;5、顾问;6、其他自定义类型字段，可通过下方接口获取到该租户的自定义员工类型的名称，参见[获取人员类型](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/contact-v3/employee_type_enum/list)。
	DepartmentHrbps        []*string           `json:"department_hrbps"`          // 部门HRBP
}

type GetDepartmentReq struct {
	req *Req
}

type GetDepartmentReqBuilder struct {
	req *Req
}

func NewGetDepartmentReqBuilder(client *Client) *GetDepartmentReqBuilder {
	builder := &GetDepartmentReqBuilder{}
	builder.req = &Req{
		PathParams:  &plugin.PathParams{},
		QueryParams: &plugin.QueryParams{},
		Client:      client,
	}
	return builder

}

func (builder *GetDepartmentReqBuilder) DepartmentId(id string) *GetDepartmentReqBuilder {
	builder.req.PathParams.Set(":department_id", id)
	return builder
}

func (builder *GetDepartmentReqBuilder) DepartmentIdType(t string) *GetDepartmentReqBuilder {
	builder.req.QueryParams.Set("department_id_type", t)
	return builder
}

func (builder *GetDepartmentReqBuilder) UserIdType(t string) *GetDepartmentReqBuilder {
	builder.req.QueryParams.Set("user_id_type", t)
	return builder
}

func (builder *GetDepartmentReqBuilder) Build() *GetDepartmentReq {
	req := &GetDepartmentReq{}
	req.req = builder.req
	return req
}

func (dept *department) Get(req *GetDepartmentReq) (DepartmentEntry, error) {
	deptEntry := DepartmentEntry{}
	id := strings.TrimSpace(req.req.PathParams.Get(":department_id"))
	if id == "" {
		return deptEntry, errors.New("部门ID不能为空")
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", strings.Replace(getDepartmentUrl, ":department_id", id, 1), req.req.QueryParams.Encode()), nil)
	if err != nil {
		return deptEntry, err
	}
	if err != nil {
		return deptEntry, err
	}
	token, err := req.req.Client.autoGetTenantAccessToken()
	if err != nil {
		return deptEntry, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := req.req.Client.http.Do(request)
	if err != nil {
		return deptEntry, err
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return deptEntry, err
	}
	var tmp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			DepartmentEntry `json:"department"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return deptEntry, err
	}
	if tmp.Code != 0 {
		return deptEntry, fmt.Errorf("from server - " + tmp.Msg)
	}
	return tmp.Data.DepartmentEntry, nil
}

type GetBatchDepartmentReq struct {
	req *Req
}

type GetBatchDepartmentReqBuilder struct {
	req *Req
}

func NewGetBatchDepartmentReqBuilder(client *Client) *GetBatchDepartmentReqBuilder {
	builder := &GetBatchDepartmentReqBuilder{}
	builder.req = &Req{
		PathParams:  &plugin.PathParams{},
		QueryParams: &plugin.QueryParams{},
		Client:      client,
	}
	return builder
}

func (builder *GetBatchDepartmentReqBuilder) DepartmentIds(ids []string) *GetBatchDepartmentReqBuilder {
	if len(ids) <= 0 {
		builder.req.QueryParams.Set("department_ids", "")
		return builder
	}
	builder.req.QueryParams.Set("department_ids", ids[0])
	for _, id := range ids[1:] {
		builder.req.QueryParams.Add("department_ids", id)
	}
	return builder
}

func (builder *GetBatchDepartmentReqBuilder) DepartmentIdType(t string) *GetBatchDepartmentReqBuilder {
	builder.req.QueryParams.Set("department_id_type", t)
	return builder
}

func (builder *GetBatchDepartmentReqBuilder) UserIdType(t string) *GetBatchDepartmentReqBuilder {
	builder.req.QueryParams.Set("user_id_type", t)
	return builder
}

func (builder *GetBatchDepartmentReqBuilder) Build() *GetBatchDepartmentReq {
	req := &GetBatchDepartmentReq{}
	req.req = builder.req
	return req
}

func (dept *department) Bath(req *GetBatchDepartmentReq) ([]DepartmentEntry, error) {
	request, err := http.NewRequest("GET", getBatchDepartmentUrl+"?"+req.req.QueryParams.Encode(), nil)
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
			Items []DepartmentEntry `json:"items"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, err
	}
	if tmp.Code != 0 {
		return nil, fmt.Errorf("from server - " + tmp.Msg)
	}
	return tmp.Data.Items, nil
}

type GetDepartmentChildrenReq struct {
	req *Req
}

type GetDepartmentChildrenReqBuilder struct {
	req *Req
}

func NewGetDepartmentChildrenReqBuilder(client *Client) *GetDepartmentChildrenReqBuilder {
	builder := &GetDepartmentChildrenReqBuilder{}
	builder.req = &Req{
		PathParams:  &plugin.PathParams{},
		QueryParams: &plugin.QueryParams{},
		Client:      client,
	}
	return builder
}

func (builder *GetDepartmentChildrenReqBuilder) DepartmentId(id string) *GetDepartmentChildrenReqBuilder {
	builder.req.PathParams.Set(":department_id", id)
	return builder
}

func (builder *GetDepartmentChildrenReqBuilder) DepartmentIdType(t string) *GetDepartmentChildrenReqBuilder {
	builder.req.QueryParams.Set("department_id_type", t)
	return builder
}

func (builder *GetDepartmentChildrenReqBuilder) UserIdType(t string) *GetDepartmentChildrenReqBuilder {
	builder.req.QueryParams.Set("user_id_type", t)
	return builder
}

func (builder *GetDepartmentChildrenReqBuilder) Fetch(fetch bool) *GetDepartmentChildrenReqBuilder {
	if fetch {
		builder.req.QueryParams.Set("fetch_children", "true")
	} else {
		builder.req.QueryParams.Set("fetch_children", "false")
	}
	return builder
}

func (builder *GetDepartmentChildrenReqBuilder) PageSize(size int) *GetDepartmentChildrenReqBuilder {
	builder.req.QueryParams.Set("page_size", strconv.Itoa(size))
	return builder
}

func (builder *GetDepartmentChildrenReqBuilder) Build() *GetDepartmentChildrenReq {
	req := &GetDepartmentChildrenReq{}
	req.req = builder.req
	return req
}

func (dept *department) moreDepartment(f *Client, req *GetDepartmentChildrenReq, pageToken string) ([]*DepartmentEntry, error) {
	id := strings.TrimSpace(req.req.PathParams.Get(":department_id"))
	if id == "" {
		return nil, errors.New("部门ID不能为空")
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s&page_token=%s", strings.Replace(getDepartmentChildrenUrl, ":department_id", id, 1), req.req.QueryParams.Encode(), pageToken), nil)
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
			HasMore   bool               `json:"has_more"`
			PageToken string             `json:"page_token"`
			Items     []*DepartmentEntry `json:"items"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, err
	}
	if tmp.Code != 0 {
		return nil, fmt.Errorf("from server - " + tmp.Msg)
	}
	var deptItems []*DepartmentEntry
	deptItems = append(deptItems, tmp.Data.Items...)
	if !tmp.Data.HasMore {
		return deptItems, nil
	}
	time.Sleep(defaultInterval)
	moreDepartments, err := dept.moreDepartment(req.req.Client, req, tmp.Data.PageToken)
	if err != nil {
		return nil, err
	}
	return append(deptItems, moreDepartments...), nil
}

func (dept *department) Children(req *GetDepartmentChildrenReq) ([]*DepartmentEntry, error) {
	id := strings.TrimSpace(req.req.PathParams.Get(":department_id"))
	if id == "" {
		return nil, errors.New("部门ID不能为空")
	}
	request, err := http.NewRequest("GET", strings.Replace(getDepartmentChildrenUrl, ":department_id", id, 1)+"?"+req.req.QueryParams.Encode(), nil)
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
			HasMore   bool               `json:"has_more"`
			PageToken string             `json:"page_token"`
			Items     []*DepartmentEntry `json:"items"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, err
	}
	if tmp.Code != 0 {
		return nil, fmt.Errorf("from server - " + tmp.Msg)
	}
	var deptItems []*DepartmentEntry
	deptItems = append(deptItems, tmp.Data.Items...)
	if !tmp.Data.HasMore {
		return deptItems, nil
	}
	time.Sleep(defaultInterval)
	moreDepartments, err := dept.moreDepartment(req.req.Client, req, tmp.Data.PageToken)
	if err != nil {
		return nil, err
	}
	return append(deptItems, moreDepartments...), nil
}
