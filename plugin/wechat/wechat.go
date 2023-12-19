package wechat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fasnow/ghttp"
	"idebug/utils"
	"net/http"
	"net/url"
	"time"
)

type apiConfig struct {
	// 获取access_token ?corpid=&corpsecret=
	getAccessTokenUrl string

	// 获取access_token权限分配 ?access_token=
	getAccessTokenDetailUrl string

	// 递归获取部门详情 ?access_token=ACCESS_TOKEN&id=
	getDepartmentListUrl string

	// 根据部门ID递归获取所有部门ID ?access_token=ACCESS_TOKEN&id=ID
	getDepartmentIdListUrl string

	// 获取单个部门详情 ?access_token=ACCESS_TOKEN&id=ID
	getDepartmentUrl string

	// 获取成员ID列表 获取企业成员的open_userid与对应的部门ID列表 ?access_token=ACCESS_TOKEN
	getUserIdListUrl string

	// 获取部门成员 ?access_token=ACCESS_TOKEN&department_id=DEPARTMENT_ID
	getDepartmentSimpleUserUrl string

	// 获取部门成员详情 ?access_token=ACCESS_TOKEN&department_id=DEPARTMENT_ID
	getDepartmentUserUrl string

	// 创建成员 ?access_token=ACCESS_TOKEN
	createUserUrl string

	// 读取成员 ?access_token=ACCESS_TOKEN&userid=USERID
	getUserUrl string

	// 更新成员 ?access_token=ACCESS_TOKEN
	updateSpecifiedUserUrl string

	// 删除成员 ?access_token=ACCESS_TOKEN&userid=USERID
	deleteSpecifiedUserUrl string

	// 手机号获取userid ?access_token=ACCESS_TOKEN
	getUserIDByPhone string

	// 邮箱获取userid ?access_token=ACCESS_TOKEN
	getUserIDByEmail string

	// 获取企业微信接口IP段 ?access_token=ACCESS_TOKEN
	getAPIDomainCIDRUrl string
}

var (
	baseUrl = "https://qyapi.weixin.qq.com"
	api     = initApi(baseUrl)
)

func initApi(baseDomain string) apiConfig {
	return apiConfig{
		getAccessTokenUrl:          baseDomain + "/cgi-bin/gettoken",
		getAccessTokenDetailUrl:    "https://open.work.weixin.qq.com/devtool/getInfoByAccessToken",
		getDepartmentListUrl:       baseDomain + "/cgi-bin/department/list",
		getDepartmentIdListUrl:     baseDomain + "/cgi-bin/department/simplelist",
		getDepartmentUrl:           baseDomain + "/cgi-bin/department/get",
		getUserIdListUrl:           baseDomain + "/cgi-bin/user/list_id",
		getDepartmentSimpleUserUrl: baseDomain + "/cgi-bin/user/simplelist",
		getDepartmentUserUrl:       baseDomain + "/cgi-bin/user/list",
		createUserUrl:              baseDomain + "/cgi-bin/user/create",
		getUserUrl:                 baseDomain + "/cgi-bin/user/get",
		updateSpecifiedUserUrl:     baseDomain + "/cgi-bin/user/update",
		deleteSpecifiedUserUrl:     baseDomain + "/cgi-bin/user/delete",
		getUserIDByPhone:           baseDomain + "/cgi-bin/user/getuserid",
		getUserIDByEmail:           baseDomain + "/cgi-bin/user/get_userid_by_email",
		getAPIDomainCIDRUrl:        baseDomain + "/cgi-bin/get_api_domain_ip",
	}
}

var roleGroup = map[int]string{
	-1: "未知管理组",
	1:  "应用",
	4:  "第三方服务商",
	8:  "通讯录管理助手",
	64: "分级管理组",
}

func SetBaseDomain(domain string) {
	baseUrl = domain
	api = initApi(baseUrl)
}

func GetBaseDomain() string {
	return baseUrl
}

type AccessTokenAuthItem struct {
	AuthApps []struct {
		AppName        string `json:"appname"`
		AppOpenid      int    `json:"appopenid"`
		ReliableDomain string `json:"reliabledomain"`
	} `json:"authapps"`
	AuthUsers []struct {
		AcctId string `json:"acctid"`
	} `json:"authusers"`
	AuthTags []struct {
		TagName   string `json:"tagname"`
		TagOpenid int    `json:"tagopenid"`
	} `json:"authtags"`
	AuthParties []struct {
		PartyName   string `json:"partyname"`
		PartyOpenid string `json:"partyopenid"`
	} `json:"authparties"`
}

type AccessTokenAuthScope struct {
	RoleGroup string `json:"rolegroup"` // 应用类型
	RoleName  string `json:"rolename"`  // 应用名称
	item      *AccessTokenAuthItem
}
type config struct {
	CorpId      *string
	CorpSecret  *string
	AccessToken *string
	ExpireIn    *int
	//authScope   *AccessTokenAuthScope
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

func NewWxClient() *Client {
	client := &Client{
		config:     &config{},
		cache:      utils.NewCache(3 * time.Second),
		http:       &ghttp.Client{},
		User:       &user{},
		Department: &department{},
	}
	client.User.client = client
	client.Department.client = client
	return client
}

func (client *Client) SetContext(ctx *context.Context) {
	client.http.Context = ctx
}

func (client *Client) StopWhenContextCanceled(enable bool) {
	client.http.StopWhenContextCanceled = enable
}

func (client *Client) Set(corpId, corpSecret string) {
	conf := &config{
		CorpId:     &corpId,
		CorpSecret: &corpSecret,
	}
	client.config = conf
	client.cache = utils.NewCache(3 * time.Second)
}

func (client *Client) SetCorpId(corpId string) {
	conf := &config{
		CorpId:     &corpId,
		CorpSecret: client.config.CorpSecret,
	}
	client.config = conf
	client.cache = utils.NewCache(3 * time.Second)
}

func (client *Client) SetAccessToken(token string) {
	client.config = &config{AccessToken: &token}
}

func (client *Client) SetCorpSecret(corpSecret string) {
	conf := &config{
		CorpId:     client.config.CorpId,
		CorpSecret: &corpSecret,
	}
	client.config = conf
	client.cache = utils.NewCache(3 * time.Second)
}

func (client *Client) GetAccessToken() (string, error) {
	token, err := client.getAccessTokenFromCache()
	if err != nil {
		return "", err
	}
	return token, nil
}

func (client *Client) GetAccessTokenFromCache() string {
	if client.config.AccessToken == nil || *client.config.AccessToken == "" {
		return ""
	}
	return *client.config.AccessToken
}

func (client *Client) GetAccessTokenFromServer() (string, error) {
	token, expire, err := client.getAccessToken()
	if err != nil {
		return "", err
	}
	client.cache.Set("accessToken", token, time.Duration(expire)*time.Second)
	client.config.AccessToken = &token
	client.config.ExpireIn = &expire
	return token, nil
}

func (client *Client) getAccessTokenFromCache() (string, error) {
	value, ok := client.cache.Get("accessToken")
	if !ok {
		token, expire, err := client.getAccessToken()
		if err != nil {
			return "", err
		}
		client.cache.Set("accessToken", token, time.Duration(expire)*time.Second)
		client.config.AccessToken = &token
		client.config.ExpireIn = &expire
		return token, nil
	} else {
		if token, ok := value.(string); ok {
			return token, nil
		}
	}
	return "", errors.New("获取access_token时出错")
}

func (client *Client) getAccessToken() (string, int, error) {
	params := url.Values{}
	params.Add("corpid", *client.config.CorpId)
	params.Add("corpsecret", *client.config.CorpSecret)
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", api.getAccessTokenUrl, params.Encode()), nil)
	if err != nil {
		return "", 0, err
	}
	request.Header.Set("User-Agent", "")
	response, err := client.http.Do(request)
	if response.StatusCode != 200 {
		return "", 0, errors.New(response.Status)
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return "", 0, err
	}
	var res struct {
		ErrCode     *int    `json:"errcode"`
		ErrMsg      *string `json:"errmsg"`
		AccessToken *string `json:"access_token"`
		ExpiresIn   *int    `json:"expires_in"`
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return "", 0, errors.New("获取 access_token 时出错")
	}
	if *res.ErrCode != 0 {
		return "", 0, fmt.Errorf("from server - " + *res.ErrMsg)
	}
	return *res.AccessToken, *res.ExpiresIn, nil
}

func (client *Client) GetConfig() *config {
	return client.config
}
