package utils

import (
	"encoding/json"
	"io"
	"os"
	"sort"
	"strings"
)

// IsFileExists 判断文件是否存在
func IsFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func ConvertToReader(data any) io.Reader {
	marshal, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	return strings.NewReader(string(marshal))
}

func ConvertString2Reader(data string) io.Reader {
	return strings.NewReader(data)
}

func StringInList(target string, strList []string) bool {
	sort.Strings(strList)
	index := sort.SearchStrings(strList, target)
	//index的取值：[0,len(str_array)]
	if index < len(strList) && strList[index] == target { //需要注意此处的判断，先判断 &&左侧的条件，如果不满足则结束此处判断，不会再进行右侧的判断
		return true
	}
	return false
}
