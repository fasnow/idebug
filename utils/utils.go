package utils

import (
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/coreos/go-semver/semver"
	"github.com/fasnow/ghttp"
	"github.com/tealeg/xlsx"
	"idebug/config"
	"idebug/plugin"
	http2 "net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Department struct {
	plugin.WxDepartment
	User     []*plugin.WxUser
	Children []*Department
}

func Banner() {
	s :=
		`	    ____   ____  __________  __  ________
	   /  _/  / __ \/ ____/ __ )/ / / / ____/
	   / /   / / / / __/ / __  / / / / / __
	 _/ /   / /_/ / /___/ /_/ / /_/ / /_/ /
	/___/  /_____/_____/_____/\____/\____/
		@github.com/fasnow version
用于通过企业微信的 corpid 和 corpsecret 自动获取access_token以调试接口`
	s = strings.Replace(s, "version", config.Version, 1)
	fmt.Println(s)
	Warning("仅用于开发人员用作接口调试,请勿用作其他非法用途")
}

func CheckUpdate() (string, string, string) {
	var httpClient = &ghttp.Client{}
	request, err := http2.NewRequest("GET", "https://api.github.com/repos/fasnow/idebug/releases/latest", nil)
	//request.Header.Set("User-Agent", "1")
	if err != nil {
		return "", "", ""
	}
	response, err := httpClient.Do(request, ghttp.Options{Timeout: 3 * time.Second})
	if err != nil {
		return "", "", ""
	}
	if response.StatusCode != 200 {
		return "", "", ""
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return "", "", ""
	}
	latestVersion, err := jsonparser.GetString(body, "tag_name")
	if err != nil {
		return "", "", ""
	}
	currentVersion, err := semver.NewVersion(config.Version[1:])
	if err != nil {
		fmt.Println(err)
	}

	v2, err := semver.NewVersion(latestVersion[1:])
	if err != nil {
		fmt.Println(err)
	}
	// 比较版本
	if v2.Compare(*currentVersion) < 1 {
		return "", "", ""
	}
	releaseUrl, err := jsonparser.GetString(body, "html_url")
	if err != nil {
		return "", "", ""
	}
	content, err := jsonparser.GetString(body, "body")
	if err != nil {
		return "", "", ""
	}
	return latestVersion, releaseUrl, content
}

// IsFileExists 判断文件是否存在
func IsFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// GenerateNewFilename 生成带时间戳的新文件名
func GenerateNewFilename(filename string) string {
	var ext string
	timestamp := time.Now().Unix()
	index := strings.LastIndex(filename, ".")
	if index != -1 {
		ext = filename[index+1:]
	}
	baseName := filename[:index] // 去除扩展名 ".html"
	newFilename := fmt.Sprintf("%s_%d.%s", baseName, timestamp, ext)
	return newFilename
}

func BuildDepartmentTree(departments []*Department) []*Department {
	departmentMap := make(map[int]*Department)
	rootDepartments := []*Department{}

	// 创建部门映射表
	for _, dept := range departments {
		departmentMap[dept.ID] = dept
	}

	// 构建部门树
	for _, dept := range departments {
		parentID := dept.ParentId
		if parentID == 0 { //根节点为0
			rootDepartments = append(rootDepartments, dept)
		} else {
			parentDept, ok := departmentMap[parentID]
			if ok {
				parentDept.Children = append(parentDept.Children, dept)
			} else {
				rootDepartments = append(rootDepartments, dept)
			}
		}
	}
	return rootDepartments
}

// PrintDepartmentTree 打印部门树
func PrintDepartmentTree(departments []*Department, level int) {
	for _, dept := range departments {
		s := ""
		if level == 0 {
			s += fmt.Sprintf("-ID:%d  %s", dept.ID, dept.Name)
		} else {
			s = fmt.Sprintf("%s-ID:%d  %s", strings.Repeat(" ", 3*level), dept.ID, dept.Name)
		}
		if dept.NameEn != "" {
			s += fmt.Sprintf("  英文名称:%s", dept.NameEn)
		}
		if len(dept.DepartmentLeader) != 0 {
			s += fmt.Sprintf("  领导ID:%s\n", strings.Join(dept.DepartmentLeader, "、"))
		} else {
			s += "\n"
		}
		fmt.Print(s)
		PrintDepartmentTree(dept.Children, level+1)
	}
}

// PrintDepartmentIdsTree 打印部门ID树
func PrintDepartmentIdsTree(departments []*Department, level int) {
	for _, dept := range departments {
		s := ""
		if level == 0 {
			s += fmt.Sprintf("-ID:%d\n", dept.ID)
		} else {
			s += fmt.Sprintf("%s-ID:%d\n", strings.Repeat(" ", 3*level), dept.ID)
		}
		fmt.Print(s)
		PrintDepartmentIdsTree(dept.Children, level+1)
	}
}

// GenerateDepartmentTreeHTML 生成包含部门信息的部门树HTML代码
func GenerateDepartmentTreeHTML(departments []*Department, level int) string {
	html := ""
	for _, dept := range departments {
		html += fmt.Sprintf("<div style=\"margin-left:%dem;\">", level)

		// 添加折叠/展开按钮
		if len(dept.Children) > 0 {
			html += fmt.Sprintf("<span class=\"toggle\" onclick=\"toggleDepartment(this)\">-</span>")
		} else {
			html += "<span class=\"empty-toggle\"></span>"
		}

		s := fmt.Sprintf("ID:%d&nbsp;&nbsp;%s", dept.ID, dept.Name)
		if dept.NameEn != "" {
			s += fmt.Sprintf("&nbsp;&nbsp;英文名称:%s", dept.NameEn)
		}
		if len(dept.DepartmentLeader) != 0 {
			s += fmt.Sprintf("&nbsp;&nbsp;领导ID:%s\n", strings.Join(dept.DepartmentLeader, "、"))
		}

		// 添加部门名称
		html += fmt.Sprintf("<span class=\"department\">%s</span>", s)

		// 递归生成子部门树
		if len(dept.Children) > 0 {
			html += GenerateDepartmentTreeHTML(dept.Children, level+1)
		}

		html += "</div>"
	}
	return html
}

// GenerateDepartmentIdsTreeHTML 生成只包含ID的部门树HTML代码
func GenerateDepartmentIdsTreeHTML(departments []*Department, level int) string {
	html := ""
	for _, dept := range departments {
		html += fmt.Sprintf("<div style=\"margin-left:%dem;\">", level)

		// 添加折叠/展开按钮
		if len(dept.Children) > 0 {
			html += fmt.Sprintf("<span class=\"toggle\" onclick=\"toggleDepartment(this)\">-</span>")
		} else {
			html += "<span class=\"empty-toggle\"></span>"
		}

		s := fmt.Sprintf("ID:%d", dept.ID)

		// 添加部门ID
		html += fmt.Sprintf("<span class=\"department\">%s</span>", s)

		// 递归生成子部门树
		if len(dept.Children) > 0 {
			html += GenerateDepartmentTreeHTML(dept.Children, level+1)
		}

		html += "</div>"
	}
	return html
}

// SaveDepartmentTreeToHTML 生成含部门信息的部门树并将其输出为HTML文档
func SaveDepartmentTreeToHTML(departments []*Department, filename string) (string, error) {
	index := strings.LastIndex(filename, ".html")
	if index == -1 {
		filename = filename + ".html"
	}
	tmp := filename
	// 判断文件是否存在
	if IsFileExists(filename) {
		newFilename := GenerateNewFilename(filename)
		filename = newFilename
	}
	file, err := os.Create(filename)
	if err != nil {
		return "", errors.New("创建HTML文件失败: " + err.Error())
	}
	defer file.Close()
	departmentTreeHTML := GenerateDepartmentTreeHTML(departments, 0)
	htmlDocument := GenerateTreeHTMLDocument(departmentTreeHTML)
	_, err = file.WriteString(htmlDocument)
	if err != nil {
		return "", errors.New("无法写入HTML内容到文件: " + err.Error())
	}
	if filename == tmp {
		return fmt.Sprintf("文件已保存至 %s", filename), nil
	}
	return fmt.Sprintf("文件 %s 已存在,已保存至 %s", tmp, filename), nil
}

// SaveDepartmentIdsTreeToHTML 生成只包含ID的部门树并将其输出为HTML文档
func SaveDepartmentIdsTreeToHTML(departments []*Department, filename string) (string, error) {
	index := strings.LastIndex(filename, ".html")
	if index == -1 {
		filename = filename + ".html"
	}
	tmp := filename
	// 判断文件是否存在
	if IsFileExists(filename) {
		newFilename := GenerateNewFilename(filename)
		filename = newFilename
	}
	file, err := os.Create(filename)
	if err != nil {
		return "", errors.New("创建HTML文件失败: " + err.Error())
	}
	defer file.Close()
	departmentTreeHTML := GenerateDepartmentIdsTreeHTML(departments, 0)
	htmlDocument := GenerateTreeHTMLDocument(departmentTreeHTML)
	_, err = file.WriteString(htmlDocument)
	if err != nil {
		return "", errors.New("无法写入HTML内容到文件: " + err.Error())
	}
	if filename == tmp {
		return fmt.Sprintf("文件已保存至 %s", filename), nil
	}
	return fmt.Sprintf("文件 %s 已存在,已保存至 %s", tmp, filename), nil
}

// GenerateDepartmentTreeWithUsersHTML 生成包含部门信息和用户信息的部门树HTML代码
func GenerateDepartmentTreeWithUsersHTML(departments []*Department, level int) string {
	html := ""
	for _, dept := range departments {
		html += fmt.Sprintf("<div style=\"margin-left:%dem;\">", level)

		s := fmt.Sprintf("ID:%d&nbsp;&nbsp;%s", dept.ID, dept.Name)
		if dept.NameEn != "" {
			s += fmt.Sprintf("&nbsp;&nbsp;英文名称:%s", dept.NameEn)
		}
		if len(dept.DepartmentLeader) != 0 {
			s += fmt.Sprintf("&nbsp;&nbsp;领导ID:%s\n", strings.Join(dept.DepartmentLeader, "、"))
		}

		// 添加折叠/展开按钮
		if len(dept.Children) > 0 {
			html += fmt.Sprintf("<span class=\"toggle\" onclick=\"toggleDepartment(this)\">-</span>")
		} else {
			html += "<span class=\"empty-toggle\"></span>"
		}
		html += fmt.Sprintf("<span class=\"department\">%s</span>", s)

		// 添加部门名称和用户列表的父级容器
		html += fmt.Sprintf("<div class=\"department-container\">")
		a := ""
		for i := 0; i < len(dept.User); i++ {
			user := dept.User[i]
			m := fmt.Sprintf("%s&nbsp;&nbsp;%s", user.UserId, user.Name)
			if user.Position != "" {
				m += fmt.Sprintf("&nbsp;&nbsp;%s", user.Position)
			}
			m += fmt.Sprintf("&nbsp;&nbsp;%s&nbsp;&nbsp;%s&nbsp;&nbsp;%s", user.Mobile, user.Email, user.QrCode)
			a += fmt.Sprintf("<li>%s</li>", m)
		}
		html += fmt.Sprintf("<ul class=\"user-list\">%s</ul>", a)

		// 递归生成子部门树
		if len(dept.Children) > 0 {
			html += GenerateDepartmentTreeWithUsersHTML(dept.Children, level+1)
		}

		html += "</div></div>"
	}
	return html
}

// GenerateTreeHTMLDocument 生成完整的HTML文档
func GenerateTreeHTMLDocument(content string) string {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				.toggle {
					margin-right: 5px;
					cursor: pointer;
				}
				.empty-toggle {
					width: 14px;
					display: inline-block;
				}
				.department {
					font-weight: bold;
				}
				.collapse {
					display: none;
				}
				ul{
					margin:0px;
				}
			</style>
			<script>
				function toggleDepartment(button) {
					var div = button.parentNode;
					var subDepartments = div.getElementsByTagName("div");
					for (var i = 0; i < subDepartments.length; i++) {
						subDepartments[i].classList.toggle("collapse");
					}
					button.textContent = button.textContent === "+" ? "-" : "+";
				}

				function expandAll() {
					var departments = document.getElementsByClassName("toggle");
					for (var i = 0; i < departments.length; i++) {
						var button = departments[i];
						var div = button.parentNode;
						var subDepartments = div.getElementsByTagName("div");
						for (var j = 0; j < subDepartments.length; j++) {
							subDepartments[j].classList.remove("collapse");
						}
						button.textContent = "-";
					}
				}

				function collapseAll() {
					var departments = document.getElementsByClassName("toggle");
					for (var i = 0; i < departments.length; i++) {
						var button = departments[i];
						var div = button.parentNode;
						var subDepartments = div.getElementsByTagName("div");
						for (var j = 0; j < subDepartments.length; j++) {
							subDepartments[j].classList.add("collapse");
						}
						button.textContent = "+";
					}
				}
			</script>
		</head>
		<body>
			<div id="departmentTree">
				<div>
					<button onclick="expandAll()">全部展开</button>
					<button onclick="collapseAll()">全部折叠</button>
				</div>
				%s
			</div>
		</body>
		</html>
	`
	return fmt.Sprintf(html, content)
}

// SaveDepartmentTreeWithUsersToHTML 生成包含部门信息和用户信息的部门树并将其输出为HTML文档
func SaveDepartmentTreeWithUsersToHTML(departments []*Department, filename string) (string, error) {
	index := strings.LastIndex(filename, ".html")
	if index == -1 {
		filename = filename + ".html"
	}
	tmp := filename
	// 判断文件是否存在
	if IsFileExists(filename) {
		newFilename := GenerateNewFilename(filename)
		filename = newFilename
	}
	file, err := os.Create(filename)
	if err != nil {
		return "", errors.New("创建HTML文件失败: " + err.Error())
	}
	defer file.Close()
	departmentTreeHTML := GenerateDepartmentTreeWithUsersHTML(departments, 0)
	htmlDocument := GenerateTreeHTMLDocument(departmentTreeHTML)
	_, err = file.WriteString(htmlDocument)
	if err != nil {
		return "", errors.New("无法写入HTML内容到文件: " + err.Error())
	}
	if filename == tmp {
		return fmt.Sprintf("文件已保存至 %s", filename), nil
	}
	return fmt.Sprintf("%s 已存在,已另存为 %s", tmp, filename), nil
}

// SaveUserTreeToExcel 生成包含部门名称的用户信息的XLSX文档
func SaveUserTreeToExcel(dept []*Department, filename string) (string, error) {
	index := strings.LastIndex(filename, ".xlsx")
	if index == -1 {
		filename = filename + ".xlsx"
	}
	tmp := filename
	// 判断文件是否存在
	if IsFileExists(filename) {
		newFilename := GenerateNewFilename(filename)
		filename = newFilename
	}
	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()

	// 创建一个工作表
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		return "", errors.New("创建工作表失败: " + err.Error())
	}

	// 设置表头
	headers := []string{"部门", "用户ID", "姓名", "性别", "电话号码", "邮箱", "职位", "微信二维码"}
	headerRow := sheet.AddRow()
	for _, header := range headers {
		cell := headerRow.AddCell()
		cell.Value = header
	}
	var data [][]string
	// 写入内容
	for _, d := range dept {
		for _, user := range d.User {
			gender := ""
			if user.Gender == "1" {
				gender = "男"
			} else if user.Gender == "0" {
				gender = "女"
			}
			s := []string{d.Name, user.UserId, user.Name, gender, user.Mobile, user.Email, user.Position, user.QrCode}
			data = append(data, s)
		}
		write(&data, d.Children)
	}

	for _, row := range data {
		dataRow := sheet.AddRow()
		for _, cellValue := range row {
			cell := dataRow.AddCell()
			cell.Value = cellValue
		}
	}
	// 保存文件
	err = file.Save(filename)
	if err != nil {
		return "", errors.New("保存 Excel 文件失败: " + err.Error())
	}
	if tmp != filename {
		return fmt.Sprintf("%s 已存在,已另存为 %s", tmp, filename), nil
	}
	return fmt.Sprintf("文件已保存至 %s", filename), nil
}

// SaveUserToExcel 生成包含所属部门ID的用户信息的XLSX文档
func SaveUserToExcel(users []plugin.WxUser, filename string) (string, error) {
	index := strings.LastIndex(filename, ".xlsx")
	if index == -1 {
		filename = filename + ".xlsx"
	}
	tmp := filename
	// 判断文件是否存在
	if IsFileExists(filename) {
		newFilename := GenerateNewFilename(filename)
		filename = newFilename
	}
	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()

	// 创建一个工作表
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		return "", errors.New("创建工作表失败: " + err.Error())
	}

	// 设置表头
	headers := []string{"所属部门ID", "用户ID", "姓名", "性别", "电话号码", "邮箱", "职位", "微信二维码"}
	headerRow := sheet.AddRow()
	for _, header := range headers {
		cell := headerRow.AddCell()
		cell.Value = header
	}
	var data [][]string
	// 写入内容
	for _, user := range users {
		var s []string
		for _, num := range user.Department {
			s = append(s, strconv.Itoa(num))
		}
		var gender string
		if user.Gender == "1" {
			gender = "男"
		} else if user.Gender == "0" {
			gender = "女"
		}
		d := []string{strings.Join(s, " "), user.UserId, user.Name, gender, user.Mobile, user.Email, user.Position, user.QrCode}
		data = append(data, d)
	}

	for _, row := range data {
		dataRow := sheet.AddRow()
		for _, cellValue := range row {
			cell := dataRow.AddCell()
			cell.Value = cellValue
		}
	}
	// 保存文件
	err = file.Save(filename)
	if err != nil {
		return "", errors.New("保存 Excel 文件失败: " + err.Error())
	}
	if tmp != filename {
		return fmt.Sprintf("%s 已存在,已另存为 %s", tmp, filename), nil
	}
	return fmt.Sprintf("文件已保存至 %s", filename), nil
}

func write(data *[][]string, dept []*Department) {
	for _, d := range dept {
		for _, user := range d.User {
			gender := ""
			if user.Gender == "1" {
				gender = "男"
			} else if user.Gender == "0" {
				gender = "女"
			}
			s := []string{d.Name, user.UserId, user.Name, gender, user.Mobile, user.Email, user.Position, user.QrCode}
			*data = append(*data, s)
		}
		write(data, d.Children)
	}
}
