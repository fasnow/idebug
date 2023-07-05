//go:build test
// +build test

// 在测试文件中添加构建标签
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chzyer/readline/runes"
	"github.com/fasnow/ghttp"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	"github.com/nsf/termbox-go"
	"idebug/plugin/feishu"
	"log"
	"reflect"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	timestamp := int64(1687651200)      // 已知的时间戳（以秒为单位）
	joinTime := time.Unix(timestamp, 0) // 使用time.Unix()函数将时间戳转换为本地时间
	fmt.Println(joinTime.String())
	//// 已知的时间戳（以秒为单位）
	//timestamp := int64(0)
	//// 使用time.Unix()函数将时间戳转换为本地时间
	//joinTime := time.Unix(timestamp, 0)
	//fmt.Println(joinTime)
	//fmt.Println(fmt.Sprintf("%-12s:", "状态"))
	//fmt.Println(fmt.Sprintf("%-12s:", "名称"))
	//fmt.Println(fmt.Sprintf("%-13s:", "ID"))
	//fmt.Println(fmt.Sprintf("%-13s:", "OPEN_ID"))
	//fmt.Println(fmt.Sprintf("%-11s:", "上级部门"))
	//fmt.Println(fmt.Sprintf("%-11s:", "主管用户"))
	//fmt.Println(fmt.Sprintf("%-11s:", "负责人"))
	//fmt.Println(fmt.Sprintf("%-11s:", "用户个数"))
	//fmt.Println(fmt.Sprintf("%-9s:", "主属用户个数"))
	//fmt.Println(fmt.Sprintf("%-12s:", "部门HRBP"))
	//regex := regexp.MustCompile("^(http|https|socks5|socks4)://[0-9.]+:[0-9]+$")
	//
	//str := "socks5://127.0.0.1:8080"
	//
	//if regex.MatchString(str) {
	//	fmt.Println("Matched")
	//} else {
	//	fmt.Println("Not matched")
	//}
	//var tmp struct {
	//	Password struct {
	//		EntEmailPassword string `json:"ent_email_password"`
	//	} `json:"password"`
	//	UserID string `json:"user_id"`
	//}
	//tmp.Password.EntEmailPassword = "\"123;'@5"
	//tmp.UserID = "userId"
	//postData, _ := json.Marshal(tmp)
	//t.Log(string(postData))
}

type twidth struct {
	r      []rune
	length int
}

func TestRuneWidth(t *testing.T) {
	ru := []twidth{
		{[]rune("☭"), 1},
		{[]rune("a"), 1},
		{[]rune("你"), 2},
		{runes.ColorFilter([]rune("☭\033[13;1m你")), 3},
	}
	for _, r := range ru {
		if w := runes.WidthAll(r.r); w != r.length {
			t.Fatal("result not expect", r.r, r.length, w)
		}
	}
}

type tagg struct {
	r      [][]rune
	e      [][]rune
	length int
}

func TestAggRunes(t *testing.T) {
	ru := []tagg{
		{
			[][]rune{[]rune("ab"), []rune("a"), []rune("abc")},
			[][]rune{[]rune("b"), []rune(""), []rune("bc")},
			1,
		},
		{
			[][]rune{[]rune("addb"), []rune("ajkajsdf"), []rune("aasdfkc")},
			[][]rune{[]rune("ddb"), []rune("jkajsdf"), []rune("asdfkc")},
			1,
		},
		{
			[][]rune{[]rune("ddb"), []rune("ajksdf"), []rune("aasdfkc")},
			[][]rune{[]rune("ddb"), []rune("ajksdf"), []rune("aasdfkc")},
			0,
		},
		{
			[][]rune{[]rune("ddb"), []rune("ddajksdf"), []rune("ddaasdfkc")},
			[][]rune{[]rune("b"), []rune("ajksdf"), []rune("aasdfkc")},
			2,
		},
	}
	for _, r := range ru {
		same, off := runes.Aggregate(r.r)
		if off != r.length {
			t.Fatal("result not expect", off)
		}
		if len(same) != off {
			t.Fatal("result not expect", same)
		}
		if !reflect.DeepEqual(r.r, r.e) {
			t.Fatal("result not expect")
		}
	}
}

func TestReadLine(t *testing.T) {
	err := termbox.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	// 在这里编写你的终端交互逻辑

	// 示例：读取并处理用户按键事件
	for {
		event := termbox.PollEvent()
		if event.Type == termbox.EventKey {
			switch event.Key {
			case termbox.KeyEsc:
				// 用户按下 ESC 键，退出循环
				return
			default:
				// 处理其他按键事件
				// 根据需要编写你的逻辑
			}
		}
	}
}

func TestFeiShu(t *testing.T) {
	//ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	//client, err := feishu.NewClient("cli_9e09b73679f5500d", "Q2MVkyEbvrUxCFZZJN8WjgjevMT61wk1")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//deptIds, groupIds, userIds, err := client.GetAuthScope()
	//if err != nil {
	//	fmt.Println(2, err)
	//	return
	//}
	//fmt.Println(deptIds)
	//fmt.Println(groupIds)
	//fmt.Println(userIds)

	// 创建 Client
	client := lark.NewClient("appID", "appSecret")

	// 创建请求对象
	req := larkcontact.NewGetDepartmentReqBuilder().
		DepartmentId(`D096`).
		Build()

	// 发起请求
	_, _ = client.Contact.Department.Get(context.Background(), req)
	_, _ = client.Contact.Department.Get(context.Background(), req)
}

func TestFeiShuGetDepartment(t *testing.T) {
	err := ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	if err != nil {
		return
	}
	client := feishu.NewClient()
	client.Set("cli_9e09b73679f5500d", "Q2MVkyEbvrUxCFZZJN8WjgjevMT61wk1")
	req := feishu.NewGetDepartmentReqBuilder(client).
		DepartmentId("od-72560b0f1a7e7151ff62809547f4a0f0").
		Build()
	if err != nil {
		return
	}
	get, err := client.Department.Get(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	marshal, err := json.Marshal(get)
	if err != nil {
		fmt.Println(err)
		return
	}
	if marshal == nil {
		fmt.Println("Failed to marshal response")
		return
	}
	if len(marshal) == 0 {
		fmt.Println("Marshal result is empty")
		return
	}
	t.Log(string(marshal))
}

func TestGetDepartmentChildren(t *testing.T) {
	err := ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	if err != nil {
		return
	}
	client := feishu.NewClient()
	client.Set("cli_9e09b73679f5500d", "Q2MVkyEbvrUxCFZZJN8WjgjevMT61wk1")
	req := feishu.NewGetDepartmentChildrenReqBuilder(client).
		DepartmentId("od-72560b0f1a7e7151ff62809547f4a0f0").
		Build()
	if err != nil {
		return
	}
	get, err := client.Department.Children(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	marshal, err := json.Marshal(get)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Log(string(marshal))
}

// 不好用,需要其他权限
func TestGetBatchDepartment(t *testing.T) {
	err := ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	if err != nil {
		return
	}
	client := feishu.NewClient()
	client.Set("cli_9e09b73679f5500d", "Q2MVkyEbvrUxCFZZJN8WjgjevMT61wk1")
	req := feishu.NewGetBatchDepartmentReqBuilder(client).
		DepartmentIds([]string{"od-72560b0f1a7e7151ff62809547f4a0f0"}).
		Build()
	if err != nil {
		return
	}
	get, err := client.Department.Bath(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	marshal, err := json.Marshal(get)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Log(string(marshal))
}

func TestFeiShuGetUsersByDepartmentId(t *testing.T) {
	err := ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	if err != nil {
		return
	}
	client := feishu.NewClient()
	client.Set("cli_9e09b73679f5500d", "Q2MVkyEbvrUxCFZZJN8WjgjevMT61wk1")
	req := feishu.NewGetUsersByDepartmentIdReqBuilder(client).DepartmentId("od-72560b0f1a7e7151ff62809547f4a0f0").Build()
	get, err := client.User.GetUsersByDepartmentId(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	marshal, err := json.Marshal(get)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Log(string(marshal))
}

func TestFeiShuGetUser(t *testing.T) {
	err := ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	if err != nil {
		return
	}
	client := feishu.NewClient()
	client.Set("cli_9e09b73679f5500d", "Q2MVkyEbvrUxCFZZJN8WjgjevMT61wk1")
	req := feishu.NewGetUserReqBuilder(client).UserId("8198878").UserIdType("user_id").Build()
	get, err := client.User.Get(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	marshal, err := json.Marshal(get)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Log(string(marshal))
}

func TestFeiShuUserEmailPasswordChange(t *testing.T) {
	err := ghttp.SetGlobalProxy("http://127.0.0.1:8080")
	if err != nil {
		return
	}
	client := feishu.NewClient()
	client.Set("cli_9e09b73679f5500d", "Q2MVkyEbvrUxCFZZJN8WjgjevMT61wk1")
	req := feishu.NewUserEmailPasswordChangeReqBuilder(client).
		PostData("fdsafdfasfadafs", "fadafsadf").
		Build()
	err = client.User.EmailPasswordUpdate(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Log("ok")
}
