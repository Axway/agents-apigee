package main

import (
	"fmt"
	"os"

	"github.com/Axway/agents-apigee/discovery/pkg/cmd"
)

func main() {
	os.Setenv("AGENTFEATURES_VERSIONCHECKER", "false")
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
