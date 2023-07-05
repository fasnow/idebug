package main

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

func main() {
	prompt := promptui.Prompt{
		Label:     "",
		Templates: &promptui.PromptTemplates{},
	}

	prompt.AllowEdit = true
	prompt.IsVimMode = true
	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		return
	}

	fmt.Printf("You entered: %s\n", result)
}
