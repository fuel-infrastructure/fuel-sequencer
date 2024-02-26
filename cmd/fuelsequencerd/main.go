package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/fuel-infrastructure/fuel-sequencer/app"
	"github.com/fuel-infrastructure/fuel-sequencer/cmd/fuelsequencerd/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
