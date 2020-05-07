package actions

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "latest"

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "prints cli version",
		Long:  "prints the current installed cli version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
)
