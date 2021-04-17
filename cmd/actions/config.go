package actions

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

var (
	configCmd = &cobra.Command{
		Use:     "config",
		Short:   "generate and view the squadron config",
		Example: "  squadron config --file squadron.yaml --file squadron.override.yaml",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return config(cwd, flagFiles)
		},
	}
)

func config(cwd string, files []string) error {
	sq, err := squadron.New(cwd, "", files)
	if err != nil {
		return err
	}

	cf, err := sq.GetConfigYAML()
	if err != nil {
		return err
	}

	fmt.Println(string(cf))
	return nil
}
