package terraform

import (
	"os"
	"path/filepath"
	"gopkg.in/yaml.v3"
)

type Document struct {
	Kind     string      `yaml:"kind"`
	Metadata Metadata    `yaml:"metadata"`
	Spec     interface{} `yaml:"spec"`
}

type Environment struct {
	Kind     string   `yaml:"kind"`
	Metadata Metadata `yaml:"metadata"`
}

type TerraformResource struct {
	Kind     string        `yaml:"kind"`
	Metadata Metadata      `yaml:"metadata"`
	Spec     TerraformSpec `yaml:"spec"`
}

type TerraformSpec struct {
	Environments    []string      `yaml:"environments"`
	EnvironmentRefs []string      `yaml:"environmentRefs"`
	Backend         BackendConfig `yaml:"backend"`
	Providers       []Provider    `yaml:"providers"`
}

type Config struct {
	Environments []Environment
	Components   []TerraformResource
	Backend      BackendConfig
	Providers    []Provider
}

type Metadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Labels      map[string]string `yaml:"labels"`
}

type BackendConfig struct {
	Type   string            `yaml:"type"`
	Config map[string]string `yaml:"config"`
}

type Provider struct {
	Name    string `yaml:"name"`
	Source  string `yaml:"source"`
	Version string `yaml:"version"`
}

func LoadConfig(deployPath string) (*Config, error) {
	var environments []Environment
	var components []TerraformResource
	
	// Load environments from deploy/environments/
	envPath := filepath.Join(deployPath, "environments")
	if envFiles, err := filepath.Glob(filepath.Join(envPath, "*.yaml")); err == nil {
		for _, envFile := range envFiles {
			data, err := os.ReadFile(envFile)
			if err != nil {
				continue
			}
			
			var env Environment
			if err := yaml.Unmarshal(data, &env); err == nil && env.Kind == "Environment" {
				environments = append(environments, env)
			}
		}
	}
	
	// Load terraform components from deploy/terraform/
	tfPath := filepath.Join(deployPath, "terraform")
	if tfFiles, err := filepath.Glob(filepath.Join(tfPath, "*.yaml")); err == nil {
		for _, tfFile := range tfFiles {
			data, err := os.ReadFile(tfFile)
			if err != nil {
				continue
			}
			
			var tfResource TerraformResource
			if err := yaml.Unmarshal(data, &tfResource); err == nil && tfResource.Kind == "Terraform" {
				components = append(components, tfResource)
			}
		}
	}
	
	config := &Config{
		Environments: environments,
		Components:   components,
	}
	
	// Set defaults for backend and providers if not specified in any component
	config.Backend = BackendConfig{
		Type: "gcs",
		Config: map[string]string{
			"bucket": "dknathalage-tf-state",
		},
	}
	
	config.Providers = []Provider{
		{
			Name:    "google",
			Source:  "hashicorp/google",
			Version: "6.46.0",
		},
		{
			Name:    "google-beta",
			Source:  "hashicorp/google-beta",
			Version: "6.46.0",
		},
	}
	
	return config, nil
}