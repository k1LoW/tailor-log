package config

type Config struct {
	WorkspaceID string   `yaml:"workspaceID"`
	Inputs      []string `yaml:"inputs"`
	Outputs     Outputs  `yaml:"outputs"`
}

type Outputs struct {
	Datadog Datadog `yaml:"datadog,omitempty,omitzero"`
}

type Datadog struct {
	Service string   `yaml:"service"`
	Tags    []string `yaml:"tags"`
}

// func Load(path string) (*Config, error) {
// 	c := &Config{}
// 	b, err := os.ReadFile(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err := yaml.Unmarshal(b, c); err != nil {
// 		return nil, err
// 	}
// 	if err := c.validate(); err != nil {
// 		return nil, err
// 	}
// 	return c, nil
// }

// func gitRoot() (string, error) {
// 	wd, err := os.Getwd()
// 	if err != nil {
// 		return "", err
// 	}
// 	for {
// 		if _, err := os.Stat(filepath.Join(wd, ".git", "config")); err == nil {
// 			return wd, nil
// 		}
// 		parent := filepath.Dir(wd)
// 		if parent == wd {
// 			break
// 		}
// 		wd = parent
// 	}
// 	return "", fmt.Errorf("not found .git directory")
// }

// func (c *Config) validate() (err error) {
// 	// TODO
// 	return err
// }
