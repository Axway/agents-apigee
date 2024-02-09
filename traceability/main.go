package main

import (
	"fmt"
	"os"

	// Required Import to setup factory for traceability transport (passivley include.  Do not remove _ (underscore))
	_ "github.com/Axway/agent-sdk/pkg/traceability"
	"github.com/Axway/agent-sdk/pkg/util/exception"

	"github.com/Axway/agents-apigee/traceability/pkg/cmd"
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
		Usage: "apigee.traceability.main",
	}.Do()
}
