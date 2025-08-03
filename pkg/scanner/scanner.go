package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

type FileScanner struct {
	rootDir string
}

func NewFileScanner(rootDir string) *FileScanner {
	return &FileScanner{rootDir: rootDir}
}

func (s *FileScanner) ScanForConfigs() ([]string, error) {
	var configFiles []string

	// Scan specific subdirectories for configs
	subdirs := []string{"terraform"}
	
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(s.rootDir, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(subdirPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			filename := entry.Name()
			if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
				relativePath := filepath.Join(subdir, filename)
				configFiles = append(configFiles, relativePath)
			}
		}
	}

	// Also scan root directory for backward compatibility
	entries, err := os.ReadDir(s.rootDir)
	if err != nil {
		return configFiles, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			configFiles = append(configFiles, filename)
		}
	}

	return configFiles, nil
}

func (s *FileScanner) GetConfigPath(filename string) string {
	return filepath.Join(s.rootDir, filename)
}
