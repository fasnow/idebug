package utils

import "fmt"

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
)

func Error(err error) {
	fmt.Println(red + "[!] " + reset + err.Error())
}

func Success(str string) {
	fmt.Println(green + "[+] " + reset + str)
}

func Info(str string) {
	fmt.Println("[!] " + str)
}
func Warning(str string) {
	fmt.Println(red + "[!] " + reset + str)
}

func Println(any string) {
	fmt.Println(any)
}
