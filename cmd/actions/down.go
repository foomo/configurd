package actions

import (
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	downCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
}

var (
	downCmd = &cobra.Command{
		Use:     "down [UNIT...]",
		Short:   "uninstalls the squadron or given units",
		Example: "  squadron down frontend backend --namespace demo",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return down(args, cwd, flagNamespace, flagFiles)
		},
	}
)

func down(args []string, cwd, namespace string, files []string) error {
	sq, err := squadron.New(cwd, namespace, files)
	if err != nil {
		return err
	}

	args, helmArgs := parseExtraArgs(args)
	units, err := parseUnitArgs(args, sq.GetUnits())
	if err != nil {
		return err
	}

	return sq.Down(units, helmArgs)
}
