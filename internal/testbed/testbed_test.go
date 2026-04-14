package testbed

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}

	git("init")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "pkg"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pkg", "lib.go"), []byte("package pkg\n"), 0o644))
	git("add", ".")
	git("commit", "-m", "init")

	return dir
}

func createTestbed(t *testing.T) *Testbed {
	t.Helper()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	projectRoot := setupGitRepo(t)
	tb, err := Create(projectRoot, "test-plan")
	require.NoError(t, err)
	return tb
}

// --- ignore tests ---

func TestLoadIgnorePatterns_Default(t *testing.T) {
	dir := t.TempDir()
	patterns, err := LoadIgnorePatterns(dir)
	require.NoError(t, err)
	assert.Equal(t, DefaultIgnorePatterns, patterns)
}

func TestLoadIgnorePatterns_CustomFile(t *testing.T) {
	dir := t.TempDir()
	content := "# comment\nfoo/\n\nbar/\n*.tmp\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".planoshignore"), []byte(content), 0o644))

	patterns, err := LoadIgnorePatterns(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo/", "bar/", "*.tmp"}, patterns)
}

func TestWriteExcludeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "excludes")
	require.NoError(t, WriteExcludeFile(path, []string{"a/", "b/"}))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "a/\nb/\n", string(data))
}

// --- Create tests ---

func TestCreate(t *testing.T) {
	tb := createTestbed(t)
	defer tb.Cleanup(false)

	assert.DirExists(t, tb.GoldenDir)
	assert.FileExists(t, filepath.Join(tb.BaseDir, ".lock"))
	assert.FileExists(t, tb.ExcludeFile)
	assert.FileExists(t, filepath.Join(tb.GoldenDir, "main.go"))
	assert.FileExists(t, filepath.Join(tb.GoldenDir, "pkg", "lib.go"))
}

func TestCreate_IncludesGitignoredPaths(t *testing.T) {
	// Regression: rsync-based Create must copy paths that git would skip
	// (e.g. nested harness repos kept outside the parent's tracked tree).
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	projectRoot := setupGitRepo(t)
	require.NoError(t, os.WriteFile(
		filepath.Join(projectRoot, ".gitignore"),
		[]byte("codespace/\n"), 0o644))

	nested := filepath.Join(projectRoot, "codespace", "app")
	require.NoError(t, os.MkdirAll(nested, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(nested, "README.md"), []byte("nested\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(nested, ".git"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(nested, ".git", "HEAD"),
		[]byte("ref: refs/heads/main\n"), 0o644))

	tb, err := Create(projectRoot, "nested")
	require.NoError(t, err)
	defer tb.Cleanup(false)

	assert.FileExists(t, filepath.Join(tb.GoldenDir, "codespace", "app", "README.md"))
	assert.FileExists(t, filepath.Join(tb.GoldenDir, "codespace", "app", ".git", "HEAD"))
}

func TestCreate_WithPlanoshignore(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	projectRoot := setupGitRepo(t)
	require.NoError(t, os.WriteFile(
		filepath.Join(projectRoot, ".planoshignore"),
		[]byte("vendor/\n"), 0o644))

	tb, err := Create(projectRoot, "custom")
	require.NoError(t, err)
	defer tb.Cleanup(false)

	data, err := os.ReadFile(tb.ExcludeFile)
	require.NoError(t, err)
	assert.Equal(t, "vendor/\n", string(data))
}

// --- CopyRuns tests ---

func TestCopyRuns(t *testing.T) {
	tb := createTestbed(t)
	defer tb.Cleanup(false)

	require.NoError(t, tb.CopyRuns(3))

	for i := 1; i <= 3; i++ {
		assert.FileExists(t, filepath.Join(tb.RunDir(i), "main.go"))
		assert.FileExists(t, filepath.Join(tb.RunDir(i), "pkg", "lib.go"))
	}
}

func TestCopyRuns_RespectsIgnore(t *testing.T) {
	tb := createTestbed(t)
	defer tb.Cleanup(false)

	// Add excluded content to golden (simulates build artifacts).
	nmDir := filepath.Join(tb.GoldenDir, "node_modules")
	require.NoError(t, os.MkdirAll(nmDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(nmDir, "dep.js"), []byte("dep"), 0o644))

	require.NoError(t, tb.CopyRuns(1))

	assert.NoDirExists(t, filepath.Join(tb.RunDir(1), "node_modules"))
	assert.FileExists(t, filepath.Join(tb.RunDir(1), "main.go"))
}

// --- ResetRun tests ---

func TestResetRun(t *testing.T) {
	tb := createTestbed(t)
	defer tb.Cleanup(false)

	require.NoError(t, tb.CopyRuns(2))

	// Modify run-2.
	modFile := filepath.Join(tb.RunDir(2), "main.go")
	require.NoError(t, os.WriteFile(modFile, []byte("package changed\n"), 0o644))

	// Add extra file to run-2.
	require.NoError(t, os.WriteFile(
		filepath.Join(tb.RunDir(2), "extra.go"),
		[]byte("package extra\n"), 0o644))

	require.NoError(t, tb.ResetRun(2))

	data, err := os.ReadFile(modFile)
	require.NoError(t, err)
	assert.Equal(t, "package main\n", string(data))
	assert.NoFileExists(t, filepath.Join(tb.RunDir(2), "extra.go"))
}

// --- UpdateGolden tests ---

func TestUpdateGolden(t *testing.T) {
	tb := createTestbed(t)
	defer tb.Cleanup(false)

	require.NoError(t, tb.CopyRuns(1))

	// Add file to run-1.
	require.NoError(t, os.WriteFile(
		filepath.Join(tb.RunDir(1), "new.go"),
		[]byte("package new\n"), 0o644))

	require.NoError(t, tb.UpdateGolden())

	data, err := os.ReadFile(filepath.Join(tb.GoldenDir, "new.go"))
	require.NoError(t, err)
	assert.Equal(t, "package new\n", string(data))
}

// --- Cleanup tests ---

func TestCleanup_Remove(t *testing.T) {
	tb := createTestbed(t)
	baseDir := tb.BaseDir

	require.NoError(t, tb.Cleanup(false))
	assert.NoDirExists(t, baseDir)
}

func TestCleanup_Keep(t *testing.T) {
	tb := createTestbed(t)
	baseDir := tb.BaseDir

	require.NoError(t, tb.Cleanup(true))

	assert.DirExists(t, baseDir)
	assert.NoFileExists(t, filepath.Join(baseDir, ".lock"))
}

// --- Lock tests ---

func TestLock_ConcurrentDetection(t *testing.T) {
	tb := createTestbed(t)
	defer tb.Cleanup(false)

	// PID 1 (init/launchd) is always alive.
	lockPath := filepath.Join(tb.BaseDir, ".lock")
	require.NoError(t, os.WriteFile(lockPath, []byte("1"), 0o644))

	err := acquireLock(tb.BaseDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "locked by PID 1")
}

func TestLock_StalePID(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	baseDir := filepath.Join(tmpHome, ".planosh", "testbed", "repo--plan")
	require.NoError(t, os.MkdirAll(baseDir, 0o755))

	lockPath := filepath.Join(baseDir, ".lock")
	require.NoError(t, os.WriteFile(lockPath, []byte("999999999"), 0o644))

	err := acquireLock(baseDir)
	assert.NoError(t, err)

	data, err := os.ReadFile(lockPath)
	require.NoError(t, err)
	assert.Equal(t, strconv.Itoa(os.Getpid()), string(data))
}

func TestLock_OwnPID(t *testing.T) {
	tb := createTestbed(t)
	defer tb.Cleanup(false)

	// Re-acquiring with own PID should succeed.
	err := acquireLock(tb.BaseDir)
	assert.NoError(t, err)
}
