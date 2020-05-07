package actions

import (
	"fmt"

	"github.com/foomo/configurd"
	"github.com/spf13/cobra"
)

func init() {
	installCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "Build service group before publishing")
	installCmd.Flags().StringVarP(&flagOutputDir, "output", "o", "default", "Specifies output directory")
	installCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
	installCmd.Flags().StringVarP(&flagService, "service", "s", "", "Specifies the service to work with")
}

var (
	flagBuild     bool
	flagOutputDir string
	flagService   string
)

var (
	installCmd = &cobra.Command{
		Use:   "install [GROUP] -n {NAMESPACE} -t {TAG} -s {SERVICE}",
		Short: "installs a group of services",
		Long:  "installs a group of services with given namespace and tag version",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := install(args[0], flagNamespace, flagTag, flagDir, flagOutputDir, flagService, flagBuild, flagVerbose)
			if err != nil {
				log.WithError(err).Fatalf("Installation failed")
			}
		},
	}
)

func install(group, namespace, tag, workDir, outputDir, service string, buildService, verbose bool) (string, error) {
	log := newLogger(verbose)
	cnf := mustNewConfigurd(log, tag, workDir)

	sis := cnf.GetServiceItems(namespace, group)
	if len(sis) == 0 {
		return "", fmt.Errorf("could not find any service for namespace: %v and group: %v", namespace, group)
	}

	// If one service is selected
	if service != "" {
		filtered := make([]configurd.ServiceItem, 1)
		for _, si := range sis {
			if si.Name == service {
				filtered[0] = si
				break
			}
		}
		sis = filtered
	}

	if len(sis) == 0 {
		return "", fmt.Errorf("no services found to install")
	}

	if buildService {
		log.Printf("Building services")
		for _, si := range sis {
			_, err := build(si.Name, tag, workDir, verbose)
			if err != nil {
				return "", err
			}
		}
	}
	output, err := cnf.Install(configurd.InstallConfiguration{
		ServiceItems: sis,
		BasePath:     workDir,
		OutputDir:    outputDir,
		Tag:          tag,
		Verbose:      verbose,
	})

	if err != nil {
		return output, outputErrorf(output, err, "could not install group: %v", group)
	}
	return output, nil
}
