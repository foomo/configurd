package actions

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

var (
	listCmd = &cobra.Command{
		Use:     "list",
		Short:   "list squadron units",
		Example: "  squadron list",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return list(cwd, flagNamespace, flagFiles)
		},
	}
)

func list(cwd, namespace string, files []string) error {
	sq, err := squadron.New(cwd, namespace, files)
	if err != nil {
		return err
	}

	for name, _ := range sq.GetUnits() {
		fmt.Println(name)
	}

	return nil
}
