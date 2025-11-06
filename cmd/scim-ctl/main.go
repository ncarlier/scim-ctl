package main

import (
	"os"

	"github.com/idf-educ/idm/scim-ctl/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}