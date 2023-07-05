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

type DepartmentEntrySimplified struct {
	ID       int `json:"id"`
	ParentId int `json:"parentid"`
	Order    int `json:"order"`
}

type DepartmentEntry struct {
	DepartmentEntrySimplified
	Name             string   `json:"name"`
	NameEn           string   `json:"name_en"`
	DepartmentLeader []string `json:"department_leader"`
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
	builder.req.QueryParams.Set("id", id)
	return builder
}

func (builder *GetDepartmentReqBuilder) Build() *GetDepartmentReq {
	req := &GetDepartmentReq{}
	req.req = builder.req
	return req
}

func (d *department) Get(req *GetDepartmentReq) (*DepartmentEntry, error) {
	token, err := req.req.Client.getAccessTokenFromCache()
	if err != nil {
		return nil, err
	}
	id := req.req.QueryParams.Get("id")
	if id == "" {
		return nil, errors.New("部门ID不能为空")
	}
	params := url.Values{}
	params.Add("access_token", token)
	params.Add("id", id)
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getDepartmentUrl, params.Encode()), nil)
	if err != nil {
		return nil, err
	}
	response, err := d.client.http.Do(request)
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
	var tmp struct {
		ErrCode    int             `json:"errcode"`
		ErrMsg     string          `json:"errmsg"`
		Department DepartmentEntry `json:"department"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, err
	}
	if tmp.ErrCode != 0 {
		return nil, fmt.Errorf("from server - " + tmp.ErrMsg)
	}
	return &tmp.Department, nil
}

type GetDepartmentListReq struct {
	req *Req
}

type GetDepartmentListReqBuilder struct {
	req *Req
}

func NewGetDepartmentListReqBuilder(client *Client) *GetDepartmentListReqBuilder {
	builder := &GetDepartmentListReqBuilder{}
	builder.req = &Req{
		PathParams:  &plugin.PathParams{},
		QueryParams: &plugin.QueryParams{},
		Client:      client,
	}
	return builder

}

func (builder *GetDepartmentListReqBuilder) DepartmentId(id string) *GetDepartmentListReqBuilder {
	builder.req.QueryParams.Set("id", id)
	return builder
}

func (builder *GetDepartmentListReqBuilder) Build() *GetDepartmentListReq {
	req := &GetDepartmentListReq{}
	req.req = builder.req
	return req
}

// GetList 递归获取
func (d *department) GetList(req *GetDepartmentListReq) ([]*DepartmentEntry, error) {
	token, err := req.req.Client.getAccessTokenFromCache()
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("access_token", token)
	id := req.req.QueryParams.Get("id")
	if id != "" {
		params.Add("id", id)
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getDepartmentListUrl, params.Encode()), nil)
	if err != nil {
		return nil, err
	}
	response, err := d.client.http.Do(request)
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
	var tmp struct {
		ErrCode    int                `json:"errcode"`
		ErrMsg     string             `json:"errmsg"`
		Department []*DepartmentEntry `json:"department"`
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, err
	}
	if tmp.ErrCode != 0 {
		return nil, fmt.Errorf("from server - " + tmp.ErrMsg)
	}
	return tmp.Department, nil
}

type GetDepartmentIdListReq struct {
	req *Req
}

type GetDepartmentIdListReqBuilder struct {
	req *Req
}

func NewGetDepartmentIdListReqBuilder(client *Client) *GetDepartmentIdListReqBuilder {
	builder := &GetDepartmentIdListReqBuilder{}
	builder.req = &Req{
		PathParams:  &plugin.PathParams{},
		QueryParams: &plugin.QueryParams{},
		Client:      client,
	}
	return builder

}

func (builder *GetDepartmentIdListReqBuilder) DepartmentId(id string) *GetDepartmentIdListReqBuilder {
	builder.req.QueryParams.Set("id", id)
	return builder
}

func (builder *GetDepartmentIdListReqBuilder) Build() *GetDepartmentIdListReq {
	req := &GetDepartmentIdListReq{}
	req.req = builder.req
	return req
}

// GetIdList 递归获取
func (d *department) GetIdList(req *GetDepartmentIdListReq) ([]*DepartmentEntrySimplified, error) {
	token, err := req.req.Client.getAccessTokenFromCache()
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("access_token", token)
	id := req.req.QueryParams.Get("id")
	if id != "" {
		params.Add("id", id)
	}
	var res struct {
		ErrCode    int                          `json:"errcode"`
		ErrMsg     string                       `json:"errmsg"`
		Department []*DepartmentEntrySimplified `json:"department_id"`
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", getDepartmentIdListUrl, params.Encode()), nil)
	if err != nil {
		return nil, err
	}
	response, err := d.client.http.Do(request)
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
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	if res.ErrCode != 0 {
		return res.Department, fmt.Errorf("from server - " + res.ErrMsg)
	}
	return res.Department, nil
}
