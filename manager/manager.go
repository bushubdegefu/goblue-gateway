package manager

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	bluegateway = &cobra.Command{
		Use:           "bluegateway",
		Short:         "bluegateway – command-line tool to aid service key managment and api gateway service",
		Long:          ``,
		Version:       "0.0.0",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func Execute() {
	if err := bluegateway.Execute(); err != nil {

		fmt.Println(err)
		os.Exit(1)
	}
}
