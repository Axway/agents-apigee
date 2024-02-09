package main

import (
	"fmt"
	"os"

	"github.com/Axway/agent-sdk/pkg/util/exception"
	"github.com/Axway/agents-apigee/discovery/pkg/cmd"
)

func main() {
	exception.Block{
		Try: func() {
			os.Setenv("AGENTFEATURES_VERSIONCHECKER", "false")
			if err := cmd.RootCmd.Execute(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
		Catch: func(error) {},
		Usage: "apigee.discovery.main",
	}.Do()
}
