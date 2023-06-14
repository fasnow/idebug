package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/fasnow/ghttp"
	"net/http"
	"net/url"
	"strconv"
)

// 获取access_token ?corpid=&corpsecret=
const getAccessTokenUrl = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"

// 获取access_token权限分配 ?access_token=
const getAccessTokenDetailUrl = "https://open.work.weixin.qq.com/devtool/getInfoByAccessToken"

// 递归获取部门详情 ?access_token=ACCESS_TOKEN&id=
const getDepartmentUrl = "https://qyapi.weixin.qq.com/cgi-bin/department/list"

// 根据部门ID递归获取所有部门ID ?access_token=ACCESS_TOKEN&id=ID
const getSubDepartmentIdUrl = "https://qyapi.weixin.qq.com/cgi-bin/department/simplelist"

// 获取单个部门详情 ?access_token=ACCESS_TOKEN&id=ID
const getSpecifiedDepartmentUrl = "https://qyapi.weixin.qq.com/cgi-bin/department/get"

// 获取成员ID列表 获取企业成员的open_userid与对应的部门ID列表 ?access_token=ACCESS_TOKEN
const getUserIdListUrl = "https://qyapi.weixin.qq.com/cgi-bin/user/list_id"

// 获取部门成员 ?access_token=ACCESS_TOKEN&department_id=DEPARTMENT_ID
const getDepartmentSimpleUserUrl = "https://qyapi.weixin.qq.com/cgi-bin/user/simplelist"

// 获取部门成员详情 ?access_token=ACCESS_TOKEN&department_id=DEPARTMENT_ID
const getDepartmentUserUrl = "https://qyapi.weixin.qq.com/cgi-bin/user/list"

// 创建成员 ?access_token=ACCESS_TOKEN
const createUserUrl = "https://qyapi.weixin.qq.com/cgi-bin/user/create"

// 读取成员 ?access_token=ACCESS_TOKEN&userid=USERID
const getSpecifiedUserUrl = "https://qyapi.weixin.qq.com/cgi-bin/user/get"

// 更新成员 ?access_token=ACCESS_TOKEN
const updateSpecifiedUserUrl = "https://qyapi.weixin.qq.com/cgi-bin/user/update"

// 删除成员 ?access_token=ACCESS_TOKEN&userid=USERID
const deleteSpecifiedUserUrl = "https://qyapi.weixin.qq.com/cgi-bin/user/delete"

// 手机号获取userid ?access_token=ACCESS_TOKEN
const getUserIDByPhone = "https://qyapi.weixin.qq.com/cgi-bin/user/getuserid"

// 邮箱获取userid ?access_token=ACCESS_TOKEN
const getUserIDByEmail = "https://qyapi.weixin.qq.com/cgi-bin/user/get_userid_by_email"

// 获取企业微信接口IP段 ?access_token=ACCESS_TOKEN
const getAPIDomainCIDRUrl = "https://qyapi.weixin.qq.com/cgi-bin/get_api_domain_ip"

var httpClient = ghttp.Client{}

var roleGroup = map[int]string{
	-1: "未知管理组",
	1:  "应用",
	4:  "第三方服务商",
	8:  "通讯录管理助手",
	64: "分级管理组",
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

type AccessTokenAuthDetail struct {
	RoleGroup string `json:"rolegroup"` // 应用类型
	RoleName  string `json:"rolename"`  // 应用名称
	AccessTokenAuthItem
}

type WxSimpleDepartment struct {
	ID       int `json:"id"`
	ParentId int `json:"parentid"`
	Order    int `json:"order"`
}

type WxDepartment struct {
	WxSimpleDepartment
	Name             string   `json:"name"`
	NameEn           string   `json:"name_en"`
	DepartmentLeader []string `json:"department_leader"`
}

type WxSimpleUser struct {
	Name       string `json:"name"`
	UserId     string `json:"userid"`
	Department []int  `json:"department"`
	OpenUserid string `json:"open_userid"`
}

type WxUser struct {
	WxSimpleUser
	Order          []int    `json:"order"`
	Position       string   `json:"position"`
	Mobile         string   `json:"mobile"`
	Gender         string   `json:"gender"`
	Email          string   `json:"email"`
	BizMail        string   `json:"biz_mail"`
	IsLeaderInDept []int    `json:"is_leader_in_dept"`
	DirectLeader   []string `json:"direct_leader"`
	Avatar         string   `json:"avatar"`
	ThumbAvatar    string   `json:"thumb_avatar"`
	Telephone      string   `json:"telephone"`
	Alias          string   `json:"alias"`
	Status         int      `json:"status"`
	Address        string   `json:"address"`
	EnglishName    string   `json:"english_name"`
	MainDepartment int      `json:"main_department"`
	Extattr        struct {
		Attrs []struct {
			Type int    `json:"type"`
			Name string `json:"name"`
			Text struct {
				Value string `json:"value"`
			} `json:"text,omitempty"`
			Web struct {
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"web,omitempty"`
		} `json:"attrs"`
	} `json:"extattr"`
	QrCode           string `json:"qr_code"`
	ExternalPosition string `json:"external_position"`
	ExternalProfile  struct {
		ExternalCorpName string `json:"external_corp_name"`
		WechatChannels   struct {
			Nickname string `json:"nickname"`
			Status   int    `json:"status"`
		} `json:"wechat_channels"`
		ExternalAttr []struct {
			Type int    `json:"type"`
			Name string `json:"name"`
			Text struct {
				Value string `json:"value"`
			} `json:"text,omitempty"`
			Web struct {
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"web,omitempty"`
			Miniprogram struct {
				Appid    string `json:"appid"`
				Pagepath string `json:"pagepath"`
				Title    string `json:"title"`
			} `json:"miniprogram,omitempty"`
		} `json:"external_attr"`
	} `json:"external_profile"`
}

type ClientInfo struct {
	CorpId         string
	CorpSecret     string
	AccessToken    string
	ExpireIn       int
	Resource       string
	Department     []string
	Member         []string
	Tag            []string
	App            []string
	ReliableDomain []string
}

type WeChat struct {
	info ClientInfo
}

func NewWxClient() (*WeChat, error) {
	return &WeChat{}, nil
}

func (w *WeChat) GetClientInfo() ClientInfo {
	return w.info
}

// GetAccessToken 获取并设置access_token 返回值说明 access_token; expire: 超时,默认7200s; errcode: 服务端返回的错误码; errmsg: 程序错误或者服务端返回的错误信息.
func (w *WeChat) GetAccessToken(corpId, corpSecret string) (string, int, int, error) {
	w.info.CorpId = corpId
	w.info.CorpSecret = corpSecret
	params := url.Values{}
	params.Add("corpid", corpId)
	params.Add("corpsecret", corpSecret)
	request, _ := http.NewRequest("GET", fmt.Sprintf("%s?%s", getAccessTokenUrl, params.Encode()), nil)
	request.Header.Set("User-Agent", "")
	response, err := httpClient.Do(request)
	if err != nil {
		return "", -1, -1, err
	}
	if response.StatusCode != 200 {
		return "", -1, -1, errors.New(response.Status)
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return "", -1, -1, err
	}
	var res struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		if len(string(body)) > 50 {
			return "", -1, -1, errors.New(string(body)[:50] + "...")
		}
		return "", -1, -1, errors.New(string(body))
	}
	if res.ErrCode != 0 {
		return "", -1, res.ErrCode, errors.New(res.ErrMsg)
	}
	w.info.AccessToken = res.AccessToken
	w.info.ExpireIn = res.ExpiresIn
	return res.AccessToken, res.ExpiresIn, 0, nil
}

// GetAccessTokenInfoByAccessToken 获取access_token权限分配并设置
func (w *WeChat) GetAccessTokenInfoByAccessToken(accessToken string) (*AccessTokenAuthDetail, error) {
	w.info.ExpireIn = 0
	w.info.Resource = ""
	w.info.Department = []string{}
	w.info.Member = []string{}
	w.info.Tag = []string{}
	w.info.App = []string{}
	w.info.ReliableDomain = []string{}
	params := url.Values{}
	params.Add("access_token", accessToken)
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getAccessTokenDetailUrl, params.Encode()), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "")
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New(response.Status)
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, err
	}
	humanMessage, _ := jsonparser.GetString(body, "result", "humanMessage")
	if humanMessage != "" {
		return nil, errors.New(humanMessage)
	}
	var res2 struct {
		Data struct {
			Roleflags int    `json:"roleflags"`
			Rolename  string `json:"rolename"`
			AccessTokenAuthItem
		} `json:"data"`
	}
	err = json.Unmarshal(body, &res2)
	if err != nil {
		return nil, err
	}
	group := roleGroup[res2.Data.Roleflags]
	if group == "" {
		group = roleGroup[-1]
	}
	w.info.Resource = fmt.Sprintf("%s (%s)", group, res2.Data.Rolename)
	for _, v := range res2.Data.AuthApps {
		if v.ReliableDomain != "" {
			w.info.ReliableDomain = append(w.info.ReliableDomain, fmt.Sprintf("%s (%s)", v.ReliableDomain, v.AppName))
		}
		w.info.App = append(w.info.App, fmt.Sprintf("%s (AgentId: %d)", v.AppName, v.AppOpenid))
	}
	for _, v := range res2.Data.AuthTags {
		w.info.Tag = append(w.info.Tag, fmt.Sprintf("%s (AgentId: %d)", v.TagName, v.TagOpenid))
	}
	for _, v := range res2.Data.AuthUsers {
		w.info.Member = append(w.info.Member, v.AcctId)
	}
	for _, v := range res2.Data.AuthParties {
		w.info.Department = append(w.info.Department, fmt.Sprintf("%s (AgentId: %s)", v.PartyName, v.PartyOpenid))
	}
	return &AccessTokenAuthDetail{RoleGroup: group, RoleName: res2.Data.Rolename, AccessTokenAuthItem: res2.Data.AccessTokenAuthItem}, nil
}

// GetAccessTokenInfo 获取access_token权限分配
func (w *WeChat) GetAccessTokenInfo() (*AccessTokenAuthDetail, error) {
	return w.GetAccessTokenInfoByAccessToken(w.info.AccessToken)
}

// GetAPIDomainCIDR 获取企业微信API域名IP段
func (w *WeChat) GetAPIDomainCIDR() ([]string, int, error) {
	return nil, -1, nil
}

// GetDepartmentListById 根据部门ID递归获取子部门信息,该接口性能较低,
// 建议改用GetSubDepartmentIdListById和GetSubDepartmentIdList
func (w *WeChat) GetDepartmentListById(id int) ([]WxDepartment, int, error) {
	var res struct {
		ErrCode    int            `json:"errcode"`
		ErrMsg     string         `json:"errmsg"`
		Department []WxDepartment `json:"department"`
	}
	params := url.Values{}
	params.Add("access_token", w.info.AccessToken)
	if id != -1 {
		params.Add("id", strconv.Itoa(id))
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getDepartmentUrl, params.Encode()), nil)
	if err != nil {
		return nil, -1, err
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, -1, err
	}
	if response.StatusCode != 200 {
		return nil, -1, errors.New(response.Status)
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, -1, err
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, -1, err
	}
	if res.ErrCode != 0 {
		return res.Department, res.ErrCode, errors.New(res.ErrMsg)
	}
	return res.Department, 0, nil
}

// GetDepartmentSimpleList 递归获取所有部门信息,该接口性能较低,
// 建议改用GetSubDepartmentIdListById和GetSubDepartmentIdListAll
func (w *WeChat) GetDepartmentSimpleList() ([]WxDepartment, int, error) {
	return w.GetDepartmentListById(-1)
}

// GetSubDepartmentIdListById 根据部门ID递归获取所有部门ID
func (w *WeChat) GetSubDepartmentIdListById(id int) ([]WxDepartment, int, error) {
	return GetSubDepartmentIdListById(w.info.AccessToken, id)
}

// GetSubDepartmentIdList 递归获取所有部门信息
func (w *WeChat) GetSubDepartmentIdList() ([]WxDepartment, int, error) {
	return w.GetSubDepartmentIdListById(-1)
}

// GetDepartmentDetail 获取单个部门详情
func (w *WeChat) GetDepartmentDetail(id int) (*WxDepartment, int, error) {
	return GetDepartmentDetail(w.info.AccessToken, id)
}

// GetDepartmentUser 获取部门成员
func (w *WeChat) GetDepartmentUser(id int, fetchChild bool) ([]WxSimpleUser, int, error) {
	return GetDepartmentSimpleUser(w.info.AccessToken, id, fetchChild)
}

// GetDepartmentUserDetail 获取部门成员详情
func (w *WeChat) GetDepartmentUserDetail(id int, fetchChild bool) ([]WxUser, int, error) {
	return GetDepartmentUser(w.info.AccessToken, id, fetchChild)
}

func (w *WeChat) createUser() {}

func (w *WeChat) readUser() {}

func (w *WeChat) updateUser() {}

func (w *WeChat) deleteUser() {}

func (w *WeChat) GetUserIDByPhone() {}

func (w *WeChat) GetUserIDByEmail() {}

func (w *WeChat) GetUserList() {}

// GetDepartmentById  根据ID递归获取所有部门详情
func GetDepartmentById(accessToken string, id int) ([]WxDepartment, int, error) {
	var res struct {
		ErrCode    int            `json:"errcode"`
		ErrMsg     string         `json:"errmsg"`
		Department []WxDepartment `json:"department"`
	}
	params := url.Values{}
	params.Add("access_token", accessToken)
	if id != -1 {
		params.Add("id", strconv.Itoa(id))
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getDepartmentUrl, params.Encode()), nil)
	if err != nil {
		return nil, -1, err
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, -1, err
	}
	if response.StatusCode != 200 {
		return nil, -1, errors.New(response.Status)
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, -1, err
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, -1, err
	}
	if res.ErrCode != 0 {
		return res.Department, res.ErrCode, errors.New(res.ErrMsg)
	}
	return res.Department, 0, nil
}

// GetDepartment  递归获取默认所有部门详情
func GetDepartment(accessToken string) ([]WxDepartment, int, error) {
	return GetDepartmentById(accessToken, -1)
}

// GetSubDepartmentIdListById 根据部门ID递归获取所有部门ID
func GetSubDepartmentIdListById(accessToken string, id int) ([]WxDepartment, int, error) {
	var res struct {
		ErrCode    int            `json:"errcode"`
		ErrMsg     string         `json:"errmsg"`
		Department []WxDepartment `json:"department_id"`
	}
	params := url.Values{}
	params.Add("access_token", accessToken)
	if id != -1 {
		params.Add("id", strconv.Itoa(id))
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getSubDepartmentIdUrl, params.Encode()), nil)
	if err != nil {
		return nil, -1, err
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, -1, err
	}
	if response.StatusCode != 200 {
		return nil, -1, errors.New(response.Status)
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, -1, err
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, -1, err
	}
	if res.ErrCode != 0 {
		return res.Department, res.ErrCode, errors.New(res.ErrMsg)
	}
	return res.Department, 0, nil
}

// GetSubDepartmentIdList 递归获取默认所有部门ID
func GetSubDepartmentIdList(accessToken string) ([]WxDepartment, int, error) {
	return GetSubDepartmentIdListById(accessToken, -1)
}

// GetDepartmentDetail 获取单个部门详情
func GetDepartmentDetail(accessToken string, id int) (*WxDepartment, int, error) {
	params := url.Values{}
	params.Add("access_token", accessToken)
	params.Add("id", strconv.Itoa(id))
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getSpecifiedDepartmentUrl, params.Encode()), nil)
	if err != nil {
		return nil, -1, err
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, -1, err
	}
	if response.StatusCode != 200 {
		return nil, -1, errors.New(response.Status)
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, -1, err
	}
	var res struct {
		ErrCode    int          `json:"errcode"`
		ErrMsg     string       `json:"errmsg"`
		Department WxDepartment `json:"department"`
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, -1, err
	}
	if res.ErrCode != 0 {
		return &res.Department, res.ErrCode, errors.New(res.ErrMsg)
	}
	return &res.Department, 0, nil
}

// GetDepartmentSimpleUser 获取所有部门成员简单信息
func GetDepartmentSimpleUser(accessToken string, id int, fetchChild bool) ([]WxSimpleUser, int, error) {
	params := url.Values{}
	params.Add("access_token", accessToken)
	params.Add("department_id", strconv.Itoa(id))
	if fetchChild {
		params.Add("fetch_child", "1")
	} else {
		params.Add("fetch_child", "0")
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getDepartmentSimpleUserUrl, params.Encode()), nil)
	if err != nil {
		return nil, -1, err
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, -1, err
	}
	if response.StatusCode != 200 {
		return nil, -1, errors.New(response.Status)
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, -1, err
	}
	var res struct {
		ErrCode  int            `json:"errcode"`
		ErrMsg   string         `json:"errmsg"`
		UserList []WxSimpleUser `json:"userlist"`
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, -1, err
	}
	if res.ErrCode != 0 {
		return res.UserList, res.ErrCode, errors.New(res.ErrMsg)
	}
	return res.UserList, 0, nil
}

// GetDepartmentUser 获取所有部门成员详情
func GetDepartmentUser(accessToken string, id int, fetchChild bool) ([]WxUser, int, error) {
	params := url.Values{}
	params.Add("access_token", accessToken)
	params.Add("department_id", strconv.Itoa(id))
	if fetchChild {
		params.Add("fetch_child", "1")
	} else {
		params.Add("fetch_child", "0")
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getDepartmentUserUrl, params.Encode()), nil)
	if err != nil {
		return nil, -1, err
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, -1, err
	}
	if response.StatusCode != 200 {
		return nil, -1, errors.New(response.Status)
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, -1, err
	}
	var res struct {
		ErrCode  int      `json:"errcode"`
		ErrMsg   string   `json:"errmsg"`
		UserList []WxUser `json:"userlist"`
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, -1, err
	}
	if res.ErrCode != 0 {
		return res.UserList, res.ErrCode, errors.New(res.ErrMsg)
	}
	return res.UserList, 0, nil
}

// GetDepartmentSpecifiedUser 获取单个成员详情
func GetDepartmentSpecifiedUser(accessToken string, userId string) (*WxUser, int, error) {
	params := url.Values{}
	params.Add("access_token", accessToken)
	params.Add("userid", userId)
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getSpecifiedUserUrl, params.Encode()), nil)
	if err != nil {
		return nil, -1, err
	}
	request.Header.Add("User-Agent", "")
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, -1, err
	}
	if response.StatusCode != 200 {
		return nil, -1, errors.New(response.Status)
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return nil, -1, err
	}
	var res struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		WxUser
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, -1, err
	}
	if res.ErrCode != 0 {
		return &res.WxUser, res.ErrCode, errors.New(res.ErrMsg)
	}
	return &res.WxUser, 0, nil
}
