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
	
	entries, err := os.ReadDir(s.rootDir)
	if err != nil {
		return nil, err
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