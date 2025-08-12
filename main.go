package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dknathalage/dkn/pkg/plugin"
	"github.com/dknathalage/dkn/pkg/plugins/terraform"
	"github.com/dknathalage/dkn/pkg/scanner"
	"github.com/urfave/cli/v2"
)

const version = "0.1.0"

func runPlugin(ctx context.Context, registry *plugin.Registry, fileScanner *scanner.FileScanner, outputDir string, pluginName string) error {
	plugin, exists := registry.Get(pluginName)
	if !exists {
		fmt.Printf("❌ Error: Plugin '%s' not found\n\n", pluginName)
		fmt.Println("Available plugins:")
		for name := range registry.All() {
			fmt.Printf("  - %s\n", name)
		}
		return fmt.Errorf("plugin not found: %s", pluginName)
	}

	// For plugins with patterns, scan for matching configs
	configPattern := plugin.ConfigFile()
	if strings.Contains(configPattern, "*") {
		configFiles, err := fileScanner.ScanForConfigs()
		if err != nil {
			return fmt.Errorf("failed to scan for config files: %w", err)
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
			return fmt.Errorf("no config files found for plugin: %s", pluginName)
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
			return fmt.Errorf("config file not found: %s", plugin.ConfigFile())
		}

		fmt.Printf("🔧 Generating with %s plugin...\n", plugin.Name())
		if err := plugin.Generate(ctx, configPath, outputDir); err != nil {
			return fmt.Errorf("failed to generate with plugin '%s': %w", pluginName, err)
		}
		fmt.Printf("✅ Successfully generated with %s plugin\n", plugin.Name())
	}
	return nil
}

func scanAndGenerate(ctx context.Context, registry *plugin.Registry, fileScanner *scanner.FileScanner, outputDir string) error {
	configFiles, err := fileScanner.ScanForConfigs()
	if err != nil {
		return fmt.Errorf("failed to scan for config files: %w", err)
	}

	if len(configFiles) == 0 {
		fmt.Println("ℹ️  No configuration files found in current directory")
		fmt.Println("\nSupported configuration patterns:")
		for name, plugin := range registry.All() {
			fmt.Printf("  - %s: %s\n", name, plugin.ConfigFile())
		}
		return nil
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
	return nil
}

func main() {
	app := &cli.App{
		Name:        "dkn",
		Usage:       "DevOps configuration generator",
		Version:     version,
		Description: "dkn scans your project directory for configuration files and generates infrastructure code using the appropriate plugins. It supports automatic detection of configuration files or targeted generation with specific plugins.",
		Commands: []*cli.Command{
			{
				Name:    "generate",
				Aliases: []string{"gen"},
				Usage:   "Generate configurations",
				Action: func(c *cli.Context) error {
					cwd, err := os.Getwd()
					if err != nil {
						return fmt.Errorf("failed to get current directory: %w", err)
					}

					registry := plugin.NewRegistry()
					registry.Register(terraform.New())

					fileScanner := scanner.NewFileScanner(cwd)
					ctx := context.Background()
					outputDir := cwd

					return scanAndGenerate(ctx, registry, fileScanner, outputDir)
				},
			},
			{
				Name:  "apply",
				Usage: "Apply configuration changes",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "Name of the component to apply (optional)",
					},
					&cli.StringFlag{
						Name:     "environment",
						Aliases:  []string{"e"},
						Usage:    "Environment to apply to",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					cwd, err := os.Getwd()
					if err != nil {
						return fmt.Errorf("failed to get current directory: %w", err)
					}

					terraformPlugin := terraform.New()
					ctx := context.Background()
					
					deployPath := "deploy"
					component := c.String("name")
					environment := c.String("environment")

					// If no component name provided, apply all components
					if component == "" {
						config, err := terraform.LoadConfig(deployPath)
						if err != nil {
							return fmt.Errorf("failed to load config: %w", err)
						}
						
						for _, comp := range config.Components {
							fmt.Printf("🚀 Applying component: %s\n", comp.Metadata.Name)
							if err := terraformPlugin.Apply(ctx, deployPath, cwd, comp.Metadata.Name, environment); err != nil {
								return fmt.Errorf("failed to apply component %s: %w", comp.Metadata.Name, err)
							}
						}
						return nil
					}

					return terraformPlugin.Apply(ctx, deployPath, cwd, component, environment)
				},
			},
		},
		Action: func(c *cli.Context) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			registry := plugin.NewRegistry()
			registry.Register(terraform.New())

			fileScanner := scanner.NewFileScanner(cwd)
			ctx := context.Background()
			outputDir := cwd

			return scanAndGenerate(ctx, registry, fileScanner, outputDir)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
