package plugin

import "context"

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
		if plugin.ConfigFile() == filename {
			return plugin, true
		}
	}
	return nil, false
}