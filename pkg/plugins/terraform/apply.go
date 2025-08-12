package terraform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (p *TerraformPlugin) Apply(ctx context.Context, deployPath string, outputDir string, component string, environment string) error {
	_, err := LoadConfig(deployPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	org, repo, err := p.getOrgAndRepo()
	if err != nil {
		return fmt.Errorf("failed to get org/repo: %w", err)
	}

	componentDir := filepath.Join(outputDir, "terraform", component)
	if _, err := os.Stat(componentDir); os.IsNotExist(err) {
		return fmt.Errorf("component directory %s does not exist. Run 'gen' command first", componentDir)
	}

	if err := p.terraformInit(componentDir, org, repo, component, environment); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	if err := p.terraformApply(componentDir, environment); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	fmt.Printf("✅ Applied Terraform changes for %s in %s environment\n", component, environment)
	return nil
}

func (p *TerraformPlugin) terraformInit(workDir, org, repo, component, environment string) error {
	prefix := fmt.Sprintf("%s/%s/%s/%s", org, repo, component, environment)
	
	cmd := exec.Command("terraform", "init", "-reconfigure", fmt.Sprintf("-backend-config=prefix=%s", prefix))
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("🔄 Initializing Terraform for %s in %s environment...\n", component, environment)
	return cmd.Run()
}

func (p *TerraformPlugin) terraformApply(workDir, environment string) error {
	tfvarsFile := filepath.Join("tfvars", environment+".tfvars")
	
	cmd := exec.Command("terraform", "apply", fmt.Sprintf("-var-file=%s", tfvarsFile), "-auto-approve")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("🚀 Applying Terraform changes for %s environment...\n", environment)
	return cmd.Run()
}