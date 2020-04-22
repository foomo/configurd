package configurd

import (
	"fmt"
	"strings"
)

type Build struct {
	Command string `yaml:"command"`
	Image   string `yaml:"image"`
	Tag     string
}

type Service struct {
	Name  string
	Build Build  `yaml:"build"`
	Chart string `yaml:"chart"`
}

func (s Service) RunBuild(log Logger, dir string, verbose bool) (string, error) {
	args := strings.Split(s.Build.Command, " ")
	if args[0] == "docker" {
		args = append(strings.Split(s.Build.Command, " "), "-t", fmt.Sprintf("%v:%v", s.Build.Image, s.Build.Tag))
	}
	log.Printf("Building service: %v", s.Name)
	return runCommand(dir, log, verbose, args...)
}
