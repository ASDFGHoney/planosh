package report

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// StepStatus represents the calibration outcome of a step.
type StepStatus string

const (
	StatusConverged StepStatus = "converged"
	StatusDiverged  StepStatus = "diverged"
	StatusStuck     StepStatus = "stuck"
	StatusFailed    StepStatus = "failed"
)

// AppliedPatch records a single patch attempt during calibration.
type AppliedPatch struct {
	StepID   int
	Retry    int
	RuleText string
	Skipped  bool
	Reason   string
}

// StepResult holds the calibration outcome of a single step.
type StepResult struct {
	StepID   int
	Name     string
	Status   StepStatus
	Retries  int
	Patches  []AppliedPatch
	Duration time.Duration
}

// Report holds the full calibration report data.
type Report struct {
	PlanName    string
	ProjectRoot string
	StartedAt   time.Time
	FinishedAt  time.Time
	Steps       []StepResult
	Runs        int
}

// DeterminismScore calculates the percentage of converged steps (0-100).
func (r *Report) DeterminismScore() float64 {
	if len(r.Steps) == 0 {
		return 100.0
	}
	converged := 0
	for _, s := range r.Steps {
		if s.Status == StatusConverged {
			converged++
		}
	}
	return float64(converged) / float64(len(r.Steps)) * 100.0
}

// ScoreInterpretation returns a human-readable interpretation of the determinism score.
func ScoreInterpretation(score float64) string {
	switch {
	case score >= 100.0:
		return "완전 수렴 — 모든 step이 결정적"
	case score >= 80.0:
		return "권장 수준 — 대부분의 step이 수렴"
	case score >= 50.0:
		return "개선 필요 — 다수의 step에서 발산 발생"
	default:
		return "하네스 부족 — 대부분의 step이 비결정적"
	}
}

func statusLabel(s StepStatus) string {
	switch s {
	case StatusConverged:
		return "수렴"
	case StatusDiverged:
		return "발산"
	case StatusStuck:
		return "stuck"
	case StatusFailed:
		return "실패"
	default:
		return string(s)
	}
}

// Generate writes a calibration-report.md to outputPath.
func (r *Report) Generate(outputPath string) error {
	var sb strings.Builder

	sb.WriteString("# 교정 리포트\n\n")
	sb.WriteString(fmt.Sprintf("- **플랜**: %s\n", r.PlanName))
	sb.WriteString(fmt.Sprintf("- **프로젝트**: %s\n", r.ProjectRoot))
	sb.WriteString(fmt.Sprintf("- **실행 횟수**: %d\n", r.Runs))
	sb.WriteString(fmt.Sprintf("- **시작**: %s\n", r.StartedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("- **종료**: %s\n", r.FinishedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("- **소요시간**: %s\n\n", r.FinishedAt.Sub(r.StartedAt).Truncate(time.Second)))

	score := r.DeterminismScore()
	sb.WriteString("## 결정성 점수\n\n")
	sb.WriteString(fmt.Sprintf("**%.0f%%** — %s\n\n", score, ScoreInterpretation(score)))

	sb.WriteString("## Step별 결과\n\n")
	sb.WriteString("| Step | 이름 | 상태 | 재시도 | 소요시간 |\n")
	sb.WriteString("|------|------|------|--------|----------|\n")
	for _, s := range r.Steps {
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %d | %s |\n",
			s.StepID, s.Name, statusLabel(s.Status), s.Retries,
			s.Duration.Truncate(time.Second)))
	}
	sb.WriteString("\n")

	hasPatches := false
	for _, s := range r.Steps {
		if len(s.Patches) > 0 {
			hasPatches = true
			break
		}
	}
	if hasPatches {
		sb.WriteString("## 적용된 패치\n\n")
		for _, s := range r.Steps {
			for _, p := range s.Patches {
				sb.WriteString(fmt.Sprintf("### Step %d — 재시도 %d\n\n", s.StepID, p.Retry))
				if p.Skipped {
					sb.WriteString(fmt.Sprintf("건너뜀: %s\n\n", p.Reason))
				} else {
					sb.WriteString("```markdown\n")
					sb.WriteString(p.RuleText)
					sb.WriteString("\n```\n\n")
				}
			}
		}
	}

	return os.WriteFile(outputPath, []byte(sb.String()), 0o644)
}
