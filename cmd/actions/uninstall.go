package actions

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	uninstallCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
}

var (
	uninstallCmd = &cobra.Command{
		Use:   "uninstall [GROUP] -n {NAMESPACE} -t {TAG}",
		Short: "uninstalls a group",
		Long:  "uninstalls a group with given namespace and tag version",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := uninstall(args[0], flagNamespace, flagVerbose)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func uninstall(group, namespace string, flagVerbose bool) (string, error) {
	cnf := mustNewConfigurd()
	sis := cnf.GetServiceItems(namespace, group)
	if len(sis) == 0 {
		return "", fmt.Errorf("could not find any service for namespace: %v and service group: %v", namespace, group)
	}
	output, err := cnf.Uninstall(log, sis, namespace, flagVerbose)
	if err != nil {
		return output, outputErrorf(output, err, "could not uninstall service group: %v", group)
	}
	return output, nil
}
