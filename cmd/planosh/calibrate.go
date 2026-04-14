package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ASDFGHoney/planosh/internal/diff"
	"github.com/ASDFGHoney/planosh/internal/discover"
	"github.com/ASDFGHoney/planosh/internal/patch"
	"github.com/ASDFGHoney/planosh/internal/report"
	"github.com/ASDFGHoney/planosh/internal/runner"
	"github.com/ASDFGHoney/planosh/internal/step"
	"github.com/ASDFGHoney/planosh/internal/testbed"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// --- lipgloss styles (D-010) ---

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	okStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	warnStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	failStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	scoreStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
)

// --- command ---

var calibrateCmd = &cobra.Command{
	Use:   "calibrate [plan-name]",
	Short: "Calibrate plan steps sequentially",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCalibrate,
}

func init() {
	f := calibrateCmd.Flags()
	f.String("plan", "", "plan name or path to plan directory")
	f.Int("runs", 3, "number of calibration runs per step")
	f.Bool("keep-testbed", false, "keep testbed after calibration")
	f.Int("max-retries", 2, "max patch retries per divergent step")
	f.String("model", "", "claude model for step execution")
	f.String("patch-model", "", "claude model for patch generation")
	f.Int("concurrency", 1, "parallel calibration runs")
	f.Duration("timeout", 30*time.Minute, "timeout per step run")
	f.Bool("dry", false, "dry run: validate plan without execution")

	rootCmd.AddCommand(calibrateCmd)
}

func runCalibrate(cmd *cobra.Command, args []string) error {
	flags := parseCalFlags(cmd)
	ctx := context.Background()

	// ── Phase 0: discovery + validation ──

	projectRoot, planPath, planName, err := resolvePlan(flags.plan, args)
	if err != nil {
		return err
	}

	stepsPath := filepath.Join(planPath, "steps.json")
	plan, err := step.Parse(stepsPath)
	if err != nil {
		return fmt.Errorf("steps.json 파싱 실패: %w", err)
	}
	if len(plan.Steps) == 0 {
		return fmt.Errorf("steps.json에 step이 없습니다")
	}

	planShPath := filepath.Join(planPath, "plan.sh")
	if _, err := os.Stat(planShPath); err != nil {
		return fmt.Errorf("plan.sh를 찾을 수 없습니다: %s", planShPath)
	}

	harnessPath := filepath.Join(planPath, "harness-for-plan.md")
	harnessContent, err := os.ReadFile(harnessPath)
	if err != nil {
		return fmt.Errorf("harness-for-plan.md 읽기 실패: %w", err)
	}

	ignorePatterns, err := testbed.LoadIgnorePatterns(projectRoot)
	if err != nil {
		return fmt.Errorf("ignore 패턴 로드 실패: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStderr(), "%s  %s (%d steps, %d runs)\n",
		headerStyle.Render("calibrate"),
		planName, len(plan.Steps), flags.runs)

	// --dry: validate and exit
	if flags.dry {
		return printDryRun(cmd, plan, planPath)
	}

	// Set MODEL env var if provided (for plan.sh to pick up)
	if flags.model != "" {
		os.Setenv("MODEL", flags.model)
		defer os.Unsetenv("MODEL")
	}

	// ── Phase 1: step-by-step calibrate loop ──

	startTime := time.Now()
	tb, err := testbed.Create(projectRoot, planName)
	if err != nil {
		return fmt.Errorf("testbed 생성 실패: %w", err)
	}
	defer tb.Cleanup(flags.keepTestbed)

	// Verify .plan/ exists in golden
	goldenPlanDir := filepath.Join(tb.GoldenDir, ".plan", planName)
	if _, err := os.Stat(goldenPlanDir); os.IsNotExist(err) {
		return fmt.Errorf(".plan/%s 디렉토리가 testbed에 없습니다 — git에 추적되는지 확인하세요", planName)
	}

	rpt := &report.Report{
		PlanName:    planName,
		ProjectRoot: projectRoot,
		StartedAt:   startTime,
		Runs:        flags.runs,
	}

	for idx, s := range plan.Steps {
		stepStart := time.Now()
		sr := report.StepResult{StepID: s.ID, Name: s.Name}

		fmt.Fprintf(cmd.OutOrStderr(), "\n%s\n",
			headerStyle.Render(fmt.Sprintf("[Step %d/%d] %s", idx+1, len(plan.Steps), s.Name)))

		// Copy runs from golden
		if err := tb.CopyRuns(flags.runs); err != nil {
			return fmt.Errorf("step %d: run 복사 실패: %w", s.ID, err)
		}

		runDirs := makeRunDirs(tb, flags.runs)
		runnerCfg := runner.Config{
			PlanShPath:  planShPath,
			PlanName:    planName,
			StepID:      s.ID,
			Concurrency: flags.concurrency,
			Timeout:     flags.timeout,
		}

		// Execute step across N runs
		fmt.Fprintf(cmd.OutOrStderr(), "  %s\n", dimStyle.Render("실행 중..."))
		result := runner.Execute(ctx, runnerCfg, runDirs)

		if result.Status == runner.StatusFailed {
			sr.Status = report.StatusFailed
			sr.Duration = time.Since(stepStart)
			rpt.Steps = append(rpt.Steps, sr)
			fmt.Fprintf(cmd.OutOrStderr(), "  %s\n", failStyle.Render("FAILED — 모든 run 실패"))
			continue
		}

		// Compare successful runs
		okDirs := successRunDirs(tb, result)
		converged, sr := tryConverge(cmd, okDirs, ignorePatterns, tb, sr, stepStart)
		if converged {
			rpt.Steps = append(rpt.Steps, sr)
			continue
		}

		// Diverged — enter retry loop
		promptRel := fmt.Sprintf("steps/%d.md", s.ID)
		goldenPromptPath := filepath.Join(goldenPlanDir, promptRel)
		patchCfg := patch.Config{Model: flags.patchModel}
		var diffResult *diff.Result

		// Get initial diffResult for retry loop
		diffResult, err = diff.Compare(okDirs, ignorePatterns)
		if err != nil {
			return fmt.Errorf("step %d: diff 실패: %w", s.ID, err)
		}

		retryConverged := false
		for retry := 1; retry <= flags.maxRetries; retry++ {
			sr.Retries = retry
			fmt.Fprintf(cmd.OutOrStderr(), "  %s\n",
				warnStyle.Render(fmt.Sprintf("패치 중 (재시도 %d/%d)...", retry, flags.maxRetries)))

			// Read current step prompt from golden
			stepPromptContent, readErr := os.ReadFile(goldenPromptPath)
			if readErr != nil {
				stepPromptContent = []byte{}
			}

			divDetails := buildDivDetails(diffResult.Divergences, okDirs)
			patchInput := patch.Input{
				Harness:     string(harnessContent),
				StepPrompt:  string(stepPromptContent),
				Divergences: divDetails,
			}

			patchResult, patchErr := patch.Generate(ctx, patchCfg, patchInput)
			if patchErr != nil {
				return fmt.Errorf("step %d: patch 생성 오류: %w", s.ID, patchErr)
			}

			applied := report.AppliedPatch{StepID: s.ID, Retry: retry}
			if patchResult.Skipped {
				applied.Skipped = true
				applied.Reason = patchResult.Reason
				sr.Patches = append(sr.Patches, applied)
				fmt.Fprintf(cmd.OutOrStderr(), "  %s %s\n",
					warnStyle.Render("STUCK"), dimStyle.Render("— 패치 생성 실패: "+patchResult.Reason))
				break
			}

			applied.RuleText = patchResult.RuleText
			sr.Patches = append(sr.Patches, applied)

			// Apply patch to step prompt in golden (D-008)
			if applyErr := patch.Apply(goldenPromptPath, patchResult.RuleText); applyErr != nil {
				return fmt.Errorf("step %d: patch 적용 실패: %w", s.ID, applyErr)
			}

			// Re-copy runs from golden (with patched prompt)
			if err := tb.CopyRuns(flags.runs); err != nil {
				return fmt.Errorf("step %d: run 재복사 실패: %w", s.ID, err)
			}

			// Re-execute
			fmt.Fprintf(cmd.OutOrStderr(), "  %s\n", dimStyle.Render("재실행 중..."))
			result = runner.Execute(ctx, runnerCfg, runDirs)

			if result.Status == runner.StatusFailed {
				sr.Status = report.StatusFailed
				fmt.Fprintf(cmd.OutOrStderr(), "  %s\n", failStyle.Render("FAILED — 재실행 후 모든 run 실패"))
				break
			}

			okDirs = successRunDirs(tb, result)
			if len(okDirs) < 2 {
				if err := tb.UpdateGolden(); err != nil {
					return fmt.Errorf("step %d: golden 업데이트 실패: %w", s.ID, err)
				}
				retryConverged = true
				break
			}

			diffResult, err = diff.Compare(okDirs, ignorePatterns)
			if err != nil {
				return fmt.Errorf("step %d: 재비교 실패: %w", s.ID, err)
			}

			if diffResult.Converged {
				retryConverged = true
				break
			}
		}

		if retryConverged {
			if err := tb.UpdateGolden(); err != nil {
				return fmt.Errorf("step %d: golden 업데이트 실패: %w", s.ID, err)
			}
			sr.Status = report.StatusConverged
			fmt.Fprintf(cmd.OutOrStderr(), "  %s %s\n",
				okStyle.Render("CONVERGED"),
				dimStyle.Render(fmt.Sprintf("(%d회 재시도 후)", sr.Retries)))
		} else if sr.Status != report.StatusFailed {
			sr.Status = report.StatusStuck
			fmt.Fprintf(cmd.OutOrStderr(), "  %s %s\n",
				warnStyle.Render("STUCK"),
				dimStyle.Render("— 최대 재시도 초과, 다음 step 진행"))
		}

		sr.Duration = time.Since(stepStart)
		rpt.Steps = append(rpt.Steps, sr)
	}

	// ── Phase 2: sync golden .plan/ back to original repo ──

	fmt.Fprintf(cmd.OutOrStderr(), "\n%s\n", dimStyle.Render("하네스 변경사항 원본에 반영 중..."))
	if err := syncPlanBack(goldenPlanDir, planPath); err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "  %s %s\n",
			warnStyle.Render("경고:"), fmt.Sprintf("rsync 실패: %v", err))
	}

	// ── Cleanup + Report ──

	rpt.FinishedAt = time.Now()
	reportPath := filepath.Join(planPath, "calibration-report.md")
	if err := rpt.Generate(reportPath); err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "  %s 리포트 생성 실패: %v\n", warnStyle.Render("경고:"), err)
	}

	printSummary(cmd, rpt, reportPath)
	return nil
}

// --- flag parsing ---

type calFlags struct {
	plan        string
	runs        int
	keepTestbed bool
	maxRetries  int
	model       string
	patchModel  string
	concurrency int
	timeout     time.Duration
	dry         bool
}

func parseCalFlags(cmd *cobra.Command) calFlags {
	f := cmd.Flags()
	plan, _ := f.GetString("plan")
	runs, _ := f.GetInt("runs")
	keepTestbed, _ := f.GetBool("keep-testbed")
	maxRetries, _ := f.GetInt("max-retries")
	model, _ := f.GetString("model")
	patchModel, _ := f.GetString("patch-model")
	concurrency, _ := f.GetInt("concurrency")
	timeout, _ := f.GetDuration("timeout")
	dry, _ := f.GetBool("dry")
	return calFlags{plan, runs, keepTestbed, maxRetries, model, patchModel, concurrency, timeout, dry}
}

// --- plan resolution ---

func resolvePlan(planFlag string, args []string) (projectRoot, planPath, planName string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", "", fmt.Errorf("working directory: %w", err)
	}

	name := planFlag
	if name == "" && len(args) > 0 {
		name = args[0]
	}

	if name == "" {
		return "", "", "", fmt.Errorf("플랜을 지정하세요: planosh calibrate <name> 또는 --plan <name>")
	}

	// If name looks like an existing directory path, use it directly.
	abs, absErr := filepath.Abs(name)
	if absErr == nil {
		if info, statErr := os.Stat(abs); statErr == nil && info.IsDir() {
			planPath = abs
			planName = filepath.Base(planPath)
			parent := filepath.Dir(planPath)
			if filepath.Base(parent) == ".plan" {
				projectRoot = filepath.Dir(parent)
			} else {
				projectRoot = filepath.Dir(planPath)
			}
			return projectRoot, planPath, planName, nil
		}
	}

	// Treat as a plan name — discover .plan/ automatically.
	disc, discErr := discover.Find(cwd)
	if discErr != nil {
		return "", "", "", fmt.Errorf("discover: %w", discErr)
	}
	if disc.PlanDir == "" {
		return "", "", "", fmt.Errorf(".plan/ 디렉토리를 찾을 수 없습니다")
	}

	planPath = filepath.Join(disc.PlanDir, name)
	if _, statErr := os.Stat(planPath); os.IsNotExist(statErr) {
		return "", "", "", fmt.Errorf("플랜 디렉토리가 존재하지 않습니다: %s", planPath)
	}

	return disc.ProjectRoot, planPath, name, nil
}

// --- helpers ---

func makeRunDirs(tb *testbed.Testbed, n int) []string {
	dirs := make([]string, n)
	for i := range dirs {
		dirs[i] = tb.RunDir(i + 1)
	}
	return dirs
}

func successRunDirs(tb *testbed.Testbed, result *runner.Result) []string {
	var dirs []string
	for _, r := range result.Runs {
		if r.Err == nil {
			dirs = append(dirs, tb.RunDir(r.RunIndex))
		}
	}
	return dirs
}

// tryConverge checks if runs converged. Returns true and updated StepResult if so.
func tryConverge(cmd *cobra.Command, okDirs []string, ignorePatterns []string, tb *testbed.Testbed, sr report.StepResult, stepStart time.Time) (bool, report.StepResult) {
	if len(okDirs) < 2 {
		// Single successful run — trivially converged.
		if len(okDirs) == 1 {
			tb.UpdateGolden()
			sr.Status = report.StatusConverged
			sr.Duration = time.Since(stepStart)
			fmt.Fprintf(cmd.OutOrStderr(), "  %s %s\n",
				okStyle.Render("CONVERGED"), dimStyle.Render("(단일 run)"))
		} else {
			sr.Status = report.StatusFailed
			sr.Duration = time.Since(stepStart)
			fmt.Fprintf(cmd.OutOrStderr(), "  %s\n", failStyle.Render("FAILED — 성공한 run 없음"))
		}
		return true, sr
	}

	diffResult, err := diff.Compare(okDirs, ignorePatterns)
	if err != nil {
		sr.Status = report.StatusFailed
		sr.Duration = time.Since(stepStart)
		fmt.Fprintf(cmd.OutOrStderr(), "  %s diff 실패: %v\n", failStyle.Render("FAILED"), err)
		return true, sr
	}

	if diffResult.Converged {
		tb.UpdateGolden()
		sr.Status = report.StatusConverged
		sr.Duration = time.Since(stepStart)
		fmt.Fprintf(cmd.OutOrStderr(), "  %s\n", okStyle.Render("CONVERGED"))
		return true, sr
	}

	// Report divergence summary
	kinds := countDivKinds(diffResult.Divergences)
	fmt.Fprintf(cmd.OutOrStderr(), "  %s %s\n",
		warnStyle.Render("DIVERGED"),
		dimStyle.Render(fmt.Sprintf("(%s)", kinds)))

	return false, sr
}

func countDivKinds(divs []diff.Divergence) string {
	counts := make(map[diff.DivergenceKind]int)
	for _, d := range divs {
		counts[d.Kind]++
	}
	var parts []string
	for k, c := range counts {
		parts = append(parts, fmt.Sprintf("%d %s", c, k))
	}
	return strings.Join(parts, ", ")
}

func buildDivDetails(divs []diff.Divergence, runDirs []string) []patch.DivergenceDetail {
	details := make([]patch.DivergenceDetail, len(divs))
	for i, d := range divs {
		details[i] = patch.DivergenceDetail{
			Kind:   d.Kind,
			Path:   d.Path,
			Detail: d.Detail,
		}
		switch d.Kind {
		case diff.KindCodePattern:
			details[i].DiffText = makeContentDiff(runDirs, d.Path)
		case diff.KindNaming:
			details[i].DiffText = fmt.Sprintf("네이밍 발산: %s — %s", d.Path, d.Detail)
		case diff.KindStructure:
			details[i].DiffText = fmt.Sprintf("구조 발산: %s", d.Detail)
		case diff.KindScopeExcess:
			runs := make([]string, len(d.Runs))
			for j, r := range d.Runs {
				runs[j] = fmt.Sprintf("run-%d", r+1)
			}
			details[i].DiffText = fmt.Sprintf("범위 초과: %s 에서만 존재 (%s)", d.Path, strings.Join(runs, ", "))
		}
	}
	return details
}

func makeContentDiff(runDirs []string, relPath string) string {
	if len(runDirs) < 2 {
		return ""
	}
	a, errA := os.ReadFile(filepath.Join(runDirs[0], relPath))
	b, errB := os.ReadFile(filepath.Join(runDirs[1], relPath))
	if errA != nil || errB != nil {
		return ""
	}
	return fmt.Sprintf("--- run-1/%s\n+++ run-2/%s\n\n[run-1]\n%s\n\n[run-2]\n%s",
		relPath, relPath,
		truncateStr(string(a), 1000),
		truncateStr(string(b), 1000))
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (truncated)"
}

func syncPlanBack(goldenPlanDir, originalPlanDir string) error {
	cmd := exec.Command("rsync", "-a", goldenPlanDir+"/", originalPlanDir+"/")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// --- output ---

func printDryRun(cmd *cobra.Command, plan *step.Plan, planPath string) error {
	w := cmd.OutOrStderr()
	fmt.Fprintf(w, "\n%s\n\n", dimStyle.Render("--- dry run: 실행하지 않고 검증만 수행 ---"))

	for _, s := range plan.Steps {
		promptPath := filepath.Join(planPath, "steps", fmt.Sprintf("%d.md", s.ID))
		exists := "OK"
		if _, err := os.Stat(promptPath); err != nil {
			exists = "MISSING"
		}
		fmt.Fprintf(w, "  Step %d: %-30s [prompt: %s]\n", s.ID, s.Name, exists)
	}

	fmt.Fprintf(w, "\n%s\n", okStyle.Render("검증 완료 — --dry 플래그를 제거하고 실행하세요"))
	return nil
}

func printSummary(cmd *cobra.Command, rpt *report.Report, reportPath string) {
	w := cmd.OutOrStderr()
	score := rpt.DeterminismScore()
	duration := rpt.FinishedAt.Sub(rpt.StartedAt).Truncate(time.Second)

	fmt.Fprintf(w, "\n%s\n", headerStyle.Render("── 교정 완료 ──"))
	fmt.Fprintf(w, "  결정성 점수: %s  %s\n",
		scoreStyle.Render(fmt.Sprintf("%.0f%%", score)),
		dimStyle.Render(report.ScoreInterpretation(score)))
	fmt.Fprintf(w, "  소요시간: %s\n", dimStyle.Render(duration.String()))

	converged, stuck, failed := 0, 0, 0
	for _, s := range rpt.Steps {
		switch s.Status {
		case report.StatusConverged:
			converged++
		case report.StatusStuck:
			stuck++
		case report.StatusFailed:
			failed++
		}
	}
	fmt.Fprintf(w, "  수렴: %s  stuck: %s  실패: %s\n",
		okStyle.Render(fmt.Sprintf("%d", converged)),
		warnStyle.Render(fmt.Sprintf("%d", stuck)),
		failStyle.Render(fmt.Sprintf("%d", failed)))

	fmt.Fprintf(w, "  리포트: %s\n", dimStyle.Render(reportPath))
}
