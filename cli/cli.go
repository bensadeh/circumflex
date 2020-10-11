package cli

import (
	"os"
	"os/exec"
	"strings"
)

func Less(output string) {
	command := exec.Command("less", "-r")
	command.Stdin = strings.NewReader(output)
	command.Stdout = os.Stdout

	err := command.Run()
	if err != nil {
		panic(err)
	}
}

func Clear() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}