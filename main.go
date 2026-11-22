package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/f-gillmann/gogituser/cmd"
)

func main() {
	err := cmd.Execute()
	if err == nil {
		return
	}

	var exitCode cmd.ExitCodeError
	if errors.As(err, &exitCode) {
		os.Exit(int(exitCode))
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
