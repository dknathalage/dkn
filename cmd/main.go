package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dknathalage/dkn/pkg/plugin"
	"github.com/dknathalage/dkn/pkg/plugins/terraform"
	"github.com/dknathalage/dkn/pkg/scanner"
)

func main() {
	if len(os.Args) < 1 {
		log.Fatal("Usage: cli [plugin-name]")
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	registry := plugin.NewRegistry()
	registry.Register(terraform.New())

	fileScanner := scanner.NewFileScanner(cwd)
	ctx := context.Background()
	outputDir := cwd

	if len(os.Args) > 1 {
		pluginName := os.Args[1]
		plugin, exists := registry.Get(pluginName)
		if !exists {
			log.Fatalf("Plugin '%s' not found", pluginName)
		}

		configPath := fileScanner.GetConfigPath(plugin.ConfigFile())
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Fatalf("Configuration file '%s' not found", plugin.ConfigFile())
		}

		if err := plugin.Generate(ctx, configPath, outputDir); err != nil {
			log.Fatalf("Failed to generate with plugin '%s': %v", pluginName, err)
		}
		return
	}

	configFiles, err := fileScanner.ScanForConfigs()
	if err != nil {
		log.Fatalf("Failed to scan for config files: %v", err)
	}

	if len(configFiles) == 0 {
		fmt.Println("No configuration files found in current directory")
		return
	}

	for _, configFile := range configFiles {
		plugin, found := registry.FindByConfigFile(configFile)
		if !found {
			fmt.Printf("⚠️  No plugin found for config file: %s\n", configFile)
			continue
		}

		configPath := fileScanner.GetConfigPath(configFile)
		fmt.Printf("🔧 Generating with %s plugin...\n", plugin.Name())

		if err := plugin.Generate(ctx, configPath, outputDir); err != nil {
			fmt.Printf("❌ Failed to generate with %s plugin: %v\n", plugin.Name(), err)
			continue
		}
	}

	fmt.Println("🎉 Code generation complete!")
}
