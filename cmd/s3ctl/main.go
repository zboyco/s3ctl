package main

import (
	"fmt"
	"os"

	"github.com/zboyco/s3ctl/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
