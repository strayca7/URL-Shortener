package main

import (
	"os"

	usctlv1 "url-shortener/pkg/usctl/v1"
)

func main() {
	if len(os.Args) < 2 {
		usctlv1.PrintGlobalHelp()
		os.Exit(0)
	}

	usctlv1.CLIHandler(os.Args...)
}
