package squadron

import (
	"bytes"
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
	Unite   bool                   `yaml:"unite,omitempty"`
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
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(sq.c); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (sq Squadron) Generate(units map[string]Unit) error {
	logrus.Infof("recreating chart output dir %q", sq.chartPath())
	if err := sq.cleanupOutput(sq.chartPath()); err != nil {
		return err
	}
	if sq.c.Unite {
		return sq.generateUmbrellaChart(units)
	}
	for uName, u := range units {
		logrus.Infof("generating %q value overrides file in %q", uName, sq.chartPath())
		if err := sq.generateValues(u.Values, sq.chartPath(), uName); err != nil {
			return err
		}
	}
	return nil
}

func (sq Squadron) generateUmbrellaChart(units map[string]Unit) error {
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

func (sq Squadron) Down(units map[string]Unit, helmArgs []string) error {
	if sq.c.Unite {
		logrus.Infof("running helm uninstall for: %s", sq.chartPath())
		_, err := util.NewHelmCommand().Args("uninstall", sq.name).
			Stdout(os.Stdout).
			Args("--namespace", sq.namespace).
			Args(helmArgs...).
			Run()
		return err
	}
	for uName, _ := range units {
		//todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		logrus.Infof("running helm uninstall for: %s", uName)
		stdErr := bytes.NewBuffer([]byte{})
		if _, err := util.NewHelmCommand().Args("uninstall", rName).
			Stderr(stdErr).
			Stdout(os.Stdout).
			Args("--namespace", sq.namespace).
			Args(helmArgs...).
			Run(); err != nil && string(bytes.TrimSpace(stdErr.Bytes())) != fmt.Sprintf("Error: uninstall: Release not loaded: %s: release: not found", uName) {
			return err
		}
	}

	return nil
}

func (sq Squadron) Diff(units map[string]Unit, helmArgs []string) (string, error) {
	if sq.c.Unite {
		logrus.Infof("running helm diff for: %s", sq.chartPath())
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
	for uName, u := range units {
		//todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		logrus.Infof("running helm diff for: %s", uName)
		manifest, err := exec.Command("helm", "get", "manifest", rName, "--namespace", sq.namespace).CombinedOutput()
		if err != nil && string(bytes.TrimSpace(manifest)) != "Error: release: not found" {
			return "", err
		}
		cmd := exec.Command("helm", "upgrade", rName, "--install", "--namespace", sq.namespace, "-f", path.Join(sq.chartPath(), uName+".yaml"), "--dry-run")
		if strings.Contains(u.Chart.Repository, "file://") {
			cmd.Args = append(cmd.Args, "/"+strings.TrimLeft(u.Chart.Repository, "file://"))
		} else {
			cmd.Args = append(cmd.Args, u.Chart.Name, "--repo", u.Chart.Repository)
		}
		template, err := cmd.Output()
		if err != nil {
			return "", err
		}
		dmp := diffmatchpatch.New()
		return dmp.DiffPrettyText(dmp.DiffMain(string(manifest), string(template), false)), nil
	}

	return "", nil
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

func (sq Squadron) Up(units map[string]Unit, helmArgs []string) error {
	if sq.c.Unite {
		logrus.Infof("running helm upgrade for chart: %s", sq.chartPath())
		_, err := util.NewHelmCommand().
			Stdout(os.Stdout).
			Args("upgrade", sq.name, sq.chartPath(), "--install").
			Args("--namespace", sq.namespace).
			Args(helmArgs...).
			Run()
		return err
	}
	for uName, u := range units {
		//todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		logrus.Infof("running helm upgrade for %s", uName)
		cmd := util.NewHelmCommand().
			Stdout(os.Stdout).
			Args("upgrade", rName, "--install").
			Args("--namespace", sq.namespace).
			Args("-f", path.Join(sq.chartPath(), uName+".yaml")).
			Args(helmArgs...)
		if strings.Contains(u.Chart.Repository, "file://") {
			cmd.Args("/" + strings.TrimLeft(u.Chart.Repository, "file://"))
		} else {
			cmd.Args(u.Chart.Name, "--repo", u.Chart.Repository)
		}
		if _, err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (sq Squadron) Template(units map[string]Unit, helmArgs []string) error {
	if sq.c.Unite {
		logrus.Infof("running helm template for chart: %s", sq.chartPath())
		_, err := util.NewHelmCommand().Args("template", sq.name, sq.chartPath()).
			Stdout(os.Stdout).
			Args("--namespace", sq.namespace).
			Args(helmArgs...).
			Run()
		return err
	}
	for uName, u := range units {
		//todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		logrus.Infof("running helm template for chart: %s", uName)
		cmd := util.NewHelmCommand().Args("template", rName).
			Stdout(os.Stdout).
			Args("--namespace", sq.namespace).
			Args("-f", path.Join(sq.chartPath(), uName+".yaml")).
			Args(helmArgs...)
		if strings.Contains(u.Chart.Repository, "file://") {
			cmd.Args("/" + strings.TrimLeft(u.Chart.Repository, "file://"))
		} else {
			cmd.Args(u.Chart.Name, "--repo", u.Chart.Repository)
		}
		if _, err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
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

func (sq Squadron) generateValues(values map[string]interface{}, vPath, vName string) error {
	if sq.GetGlobal() != nil {
		values["global"] = sq.GetGlobal()
	}
	if err := util.GenerateYaml(path.Join(vPath, vName+".yaml"), values); err != nil {
		return err
	}
	return nil
}
