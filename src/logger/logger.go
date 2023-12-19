package logger

import (
	"fmt"
	"github.com/mattn/go-colorable"
	"os"
	"path/filepath"
	"runtime"
)

var Writer = colorable.NewColorableStdout()

const (
	reset = "\033[0m"
	red   = "\033[31m"
	green = "\033[32m"
	blue  = "\033[34m"
)

func Error(err error) {
	fmt.Fprintln(Writer, red+"[!] "+reset+err.Error())
}

func Success(str string) {
	fmt.Fprintln(Writer, green+"[+] "+reset+str)
}

func Info(str string) {
	fmt.Fprintln(Writer, "[+] "+str)
}
func Warning(str string) {
	fmt.Fprintln(Writer, red+"[!] "+reset+str)
}

func Println(any string) {
	fmt.Fprintln(Writer, Writer, any)
}

func Notice(str string) {
	fmt.Fprintln(Writer, green+"[+] "+str+reset)
}

func ModuleSelectedV1(module string) {
	if module != "" {
		fmt.Fprint(Writer, fmt.Sprintf("idebug %s >", red+module+reset))
		return
	}
	fmt.Fprint(Writer, "idebug >")
}

func ModuleSelectedV2(module string) string {
	if module != "" {
		return fmt.Sprintf("idebug %s >", red+module+reset)
	}
	return "idebug >"
}

func ModuleSelectedV3(module string) string {
	if module != "" {
		return fmt.Sprintf("idebug %s >", module)
	}
	return "idebug >"
}

func FormatError(err error) error {
	pc, _, line, ok := runtime.Caller(1)
	if !ok {
		return err
	}
	dir, e := os.Getwd()
	if e != nil {
		return err
	}
	dir = filepath.ToSlash(dir)
	funcName := runtime.FuncForPC(pc).Name()
	return fmt.Errorf("%s %d line: %s", funcName, line, err)
}
