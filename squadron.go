package squadron

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/kylelemons/godebug/pretty"
	"github.com/logrusorgru/aurora"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sirupsen/logrus"

	"github.com/foomo/squadron/util"
)

const (
	defaultOutputDir  = ".squadron"
	chartApiVersionV2 = "v2"
	defaultChartType  = "application" // application or library
	chartFile         = "Chart.yaml"
	valuesFile        = "values.yaml"
)

type Configuration struct {
	Name    string                 `yaml:"name,omitempty"`
	Version string                 `yaml:"version,omitempty"`
	Prefix  string                 `yaml:"prefix,omitempty"`
	Global  map[string]interface{} `yaml:"global,omitempty"`
	Units   map[string]Unit        `yaml:"squadron,omitempty"`
}

type Squadron struct {
	name      string
	basePath  string
	namespace string
	c         Configuration
}

func New(basePath, namespace string, files []string) (*Squadron, error) {
	sq := Squadron{
		basePath:  basePath,
		namespace: namespace,
		c:         Configuration{},
	}

	tv := TemplateVars{}
	if err := mergeSquadronFiles(files, &sq.c, tv); err != nil {
		return nil, err
	}

	sq.name = filepath.Base(basePath)
	if sq.c.Name != "" {
		sq.name = sq.c.Name
	}
	return &sq, nil
}

func (sq Squadron) GetUnits() map[string]Unit {
	return sq.c.Units
}

func (sq Squadron) GetGlobal() map[string]interface{} {
	return sq.c.Global
}

func (sq Squadron) GetConfigYAML() ([]byte, error) {
	return yaml.Marshal(sq.c)
}

func (sq Squadron) Generate(units map[string]Unit) error {
	logrus.Infof("recreating chart output dir %q", sq.chartPath())
	if err := sq.cleanupOutput(sq.chartPath()); err != nil {
		return err
	}
	logrus.Infof("generating chart %q files in %q", sq.name, sq.chartPath())
	if err := sq.generateChart(units, sq.chartPath(), sq.name, sq.c.Version); err != nil {
		return err
	}
	logrus.Infof("running helm dependency update for chart: %v", sq.chartPath())
	_, err := util.NewHelmCommand().UpdateDependency(sq.name, sq.chartPath())
	return err
}

func (sq Squadron) Package() error {
	logrus.Infof("running helm package for chart: %v", sq.chartPath())
	_, err := util.NewHelmCommand().Package(sq.name, sq.chartPath(), sq.basePath)
	return err
}

func (sq Squadron) Down(helmArgs []string) error {
	logrus.Infof("running helm uninstall for chart: %v", sq.chartPath())
	_, err := util.NewHelmCommand().Args("uninstall", sq.name).
		Stdout(os.Stdout).
		Args("--namespace", sq.namespace).
		Args(helmArgs...).
		Run()
	return err
}

func (sq Squadron) Diff(helmArgs []string) (string, error) {
	logrus.Infof("running helm diff for chart: %v", sq.chartPath())
	manifest, err := exec.Command("helm", "get", "manifest", sq.name, "--namespace", sq.namespace).Output()
	if err != nil {
		return "", err
	}
	template, err := exec.Command("helm", "upgrade", sq.name, sq.chartPath(), "--namespace", sq.namespace, "--dry-run").Output()
	if err != nil {
		return "", err
	}
	dmp := diffmatchpatch.New()
	return dmp.DiffPrettyText(dmp.DiffMain(string(manifest), string(template), false)), nil
}

func (sq Squadron) computeDiff(formatter aurora.Aurora, a interface{}, b interface{}) string {
	diffs := make([]string, 0)
	for _, s := range strings.Split(pretty.Compare(a, b), "\n") {
		switch {
		case strings.HasPrefix(s, "+"):
			diffs = append(diffs, formatter.Bold(formatter.Green(s)).String())
		case strings.HasPrefix(s, "-"):
			diffs = append(diffs, formatter.Bold(formatter.Red(s)).String())
		}
	}
	return strings.Join(diffs, "\n")
}

func (sq Squadron) Up(helmArgs []string) error {
	logrus.Infof("running helm upgrade for chart: %v", sq.chartPath())
	_, err := util.NewHelmCommand().
		Stdout(os.Stdout).
		Args("upgrade", sq.name, sq.chartPath(), "--install").
		Args("--namespace", sq.namespace).
		Args(helmArgs...).
		Run()
	return err
}

func (sq Squadron) Template(helmArgs []string) (string, error) {
	logrus.Infof("running helm template for chart: %v", sq.chartPath())
	return util.NewHelmCommand().Args("template", sq.name, sq.chartPath()).
		Args("--namespace", sq.namespace).
		Args(helmArgs...).
		Run()
}

func (sq Squadron) chartPath() string {
	return path.Join(sq.basePath, defaultOutputDir, sq.name)
}

func (sq Squadron) cleanupOutput(chartPath string) error {
	if _, err := os.Stat(chartPath); err == nil {
		if err := os.RemoveAll(chartPath); err != nil {
			logrus.Warnf("could not delete chart output directory: %q", err)
		}
	}
	if _, err := os.Stat(chartPath); os.IsNotExist(err) {
		if err := os.MkdirAll(chartPath, 0744); err != nil {
			return fmt.Errorf("could not create chart output directory: %w", err)
		}
	}
	return nil
}

func (sq Squadron) generateChart(units map[string]Unit, chartPath, chartName, version string) error {
	chart := newChart(chartName, version)
	values := map[string]interface{}{}
	if sq.GetGlobal() != nil {
		values["global"] = sq.GetGlobal()
	}
	for name, unit := range units {
		chart.addDependency(name, unit.Chart)
		values[name] = unit.Values
	}
	if err := chart.generate(chartPath, values); err != nil {
		return err
	}
	return nil
}
