package diff

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DivergenceKind classifies the type of divergence between runs.
type DivergenceKind string

const (
	KindNaming      DivergenceKind = "naming"
	KindStructure   DivergenceKind = "structure"
	KindCodePattern DivergenceKind = "code_pattern"
	KindScopeExcess DivergenceKind = "scope_excess"
)

// Divergence represents a single detected difference between runs.
type Divergence struct {
	Kind   DivergenceKind
	Path   string // relative file or directory path
	Runs   []int  // 0-indexed run indices
	Detail string
}

// Result holds the comparison outcome for N runs.
type Result struct {
	Converged   bool
	Divergences []Divergence
}

// rawDivergence holds comparison data before classification.
type rawDivergence struct {
	path     string
	present  []int    // 0-indexed run indices that have this file
	allRuns  bool     // true when present in every run
	contents [][]byte // per-run content (set only when allRuns is true)
}

// Compare performs a 2-stage comparison of N run directories.
//
//   - Stage 1: file list comparison (excluding ignorePatterns).
//   - Stage 2: content diff for shared files (whitespace-only diffs ignored).
//
// Convergence: all runs identical in Stage 1 + Stage 2.
// The comparison is filesystem-based; git state inside runs is irrelevant.
func Compare(runDirs []string, ignorePatterns []string) (*Result, error) {
	if len(runDirs) < 2 {
		return &Result{Converged: true}, nil
	}

	for i, dir := range runDirs {
		info, err := os.Stat(dir)
		if err != nil {
			return nil, fmt.Errorf("run-%d (%s): %w", i+1, dir, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("run-%d (%s): not a directory", i+1, dir)
		}
	}

	// Collect file paths from each run.
	fileSets := make([]map[string]struct{}, len(runDirs))
	for i, dir := range runDirs {
		files, err := collectFiles(dir, ignorePatterns)
		if err != nil {
			return nil, fmt.Errorf("collecting files from run-%d: %w", i+1, err)
		}
		fileSets[i] = files
	}

	// Build union of all file paths.
	union := make(map[string]struct{})
	for _, fs := range fileSets {
		for f := range fs {
			union[f] = struct{}{}
		}
	}

	var rawDivs []rawDivergence
	for _, p := range sortedKeys(union) {
		present := runsContaining(fileSets, p)

		if len(present) == len(runDirs) {
			// Stage 2: file in all runs — compare content.
			contents, err := readFileFromRuns(runDirs, p)
			if err != nil {
				return nil, err
			}
			if !contentEqual(contents) {
				rawDivs = append(rawDivs, rawDivergence{
					path: p, present: present, allRuns: true, contents: contents,
				})
			}
		} else {
			// Stage 1: file not in all runs.
			rawDivs = append(rawDivs, rawDivergence{
				path: p, present: present, allRuns: false,
			})
		}
	}

	divergences := classify(rawDivs, len(runDirs), runDirs, fileSets)
	return &Result{
		Converged:   len(divergences) == 0,
		Divergences: divergences,
	}, nil
}

// collectFiles walks dir and returns relative paths of regular files,
// excluding those matching ignorePatterns. .git is always skipped.
func collectFiles(dir string, ignorePatterns []string) (map[string]struct{}, error) {
	files := make(map[string]struct{})
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		// Always skip .git so git commits don't affect comparison.
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if matchesIgnore(rel, info.IsDir(), ignorePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode().IsRegular() {
			files[filepath.ToSlash(rel)] = struct{}{}
		}
		return nil
	})
	return files, err
}

// matchesIgnore checks if a relative path matches any ignore pattern.
// Simplified rsync --exclude semantics:
//   - "dir/" matches any component named "dir"
//   - "*.ext" matches file names via glob
//   - "exact" matches file names exactly
func matchesIgnore(relPath string, isDir bool, patterns []string) bool {
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	name := parts[len(parts)-1]

	for _, pat := range patterns {
		if pat == "" {
			continue
		}
		if strings.HasSuffix(pat, "/") {
			dirName := strings.TrimSuffix(pat, "/")
			for _, part := range parts {
				if part == dirName {
					return true
				}
			}
			continue
		}
		if strings.ContainsAny(pat, "*?[") {
			if matched, _ := filepath.Match(pat, name); matched {
				return true
			}
			continue
		}
		if name == pat {
			return true
		}
	}
	return false
}

// NormalizeWhitespace trims trailing whitespace per line,
// normalizes line endings to \n, and trims trailing empty lines.
func NormalizeWhitespace(content []byte) string {
	s := strings.ReplaceAll(string(content), "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

func contentEqual(contents [][]byte) bool {
	if len(contents) < 2 {
		return true
	}
	ref := NormalizeWhitespace(contents[0])
	for _, c := range contents[1:] {
		if NormalizeWhitespace(c) != ref {
			return false
		}
	}
	return true
}

func readFileFromRuns(runDirs []string, relPath string) ([][]byte, error) {
	contents := make([][]byte, len(runDirs))
	for i, dir := range runDirs {
		data, err := os.ReadFile(filepath.Join(dir, relPath))
		if err != nil {
			return nil, fmt.Errorf("reading %s from run-%d: %w", relPath, i+1, err)
		}
		contents[i] = data
	}
	return contents, nil
}

func runsContaining(fileSets []map[string]struct{}, path string) []int {
	var runs []int
	for i, fs := range fileSets {
		if _, ok := fs[path]; ok {
			runs = append(runs, i)
		}
	}
	return runs
}

func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
