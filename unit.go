package squadron

type Unit struct {
	Chart  ChartDependency        `yaml:"chart,omitempty"`
	Builds map[string]Build       `yaml:"builds,omitempty"`
	Values map[string]interface{} `yaml:"values,omitempty"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Build ...
func (u *Unit) Build(always bool) error {
	for _, build := range u.Builds {
		if exists, err := build.Exists(); err != nil {
			return err
		} else if !exists || always {
			if err := build.Build(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Push ...
func (u *Unit) Push() error {
	for _, build := range u.Builds {
		if err := build.Push(); err != nil {
			return err
		}
	}
	return nil
}
