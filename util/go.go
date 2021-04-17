package util

import (
	"path/filepath"
	"strings"
)

type GoCmd struct {
	Cmd
}

func NewGoCommand() *GoCmd {
	return &GoCmd{*NewCommand("go")}
}

func (c GoCmd) Build(workDir, output, input string, flags ...string) *Cmd {
	relInput := strings.TrimPrefix(input, workDir+string(filepath.Separator))
	return c.Args("build", "-o", output).Cwd(workDir).Args(flags...).Args(relInput)
}
