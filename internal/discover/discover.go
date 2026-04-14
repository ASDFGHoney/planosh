package discover

import (
	"fmt"
	"os"
	"path/filepath"
)

// Result holds the .plan/ discovery outcome.
type Result struct {
	ProjectRoot string // parent of .plan/, or startDir if not found
	PlanDir     string // absolute path to .plan/; empty if not found
}

// PlanPath returns the absolute path for a named plan within .plan/.
func (r *Result) PlanPath(name string) (string, error) {
	if r.PlanDir == "" {
		return "", fmt.Errorf(".plan/ directory not discovered")
	}
	return filepath.Join(r.PlanDir, name), nil
}

// Find locates .plan/ by walking up from startDir toward the git root (D-007).
//
// Rules:
//   - Start at startDir, walk parent directories
//   - First .plan/ directory found (closest) wins
//   - Stop at git root boundary
//   - If no .plan/ found, ProjectRoot = startDir
func Find(startDir string) (*Result, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	if _, err := os.Stat(abs); err != nil {
		return nil, fmt.Errorf("start directory: %w", err)
	}

	ceiling := gitRoot(abs)

	dir := abs
	for {
		candidate := filepath.Join(dir, ".plan")
		if isDir(candidate) {
			return &Result{
				ProjectRoot: dir,
				PlanDir:     candidate,
			}, nil
		}

		if dir == ceiling {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // filesystem root
		}
		dir = parent
	}

	return &Result{ProjectRoot: abs}, nil
}

// gitRoot walks up from dir to find the nearest .git directory or file.
// Returns filesystem root if no git repository is found.
func gitRoot(dir string) string {
	d := dir
	for {
		gitPath := filepath.Join(d, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			return d
		}
		d = parent
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
