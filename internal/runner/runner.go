package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/sync/errgroup"
)

// Status represents the outcome of parallel execution.
type Status string

const (
	StatusOK      Status = "OK"      // all runs succeeded
	StatusPartial Status = "PARTIAL" // some runs excluded after retry failure
	StatusFailed  Status = "FAILED"  // all runs failed
)

// Config holds configuration for parallel plan.sh execution.
type Config struct {
	PlanShPath  string        // absolute path to plan.sh
	PlanName    string        // plan name (for PLAN_DIR)
	StepID      int           // step number (--from and --to)
	Concurrency int           // max parallel runs (default 3)
	Timeout     time.Duration // per-run timeout; 0 = no timeout
}

// RunResult holds the outcome of a single plan.sh execution.
type RunResult struct {
	RunIndex int           // 1-based run index
	ExitCode int           // 0 = success, -1 = killed/error
	Stdout   string
	Stderr   string
	Duration time.Duration
	Err      error
}

// Result holds the outcome of parallel execution across all runs.
type Result struct {
	StepID int
	Runs   []RunResult
	Status Status
}

// Execute runs plan.sh for a single step across N run directories in parallel.
//
// Concurrency is limited by Config.Concurrency (default 3).
// Failure handling:
//   - All succeed → StatusOK
//   - All fail    → StatusFailed (no retry)
//   - Mixed       → retry failed runs once; still-failed runs are excluded (StatusPartial)
func Execute(ctx context.Context, cfg Config, runDirs []string) *Result {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 3
	}

	// Phase 1: run all in parallel.
	results := runAll(ctx, cfg, toSpecs(runDirs))

	var failedIdx []int
	successCount := 0
	for i, r := range results {
		if r.Err != nil {
			failedIdx = append(failedIdx, i)
		} else {
			successCount++
		}
	}

	if len(failedIdx) == 0 {
		return &Result{StepID: cfg.StepID, Runs: results, Status: StatusOK}
	}
	if successCount == 0 {
		return &Result{StepID: cfg.StepID, Runs: results, Status: StatusFailed}
	}

	// Phase 2: retry failed runs once.
	retrySpecs := make([]runSpec, len(failedIdx))
	for i, idx := range failedIdx {
		retrySpecs[i] = runSpec{dir: runDirs[idx], index: idx + 1}
	}
	retried := runAll(ctx, cfg, retrySpecs)
	for i, idx := range failedIdx {
		results[idx] = retried[i]
	}

	status := StatusOK
	for _, r := range results {
		if r.Err != nil {
			status = StatusPartial
			break
		}
	}

	return &Result{StepID: cfg.StepID, Runs: results, Status: status}
}

type runSpec struct {
	dir   string
	index int // 1-based
}

func toSpecs(runDirs []string) []runSpec {
	specs := make([]runSpec, len(runDirs))
	for i, d := range runDirs {
		specs[i] = runSpec{dir: d, index: i + 1}
	}
	return specs
}

func runAll(ctx context.Context, cfg Config, specs []runSpec) []RunResult {
	results := make([]RunResult, len(specs))
	g := new(errgroup.Group)
	g.SetLimit(cfg.Concurrency)

	for i, s := range specs {
		g.Go(func() error {
			results[i] = runOne(ctx, cfg, s.dir, s.index)
			return nil
		})
	}
	g.Wait()
	return results
}

func runOne(ctx context.Context, cfg Config, runDir string, runIndex int) RunResult {
	start := time.Now()

	var runCtx context.Context
	var cancel context.CancelFunc
	if cfg.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, cfg.Timeout)
	} else {
		runCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	cmd := exec.CommandContext(runCtx, "bash", cfg.PlanShPath,
		fmt.Sprintf("--from=%d", cfg.StepID),
		fmt.Sprintf("--to=%d", cfg.StepID),
		"--testbed")
	cmd.Dir = runDir
	cmd.Env = append(os.Environ(),
		"PROJECT_ROOT="+runDir,
		"PLAN_DIR="+filepath.Join(runDir, ".plan", cfg.PlanName),
	)
	cmd.WaitDelay = time.Second

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return RunResult{
		RunIndex: runIndex,
		ExitCode: exitCode,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Duration: duration,
		Err:      err,
	}
}
