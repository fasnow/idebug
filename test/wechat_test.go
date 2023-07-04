//go:build test
// +build test

package test

import (
	"encoding/json"
	"github.com/fasnow/ghttp"
	"idebug/plugin/wechat"
	"testing"
)

func TestGetAccessToken(t *testing.T) {
	client := wechat.NewWxClient()
	client.Set("wwa845e703b23918a8", "ZRBNETOv2ywaU0sH_1UJ--dqaYXjqKqLFC_G2s0rWUg")
	token, err := client.GetAccessToken()
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(token)
}

func TestGetDepartment(t *testing.T) {
	ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	client := wechat.NewWxClient()
	client.Set("wx7b47806a86fe27db", "UXbKVR_5SlNbeo5uZnyTOlIel_MhNnqQ7vYXsdKYw4zlxgSkjf63_Vxd35QpIsEp")
	req := wechat.NewGetDepartmentReqBuilder(client).DepartmentId("8221").Build()
	get, err := client.Department.Get(req)
	if err != nil {
		t.Log(err)
		return
	}
	marshal, err := json.Marshal(&get)
	if marshal == nil {
		t.Log("Failed to marshal response")
		return
	}
	if len(marshal) == 0 {
		t.Log("Marshal result is empty")
		return
	}
	a := string(marshal)
	t.Log(a)
}

func TestGetDepartmentList(t *testing.T) {
	ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	client := wechat.NewWxClient()
	client.Set("wx7b47806a86fe27db", "UXbKVR_5SlNbeo5uZnyTOlIel_MhNnqQ7vYXsdKYw4zlxgSkjf63_Vxd35QpIsEp")
	req := wechat.NewGetDepartmentListReqBuilder(client).DepartmentId("3740").Build()
	get, err := client.Department.GetList(req)
	if err != nil {
		t.Log(err)
		return
	}
	marshal, err := json.Marshal(&get)
	if marshal == nil {
		t.Log("Failed to marshal response")
		return
	}
	if len(marshal) == 0 {
		t.Log("Marshal result is empty")
		return
	}
	a := string(marshal)
	t.Log(a)
}

func TestGetDepartmentIdList(t *testing.T) {
	ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	client := wechat.NewWxClient()
	client.Set("wx7b47806a86fe27db", "UXbKVR_5SlNbeo5uZnyTOlIel_MhNnqQ7vYXsdKYw4zlxgSkjf63_Vxd35QpIsEp")
	req := wechat.NewGetDepartmentIdListReqBuilder(client).Build()
	get, err := client.Department.GetIdList(req)
	if err != nil {
		t.Log(err)
		return
	}
	marshal, err := json.Marshal(&get)
	if marshal == nil {
		t.Log("Failed to marshal response")
		return
	}
	if len(marshal) == 0 {
		t.Log("Marshal result is empty")
		return
	}
	a := string(marshal)
	t.Log(a)
}

func TestGetUser(t *testing.T) {
	ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	client := wechat.NewWxClient()
	client.Set("wx7b47806a86fe27db", "UXbKVR_5SlNbeo5uZnyTOlIel_MhNnqQ7vYXsdKYw4zlxgSkjf63_Vxd35QpIsEp")
	req := wechat.NewGetUserReqBuilder(client).UserId("peter").Build()
	get, err := client.User.Get(req)
	if err != nil {
		t.Log(err)
		return
	}
	marshal, err := json.Marshal(&get)
	if marshal == nil {
		t.Log("Failed to marshal response")
		return
	}
	if len(marshal) == 0 {
		t.Log("Marshal result is empty")
		return
	}
	a := string(marshal)
	t.Log(a)
}

func TestGetUsersByDepartmentId(t *testing.T) {
	ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	client := wechat.NewWxClient()
	client.Set("wx7b47806a86fe27db", "UXbKVR_5SlNbeo5uZnyTOlIel_MhNnqQ7vYXsdKYw4zlxgSkjf63_Vxd35QpIsEp")
	req := wechat.NewGetUsersByDepartmentIdReqBuilder(client).DepartmentId("45460").Fetch(true).Build()
	get, err := client.User.GetUsersByDepartmentId(req)
	if err != nil {
		t.Log(err)
		return
	}
	marshal, err := json.Marshal(&get)
	if marshal == nil {
		t.Log("Failed to marshal response")
		return
	}
	if len(marshal) == 0 {
		t.Log("Marshal result is empty")
		return
	}
	a := string(marshal)
	t.Log(a)
}

func TestGetUsersSimplifiedByDepartmentId(t *testing.T) {
	ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	client := wechat.NewWxClient()
	client.Set("wx7b47806a86fe27db", "UXbKVR_5SlNbeo5uZnyTOlIel_MhNnqQ7vYXsdKYw4zlxgSkjf63_Vxd35QpIsEp")
	req := wechat.NewGetUsersSimplifiedByDepartmentIdReqBuilder(client).DepartmentId("45460").Fetch(true).Build()
	get, err := client.User.GetUsersSimplifiedByDepartmentId(req)
	if err != nil {
		t.Log(err)
		return
	}
	marshal, err := json.Marshal(&get)
	if marshal == nil {
		t.Log("Failed to marshal response")
		return
	}
	if len(marshal) == 0 {
		t.Log("Marshal result is empty")
		return
	}
	a := string(marshal)
	t.Log(a)
}
