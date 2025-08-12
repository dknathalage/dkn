package plugin

import (
	"context"
	"path/filepath"
	"strings"
)

type Plugin interface {
	Name() string
	ConfigFile() string
	Generate(ctx context.Context, configPath string, outputDir string) error
}

type Registry struct {
	plugins map[string]Plugin
}

func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

func (r *Registry) Register(plugin Plugin) {
	r.plugins[plugin.Name()] = plugin
}

func (r *Registry) Get(name string) (Plugin, bool) {
	plugin, exists := r.plugins[name]
	return plugin, exists
}

func (r *Registry) All() map[string]Plugin {
	return r.plugins
}

func (r *Registry) FindByConfigFile(filename string) (Plugin, bool) {
	for _, plugin := range r.plugins {
		configPattern := plugin.ConfigFile()
		
		// Handle exact matches (for terraform/terraform.yaml)
		if configPattern == filename {
			return plugin, true
		}
		
		// Handle glob patterns
		if strings.Contains(configPattern, "*") {
			matched, err := filepath.Match(configPattern, filename)
			if err == nil && matched {
				return plugin, true
			}
			
			// Also try with relative path patterns for multi-level wildcards
			// e.g., deploy/*/*.yaml should match deploy/terraform/postgres.yaml
			if strings.Contains(configPattern, "/*/") {
				// For patterns like "deploy/*/*.yaml", use a simple contains check for now
				basePath := strings.Split(configPattern, "/*")[0] // "deploy"
				if strings.HasPrefix(filename, basePath) && strings.HasSuffix(filename, ".yaml") {
					return plugin, true
				}
			}
		}
	}
	return nil, false
}