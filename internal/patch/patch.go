package patch

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ASDFGHoney/planosh/internal/diff"
)

const (
	defaultModel   = "sonnet"
	defaultTimeout = 5 * time.Minute
)

// Config holds configuration for patch generation.
type Config struct {
	ClaudePath string        // path to claude CLI; empty = "claude"
	Model      string        // claude model for patch generation; empty = "sonnet"
	Timeout    time.Duration // per-call timeout; 0 = 5 minutes
}

// DivergenceDetail combines a divergence classification with its diff text.
type DivergenceDetail struct {
	Kind     diff.DivergenceKind
	Path     string
	Detail   string // from diff.Divergence.Detail
	DiffText string // formatted diff text (unified diff or content comparison)
}

// Input holds all context needed to generate a patch.
type Input struct {
	Harness     string              // full harness-for-plan.md content
	StepPrompt  string              // current step prompt content
	Divergences []DivergenceDetail  // divergences with diff text
}

// Result holds the outcome of patch generation.
type Result struct {
	RuleText string // markdown rule text to append to step prompt
	Skipped  bool   // true if generation failed or returned empty
	Reason   string // reason for skip (set when Skipped is true)
}

// Generate calls claude -p to produce patch rules for a divergent step.
// On failure or empty/unparseable response, returns a Result with Skipped=true
// so the caller can mark the step as stuck.
func Generate(ctx context.Context, cfg Config, input Input) (*Result, error) {
	claudePath := cfg.ClaudePath
	if claudePath == "" {
		claudePath = "claude"
	}
	model := cfg.Model
	if model == "" {
		model = defaultModel
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	if len(input.Divergences) == 0 {
		return &Result{Skipped: true, Reason: "발산 없음"}, nil
	}

	prompt := buildPrompt(input)

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, claudePath, "-p", prompt, "--model", model)
	cmd.WaitDelay = time.Second

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		return &Result{
			Skipped: true,
			Reason:  fmt.Sprintf("claude -p 실행 실패: %v (stderr: %s)", err, strings.TrimSpace(stderrBuf.String())),
		}, nil
	}

	ruleText, err := parseResponse(stdoutBuf.String())
	if err != nil {
		return &Result{
			Skipped: true,
			Reason:  fmt.Sprintf("응답 파싱 실패: %v", err),
		}, nil
	}

	return &Result{RuleText: ruleText}, nil
}

// Apply appends rule text to a step prompt file.
// The rules are added under a "## 아키텍처 제약 (자동 생성)" section.
// This function ONLY modifies the given step prompt file (D-008:
// 글로벌 하네스 harness-for-plan.md는 절대 수정하지 않음).
func Apply(stepPromptPath string, ruleText string) error {
	existing, err := os.ReadFile(stepPromptPath)
	if err != nil {
		return fmt.Errorf("step 프롬프트 읽기 실패: %w", err)
	}

	content := strings.TrimRight(string(existing), "\n")

	var sb strings.Builder
	sb.WriteString(content)
	sb.WriteString("\n\n## 아키텍처 제약 (자동 생성)\n\n")
	sb.WriteString(ruleText)
	sb.WriteString("\n")

	if err := os.WriteFile(stepPromptPath, []byte(sb.String()), 0o644); err != nil {
		return fmt.Errorf("step 프롬프트 쓰기 실패: %w", err)
	}

	return nil
}
