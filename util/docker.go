package util

import (
	"fmt"
)

type DockerCmd struct {
	Cmd
	Options []string
}

func NewDockerCommand() *DockerCmd {
	return &DockerCmd{*NewCommand("docker"), []string{}}
}

func (c *DockerCmd) Build(workDir string) *Cmd {
	return c.Cwd(workDir).Args("build", ".")
}

func (c *DockerCmd) Push(image, tag string) (string, error) {
	return c.Args("push", fmt.Sprintf("%s:%s", image, tag)).Run()
}
