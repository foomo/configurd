package util

type HelmCmd struct {
	Cmd
}

func NewHelmCommand() *HelmCmd {
	return &HelmCmd{*NewCommand("helm")}
}

func (c HelmCmd) UpdateDependency(chart, chartPath string) (string, error) {
	return c.Base().Args("dependency", "update", chartPath).Run()
}

func (c HelmCmd) Package(chart, chartPath, destPath string) (string, error) {
	return c.Base().Args("package", chartPath, "--destination", destPath).Run()
}
