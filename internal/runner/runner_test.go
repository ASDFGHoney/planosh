package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeMockScript creates a bash script in dir and returns its absolute path.
func writeMockScript(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o755))
	return path
}

// makeRunDirs creates run-1..run-n directories under base and returns their paths.
func makeRunDirs(t *testing.T, base string, n int) []string {
	t.Helper()
	dirs := make([]string, n)
	for i := range n {
		d := filepath.Join(base, fmt.Sprintf("run-%d", i+1))
		require.NoError(t, os.MkdirAll(d, 0o755))
		dirs[i] = d
	}
	return dirs
}

func TestExecute_AllSucceed(t *testing.T) {
	tmp := t.TempDir()
	script := writeMockScript(t, tmp, "plan.sh", `#!/bin/bash
echo "step ok"
exit 0
`)
	runDirs := makeRunDirs(t, tmp, 3)

	result := Execute(context.Background(), Config{
		PlanShPath:  script,
		PlanName:    "test-plan",
		StepID:      1,
		Concurrency: 3,
	}, runDirs)

	assert.Equal(t, StatusOK, result.Status)
	assert.Equal(t, 1, result.StepID)
	require.Len(t, result.Runs, 3)
	for _, r := range result.Runs {
		assert.NoError(t, r.Err)
		assert.Equal(t, 0, r.ExitCode)
		assert.Contains(t, r.Stdout, "step ok")
		assert.True(t, r.Duration > 0)
	}
}

func TestExecute_AllFail_NoRetry(t *testing.T) {
	tmp := t.TempDir()
	script := writeMockScript(t, tmp, "plan.sh", `#!/bin/bash
echo "error detail" >&2
exit 1
`)
	runDirs := makeRunDirs(t, tmp, 3)

	result := Execute(context.Background(), Config{
		PlanShPath:  script,
		PlanName:    "test-plan",
		StepID:      2,
		Concurrency: 3,
	}, runDirs)

	assert.Equal(t, StatusFailed, result.Status)
	assert.Equal(t, 2, result.StepID)
	for _, r := range result.Runs {
		assert.Error(t, r.Err)
		assert.Equal(t, 1, r.ExitCode)
		assert.Contains(t, r.Stderr, "error detail")
	}
}

func TestExecute_PartialFailure_RetrySucceeds(t *testing.T) {
	tmp := t.TempDir()
	// Script fails once if .fail-once marker exists, then succeeds on retry.
	script := writeMockScript(t, tmp, "plan.sh", `#!/bin/bash
marker="$PROJECT_ROOT/.fail-once"
if [ -f "$marker" ]; then
    rm "$marker"
    echo "first attempt failed" >&2
    exit 1
fi
echo "success"
exit 0
`)
	runDirs := makeRunDirs(t, tmp, 3)

	// Plant marker in run-2 only.
	require.NoError(t, os.WriteFile(filepath.Join(runDirs[1], ".fail-once"), []byte(""), 0o644))

	result := Execute(context.Background(), Config{
		PlanShPath:  script,
		PlanName:    "test-plan",
		StepID:      1,
		Concurrency: 3,
	}, runDirs)

	// run-2 failed first, was retried, succeeded → all OK.
	assert.Equal(t, StatusOK, result.Status)
	for _, r := range result.Runs {
		assert.NoError(t, r.Err)
		assert.Equal(t, 0, r.ExitCode)
	}
}

func TestExecute_PartialFailure_RetryFails(t *testing.T) {
	tmp := t.TempDir()
	// Script always fails if .fail-always marker exists.
	script := writeMockScript(t, tmp, "plan.sh", `#!/bin/bash
if [ -f "$PROJECT_ROOT/.fail-always" ]; then
    echo "permanent failure" >&2
    exit 1
fi
echo "success"
exit 0
`)
	runDirs := makeRunDirs(t, tmp, 3)

	// Plant persistent marker in run-2.
	require.NoError(t, os.WriteFile(filepath.Join(runDirs[1], ".fail-always"), []byte(""), 0o644))

	result := Execute(context.Background(), Config{
		PlanShPath:  script,
		PlanName:    "test-plan",
		StepID:      1,
		Concurrency: 3,
	}, runDirs)

	assert.Equal(t, StatusPartial, result.Status)

	// run-1 and run-3 succeeded.
	assert.NoError(t, result.Runs[0].Err)
	assert.NoError(t, result.Runs[2].Err)

	// run-2 still failed after retry.
	assert.Error(t, result.Runs[1].Err)
	assert.Equal(t, 1, result.Runs[1].ExitCode)
	assert.Equal(t, 2, result.Runs[1].RunIndex)
}

func TestExecute_Timeout(t *testing.T) {
	tmp := t.TempDir()
	script := writeMockScript(t, tmp, "plan.sh", `#!/bin/bash
echo "starting"
# Busy-wait using bash builtins only (no child process to outlive kill).
end=$((SECONDS + 30))
while [ $SECONDS -lt $end ]; do :; done
echo "should not reach here"
`)
	runDirs := makeRunDirs(t, tmp, 1)

	start := time.Now()
	result := Execute(context.Background(), Config{
		PlanShPath:  script,
		PlanName:    "test-plan",
		StepID:      1,
		Concurrency: 1,
		Timeout:     500 * time.Millisecond,
	}, runDirs)
	elapsed := time.Since(start)

	assert.Equal(t, StatusFailed, result.Status)
	require.Len(t, result.Runs, 1)
	assert.Error(t, result.Runs[0].Err)
	assert.NotEqual(t, 0, result.Runs[0].ExitCode)
	assert.NotContains(t, result.Runs[0].Stdout, "should not reach here")

	// Should finish well before the 30s busy-wait.
	assert.Less(t, elapsed, 5*time.Second)
}

func TestExecute_ConcurrencyLimit(t *testing.T) {
	tmp := t.TempDir()
	// Each script atomically increments/decrements a shared counter via mkdir lock.
	// Records the maximum concurrent count.
	script := writeMockScript(t, tmp, "plan.sh", `#!/bin/bash
shared="$(dirname "$PROJECT_ROOT")/.concurrent"
mkdir -p "$shared"

# Increment active count.
while ! mkdir "$shared/lock" 2>/dev/null; do sleep 0.01; done
active=$(cat "$shared/active" 2>/dev/null || echo 0)
active=$((active + 1))
echo "$active" > "$shared/active"
max=$(cat "$shared/max" 2>/dev/null || echo 0)
[ "$active" -gt "$max" ] && echo "$active" > "$shared/max"
rmdir "$shared/lock"

sleep 0.15

# Decrement active count.
while ! mkdir "$shared/lock" 2>/dev/null; do sleep 0.01; done
active=$(cat "$shared/active" 2>/dev/null || echo 0)
active=$((active - 1))
echo "$active" > "$shared/active"
rmdir "$shared/lock"

exit 0
`)
	runDirs := makeRunDirs(t, tmp, 6)

	result := Execute(context.Background(), Config{
		PlanShPath:  script,
		PlanName:    "test-plan",
		StepID:      1,
		Concurrency: 2,
	}, runDirs)

	assert.Equal(t, StatusOK, result.Status)

	// Read the recorded max concurrency.
	sharedDir := filepath.Join(tmp, ".concurrent")
	maxBytes, err := os.ReadFile(filepath.Join(sharedDir, "max"))
	require.NoError(t, err)
	maxConcurrent, err := strconv.Atoi(strings.TrimSpace(string(maxBytes)))
	require.NoError(t, err)

	assert.LessOrEqual(t, maxConcurrent, 2, "concurrency exceeded limit")
	assert.GreaterOrEqual(t, maxConcurrent, 1, "no parallelism detected")
}

func TestExecute_EnvVars(t *testing.T) {
	tmp := t.TempDir()
	script := writeMockScript(t, tmp, "plan.sh", `#!/bin/bash
echo "PROJECT_ROOT=$PROJECT_ROOT"
echo "PLAN_DIR=$PLAN_DIR"
echo "CWD=$(pwd)"
echo "ARGS=$@"
exit 0
`)
	runDirs := makeRunDirs(t, tmp, 1)

	result := Execute(context.Background(), Config{
		PlanShPath:  script,
		PlanName:    "my-plan",
		StepID:      3,
		Concurrency: 1,
	}, runDirs)

	require.Equal(t, StatusOK, result.Status)
	out := result.Runs[0].Stdout

	assert.Contains(t, out, "PROJECT_ROOT="+runDirs[0])
	assert.Contains(t, out, "PLAN_DIR="+filepath.Join(runDirs[0], ".plan", "my-plan"))
	assert.Contains(t, out, "ARGS=--from=3 --to=3 --testbed")

	// CWD may resolve symlinks (e.g. /var → /private/var on macOS).
	realDir, err := filepath.EvalSymlinks(runDirs[0])
	require.NoError(t, err)
	assert.Contains(t, out, "CWD="+realDir)
}

func TestExecute_DefaultConcurrency(t *testing.T) {
	tmp := t.TempDir()
	script := writeMockScript(t, tmp, "plan.sh", `#!/bin/bash
exit 0
`)
	runDirs := makeRunDirs(t, tmp, 3)

	// Concurrency 0 → default to 3.
	result := Execute(context.Background(), Config{
		PlanShPath: script,
		PlanName:   "test-plan",
		StepID:     1,
	}, runDirs)

	assert.Equal(t, StatusOK, result.Status)
	assert.Len(t, result.Runs, 3)
}
