package diff

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := filepath.Join(dir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0755))
		require.NoError(t, os.WriteFile(full, []byte(content), 0644))
	}
}

func makeRun(t *testing.T, base string, n int, files map[string]string) string {
	t.Helper()
	dir := filepath.Join(base, fmt.Sprintf("run-%d", n))
	require.NoError(t, os.MkdirAll(dir, 0755))
	writeFiles(t, dir, files)
	return dir
}

func TestSingleRun(t *testing.T) {
	tmp := t.TempDir()
	r1 := makeRun(t, tmp, 1, map[string]string{
		"main.go": "package main\n",
	})

	result, err := Compare([]string{r1}, nil)
	require.NoError(t, err)
	assert.True(t, result.Converged)
}

func TestConvergence(t *testing.T) {
	tmp := t.TempDir()
	files := map[string]string{
		"src/main.go":   "package main\n\nfunc main() {}\n",
		"src/config.go": "package main\n\nvar Config = \"default\"\n",
	}
	r1 := makeRun(t, tmp, 1, files)
	r2 := makeRun(t, tmp, 2, files)
	r3 := makeRun(t, tmp, 3, files)

	result, err := Compare([]string{r1, r2, r3}, nil)
	require.NoError(t, err)
	assert.True(t, result.Converged)
	assert.Empty(t, result.Divergences)
}

func TestWhitespaceIgnored(t *testing.T) {
	tmp := t.TempDir()
	r1 := makeRun(t, tmp, 1, map[string]string{
		"main.go": "package main\n\nfunc main() {}\n",
	})
	r2 := makeRun(t, tmp, 2, map[string]string{
		"main.go": "package main  \r\n\r\nfunc main() {}  \n\n",
	})

	result, err := Compare([]string{r1, r2}, nil)
	require.NoError(t, err)
	assert.True(t, result.Converged)
	assert.Empty(t, result.Divergences)
}

func TestCodePatternDivergence(t *testing.T) {
	tmp := t.TempDir()
	r1 := makeRun(t, tmp, 1, map[string]string{
		"main.go": "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfor i := 0; i < 10; i++ {\n\t\tfmt.Println(i)\n\t}\n}\n",
	})
	r2 := makeRun(t, tmp, 2, map[string]string{
		"main.go": "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tnums := []int{0,1,2,3,4,5,6,7,8,9}\n\tfor _, n := range nums {\n\t\tfmt.Println(n)\n\t}\n}\n",
	})

	result, err := Compare([]string{r1, r2}, nil)
	require.NoError(t, err)
	assert.False(t, result.Converged)
	require.Len(t, result.Divergences, 1)

	d := result.Divergences[0]
	assert.Equal(t, KindCodePattern, d.Kind)
	assert.Equal(t, "main.go", d.Path)
	assert.Equal(t, []int{0, 1}, d.Runs)
	assert.Contains(t, d.Detail, "2개 변형")
}

func TestNamingDivergence(t *testing.T) {
	tmp := t.TempDir()
	shared := "package src\n\nfunc FormatDate() string { return \"2006-01-02\" }\n\nfunc ParseDate() {}\n"

	r1 := makeRun(t, tmp, 1, map[string]string{
		"src/main.go":  "package src\n",
		"src/utils.go": shared,
	})
	r2 := makeRun(t, tmp, 2, map[string]string{
		"src/main.go":    "package src\n",
		"src/helpers.go": shared,
	})

	result, err := Compare([]string{r1, r2}, nil)
	require.NoError(t, err)
	assert.False(t, result.Converged)
	require.Len(t, result.Divergences, 1)

	d := result.Divergences[0]
	assert.Equal(t, KindNaming, d.Kind)
	assert.Equal(t, "src", d.Path)
	assert.Contains(t, d.Detail, "utils.go")
	assert.Contains(t, d.Detail, "helpers.go")
}

func TestStructureDivergence(t *testing.T) {
	tmp := t.TempDir()
	r1 := makeRun(t, tmp, 1, map[string]string{
		"main.go":                 "package main\n",
		"internal/models/user.go": "package models\n\ntype User struct{}\n",
	})
	r2 := makeRun(t, tmp, 2, map[string]string{
		"main.go":            "package main\n",
		"pkg/models/user.go": "package models\n\ntype User struct{}\n",
	})

	result, err := Compare([]string{r1, r2}, nil)
	require.NoError(t, err)
	assert.False(t, result.Converged)

	var structure []Divergence
	for _, d := range result.Divergences {
		if d.Kind == KindStructure {
			structure = append(structure, d)
		}
	}
	require.Len(t, structure, 2)
	assert.Equal(t, "internal/models/user.go", structure[0].Path)
	assert.Equal(t, "pkg/models/user.go", structure[1].Path)
}

func TestScopeExcess(t *testing.T) {
	tmp := t.TempDir()
	r1 := makeRun(t, tmp, 1, map[string]string{
		"src/main.go": "package main\n",
	})
	r2 := makeRun(t, tmp, 2, map[string]string{
		"src/main.go":  "package main\n",
		"src/extra.go": "package main\n\nfunc Extra() {}\n",
	})

	result, err := Compare([]string{r1, r2}, nil)
	require.NoError(t, err)
	assert.False(t, result.Converged)
	require.Len(t, result.Divergences, 1)

	d := result.Divergences[0]
	assert.Equal(t, KindScopeExcess, d.Kind)
	assert.Equal(t, "src/extra.go", d.Path)
	assert.Equal(t, []int{1}, d.Runs)
}

func TestIgnorePatterns(t *testing.T) {
	tmp := t.TempDir()
	r1 := makeRun(t, tmp, 1, map[string]string{
		"src/main.go":             "package main\n",
		"node_modules/foo/bar.js": "module.exports = 'v1';\n",
		"yarn.lock":               "lockfile v1\n",
	})
	r2 := makeRun(t, tmp, 2, map[string]string{
		"src/main.go":             "package main\n",
		"node_modules/foo/bar.js": "module.exports = 'v2';\n",
		"yarn.lock":               "lockfile v2\n",
	})

	patterns := []string{"node_modules/", "*.lock"}
	result, err := Compare([]string{r1, r2}, patterns)
	require.NoError(t, err)
	assert.True(t, result.Converged)
	assert.Empty(t, result.Divergences)
}
