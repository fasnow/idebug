package utils

import (
	"fmt"
	"github.com/mattn/go-colorable"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
)

func Error(err error) {
	writer := colorable.NewColorableStdout()
	fmt.Fprintln(writer, red+"[!] "+reset+err.Error())
}

func Success(str string) {
	writer := colorable.NewColorableStdout()
	fmt.Fprintln(writer, green+"[+] "+reset+str)
}

func Info(str string) {
	writer := colorable.NewColorableStdout()
	fmt.Fprintln(writer, "[!] "+str)
}
func Warning(str string) {
	writer := colorable.NewColorableStdout()
	fmt.Fprintln(writer, red+"[!] "+reset+str)
}

func Println(any string) {
	writer := colorable.NewColorableStdout()
	fmt.Fprintln(writer, writer, any)
}
