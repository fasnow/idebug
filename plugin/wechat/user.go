package wechat

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fasnow/ghttp"
	"idebug/plugin"
	"net/http"
	"net/url"
)

type UserEntrySimplified struct {
	Name       string `json:"name"`
	UserId     string `json:"userid"`
	Department []int  `json:"department"`
	OpenUserid string `json:"open_userid"`
}

type UserEntry struct {
	UserEntrySimplified
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

type GetUserReq struct {
	req *Req
}

type GetUserReqBuilder struct {
	req *Req
}

func NewGetUserReqBuilder(client *Client) *GetUserReqBuilder {
	builder := &GetUserReqBuilder{}
	builder.req = &Req{
		Client:      client,
		QueryParams: &plugin.QueryParams{},
		PathParams:  &plugin.PathParams{},
	}
	return builder
}

func (builder *GetUserReqBuilder) UserId(id string) *GetUserReqBuilder {
	builder.req.QueryParams.Set("userid", id)
	return builder
}

func (builder *GetUserReqBuilder) Build() *GetUserReq {
	req := &GetUserReq{}
	req.req = builder.req
	return req
}

func (u *user) Get(req *GetUserReq) (*UserEntry, error) {
	token, err := req.req.Client.getAccessTokenFromCache()
	if err != nil {
		return nil, err
	}
	id := req.req.QueryParams.Get("userid")
	if id == "" {
		return nil, errors.New("用户ID不能为空")
	}
	params := url.Values{}
	params.Add("access_token", token)
	params.Add("userid", id)
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getUserUrl, params.Encode()), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("User-Agent", "")
	response, err := u.client.http.Do(request)
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
	var res struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		*UserEntry
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	if res.ErrCode != 0 {
		return res.UserEntry, fmt.Errorf("from server - " + res.ErrMsg)
	}
	return res.UserEntry, nil
}

type GetUsersByDepartmentIdReq struct {
	req *Req
}

type GetUsersByDepartmentIdReqBuilder struct {
	req *Req
}

func NewGetUsersByDepartmentIdReqBuilder(client *Client) *GetUsersByDepartmentIdReqBuilder {
	builder := &GetUsersByDepartmentIdReqBuilder{}
	builder.req = &Req{
		Client:      client,
		QueryParams: &plugin.QueryParams{},
		PathParams:  &plugin.PathParams{},
	}
	return builder
}

func (builder *GetUsersByDepartmentIdReqBuilder) DepartmentId(id string) *GetUsersByDepartmentIdReqBuilder {
	builder.req.QueryParams.Set("department_id", id)
	return builder
}

func (builder *GetUsersByDepartmentIdReqBuilder) Fetch(fetch bool) *GetUsersByDepartmentIdReqBuilder {
	if fetch {
		builder.req.QueryParams.Set("fetch_child", "1")
	}
	return builder
}

func (builder *GetUsersByDepartmentIdReqBuilder) Build() *GetUsersByDepartmentIdReq {
	req := &GetUsersByDepartmentIdReq{}
	req.req = builder.req
	return req
}

func (u *user) GetUsersByDepartmentId(req *GetUsersByDepartmentIdReq) ([]*UserEntry, error) {
	token, err := req.req.Client.getAccessTokenFromCache()
	if err != nil {
		return nil, err
	}
	id := req.req.QueryParams.Get("department_id")
	if id == "" {
		return nil, errors.New("部门ID不能为空")
	}
	params := url.Values{}
	params.Add("access_token", token)
	params.Add("department_id", id)
	fetch := req.req.QueryParams.Get("fetch_child")
	if fetch == "1" {
		params.Add("fetch_child", fetch)
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getDepartmentUserUrl, params.Encode()), nil)
	if err != nil {
		return nil, err
	}
	response, err := u.client.http.Do(request)
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
	var res struct {
		ErrCode  int          `json:"errcode"`
		ErrMsg   string       `json:"errmsg"`
		UserList []*UserEntry `json:"userlist"`
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	if res.ErrCode != 0 {
		return res.UserList, fmt.Errorf("from server - " + res.ErrMsg)
	}
	return res.UserList, nil
}

type GetUsersSimplifiedByDepartmentIdReq struct {
	req *Req
}

type GetUsersSimplifiedByDepartmentIdReqBuilder struct {
	req *Req
}

func NewGetUsersSimplifiedByDepartmentIdReqBuilder(client *Client) *GetUsersSimplifiedByDepartmentIdReqBuilder {
	builder := &GetUsersSimplifiedByDepartmentIdReqBuilder{}
	builder.req = &Req{
		Client:      client,
		QueryParams: &plugin.QueryParams{},
		PathParams:  &plugin.PathParams{},
	}
	return builder
}

func (builder *GetUsersSimplifiedByDepartmentIdReqBuilder) DepartmentId(id string) *GetUsersSimplifiedByDepartmentIdReqBuilder {
	builder.req.QueryParams.Set("department_id", id)
	return builder
}

func (builder *GetUsersSimplifiedByDepartmentIdReqBuilder) Fetch(fetch bool) *GetUsersSimplifiedByDepartmentIdReqBuilder {
	if fetch {
		builder.req.QueryParams.Set("fetch_child", "1")
	}
	return builder
}

func (builder *GetUsersSimplifiedByDepartmentIdReqBuilder) Build() *GetUsersSimplifiedByDepartmentIdReq {
	req := &GetUsersSimplifiedByDepartmentIdReq{}
	req.req = builder.req
	return req
}

func (u *user) GetUsersSimplifiedByDepartmentId(req *GetUsersSimplifiedByDepartmentIdReq) ([]*UserEntrySimplified, error) {
	token, err := req.req.Client.getAccessTokenFromCache()
	if err != nil {
		return nil, err
	}
	id := req.req.QueryParams.Get("department_id")
	if id == "" {
		return nil, errors.New("部门ID不能为空")
	}
	params := url.Values{}
	params.Add("access_token", token)
	params.Add("department_id", id)
	fetch := req.req.QueryParams.Get("fetch_child")
	if fetch == "1" {
		params.Add("fetch_child", fetch)
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getDepartmentSimpleUserUrl, params.Encode()), nil)
	if err != nil {
		return nil, err
	}
	response, err := u.client.http.Do(request)
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
	var res struct {
		ErrCode  int                    `json:"errcode"`
		ErrMsg   string                 `json:"errmsg"`
		UserList []*UserEntrySimplified `json:"userlist"`
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	if res.ErrCode != 0 {
		return res.UserList, fmt.Errorf("from server - " + res.ErrMsg)
	}
	return res.UserList, nil
}
