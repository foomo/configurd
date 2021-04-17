package actions

import (
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	templateCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
}

var (
	templateCmd = &cobra.Command{
		Use:     "template [UNIT...]",
		Short:   "render chart templates locally and display the output",
		Example: "  squadron template frontend backend --namespace demo",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return template(args, cwd, flagNamespace, flagFiles)
		},
	}
)

func template(args []string, cwd, namespace string, files []string) error {
	sq, err := squadron.New(cwd, namespace, files)
	if err != nil {
		return err
	}

	args, helmArgs := parseExtraArgs(args)
	units, err := parseUnitArgs(args, sq.GetUnits())
	if err != nil {
		return err
	}

	if err := sq.Generate(sq.GetUnits()); err != nil {
		return err
	} else if err := sq.Template(units, helmArgs); err != nil {
		return err
	}

	return nil
}
