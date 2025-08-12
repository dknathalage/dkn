package terraform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type TerraformPlugin struct{}

type GenerateContext struct {
	Component    string
	Environments []string
	OutputDir    string
	Org          string
	Repo         string
}

func New() *TerraformPlugin {
	return &TerraformPlugin{}
}

func (p *TerraformPlugin) Name() string {
	return "terraform"
}

func (p *TerraformPlugin) ConfigFile() string {
	return "deploy/*/*.yaml"
}

func (p *TerraformPlugin) Generate(ctx context.Context, configPath string, outputDir string) error {
	return p.Gen(ctx, "deploy", outputDir)
}

func (p *TerraformPlugin) getOrgAndRepo() (string, string, error) {
	// Check if we're in a testing environment
	if os.Getenv("GO_TEST_MODE") != "" {
		return "test-org", "test-repo", nil
	}

	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("git remote origin is not set. Please set up your git remote origin or use 'gh repo create' to create and connect a GitHub repository")
	}

	remoteURL := strings.TrimSpace(string(output))

	// Handle both SSH and HTTPS formats
	// SSH: git@github.com:org/repo.git
	// HTTPS: https://github.com/org/repo.git
	var org, repo string

	if strings.HasPrefix(remoteURL, "git@") {
		// SSH format
		re := regexp.MustCompile(`git@[^:]+:([^/]+)/(.+)\.git$`)
		matches := re.FindStringSubmatch(remoteURL)
		if len(matches) == 3 {
			org = matches[1]
			repo = matches[2]
		}
	} else if strings.HasPrefix(remoteURL, "https://") {
		// HTTPS format
		re := regexp.MustCompile(`https://[^/]+/([^/]+)/(.+)\.git$`)
		matches := re.FindStringSubmatch(remoteURL)
		if len(matches) == 3 {
			org = matches[1]
			repo = matches[2]
		}
	}

	if org == "" || repo == "" {
		return "", "", fmt.Errorf("unable to parse git remote URL '%s'. Please ensure your git remote origin is set to a valid GitHub repository URL, or use 'gh repo create' to set up your repository", remoteURL)
	}

	return org, repo, nil
}

