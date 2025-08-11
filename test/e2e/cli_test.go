package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_AllPlugins(t *testing.T) {
	tempDir := t.TempDir()
	
	configContent := `components:
  - api
  - worker
environments:
  - dev
  - prod`

	terraformDir := filepath.Join(tempDir, "terraform")
	if err := os.MkdirAll(terraformDir, 0755); err != nil {
		t.Fatalf("Failed to create terraform directory: %v", err)
	}
	configPath := filepath.Join(terraformDir, "terraform.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	codegenPath := buildCLI(t)
	
	cmd := exec.Command(codegenPath)
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "GO_TEST_MODE=1")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "🔧 Generating with terraform plugin") {
		t.Error("Expected terraform plugin execution message")
	}
	
	if !strings.Contains(string(output), "✅ Generated Terraform configuration") {
		t.Error("Expected terraform generation success message")
	}
	
	if !strings.Contains(string(output), "🎉 Code generation complete!") {
		t.Error("Expected completion message")
	}

	terraformOutputDir := filepath.Join(tempDir, "terraform")
	if _, err := os.Stat(terraformOutputDir); os.IsNotExist(err) {
		t.Error("Terraform directory should be created")
	}
}

func TestCLI_SpecificPlugin(t *testing.T) {
	tempDir := t.TempDir()
	
	configContent := `components:
  - service
environments:
  - test`

	terraformDir := filepath.Join(tempDir, "terraform")
	if err := os.MkdirAll(terraformDir, 0755); err != nil {
		t.Fatalf("Failed to create terraform directory: %v", err)
	}
	configPath := filepath.Join(terraformDir, "terraform.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	codegenPath := buildCLI(t)
	
	cmd := exec.Command(codegenPath, "terraform")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "GO_TEST_MODE=1")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "✅ Generated Terraform configuration") {
		t.Error("Expected terraform generation success message")
	}

	terraformOutputDir := filepath.Join(tempDir, "terraform", "service")
	if _, err := os.Stat(terraformOutputDir); os.IsNotExist(err) {
		t.Error("Component directory should be created")
	}
}

func TestCLI_NoConfigFiles(t *testing.T) {
	tempDir := t.TempDir()
	
	codegenPath := buildCLI(t)
	
	cmd := exec.Command(codegenPath)
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "No configuration files found") {
		t.Error("Expected no config files message")
	}
}

func TestCLI_InvalidPlugin(t *testing.T) {
	tempDir := t.TempDir()
	
	codegenPath := buildCLI(t)
	
	cmd := exec.Command(codegenPath, "nonexistent")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	
	if err == nil {
		t.Errorf("Expected CLI to fail with invalid plugin name, but got success. Output: %s", output)
	}
	
	if !strings.Contains(string(output), "Plugin 'nonexistent' not found") {
		t.Errorf("Expected error message about plugin not found, got: %s", output)
	}
}

func buildCLI(t *testing.T) string {
	ctx := context.Background()
	
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	projectRoot := filepath.Join(wd, "..", "..")
	
	codegenPath := filepath.Join(projectRoot, "codegen-test")
	
	cmd := exec.CommandContext(ctx, "go", "build", "-o", codegenPath, ".")
	cmd.Dir = projectRoot
	
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	
	t.Cleanup(func() {
		os.Remove(codegenPath)
	})
	
	return codegenPath
}