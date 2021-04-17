package actions

import (
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	buildCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "pushes built squadron units to the registry")
}

var (
	buildCmd = &cobra.Command{
		Use:     "build [UNIT...]",
		Short:   "build or rebuild squadron units",
		Example: "  squadron build frontend backend",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return build(args, cwd, flagFiles, flagPush)
		},
	}
)

func build(args []string, cwd string, files []string, push bool) error {
	sq, err := squadron.New(cwd, "", files)
	if err != nil {
		return err
	}

	units, err := parseUnitArgs(args, sq.GetUnits())
	if err != nil {
		return err
	}

	for _, unit := range units {
		if err := unit.Build(); err != nil {
			return err
		}
	}

	if push {
		for _, unit := range units {
			if err := unit.Push(); err != nil {
				return err
			}
		}
	}

	return nil
}
