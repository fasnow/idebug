package cmd

import (
	"fmt"
	"strings"
)

type flag struct {
	name  string
	value string
	usage string
}

type command struct {
	flag         []flag
	name         string
	arg          string
	usage        string
	childCommand []command
}

var usageSort = []string{"help", "run", "info", "set", "user", "dp", "exit"}

var usageMap = map[string]command{
	"help": {
		name: "help,--help,-h", arg: "", usage: "查看帮助信息",
		childCommand: []command{}},
	"run": {
		name: "run", arg: "", usage: "通过当前设置 获取/刷新 access_token",
		childCommand: []command{}},
	"info": {
		name: "info", arg: "", usage: "查看当前设置",
		childCommand: []command{}},
	"exit": {
		name: "exit", arg: "", usage: "退出程序",
		childCommand: []command{}},
	"set": {
		name: "set",
		childCommand: []command{
			{name: "proxy", arg: "", usage: "代理 [socks5|http]://host:port"},
			{name: "corpid", arg: "", usage: "corpid"},
			{name: "corpsecret", arg: "", usage: "corpsecret"},
		}},
	"dp": {
		name: "dp",
		childCommand: []command{
			//{name: "add", arg: "<data>", usage: "新建部门"},
			//{name: "update", arg: "<data>", usage: "更新部门"},
			//{name: "delete", arg: "<id>", usage: "删除部门 (注：不能删除根部门；不能删除含有子部门、成员的部门)"},
			{name: "tree", arg: "", usage: "根据ID递归获取部门树,不包含用户信息,不提供则递归获取默认部门 (保存至dept_tree.html)"},
			{name: "list", arg: "", usage: "根据部门ID递归获取子部门ID,不提供则递归获取默认部门 (保存至dept_tree_ids.html)"},
			{name: "detail", arg: "", usage: "根据部门ID获取部门详情,包含部门ID,中文名称,上级部门ID,英文名称,部门领导ID"},
			{name: "--print", arg: "", usage: "是否打印过程信息 (默认 否),仅对 tree 和 list 生效"},
		},
	},
	"user": {
		name: "user",
		flag: []flag{
			{name: "--uid", value: "", usage: "根据用户ID查看用户详情"},
			{name: "--did", value: "", usage: "根据部门ID查看所有用户详情 (保存至dept_users.xlsx)"},
			{name: "--re", value: "", usage: "是否递归查询 (默认 否),仅对 --did 生效"},
			{name: "--print", value: "", usage: "输出到控制台数目的数目 (默认 5),仅对 --did 生效"},
			{name: "--dump", value: "", usage: "导出所有部门和用户 (保存至dept_tree_users.html和dept_tree_users.xlsx)"},
		},
		childCommand: []command{},
	},
}

func ShowUsageByName(name string) {
	//cmd := usageMap[name]
	//s := fmt.Sprintf("Usage:\n  %s [command]\nCommand:\n", cmd.name)
	//for _, v := range cmd.childCommand {
	//	s += fmt.Sprintf("  %-20s%s\n", v.name, v.usage)
	//}
	//fmt.Print(s)
}

func ShowAllUsage() {
	s := "Usage:\n  [command] argument\n  [command] [option]\nCommands/Options:\n"
	for _, v := range usageSort {
		value, ok := usageMap[v]
		if !ok {
			continue
		}

		if len(value.childCommand) == 0 {
			s += fmt.Sprintf("  %-17s%s\n", value.name, value.usage)
		} else {
			s += fmt.Sprintf("  %-17s\n", value.name)
		}
		for _, f := range value.flag {
			s += fmt.Sprintf("    %-15s%-15s\n", f.name, f.usage)
		}
		for _, vv := range value.childCommand {
			s += fmt.Sprintf("  %-17s%s\n", strings.Repeat("  ", 1)+vv.name, vv.usage)
			s = showSubCommandUsage(vv.childCommand, 2, s)
		}
	}
	fmt.Print(s)
}

func showSubCommandUsage(child []command, indent int, s string) string {
	for _, vv := range child {
		s += fmt.Sprintf("  %-17s%s\n", strings.Repeat("  ", indent)+vv.name, vv.usage)
		s = showSubCommandUsage(vv.childCommand, indent+1, s)
	}
	return s
}
