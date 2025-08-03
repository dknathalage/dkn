package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dknathalage/dkn/pkg/plugin"
	"github.com/dknathalage/dkn/pkg/plugins/terraform"
	"github.com/dknathalage/dkn/pkg/scanner"
)

const version = "0.1.0"

func printUsage() {
	fmt.Printf(`dkn - DevOps configuration generator

USAGE:
    dkn [OPTIONS] [PLUGIN]

OPTIONS:
    -h, --help      Show this help message
    -v, --version   Show version information

PLUGINS:
    terraform       Generate Terraform configurations

EXAMPLES:
    dkn                 # Scan and generate all configurations
    dkn terraform       # Generate only Terraform configurations
    dkn -h             # Show this help
    dkn -v             # Show version

DESCRIPTION:
    dkn scans your project directory for configuration files and generates
    infrastructure code using the appropriate plugins. It supports automatic
    detection of configuration files or targeted generation with specific plugins.

`)
}

func main() {
	var showHelp bool
	var showVersion bool
	
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	
	flag.Usage = printUsage
	flag.Parse()

	if showHelp {
		printUsage()
		return
	}

	if showVersion {
		fmt.Printf("dkn version %s\n", version)
		return
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

	args := flag.Args()
	if len(args) > 0 {
		pluginName := args[0]
		plugin, exists := registry.Get(pluginName)
		if !exists {
			fmt.Printf("❌ Error: Plugin '%s' not found\n\n", pluginName)
			fmt.Println("Available plugins:")
			for name := range registry.All() {
				fmt.Printf("  - %s\n", name)
			}
			fmt.Println("\nUse 'dkn -h' for more information.")
			os.Exit(1)
		}

		// For plugins with patterns, scan for matching configs
		configPattern := plugin.ConfigFile()
		if strings.Contains(configPattern, "*") {
			configFiles, err := fileScanner.ScanForConfigs()
			if err != nil {
				log.Fatalf("Failed to scan for config files: %v", err)
			}

			var matchingConfigs []string
			for _, configFile := range configFiles {
				if matchedPlugin, found := registry.FindByConfigFile(configFile); found && matchedPlugin == plugin {
					matchingConfigs = append(matchingConfigs, configFile)
				}
			}

			if len(matchingConfigs) == 0 {
				fmt.Printf("❌ Error: No configuration files found for plugin '%s'\n", pluginName)
				fmt.Printf("Expected config pattern: %s\n", configPattern)
				fmt.Println("\nUse 'dkn -h' for more information.")
				os.Exit(1)
			}

			for _, configFile := range matchingConfigs {
				configPath := fileScanner.GetConfigPath(configFile)
				fmt.Printf("🔧 Generating %s with %s plugin...\n", configFile, plugin.Name())
				if err := plugin.Generate(ctx, configPath, outputDir); err != nil {
					fmt.Printf("❌ Failed to generate with %s plugin for %s: %v\n", plugin.Name(), configFile, err)
					continue
				}
				fmt.Printf("✅ Successfully generated %s\n", configFile)
			}
		} else {
			// For plugins with exact config paths (like terraform/terraform.yaml)
			configPath := fileScanner.GetConfigPath(plugin.ConfigFile())
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				fmt.Printf("❌ Error: Configuration file '%s' not found\n", plugin.ConfigFile())
				fmt.Printf("Please create the required configuration file or use 'dkn' without arguments to scan for all configs.\n")
				fmt.Println("\nUse 'dkn -h' for more information.")
				os.Exit(1)
			}

			fmt.Printf("🔧 Generating with %s plugin...\n", plugin.Name())
			if err := plugin.Generate(ctx, configPath, outputDir); err != nil {
				fmt.Printf("❌ Failed to generate with plugin '%s': %v\n", pluginName, err)
				os.Exit(1)
			}
			fmt.Printf("✅ Successfully generated with %s plugin\n", plugin.Name())
		}
		return
	}

	configFiles, err := fileScanner.ScanForConfigs()
	if err != nil {
		log.Fatalf("Failed to scan for config files: %v", err)
	}

	if len(configFiles) == 0 {
		fmt.Println("ℹ️  No configuration files found in current directory")
		fmt.Println("\nSupported configuration patterns:")
		for name, plugin := range registry.All() {
			fmt.Printf("  - %s: %s\n", name, plugin.ConfigFile())
		}
		fmt.Println("\nUse 'dkn -h' for more information.")
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
		fmt.Printf("✅ Successfully generated with %s plugin\n", plugin.Name())
	}

	fmt.Println("🎉 Code generation complete!")
}
