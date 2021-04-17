package actions

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "latest"

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
)
