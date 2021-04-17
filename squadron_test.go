package squadron_test

import (
	"path"
	"testing"

	"github.com/foomo/squadron"
	testutils "github.com/foomo/squadron/tests/utils"
	"github.com/foomo/squadron/util"
)

func TestConfigSimpleSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-simple", "squadron.yaml"),
		},
		path.Join("testdata", "config-simple", "squadron.yaml.snapshot"),
	)
}

func TestConfigOverrideSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-override", "squadron.yaml"),
			path.Join("testdata", "config-override", "squadron.override.yaml"),
		},
		path.Join("testdata", "config-override", "squadron.yaml.snapshot"),
	)
}

func TestConfigGlobalSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-global", "squadron.yaml"),
			path.Join("testdata", "config-global", "squadron.override.yaml"),
		},
		path.Join("testdata", "config-global", "squadron.yaml.snapshot"),
	)
}

func TestConfigTemplateSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-template", "squadron.yaml"),
		},
		path.Join("testdata", "config-template", "squadron.yaml.snapshot"),
	)
}

func testConfigSnapshot(t *testing.T, configs []string, snapshot string) {
	var cwd string
	testutils.Must(t, util.ValidatePath(".", &cwd))

	sq, err := squadron.New(cwd, "", configs)
	testutils.Must(t, err, "failed to init squadron")

	cf, err := sq.GetConfigYAML()
	testutils.Must(t, err, "failed to parse config")

	testutils.MustCheckSnapshot(t, snapshot, cf)
}
