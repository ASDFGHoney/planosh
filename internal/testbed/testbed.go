package testbed

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Testbed manages the lifecycle of a calibration testbed (D-006).
type Testbed struct {
	BaseDir     string
	GoldenDir   string
	ProjectRoot string
	PlanName    string
	ExcludeFile string
}

// Create clones projectRoot into ~/.planosh/testbed/{repo}--{planName}/golden/.
func Create(projectRoot, planName string) (*Testbed, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("user home dir: %w", err)
	}

	repo := filepath.Base(projectRoot)
	baseDir := filepath.Join(homeDir, ".planosh", "testbed", repo+"--"+planName)
	goldenDir := filepath.Join(baseDir, "golden")
	excludeFile := filepath.Join(baseDir, ".excludes")

	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create testbed dir: %w", err)
	}

	if err := acquireLock(baseDir); err != nil {
		return nil, err
	}

	// Release lock if anything fails after acquisition.
	ok := false
	defer func() {
		if !ok {
			releaseLock(baseDir)
		}
	}()

	// Remove stale golden from a previous crashed run.
	if _, statErr := os.Stat(goldenDir); statErr == nil {
		if err := os.RemoveAll(goldenDir); err != nil {
			return nil, fmt.Errorf("clean stale golden: %w", err)
		}
	}

	cmd := exec.Command("git", "clone", "--local", projectRoot, goldenDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git clone: %s: %w", strings.TrimSpace(string(out)), err)
	}

	patterns, err := LoadIgnorePatterns(projectRoot)
	if err != nil {
		return nil, err
	}
	if err := WriteExcludeFile(excludeFile, patterns); err != nil {
		return nil, err
	}

	ok = true
	return &Testbed{
		BaseDir:     baseDir,
		GoldenDir:   goldenDir,
		ProjectRoot: projectRoot,
		PlanName:    planName,
		ExcludeFile: excludeFile,
	}, nil
}

// RunDir returns the path for the i-th run directory.
func (t *Testbed) RunDir(i int) string {
	return filepath.Join(t.BaseDir, fmt.Sprintf("run-%d", i))
}

// CopyRuns copies golden to run-1 through run-n.
func (t *Testbed) CopyRuns(n int) error {
	for i := 1; i <= n; i++ {
		if err := t.syncFromGolden(i); err != nil {
			return err
		}
	}
	return nil
}

// ResetRun restores run-i from golden.
func (t *Testbed) ResetRun(i int) error {
	return t.syncFromGolden(i)
}

// UpdateGolden copies run-1 back to golden after step convergence.
func (t *Testbed) UpdateGolden() error {
	src := t.RunDir(1) + "/"
	dst := t.GoldenDir + "/"
	return t.rsync(src, dst, "update golden")
}

// Cleanup releases the lock and optionally removes the testbed directory.
func (t *Testbed) Cleanup(keep bool) error {
	releaseLock(t.BaseDir)
	if keep {
		return nil
	}
	if err := os.RemoveAll(t.BaseDir); err != nil {
		return fmt.Errorf("remove testbed: %w", err)
	}
	return nil
}

func (t *Testbed) syncFromGolden(i int) error {
	dst := t.RunDir(i) + "/"
	src := t.GoldenDir + "/"

	if err := os.MkdirAll(t.RunDir(i), 0o755); err != nil {
		return fmt.Errorf("create run-%d dir: %w", i, err)
	}
	return t.rsync(src, dst, fmt.Sprintf("run-%d", i))
}

func (t *Testbed) rsync(src, dst, label string) error {
	cmd := exec.Command("rsync", "-a", "--delete",
		"--exclude-from", t.ExcludeFile, src, dst)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("rsync %s: %s: %w", label, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// --- lock ---

func acquireLock(baseDir string) error {
	lockPath := filepath.Join(baseDir, ".lock")

	data, err := os.ReadFile(lockPath)
	if err == nil {
		pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
		if parseErr == nil && pid != os.Getpid() && !isStale(pid) {
			return fmt.Errorf("testbed locked by PID %d", pid)
		}
	}

	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		return fmt.Errorf("write lock: %w", err)
	}
	return nil
}

func releaseLock(baseDir string) {
	os.Remove(filepath.Join(baseDir, ".lock"))
}

func isStale(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == syscall.ESRCH
}
