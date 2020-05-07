package configurd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ServiceItem struct {
	Name      string
	Overrides interface{}
	namespace string
	group     string
	chart     string
}

type JobItem struct {
	Name      string
	Overrides interface{}
	namespace string
	group     string
	chart     string
}

func generateYaml(_ *logrus.Logger, path string, data interface{}) error {
	out, marshalErr := yaml.Marshal(data)
	if marshalErr != nil {
		return marshalErr
	}
	file, crateErr := os.Create(path)
	if crateErr != nil {
		return crateErr
	}
	defer file.Close()
	_, writeErr := file.Write(out)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func generate(log *logrus.Logger, si ServiceItem, basePath, outputDir string) error {
	outputPath := path.Join(basePath, defaultOutputDir, outputDir, si.Name)
	log.Infof("Creating dir: %q", path.Join(outputDir, si.Name))
	if err := os.MkdirAll(outputPath, 0744); err != nil {
		return fmt.Errorf("could not create output dir: %w", err)
	}

	log.Infof("Copying chart: %v to dir: %q", si.chart, path.Join(outputDir, si.Name))
	chartPath := path.Join(basePath, defaultChartDir, si.chart)
	err := filepath.Walk(chartPath, func(source string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if chartPath == source {
			return nil
		}

		destination := path.Join(outputPath, info.Name())
		if err := copy.Copy(source, destination); err != nil {
			return err
		}
		if info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not copy template files: %w", err)
	}

	log.Printf("Generating yaml file: %q", path.Join(outputDir, si.Name, defaultOverridesFile))
	err = generateYaml(log, path.Join(outputPath, defaultOverridesFile), si.Overrides)
	if err != nil {
		return fmt.Errorf("could not generate %v: %w", defaultOverridesFile, err)
	}
	return nil
}

func helmInstall(log *logrus.Logger, si ServiceItem, service Service, outputDir string) (string, error) {
	log.Infof("Running helm install for service: %v", si.Name)
	chartPath := path.Join(outputDir, si.Name)
	cmd := []string{
		"helm", "upgrade", si.Name, chartPath,
		"-f", path.Join(chartPath, defaultOverridesFile),
		"-n", si.namespace,
		"--install",
		"--set", fmt.Sprintf("group=%v", si.group),
		"--set", fmt.Sprintf("image.repository=%s", service.Image),
		"--set", fmt.Sprintf("image.tag=%s", service.Tag),
		"--set", fmt.Sprintf("metadata.name=%s", service.Name),
		"--set", fmt.Sprintf("metadata.component=%s", si.group),
		"--set", fmt.Sprintf("metadata.namespace=%s", si.namespace),
	}

	output, err := runCommand("", cmd...)

	if err != nil {
		return output, fmt.Errorf("could not install a helm chart for service %v", si.Name)
	}
	return output, nil
}

func helmUninstall(log *logrus.Logger, si ServiceItem) (string, error) {
	log.Infof("Running helm uninstall for service: %v", si.Name)
	cmd := []string{
		"helm",
		"uninstall",
		"-n", si.namespace,
		si.Name,
	}
	output, err := runCommand("", cmd...)

	if err != nil {
		return output, fmt.Errorf("could not uninstall a helm chart for service: %v, namespace: %v", si.Name, si.namespace)
	}
	return output, nil
}

type InstallConfiguration struct {
	ServiceItems []ServiceItem
	BasePath     string
	OutputDir    string
	Tag          string
	Verbose      bool
}

func (c Configurd) Install(cnf InstallConfiguration) (string, error) {
	logger := c.config.Log

	logger.Infof("Installing services")
	outputPath := path.Join(cnf.BasePath, defaultOutputDir, cnf.OutputDir)

	logger.Infof("Entering dir: %q", path.Join(cnf.BasePath, defaultOutputDir))
	logger.Infof("Removing dir: %q", cnf.OutputDir)
	if err := os.RemoveAll(outputPath); err != nil {
		return "", fmt.Errorf("could not clean workdir directory: %w", err)
	}

	logger.Printf("Creating dir: %q", cnf.OutputDir)
	if err := os.MkdirAll(outputPath, 0744); err != nil {
		return "", fmt.Errorf("could not create a workdir directory: %w", err)
	}
	for _, si := range cnf.ServiceItems {
		err := generate(logger, si, cnf.BasePath, cnf.OutputDir)
		if err != nil {
			return "", err
		}
	}

	var output []string
	for _, si := range cnf.ServiceItems {
		s, err := c.Service(si.Name)
		if err != nil {
			return "", err
		}
		out, err := helmInstall(logger, si, s, outputPath)
		if err != nil {
			return out, err
		}
		logger.Trace(out)
		output = append(output, out)
	}

	return strings.Join(output, "\n"), nil
}

func (c Configurd) Uninstall(sis []ServiceItem, namespace string, verbose bool) (string, error) {
	logger := c.config.Log

	var outputs []string
	for _, si := range sis {
		out, err := helmUninstall(logger, si)
		if err != nil {
			return out, err
		}
		outputs = append(outputs, out)
		logger.Trace(out)
	}
	return strings.Join(outputs, "\n"), nil
}
