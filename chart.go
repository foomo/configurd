package squadron

import (
	"fmt"
	"io/ioutil"
	"path"

	"gopkg.in/yaml.v3"

	"github.com/foomo/squadron/util"
)

type ChartDependency struct {
	Name       string `yaml:"name,omitempty"`
	Repository string `yaml:"repository,omitempty"`
	Version    string `yaml:"version,omitempty"`
	Alias      string `yaml:"alias,omitempty"`
}

func (cd *ChartDependency) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag == "!!map" {
		type wrapper ChartDependency
		return value.Decode((*wrapper)(cd))
	}
	if value.Tag == "!!str" {
		var vString string
		if err := value.Decode(&vString); err != nil {
			return err
		}
		localChart, err := loadChart(path.Join(vString, chartFile))
		if err != nil {
			return fmt.Errorf("unable to load chart path %q, error: %v", vString, err)
		}
		cd.Name = localChart.Name
		cd.Repository = fmt.Sprintf("file://%v", vString)
		cd.Version = localChart.Version
		return nil
	}
	return fmt.Errorf("unsupported node tag type for %T: %q", cd, value.Tag)
}

type Chart struct {
	APIVersion   string            `yaml:"apiVersion"`
	Name         string            `yaml:"name,omitempty"`
	Description  string            `yaml:"description,omitempty"`
	Type         string            `yaml:"type,omitempty"`
	Version      string            `yaml:"version,omitempty"`
	Dependencies []ChartDependency `yaml:"dependencies,omitempty"`
}

func newChart(name, version string) *Chart {
	return &Chart{
		APIVersion:  chartApiVersionV2,
		Name:        name,
		Description: fmt.Sprintf("A helm parent chart for squadron %v", name),
		Type:        defaultChartType,
		Version:     version,
	}
}

func (c *Chart) addDependency(alias string, cd ChartDependency) {
	cd.Alias = alias
	c.Dependencies = append(c.Dependencies, cd)
}

func loadChart(path string) (*Chart, error) {
	c := Chart{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error while opening file: %v", err)
	}
	if err := yaml.Unmarshal(file, &c); err != nil {
		return nil, fmt.Errorf("error while unmarshalling template file: %s", err)
	}
	return &c, nil
}

func (c Chart) generate(chartPath string, overrides interface{}) error {
	// generate Chart.yaml
	if err := util.GenerateYaml(path.Join(chartPath, chartFile), c); err != nil {
		return err
	}
	// generate values.yaml
	if err := util.GenerateYaml(path.Join(chartPath, valuesFile), overrides); err != nil {
		return err
	}
	return nil
}
