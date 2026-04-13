package testbed

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultIgnorePatterns are the built-in .planoshignore defaults (D-013).
var DefaultIgnorePatterns = []string{
	"node_modules/",
	".next/",
	".nuxt/",
	"dist/",
	"build/",
	"*.lock",
	"package-lock.json",
	"yarn.lock",
	"pnpm-lock.yaml",
	".claude/",
}

// LoadIgnorePatterns reads .planoshignore from projectRoot.
// Returns DefaultIgnorePatterns when the file does not exist.
func LoadIgnorePatterns(projectRoot string) ([]string, error) {
	path := filepath.Join(projectRoot, ".planoshignore")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return DefaultIgnorePatterns, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open .planoshignore: %w", err)
	}
	defer f.Close()

	var patterns []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read .planoshignore: %w", err)
	}
	return patterns, nil
}

// WriteExcludeFile writes patterns to path for rsync --exclude-from.
func WriteExcludeFile(path string, patterns []string) error {
	content := strings.Join(patterns, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write exclude file: %w", err)
	}
	return nil
}
