package actions

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/util"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use: "squadron",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logrus.SetLevel(logrus.InfoLevel)
			if flagVerbose {
				logrus.SetLevel(logrus.TraceLevel)
			}
			var err error
			if cmd.Name() == "help" || cmd.Name() == "init" || cmd.Name() == "version" {
				return nil
			}
			// cwd
			if err = util.ValidatePath(".", &cwd); err != nil {
				return err
			}
			return nil
		},
	}

	cwd           string
	flagVerbose   bool
	flagNamespace string
	flagBuild     bool
	flagPush      bool
	flagDiff      bool
	flagFiles     []string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "show more output")
	rootCmd.PersistentFlags().StringSliceVarP(&flagFiles, "file", "f", []string{"squadron.yaml"}, "specify alternative squadron files")

	rootCmd.AddCommand(upCmd, downCmd, buildCmd, listCmd, generateCmd, configCmd, versionCmd, completionCmd, templateCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

// parseExtraArgs ...
func parseExtraArgs(args []string) (out []string, extraArgs []string) {
	for i, arg := range args {
		if strings.HasPrefix(arg, "--") && i > 0 {
			return args[:i], args[i:]
		} else if strings.HasPrefix(arg, "--") {
			return nil, args
		}
	}
	return args, nil
}

// parseUnitArgs helper
func parseUnitArgs(args []string, units map[string]squadron.Unit) (map[string]squadron.Unit, error) {
	if len(args) == 0 {
		return units, nil
	}
	ret := map[string]squadron.Unit{}
	for _, arg := range args {
		if unit, ok := units[arg]; ok {
			ret[arg] = unit
		} else {
			return nil, errors.Errorf("unknown unit name %s", arg)
		}
	}
	return ret, nil
}
