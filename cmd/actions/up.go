package actions

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	upCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
	upCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "builds or rebuilds units")
	upCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "pushes units to the registry")
	upCmd.Flags().BoolVar(&flagDiff, "diff", false, "preview upgrade as a coloured diff")
}

var (
	upCmd = &cobra.Command{
		Use:     "up [UNIT...]",
		Short:   "installs the squadron or given units",
		Example: "  squadron up frontend backend --namespace demo --build --push -- --dry-run",
		RunE: func(cmd *cobra.Command, args []string) error {
			return up(args, cwd, flagNamespace, flagBuild, flagPush, flagDiff, flagFiles)
		},
	}
)

func up(args []string, cwd, namespace string, build, push, diff bool, files []string) error {
	sq, err := squadron.New(cwd, namespace, files)
	if err != nil {
		return err
	}

	args, helmArgs := parseExtraArgs(args)
	units, err := parseUnitArgs(args, sq.GetUnits())

	if err != nil {
		return err
	}

	if build {
		for _, unit := range units {
			if err := unit.Build(); err != nil {
				return err
			}
		}
	}

	if push {
		for _, unit := range units {
			if err := unit.Push(); err != nil {
				return err
			}
		}
	}

	if err := sq.Generate(units); err != nil {
		return err
	}

	if !diff {
		return sq.Up(units, helmArgs)
	} else if out, err := sq.Diff(units, helmArgs); err != nil {
		return err
	} else {
		fmt.Println(out)
	}

	return nil
}
