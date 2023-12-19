package cmd

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/coreos/go-semver/semver"
	"github.com/fasnow/ghttp"
	"github.com/xuri/excelize/v2"
	"idebug/config"
	"idebug/logger"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// generateNewFilename 生成带时间戳的新文件名
func generateNewFilename(filename string) string {
	var ext string
	t := time.Now()
	loc, _ := time.LoadLocation("Asia/Shanghai")             // 设置时区为中国
	chinaTime := t.In(loc)                                   // 转换为中国时间
	formattedTime := chinaTime.Format("2006_01_02_15_04_05") // 格式化时间
	index := strings.LastIndex(filename, ".")
	if index != -1 {
		ext = filename[index+1:]
	}
	baseName := filename[:index] // 去除扩展名
	newFilename := fmt.Sprintf("%s_%s.%s", baseName, formattedTime, ext)
	return newFilename
}

// 保存数据至excel,header长度要和数据列数匹配
func saveToExcel(header []any, data [][]any, filename string) error {
	file := excelize.NewFile()
	data = append([][]any{header}, data...)
	// 添加数据
	for i := 0; i < len(data); i++ {
		row := data[i]
		startCell, err := excelize.JoinCellName("A", i+1)
		if err != nil {
			return err
		}
		if i == 0 {
			// 首行大写
			for j := 0; j < len(row); j++ {
				if value, ok := row[j].(string); ok {
					row[j] = strings.ToUpper(value)
				}
			}
			if err = file.SetSheetRow("Sheet1", startCell, &row); err != nil {
				return err
			}
			continue
		}
		if err = file.SetSheetRow("Sheet1", startCell, &row); err != nil {
			return err
		}
	}

	// 表头颜色填充
	headerStyle, err := file.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#d0cece"}, Pattern: 1, Shading: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	if err != nil {

	}
	err = file.SetCellStyle("Sheet1", "A1", columnNumberToName(len(data[0]))+"1", headerStyle)
	if err != nil {
		return err
	}

	// 添加边框
	dataStyle, err := file.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern"},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
		},
		Border: []excelize.Border{
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	if err != nil {
		return err
	}
	err = file.SetCellStyle("Sheet1", "A1", columnNumberToName(len(data[0]))+strconv.Itoa(len(data)), dataStyle)
	if err != nil {
		return err
	}

	if err := file.SaveAs(filename); err != nil {
		return err
	}
	return nil
}

func columnNumberToName(n int) string {
	name := ""
	for n > 0 {
		n--
		name = string(byte(n%26)+'A') + name
		n /= 26
	}
	return name
}

func Banner() {
	s :=
		`	    ____   ____  __________  __  ________
	   /  _/  / __ \/ ____/ __ )/ / / / ____/
	   / /   / / / / __/ / __  / / / / / __
	 _/ /   / /_/ / /___/ /_/ / /_/ / /_/ /
	/___/  /_____/_____/_____/\____/\____/
		@github.com/fasnow version
企业微信、企业飞书接口调用工具`
	s = strings.Replace(s, "version", config.Version, 1)
	fmt.Println(s)
	logger.Warning("仅用于开发人员用作接口调试,请勿用作其他非法用途")
}

// CheckUpdate latestVersion, releaseUrl, publishTime, content
func CheckUpdate() (string, string, string, string) {
	var httpClient = &ghttp.Client{}
	request, err := http.NewRequest("GET", "https://api.github.com/repos/fasnow/idebug/releases/latest", nil)
	//request.Header.Set("User-Agent", "1")
	if err != nil {
		return "", "", "", ""
	}
	response, err := httpClient.Do(request, ghttp.Options{Timeout: 3 * time.Second})
	if err != nil {
		return "", "", "", ""
	}
	if response.StatusCode != 200 {
		return "", "", "", ""
	}
	body, err := ghttp.GetResponseBody(response.Body)
	if err != nil {
		return "", "", "", ""
	}
	latestVersion, err := jsonparser.GetString(body, "tag_name")
	if err != nil {
		return "", "", "", ""
	}
	currentVersion, err := semver.NewVersion(config.Version[1:])
	if err != nil {
		fmt.Println(err)
	}

	publishTime, err := jsonparser.GetString(body, "published_at")

	v2, err := semver.NewVersion(latestVersion[1:])
	if err != nil {
		fmt.Println(err)
	}
	// 比较版本
	if v2.Compare(*currentVersion) < 1 {
		return "", "", "", ""
	}
	releaseUrl, err := jsonparser.GetString(body, "html_url")
	if err != nil {
		return "", "", "", ""
	}
	content, err := jsonparser.GetString(body, "body")
	if err != nil {
		return "", "", "", ""
	}
	return latestVersion, releaseUrl, publishTime, content
}
