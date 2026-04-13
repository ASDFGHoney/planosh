package step

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_Valid(t *testing.T) {
	f := writeJSON(t, `{
		"plan_name": "calibrate-mvp",
		"prd": "~/docs/prd.md",
		"created": "2026-04-13",
		"steps": [
			{
				"id": 1,
				"name": "스캐폴딩",
				"prompt": "step-1.md",
				"verify": [
					{"name": "빌드", "run": "go build ./..."}
				],
				"commit": "feat: init"
			},
			{
				"id": 2,
				"name": "파서",
				"prompt": "step-2.md",
				"verify": [
					{"name": "테스트A", "run": "go test ./a/"},
					{"name": "테스트B", "run": "go test ./b/"}
				],
				"commit": "feat: parser"
			}
		]
	}`)

	plan, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, "calibrate-mvp", plan.PlanName)
	assert.Equal(t, "~/docs/prd.md", plan.PRD)
	assert.Equal(t, "2026-04-13", plan.Created)
	require.Len(t, plan.Steps, 2)

	s1 := plan.Steps[0]
	assert.Equal(t, 1, s1.ID)
	assert.Equal(t, "스캐폴딩", s1.Name)
	assert.Equal(t, "step-1.md", s1.Prompt)
	require.Len(t, s1.Verify, 1)
	assert.Equal(t, "빌드", s1.Verify[0].Name)
	assert.Equal(t, "go build ./...", s1.Verify[0].Run)
	assert.Equal(t, "feat: init", s1.Commit)

	s2 := plan.Steps[1]
	assert.Equal(t, 2, s2.ID)
	require.Len(t, s2.Verify, 2)
}

func TestParse_EmptySteps(t *testing.T) {
	f := writeJSON(t, `{
		"plan_name": "empty",
		"steps": []
	}`)

	plan, err := Parse(f)
	require.NoError(t, err)
	assert.Equal(t, "empty", plan.PlanName)
	assert.Empty(t, plan.Steps)
}

func TestParse_InvalidJSON(t *testing.T) {
	f := writeJSON(t, `{not valid json`)

	_, err := Parse(f)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse steps.json")
}

func TestParse_FileNotFound(t *testing.T) {
	_, err := Parse("/nonexistent/steps.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read steps.json")
}

// --- helpers ---

func writeJSON(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "steps.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}
