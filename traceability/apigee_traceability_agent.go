package main

import (
	"fmt"
	"os"

	// Required Import to setup factory for traceability transport (passivley include.  Do not remove _ (underscore))
	_ "github.com/Axway/agent-sdk/pkg/traceability"

	// CHANGE_HERE - Change the import path(s) below to reference packages correctly
	"github.com/Axway/agents-apigee/traceability/pkg/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
