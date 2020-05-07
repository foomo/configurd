package configurd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/foomo/configurd/exampledata"
	"github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFileExt = ".yml"
	defaultServiceDir    = "configurd/services"
	defaultNamespaceDir  = "configurd/namespaces"
	defaultChartDir      = "configurd/charts"
	defaultOutputDir     = "configurd/.workdir"
	defaultOverridesFile = "overrides.yaml"
)

var (
	ErrServiceNotFound    = errors.New("service not found")
	ErrBuildNotConfigured = errors.New("build parameter was not configured")
)

type Group struct {
	name     string
	Services map[string]ServiceItem
	Jobs     map[string]JobItem
}

type Namespace struct {
	name   string
	groups []Group
}

type Config struct {
	Tag      string
	BasePath string
	Log      *logrus.Logger
}

type Configurd struct {
	config     Config
	Services   []Service
	Templates  []string
	Namespaces []Namespace
}

type Service struct {
	Name  string `yaml:"-"`
	Image string `yaml:"image"`
	Tag   string `yaml:"tag"`
	Build string `yaml:"build"`
	Chart string `yaml:"chart"`
}

type serviceLoader func(string) (Service, error)

func relativePath(path, basePath string) string {
	return strings.Replace(path, basePath+"/", "", -1)
}

func New(config Config) (Configurd, error) {
	log := config.Log
	log.Infof("Parsing configuration files")
	log.Infof("Entering dir: %q", config.BasePath)

	c := Configurd{
		config: config,
	}

	serviceDir := path.Join(config.BasePath, defaultServiceDir)
	err := filepath.Walk(serviceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, defaultConfigFileExt) {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			name := strings.TrimSuffix(info.Name(), defaultConfigFileExt)
			log.Infof("Loading service: %v, from: %q", name, relativePath(path, config.BasePath))
			svc, err := loadService(file, name, config.Tag)
			if err != nil {
				return err
			}
			c.Services = append(c.Services, svc)
		}
		return nil
	})

	if err != nil {
		return Configurd{}, err
	}

	c.Namespaces, err = loadNamespaces(log, c.Service, config.BasePath)

	if err != nil {
		return Configurd{}, err
	}

	return c, nil
}

func loadNamespaces(log *logrus.Logger, sl serviceLoader, basePath string) ([]Namespace, error) {
	var nss []Namespace
	namespaceDir := path.Join(basePath, defaultNamespaceDir)
	err := filepath.Walk(namespaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != namespaceDir {
			log.Infof("Loading namespace: %v, from: %q", info.Name(), relativePath(path, basePath))
			gs, err := loadGroups(log, sl, basePath, info.Name())
			if err != nil {
				return err
			}
			ns := Namespace{
				name:   info.Name(),
				groups: gs,
			}
			nss = append(nss, ns)
		}
		return nil
	})
	return nss, err
}

func loadGroups(log *logrus.Logger, sl serviceLoader, basePath, namespace string) ([]Group, error) {
	var gs []Group
	groupPath := path.Join(basePath, defaultNamespaceDir, namespace)
	err := filepath.Walk(groupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, defaultConfigFileExt)) {
			name := strings.TrimSuffix(info.Name(), defaultConfigFileExt)
			log.Infof("Loading group: %v, from: %q", name, relativePath(path, basePath))
			g, err := loadGroup(log, sl, path, namespace, name)
			if err != nil {
				return err
			}
			gs = append(gs, g)
		}
		return nil
	})
	return gs, err
}

func loadGroup(log *logrus.Logger, sl serviceLoader, path, namespace, group string) (Group, error) {
	file, err := os.Open(path)
	if err != nil {
		return Group{}, err
	}
	defer file.Close()

	var wrapper struct {
		Group Group `yaml:"group"`
	}
	if err := yaml.NewDecoder(file).Decode(&wrapper); err != nil {
		return Group{}, fmt.Errorf("could not decode group: %w", err)
	}

	for name := range wrapper.Group.Services {
		log.Infof("Loading group item: %v", name)
		svc, err := sl(name)
		if err != nil {
			return Group{}, err
		}
		wrapper.Group.Services[name] = loadServiceItem(wrapper.Group.Services[name], svc.Name, namespace, group, svc.Chart)
	}
	wrapper.Group.name = group
	return wrapper.Group, nil
}

func loadServiceItem(item ServiceItem, service, namespace, group, chart string) ServiceItem {
	item.Name = service
	item.namespace = namespace
	item.group = group
	item.chart = chart
	return item
}

func (c Configurd) Service(name string) (Service, error) {
	for _, svc := range c.Services {
		if svc.Name == name {
			return svc, nil
		}
	}
	return Service{}, ErrServiceNotFound
}

func (c Configurd) Build(name string) (string, error) {
	l := c.config.Log
	s, err := c.Service(name)
	if err != nil {
		return "", err
	}

	if s.Build == "" {
		return "", ErrBuildNotConfigured
	}
	args := strings.Split(s.Build, " ")
	if args[0] == "docker" {
		args = append(strings.Split(s.Build, " "), "-t", fmt.Sprintf("%v:%v", s.Image, s.Tag))
	}
	l.Infof("Building service: %v", s.Name)

	output, err := runCommand(c.config.BasePath, args...)
	if err != nil {
		return output, err
	}

	l.Trace(output)

	return output, err
}

func (c Configurd) GetServiceItems(namespace, group string) []ServiceItem {
	var sis []ServiceItem
	for _, ns := range c.Namespaces {
		for _, g := range ns.groups {
			for _, si := range g.Services {
				if (namespace == "" || namespace == ns.name) && (group == "" || group == g.name) {
					sis = append(sis, si)
				}
			}
		}
	}
	return sis
}

func loadService(reader io.Reader, name, defaultTag string) (Service, error) {
	var wrapper struct {
		Service Service `yaml:"service"`
	}
	if err := yaml.NewDecoder(reader).Decode(&wrapper); err != nil {
		return Service{}, fmt.Errorf("could not decode service: %w", err)
	}
	wrapper.Service.Name = name
	if wrapper.Service.Tag == "" {
		wrapper.Service.Tag = defaultTag
	}
	return wrapper.Service, nil
}

func Init(log *logrus.Logger, dir string, _ bool) (string, error) {
	log.Infof("Downloading example configuration into dir: %q", dir)
	return "done with export", exampledata.RestoreAssets(dir, "")
}

func runCommand(cwd string, command ...string) (string, error) {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	output := strings.Replace(string(out), "\n", "\n\t", -1)
	if out == nil && err != nil {
		output = err.Error()
	}
	return output, err
}
