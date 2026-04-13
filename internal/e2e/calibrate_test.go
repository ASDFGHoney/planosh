package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// binaryPath holds the path to the compiled planosh binary.
var binaryPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "planosh-e2e-bin-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmpdir: %v\n", err)
		os.Exit(1)
	}
	binaryPath = filepath.Join(tmp, "planosh")

	cmd := exec.Command("go", "build", "-o", binaryPath,
		"github.com/ASDFGHoney/planosh/cmd/planosh")
	cmd.Dir = filepath.Join("..", "..")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n%s\n", err, out)
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

// ---------------------------------------------------------------------------
// fixtures
// ---------------------------------------------------------------------------

type stepFixture struct {
	ID     int
	Name   string
	Prompt string
}

// setupProject creates a git-initialised project with .plan/{planName}/.
// Returns projectRoot (the git repo) and planDir (.plan/{planName}).
func setupProject(t *testing.T, planName string, steps []stepFixture, planShScript string) (string, string) {
	t.Helper()
	projectRoot := t.TempDir()

	// Seed file so git has something to commit.
	writeFile(t, filepath.Join(projectRoot, "main.go"), "package main\n")

	planDir := filepath.Join(projectRoot, ".plan", planName)
	stepsDir := filepath.Join(planDir, "steps")
	require.NoError(t, os.MkdirAll(stepsDir, 0o755))

	// steps.json
	type jsonVerify struct {
		Name string `json:"name"`
		Run  string `json:"run"`
	}
	type jsonStep struct {
		ID     int          `json:"id"`
		Name   string       `json:"name"`
		Prompt string       `json:"prompt"`
		Verify []jsonVerify `json:"verify"`
		Commit string       `json:"commit"`
	}
	type jsonPlan struct {
		PlanName string     `json:"plan_name"`
		PRD      string     `json:"prd"`
		Created  string     `json:"created"`
		Steps    []jsonStep `json:"steps"`
	}
	jp := jsonPlan{PlanName: planName, PRD: "prd.md", Created: "2026-04-13"}
	for _, s := range steps {
		jp.Steps = append(jp.Steps, jsonStep{
			ID:     s.ID,
			Name:   s.Name,
			Prompt: fmt.Sprintf("%d.md", s.ID),
			Commit: fmt.Sprintf("step %d", s.ID),
		})
		writeFile(t, filepath.Join(stepsDir, fmt.Sprintf("%d.md", s.ID)), s.Prompt)
	}
	data, err := json.MarshalIndent(jp, "", "  ")
	require.NoError(t, err)
	writeFile(t, filepath.Join(planDir, "steps.json"), string(data))

	// plan.sh
	writeFile(t, filepath.Join(planDir, "plan.sh"), planShScript)
	require.NoError(t, os.Chmod(filepath.Join(planDir, "plan.sh"), 0o755))

	// harness-for-plan.md
	writeFile(t, filepath.Join(planDir, "harness-for-plan.md"),
		"# 하네스\n\n- 테스트 하네스\n")

	// git init + commit
	gitRun(t, projectRoot, "init")
	gitRun(t, projectRoot, "add", ".")
	gitRun(t, projectRoot, "commit", "-m", "init")

	return projectRoot, planDir
}

// writeMockClaude creates a mock claude executable under dir and returns
// the directory (to prepend to PATH).
func writeMockClaude(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "claude")
	writeFile(t, path, script)
	require.NoError(t, os.Chmod(path, 0o755))
	return dir
}

// runCalibrate runs `planosh calibrate` with the given args.
// HOME is redirected so testbed goes to a temp directory.
func runCalibrate(t *testing.T, projectRoot string, mockClaudeDir string, extraArgs ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	tmpHome := t.TempDir()

	args := append([]string{"calibrate"}, extraArgs...)
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = projectRoot

	env := []string{
		"HOME=" + tmpHome,
		"PATH=" + mockClaudeDir + ":" + os.Getenv("PATH"),
		// git needs these for clone operations
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	}
	// Preserve necessary system env vars.
	for _, e := range os.Environ() {
		key := strings.SplitN(e, "=", 2)[0]
		switch key {
		case "HOME", "PATH", "GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL",
			"GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL":
			continue // overridden above
		default:
			env = append(env, e)
		}
	}
	cmd.Env = env

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
}

// ---------------------------------------------------------------------------
// mock plan.sh scripts
// ---------------------------------------------------------------------------

// planShHappy creates deterministic identical files per step.
const planShHappy = `#!/bin/bash
for arg in "$@"; do
    case "$arg" in --from=*) STEP_ID="${arg#*=}" ;; esac
done
mkdir -p "$PROJECT_ROOT/internal"
cat > "$PROJECT_ROOT/internal/step${STEP_ID}.go" <<GOEOF
package internal

// Step${STEP_ID} — deterministic output
func Step${STEP_ID}() string { return "ok" }
GOEOF
exit 0
`

// planShDivergeThenConverge produces different files initially,
// identical files after patch.Apply adds the marker section.
const planShDivergeThenConverge = `#!/bin/bash
for arg in "$@"; do
    case "$arg" in --from=*) STEP_ID="${arg#*=}" ;; esac
done
PROMPT="$PLAN_DIR/steps/${STEP_ID}.md"

if grep -q "아키텍처 제약" "$PROMPT" 2>/dev/null; then
    # Post-patch: deterministic output
    cat > "$PROJECT_ROOT/output.go" <<'GOEOF'
package main

func Output() string { return "converged" }
GOEOF
else
    # Pre-patch: vary by run dir name
    RUN_NAME=$(basename "$PROJECT_ROOT")
    cat > "$PROJECT_ROOT/output.go" <<GOEOF
package main

// variant: ${RUN_NAME}
func Output() string { return "${RUN_NAME}" }
GOEOF
fi
exit 0
`

// planShAlwaysDiverge always produces different files, ignoring patches.
const planShAlwaysDiverge = `#!/bin/bash
RUN_NAME=$(basename "$PROJECT_ROOT")
cat > "$PROJECT_ROOT/output.go" <<GOEOF
package main

// variant: ${RUN_NAME}
func Output() string { return "${RUN_NAME}" }
GOEOF
exit 0
`

// planShFail always fails.
const planShFail = `#!/bin/bash
echo "step execution error" >&2
exit 1
`

// ---------------------------------------------------------------------------
// mock claude scripts
// ---------------------------------------------------------------------------

const mockClaudeWithRules = `#!/bin/bash
cat <<'EOF'
---RULES---
- 출력 파일은 반드시 output.go
- 내용은 converged 문자열 고정
---END---
EOF
`

const mockClaudeEmpty = `#!/bin/bash
echo ""
`

// ---------------------------------------------------------------------------
// Test: Happy path — all steps converge on first execution
// ---------------------------------------------------------------------------

func TestCalibrate_HappyPath(t *testing.T) {
	steps := []stepFixture{
		{ID: 1, Name: "스캐폴딩", Prompt: "# Step 1\n\ninternal/step1.go 생성\n"},
		{ID: 2, Name: "핵심 모듈", Prompt: "# Step 2\n\ninternal/step2.go 생성\n"},
		{ID: 3, Name: "통합", Prompt: "# Step 3\n\ninternal/step3.go 생성\n"},
	}

	projectRoot, planDir := setupProject(t, "happy", steps, planShHappy)
	// claude mock not needed for happy path (no divergence → no patch),
	// but provide one anyway so PATH resolution doesn't fail.
	claudeDir := writeMockClaude(t, mockClaudeEmpty)

	_, stderr, exitCode := runCalibrate(t, projectRoot, claudeDir,
		planDir, "--runs=3", "--concurrency=3")

	// Exit code 0.
	assert.Equal(t, 0, exitCode, "stderr:\n%s", stderr)

	// All steps should converge.
	assert.Contains(t, stderr, "CONVERGED")
	assert.NotContains(t, stderr, "DIVERGED")
	assert.NotContains(t, stderr, "FAILED")
	assert.NotContains(t, stderr, "STUCK")

	// 100% determinism score.
	assert.Contains(t, stderr, "100%")
	assert.Contains(t, stderr, "완전 수렴")

	// Calibration report generated.
	reportPath := filepath.Join(planDir, "calibration-report.md")
	assert.FileExists(t, reportPath)

	reportData, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	report := string(reportData)
	assert.Contains(t, report, "100%")
	assert.Contains(t, report, "수렴")
	assert.Contains(t, report, "스캐폴딩")
	assert.Contains(t, report, "핵심 모듈")
	assert.Contains(t, report, "통합")
}

// ---------------------------------------------------------------------------
// Test: Diverge → patch → converge
// ---------------------------------------------------------------------------

func TestCalibrate_DivergeAndConverge(t *testing.T) {
	steps := []stepFixture{
		{ID: 1, Name: "발산후수렴", Prompt: "# Step 1\n\noutput.go 생성\n"},
	}

	projectRoot, planDir := setupProject(t, "divconv", steps, planShDivergeThenConverge)
	claudeDir := writeMockClaude(t, mockClaudeWithRules)

	_, stderr, exitCode := runCalibrate(t, projectRoot, claudeDir,
		planDir, "--runs=3", "--max-retries=2", "--concurrency=3")

	assert.Equal(t, 0, exitCode, "stderr:\n%s", stderr)

	// Should first diverge, then converge after patch.
	assert.Contains(t, stderr, "DIVERGED")
	assert.Contains(t, stderr, "패치 중")
	assert.Contains(t, stderr, "CONVERGED")

	// Score = 100% (1 step, converged after retry).
	assert.Contains(t, stderr, "100%")

	// Report should record the patch.
	reportPath := filepath.Join(planDir, "calibration-report.md")
	reportData, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	report := string(reportData)
	assert.Contains(t, report, "적용된 패치")
	assert.Contains(t, report, "output.go")

	// Step prompt should have been patched and synced back (D-008).
	stepPromptPath := filepath.Join(planDir, "steps", "1.md")
	promptData, err := os.ReadFile(stepPromptPath)
	require.NoError(t, err)
	assert.Contains(t, string(promptData), "아키텍처 제약 (자동 생성)")

	// Harness must NOT be modified (D-008).
	harnessData, err := os.ReadFile(filepath.Join(planDir, "harness-for-plan.md"))
	require.NoError(t, err)
	assert.Equal(t, "# 하네스\n\n- 테스트 하네스\n", string(harnessData))
}

// ---------------------------------------------------------------------------
// Test: Stuck — max retries exceeded, step continues
// ---------------------------------------------------------------------------

func TestCalibrate_Stuck(t *testing.T) {
	steps := []stepFixture{
		{ID: 1, Name: "항상발산", Prompt: "# Step 1\n\noutput.go 생성\n"},
		{ID: 2, Name: "정상수렴", Prompt: "# Step 2\n\ninternal/step2.go 생성\n"},
	}

	// Step 1 always diverges; Step 2 converges normally.
	// plan.sh dispatches by step ID.
	combinedPlanSh := `#!/bin/bash
for arg in "$@"; do
    case "$arg" in --from=*) STEP_ID="${arg#*=}" ;; esac
done

if [ "$STEP_ID" = "1" ]; then
    # Always diverge
    RUN_NAME=$(basename "$PROJECT_ROOT")
    cat > "$PROJECT_ROOT/output.go" <<GOEOF
package main
// variant: ${RUN_NAME}
func Output() string { return "${RUN_NAME}" }
GOEOF
else
    # Converge
    mkdir -p "$PROJECT_ROOT/internal"
    cat > "$PROJECT_ROOT/internal/step2.go" <<'GOEOF'
package internal
func Step2() string { return "ok" }
GOEOF
fi
exit 0
`

	projectRoot, planDir := setupProject(t, "stuck", steps, combinedPlanSh)
	claudeDir := writeMockClaude(t, mockClaudeWithRules)

	_, stderr, exitCode := runCalibrate(t, projectRoot, claudeDir,
		planDir, "--runs=3", "--max-retries=2", "--concurrency=3")

	assert.Equal(t, 0, exitCode, "stderr:\n%s", stderr)

	// Step 1: stuck after max retries.
	assert.Contains(t, stderr, "STUCK")
	// Step 2: converges normally — calibrate continues past stuck step.
	// Count CONVERGED occurrences — at least one for step 2.
	assert.Contains(t, stderr, "CONVERGED")

	// Score should be 50% (1 of 2 steps converged).
	assert.Contains(t, stderr, "50%")

	// Report records stuck status.
	reportPath := filepath.Join(planDir, "calibration-report.md")
	reportData, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	report := string(reportData)
	assert.Contains(t, report, "stuck")
	assert.Contains(t, report, "수렴")
	assert.Contains(t, report, "50%")
}

// ---------------------------------------------------------------------------
// Test: Execution failure — plan.sh fails for all runs
// ---------------------------------------------------------------------------

func TestCalibrate_ExecutionFailure(t *testing.T) {
	steps := []stepFixture{
		{ID: 1, Name: "실패스텝", Prompt: "# Step 1\n\n실패 테스트\n"},
		{ID: 2, Name: "정상스텝", Prompt: "# Step 2\n\ninternal/step2.go 생성\n"},
	}

	// Step 1 fails; Step 2 succeeds.
	combinedPlanSh := `#!/bin/bash
for arg in "$@"; do
    case "$arg" in --from=*) STEP_ID="${arg#*=}" ;; esac
done

if [ "$STEP_ID" = "1" ]; then
    echo "step execution error" >&2
    exit 1
else
    mkdir -p "$PROJECT_ROOT/internal"
    cat > "$PROJECT_ROOT/internal/step2.go" <<'GOEOF'
package internal
func Step2() string { return "ok" }
GOEOF
    exit 0
fi
`

	projectRoot, planDir := setupProject(t, "fail", steps, combinedPlanSh)
	claudeDir := writeMockClaude(t, mockClaudeEmpty)

	_, stderr, exitCode := runCalibrate(t, projectRoot, claudeDir,
		planDir, "--runs=3", "--concurrency=3")

	// Calibrate itself succeeds (it records failure per step, doesn't abort).
	assert.Equal(t, 0, exitCode, "stderr:\n%s", stderr)

	// Step 1: FAILED.
	assert.Contains(t, stderr, "FAILED")
	// Step 2: CONVERGED — continues past failed step.
	assert.Contains(t, stderr, "CONVERGED")

	// Score = 50% (1 converged, 1 failed).
	assert.Contains(t, stderr, "50%")

	// Report records failure.
	reportPath := filepath.Join(planDir, "calibration-report.md")
	reportData, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	report := string(reportData)
	assert.Contains(t, report, "실패")
	assert.Contains(t, report, "수렴")
}

// ---------------------------------------------------------------------------
// Test: Dry run — validates plan without execution
// ---------------------------------------------------------------------------

func TestCalibrate_DryRun(t *testing.T) {
	steps := []stepFixture{
		{ID: 1, Name: "스캐폴딩", Prompt: "# Step 1\n\n생성\n"},
		{ID: 2, Name: "핵심모듈", Prompt: "# Step 2\n\n생성\n"},
	}

	projectRoot, planDir := setupProject(t, "dryrun", steps, planShHappy)
	claudeDir := writeMockClaude(t, mockClaudeEmpty)

	_, stderr, exitCode := runCalibrate(t, projectRoot, claudeDir,
		planDir, "--runs=3", "--dry")

	assert.Equal(t, 0, exitCode, "stderr:\n%s", stderr)
	assert.Contains(t, stderr, "dry run")
	assert.Contains(t, stderr, "스캐폴딩")
	assert.Contains(t, stderr, "핵심모듈")
	assert.Contains(t, stderr, "OK")

	// No report should be generated in dry mode.
	reportPath := filepath.Join(planDir, "calibration-report.md")
	assert.NoFileExists(t, reportPath)
}
