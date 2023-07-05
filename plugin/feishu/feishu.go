package feishu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fasnow/ghttp"
	"idebug/plugin"
	"idebug/utils"
	"net/http"
	"time"
)

const (
	getTenantAccessTokenUrl    = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal" // 应用将代表租户（企业或团队）执行对应的操作，例如获取一个通讯录用户的信息。API 所能操作的数据资源范围受限于应用的身份所能操作的资源范围。由于商店应用会为多家企业提供服务，所以需要先获取对应企业的授权访问凭证 tenant_access_token，并使用该访问凭证来调用 API 访问企业的数据或者资源
	getAppAccessTokenUrl       = "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal"
	getUserAccessToken         = "" // 应用以用户的身份进行相关的操作，访问的数据范围、可以执行的操作将会受到该用户的权限影响。
	getAuthScopeUrl            = "https://open.feishu.cn/open-apis/contact/v3/scopes"
	getDepartmentUrl           = "https://open.feishu.cn/open-apis/contact/v3/departments/:department_id"
	getBatchDepartmentUrl      = "https://open.feishu.cn/open-apis/contact/v3/departments/batch"
	getDepartmentChildrenUrl   = "https://open.feishu.cn/open-apis/contact/v3/departments/:department_id/children"
	getUserUrl                 = "https://open.feishu.cn/open-apis/contact/v3/users/:user_id"
	getUsersIdUrl              = "https://open.feishu.cn/open-apis/contact/v3/users/find_by_department"
	userEmailPasswordChangeUrl = "https://open.feishu.cn/open-apis/admin/v1/password/reset"
)

var defaultInterval = 200 * time.Millisecond

type config struct {
	AppId             *string
	AppSecret         *string
	TenantAccessToken *string
	DepartmentScope   map[string]string
	GroupScope        map[string]string
	UserScope         map[string]string
}

type department struct {
	client *Client
}

type user struct {
	client *Client
}

type Client struct {
	config     *config
	Department *department
	User       *user
	cache      *utils.Cache // 保存access_token
	http       *ghttp.Client
}

func NewClient() *Client {
	f := &Client{
		config: &config{},
		cache:  utils.NewCache(3 * time.Second),
		http:   &ghttp.Client{},
	}
	return f
}

func (client *Client) SetContext(ctx *context.Context) {
	client.http.Context = ctx
}

func (client *Client) Set(appId, appSecret string) {
	conf := &config{
		AppId:     &appId,
		AppSecret: &appSecret,
	}
	client.config = conf
	client.cache = utils.NewCache(3 * time.Second)
}

func (client *Client) SetAppId(appId string) {
	conf := &config{AppSecret: client.config.AppSecret, AppId: &appId}
	client.config = conf
	client.cache = utils.NewCache(3 * time.Second)
}

func (client *Client) SetAppSecret(appSecret string) {
	conf := &config{AppSecret: &appSecret, AppId: client.config.AppId}
	client.config = conf
	client.cache = utils.NewCache(3 * time.Second)
}

// SetTenantAccessTokenFromServer 设置新的tenant_access_token
func (client *Client) SetTenantAccessTokenFromServer() error {
	_, err := client.getNewTenantAccessToken()
	if err != nil {
		return err
	}
	return nil
}

// GetNewAuthScope 获取新的tenant_access_token和新的权限范围,并设置两者的缓存,返回新的config
func (client *Client) GetNewAuthScope(req *GetAuthScopeReq) (*config, error) {
	client.config.DepartmentScope = map[string]string{}
	client.config.UserScope = map[string]string{}
	departmentScope, _, userScope, err := client.getNewAuthScope(req)
	if err != nil {
		return nil, err
	}
	didType := req.req.QueryParams.Get("department_id_type")
	uidType := req.req.QueryParams.Get("user_id_type")
	for _, deptId := range departmentScope {
		//if client.http.Context!=nil && client.http.Context{
		//
		//}
		req1 := NewGetDepartmentReqBuilder(client).
			DepartmentIdType(didType).
			DepartmentId(*deptId).
			Build()
		dpetInfo, err := client.Department.Get(req1)
		if err != nil {
			client.config.DepartmentScope[*deptId] = ""
			time.Sleep(defaultInterval)
			continue
		}
		client.config.DepartmentScope[*deptId] = dpetInfo.Name
		time.Sleep(defaultInterval)
	}
	for _, userId := range userScope {
		req1 := NewGetUserReqBuilder(client).
			UserId(*userId).
			UserIdType(uidType).
			Build()
		userInfo, err := client.User.Get(req1)
		if err != nil {
			client.config.UserScope[*userId] = ""
			time.Sleep(defaultInterval)
			continue
		}
		client.config.UserScope[*userId] = userInfo.Name
		time.Sleep(defaultInterval)
	}

	return client.config, nil
}

// getNewAuthScope dpetIds,gids,uids,error：获取新的tenant_access_token并设置两缓存,返回新的权限范围,
func (client *Client) getNewAuthScope(req *GetAuthScopeReq) ([]*string, []*string, []*string, error) {
	req.req.QueryParams.Set("page_size", "100")
	request, err := http.NewRequest("GET", getAuthScopeUrl+"?"+req.req.QueryParams.Encode(), nil)
	if err != nil {
		return nil, nil, nil, err
	}
	token, err := client.getNewTenantAccessToken()
	if err != nil {
		return nil, nil, nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	if err != nil {
		return nil, nil, nil, err
	}
	response, err := client.http.Do(request)
	if err != nil {
		return nil, nil, nil, err
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, nil, nil, err
	}
	var tmp struct {
		Code int `json:"code"`
		Data struct {
			DepartmentIds []*string `json:"department_ids"`
			GroupIds      []*string `json:"group_ids"`
			HasMore       bool      `json:"has_more"`
			PageToken     string    `json:"page_token"`
			UserIds       []*string `json:"user_ids"`
		} `json:"data"`
		Msg string `json:"msg"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, nil, nil, err
	}
	if tmp.Code != 0 {
		return nil, nil, nil, fmt.Errorf("from server - " + tmp.Msg)
	}
	var (
		departmentIds []*string
		groupIds      []*string
		userIds       []*string
	)
	departmentIds = append(departmentIds, tmp.Data.DepartmentIds...)
	groupIds = append(groupIds, tmp.Data.GroupIds...)
	userIds = append(userIds, tmp.Data.UserIds...)
	if !tmp.Data.HasMore {
		return departmentIds, groupIds, userIds, nil
	}
	time.Sleep(defaultInterval)
	deptIds, groupIds, userIds, err := client.getAuthScopeMore(tmp.Data.PageToken, req)
	if err != nil {
		return nil, nil, nil, err
	}
	departmentIds = append(departmentIds, deptIds...)
	groupIds = append(groupIds, groupIds...)
	userIds = append(userIds, userIds...)
	return departmentIds, groupIds, userIds, nil
}

// GetAuthScopeFromCache 从缓存中取出token的权限范围,不涉及tenant_access_token更新
func (client *Client) GetAuthScopeFromCache() *config {
	return client.config
}

// GetTenantAccessTokenFromCache 从缓存中取出token,不涉及tenant_access_token更新
func (client *Client) GetTenantAccessTokenFromCache() string {
	value, ok := client.cache.Get("tenantAccessToken")
	if ok {
		if token, ok := value.(string); ok {
			return token
		}
	}
	return ""
}

func (client *Client) getAccessTokenByUrl(url string) (string, int, error) {
	postData := map[string]string{
		"app_id":     *client.config.AppId,
		"app_secret": *client.config.AppSecret,
	}
	request, err := http.NewRequest("POST", url, utils.ConvertToReader(postData))
	if err != nil {
		return "", 0, err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		return "", 0, err
	}
	response, err := client.http.Do(request)
	if err != nil {
		return "", 0, err
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return "", 0, err
	}
	var tmp struct {
		Code              int    `json:"code"`
		Expire            int    `json:"expire"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return "", 0, err
	}
	if tmp.Code != 0 {
		return "", 0, fmt.Errorf("from server - " + tmp.Msg)
	}
	return tmp.TenantAccessToken, tmp.Expire, nil
}

func (client *Client) getAuthScopeMore(pageToken string, req *GetAuthScopeReq) ([]*string, []*string, []*string, error) {
	req.req.QueryParams.Set("page_token", pageToken)
	request, err := http.NewRequest("GET", getAuthScopeUrl+"?"+req.req.QueryParams.Encode(), nil)
	if err != nil {
		return nil, nil, nil, err
	}
	token, err := client.autoGetTenantAccessToken()
	if err != nil {
		return nil, nil, nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	if err != nil {
		return nil, nil, nil, err
	}
	response, err := client.http.Do(request)
	if err != nil {
		return nil, nil, nil, err
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, nil, nil, err
	}
	var tmp struct {
		Code int `json:"code"`
		Data struct {
			DepartmentIds []*string `json:"department_ids"`
			GroupIds      []*string `json:"group_ids"`
			HasMore       bool      `json:"has_more"`
			PageToken     string    `json:"page_token"`
			UserIds       []*string `json:"user_ids"`
		} `json:"data"`
		Msg string `json:"msg"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, nil, nil, err
	}
	if tmp.Code != 0 {
		return nil, nil, nil, fmt.Errorf("from server - " + tmp.Msg)
	}
	var (
		departmentIds []*string
		groupIds      []*string
		userIds       []*string
	)
	departmentIds = append(departmentIds, tmp.Data.GroupIds...)
	groupIds = append(groupIds, tmp.Data.GroupIds...)
	userIds = append(userIds, tmp.Data.UserIds...)
	if !tmp.Data.HasMore {
		return departmentIds, groupIds, userIds, nil
	}
	time.Sleep(defaultInterval)
	return client.getAuthScopeMore(tmp.Data.PageToken, req)
}

// 从缓存中取出tenant_access_token,没有缓存的话则添加新的
func (client *Client) autoGetTenantAccessToken() (string, error) {
	value, ok := client.cache.Get("tenantAccessToken")
	if !ok {
		return client.getNewTenantAccessToken()
	} else {
		if token, ok := value.(string); ok {
			return token, nil
		}
	}
	return "", errors.New("获取tenant_access_token时出错")
}

// 获取新的tenant_access_token并设置缓存
func (client *Client) getNewTenantAccessToken() (string, error) {
	token, expire, err := client.getAccessTokenByUrl(getTenantAccessTokenUrl)
	if err != nil {
		return "", err
	}
	client.cache.Set("tenantAccessToken", token, time.Duration(expire)*time.Second)
	client.config.TenantAccessToken = &token
	return token, nil
}

type GetAuthScopeReqBuilder struct {
	req *Req
}

type GetAuthScopeReq struct {
	req *Req
}

func NewGetAuthScopeReqBuilder(client *Client) *GetAuthScopeReqBuilder {
	builder := &GetAuthScopeReqBuilder{}
	builder.req = &Req{
		PathParams:  &plugin.PathParams{},
		QueryParams: &plugin.QueryParams{},
		Client:      client,
	}
	return builder
}

func (builder *GetAuthScopeReqBuilder) UserIdType(t string) *GetAuthScopeReqBuilder {
	builder.req.QueryParams.Set("user_id_type", t)
	return builder
}

func (builder *GetAuthScopeReqBuilder) DepartmentIdType(t string) *GetAuthScopeReqBuilder {
	builder.req.QueryParams.Set("department_id_type", t)
	return builder
}

func (builder *GetAuthScopeReqBuilder) Build() *GetAuthScopeReq {
	req := &GetAuthScopeReq{}
	req.req = builder.req
	return req
}
