package main

import (
	"os"

	"github.com/ncarlier/scim-ctl/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
