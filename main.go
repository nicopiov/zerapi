package main

import (
	"fmt"
	"os"

	"github.com/nicopiov/zerapi/cmd"
	"github.com/nicopiov/zerapi/internal/util"
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", util.Error("error:"), err)
		os.Exit(1)
	}
}
