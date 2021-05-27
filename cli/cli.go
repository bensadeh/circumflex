package cli

import (
	"os"
	"os/exec"
	"strings"
)

func Less(input string) {
	command := exec.Command("less", "-r")
	command.Stdin = strings.NewReader(input)
	command.Stdout = os.Stdout

	if err := command.Run(); err != nil {
		panic(err)
	}
}

func ClearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}
