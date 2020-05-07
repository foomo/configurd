package actions

import (
	"path"

	"github.com/foomo/configurd"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [NAME]",
	Short: "initializes an example",
	Long:  "initializes an example project with configurd",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := initialize(args[0], flagDir, flagVerbose)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func initialize(name, dir string, flagVerbose bool) (string, error) {
	output, err := configurd.Init(log, path.Join(dir, name), flagVerbose)
	if err != nil {
		return output, outputErrorf(output, err, "could not initialize an example configuraiton")
	}
	return output, nil
}
