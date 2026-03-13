package embed

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahtwr/cw/internal/paths"
)

//go:embed all:files
var embeddedFS embed.FS

var installDir string

// Dir returns the base install directory.
func Dir() string {
	return installDir
}

// PluginDir returns the path to the extracted plugins directory.
func PluginDir() string {
	return filepath.Join(installDir, "plugins")
}

// ModesDir returns the path to the extracted modes directory.
func ModesDir() string {
	return filepath.Join(installDir, "modes")
}

// HooksDir returns the path to the extracted hooks directory.
func HooksDir() string {
	return filepath.Join(installDir, "hooks")
}

// Install extracts embedded files to the platform-specific data directory.
func Install() error {
	return installToDir(paths.DataDir())
}

func installToDir(dest string) error {
	installDir = dest

	// Build set of embedded file paths
	embedded := make(map[string]bool)
	if err := fs.WalkDir(embeddedFS, "files", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel("files", path)
			embedded[rel] = true
		}
		return nil
	}); err != nil {
		return err
	}

	// Write all embedded files
	if err := fs.WalkDir(embeddedFS, "files", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel("files", path)
		target := filepath.Join(dest, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := embeddedFS.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		return os.WriteFile(target, data, 0o755)
	}); err != nil {
		return err
	}

	// Auto-generate .md plugin stubs from .sh scripts
	generated, err := generatePluginsFromScripts(dest)
	if err != nil {
		return err
	}
	for _, p := range generated {
		rel, _ := filepath.Rel(dest, p)
		embedded[rel] = true
	}

	// Clean up files on disk that are no longer embedded
	managedDirs := []string{"plugins", "modes", "hooks", "scripts"}
	for _, dir := range managedDirs {
		dirPath := filepath.Join(dest, dir)
		filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(dest, path)
			if !embedded[rel] {
				os.Remove(path)
			}
			return nil
		})
	}

	return nil
}

// generatePluginsFromScripts creates .md plugin stubs for .sh scripts
// that don't already have a hand-written .md file in the embedded files.
func generatePluginsFromScripts(dest string) ([]string, error) {
	scriptsDir := filepath.Join(dest, "scripts")
	commandsDir := filepath.Join(dest, "plugins", "commands")

	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		return nil, err
	}

	var generated []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sh") {
			continue
		}

		name := strings.TrimSuffix(e.Name(), ".sh")
		mdPath := filepath.Join(commandsDir, name+".md")

		// Check if a hand-written .md exists in embedded files
		embeddedMD := filepath.Join("files", "plugins", "commands", name+".md")
		if _, err := embeddedFS.Open(embeddedMD); err == nil {
			continue // hand-written .md exists, don't overwrite
		}

		// Extract description from script's "# description:" comment
		desc := extractScriptDescription(filepath.Join(scriptsDir, e.Name()))

		md := fmt.Sprintf("---\ndescription: %s\n---\n\nDone. Say nothing else.\n", desc)
		if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
			return nil, err
		}
		generated = append(generated, mdPath)
	}

	return generated, nil
}

func extractScriptDescription(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "Direct command."
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# description: ") {
			return strings.TrimPrefix(line, "# description: ")
		}
	}
	return "Direct command."
}
