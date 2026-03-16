package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ahtwr/iara/internal/paths"
)

// Metadata holds iara project metadata persisted in .iara/metadata.json.
type Metadata struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Instructions string `json:"instructions"`
}

// SaveMetadata validates and writes metadata to <projectDir>/.iara/metadata.json.
func SaveMetadata(projectDir string, raw string) error {
	var m Metadata
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if m.Title == "" {
		return fmt.Errorf("title is required")
	}
	if m.Description == "" {
		return fmt.Errorf("description is required")
	}
	if m.Instructions == "" {
		return fmt.Errorf("instructions is required")
	}

	cwDir := filepath.Join(projectDir, ".iara")
	if err := os.MkdirAll(cwDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(cwDir, "metadata.json"), data, 0644)
}

// LoadMetadata reads .iara/metadata.json for a project by name.
func LoadMetadata(name string) (*Metadata, error) {
	dir := filepath.Join(paths.ProjectsDir(), name)
	return LoadMetadataAt(dir)
}

// LoadMetadataAt reads .iara/metadata.json from a project directory.
func LoadMetadataAt(projectDir string) (*Metadata, error) {
	data, err := os.ReadFile(filepath.Join(projectDir, ".iara", "metadata.json"))
	if err != nil {
		return nil, err
	}
	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// HasMetadata checks if a project has .iara/metadata.json.
func HasMetadata(name string) bool {
	dir := filepath.Join(paths.ProjectsDir(), name)
	_, err := os.Stat(filepath.Join(dir, ".iara", "metadata.json"))
	return err == nil
}
