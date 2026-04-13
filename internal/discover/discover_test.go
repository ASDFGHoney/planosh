package discover

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFind_PlanInCwd(t *testing.T) {
	root := setupGitRepo(t)
	mkDir(t, root, ".plan", "my-plan")

	res, err := Find(root)
	require.NoError(t, err)
	assert.Equal(t, root, res.ProjectRoot)
	assert.Equal(t, filepath.Join(root, ".plan"), res.PlanDir)
}

func TestFind_PlanInParent(t *testing.T) {
	root := setupGitRepo(t)
	mkDir(t, root, ".plan", "my-plan")
	sub := mkDir(t, root, "sub", "deep")

	res, err := Find(sub)
	require.NoError(t, err)
	assert.Equal(t, root, res.ProjectRoot)
	assert.Equal(t, filepath.Join(root, ".plan"), res.PlanDir)
}

func TestFind_NoPlan_FallbackToCwd(t *testing.T) {
	root := setupGitRepo(t)
	sub := mkDir(t, root, "sub")

	res, err := Find(sub)
	require.NoError(t, err)
	assert.Equal(t, sub, res.ProjectRoot)
	assert.Empty(t, res.PlanDir)
}

func TestFind_NestedPlan_ClosestWins(t *testing.T) {
	root := setupGitRepo(t)
	mkDir(t, root, ".plan", "outer-plan")
	inner := mkDir(t, root, "sub")
	mkDir(t, inner, ".plan", "inner-plan")

	res, err := Find(inner)
	require.NoError(t, err)
	assert.Equal(t, inner, res.ProjectRoot)
	assert.Equal(t, filepath.Join(inner, ".plan"), res.PlanDir)
}

func TestFind_StopsAtGitRoot(t *testing.T) {
	// parent/ has .plan/ but is ABOVE git root — should NOT be found
	parent := t.TempDir()
	mkDir(t, parent, ".plan", "unreachable")

	gitRepo := mkDir(t, parent, "repo")
	mkDir(t, gitRepo, ".git")
	sub := mkDir(t, gitRepo, "sub")

	res, err := Find(sub)
	require.NoError(t, err)
	assert.Equal(t, sub, res.ProjectRoot)
	assert.Empty(t, res.PlanDir)
}

func TestFind_PlanAtGitRoot(t *testing.T) {
	root := setupGitRepo(t)
	mkDir(t, root, ".plan", "at-root")
	sub := mkDir(t, root, "a", "b", "c")

	res, err := Find(sub)
	require.NoError(t, err)
	assert.Equal(t, root, res.ProjectRoot)
	assert.Equal(t, filepath.Join(root, ".plan"), res.PlanDir)
}

func TestFind_InvalidStartDir(t *testing.T) {
	_, err := Find("/nonexistent/path/xyz")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start directory")
}

func TestResult_PlanPath(t *testing.T) {
	r := &Result{
		ProjectRoot: "/project",
		PlanDir:     "/project/.plan",
	}
	path, err := r.PlanPath("calibrate-mvp")
	require.NoError(t, err)
	assert.Equal(t, "/project/.plan/calibrate-mvp", path)
}

func TestResult_PlanPath_NoPlanDir(t *testing.T) {
	r := &Result{ProjectRoot: "/project"}
	_, err := r.PlanPath("any")
	require.Error(t, err)
}

// --- helpers ---

func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))
	return dir
}

func mkDir(t *testing.T, base string, parts ...string) string {
	t.Helper()
	dir := filepath.Join(append([]string{base}, parts...)...)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	return dir
}
