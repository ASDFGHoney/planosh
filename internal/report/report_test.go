package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeterminismScore(t *testing.T) {
	tests := []struct {
		name     string
		steps    []StepResult
		expected float64
	}{
		{"empty", nil, 100.0},
		{"all converged", []StepResult{
			{Status: StatusConverged},
			{Status: StatusConverged},
			{Status: StatusConverged},
		}, 100.0},
		{"all failed", []StepResult{
			{Status: StatusFailed},
			{Status: StatusFailed},
		}, 0.0},
		{"half converged", []StepResult{
			{Status: StatusConverged},
			{Status: StatusStuck},
		}, 50.0},
		{"mixed statuses", []StepResult{
			{Status: StatusConverged},
			{Status: StatusConverged},
			{Status: StatusStuck},
			{Status: StatusFailed},
		}, 50.0},
		{"single converged", []StepResult{
			{Status: StatusConverged},
		}, 100.0},
		{"single failed", []StepResult{
			{Status: StatusFailed},
		}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Report{Steps: tt.steps}
			assert.InDelta(t, tt.expected, r.DeterminismScore(), 0.01)
		})
	}
}

func TestScoreInterpretation(t *testing.T) {
	tests := []struct {
		score    float64
		contains string
	}{
		{100.0, "완전 수렴"},
		{90.0, "권장"},
		{80.0, "권장"},
		{79.9, "개선 필요"},
		{60.0, "개선 필요"},
		{50.0, "개선 필요"},
		{49.9, "하네스 부족"},
		{30.0, "하네스 부족"},
		{0.0, "하네스 부족"},
	}

	for _, tt := range tests {
		t.Run(tt.contains, func(t *testing.T) {
			assert.Contains(t, ScoreInterpretation(tt.score), tt.contains)
		})
	}
}

func TestStatusLabel(t *testing.T) {
	assert.Equal(t, "수렴", statusLabel(StatusConverged))
	assert.Equal(t, "발산", statusLabel(StatusDiverged))
	assert.Equal(t, "stuck", statusLabel(StatusStuck))
	assert.Equal(t, "실패", statusLabel(StatusFailed))
	assert.Equal(t, "unknown", statusLabel(StepStatus("unknown")))
}

func TestGenerate(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "calibration-report.md")

	rpt := &Report{
		PlanName:    "test-plan",
		ProjectRoot: "/home/user/project",
		StartedAt:   time.Date(2026, 4, 13, 10, 0, 0, 0, time.UTC),
		FinishedAt:  time.Date(2026, 4, 13, 10, 5, 30, 0, time.UTC),
		Runs:        3,
		Steps: []StepResult{
			{StepID: 1, Name: "init", Status: StatusConverged, Retries: 0, Duration: 30 * time.Second},
			{StepID: 2, Name: "routes", Status: StatusConverged, Retries: 1, Duration: 2 * time.Minute,
				Patches: []AppliedPatch{
					{StepID: 2, Retry: 1, RuleText: "- 파일명은 routes.go로 고정"},
				}},
			{StepID: 3, Name: "auth", Status: StatusStuck, Retries: 2, Duration: 5 * time.Minute,
				Patches: []AppliedPatch{
					{StepID: 3, Retry: 1, RuleText: "- jwt 패키지 사용"},
					{StepID: 3, Retry: 2, Skipped: true, Reason: "응답 파싱 실패"},
				}},
		},
	}

	require.NoError(t, rpt.Generate(outPath))

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	content := string(data)

	// Header
	assert.Contains(t, content, "# 교정 리포트")
	assert.Contains(t, content, "test-plan")
	assert.Contains(t, content, "/home/user/project")

	// Score: 2/3 = 66.67% → rounds to 67%
	assert.Contains(t, content, "## 결정성 점수")
	assert.Contains(t, content, "67%")
	assert.Contains(t, content, "개선 필요")

	// Step table
	assert.Contains(t, content, "| 1 | init | 수렴 | 0 |")
	assert.Contains(t, content, "| 2 | routes | 수렴 | 1 |")
	assert.Contains(t, content, "| 3 | auth | stuck | 2 |")

	// Patches section
	assert.Contains(t, content, "## 적용된 패치")
	assert.Contains(t, content, "파일명은 routes.go로 고정")
	assert.Contains(t, content, "jwt 패키지 사용")
	assert.Contains(t, content, "건너뜀: 응답 파싱 실패")

	// Duration
	assert.Contains(t, content, "5m30s")
}

func TestGenerateNoPatches(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "report.md")

	rpt := &Report{
		PlanName:    "simple",
		ProjectRoot: "/tmp/test",
		StartedAt:   time.Date(2026, 4, 13, 10, 0, 0, 0, time.UTC),
		FinishedAt:  time.Date(2026, 4, 13, 10, 1, 0, 0, time.UTC),
		Runs:        3,
		Steps: []StepResult{
			{StepID: 1, Name: "only-step", Status: StatusConverged, Duration: 60 * time.Second},
		},
	}

	require.NoError(t, rpt.Generate(outPath))

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "100%")
	assert.Contains(t, content, "완전 수렴")
	assert.False(t, strings.Contains(content, "## 적용된 패치"))
}

func TestGenerateAllFailed(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "report.md")

	rpt := &Report{
		PlanName:    "failing",
		ProjectRoot: "/tmp/fail",
		StartedAt:   time.Now(),
		FinishedAt:  time.Now(),
		Runs:        3,
		Steps: []StepResult{
			{StepID: 1, Name: "step-a", Status: StatusFailed},
			{StepID: 2, Name: "step-b", Status: StatusFailed},
		},
	}

	require.NoError(t, rpt.Generate(outPath))

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "0%")
	assert.Contains(t, content, "하네스 부족")
	assert.Contains(t, content, "실패")
}
