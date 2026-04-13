package patch

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ASDFGHoney/planosh/internal/diff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeMockClaude creates a mock claude script and returns its path.
func writeMockClaude(t *testing.T, dir, script string) string {
	t.Helper()
	path := filepath.Join(dir, "mock-claude")
	require.NoError(t, os.WriteFile(path, []byte(script), 0o755))
	return path
}

// writeFile creates a file with the given content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func sampleInput() Input {
	return Input{
		Harness:    "# Global Harness\n\n- 모든 파일은 snake_case\n",
		StepPrompt: "# Step 1: 초기 스캐폴딩\n\ncmd/main.go 생성\n",
		Divergences: []DivergenceDetail{
			{
				Kind:     diff.KindNaming,
				Path:     "internal/util",
				Detail:   "같은 위치, 다른 이름: utils.go vs helpers.go",
				DiffText: "--- run-1/internal/util/utils.go\n+++ run-2/internal/util/helpers.go",
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Generate tests
// ---------------------------------------------------------------------------

func TestGenerate_Success(t *testing.T) {
	tmp := t.TempDir()
	claude := writeMockClaude(t, tmp, `#!/bin/bash
cat <<'RESP'
Some preamble text.

---RULES---
- 파일명은 반드시 utils.go를 사용한다
- 헬퍼 함수는 internal/util 패키지에 배치한다
---END---
RESP
`)

	result, err := Generate(context.Background(), Config{
		ClaudePath: claude,
		Model:      "sonnet",
	}, sampleInput())

	require.NoError(t, err)
	assert.False(t, result.Skipped)
	assert.Contains(t, result.RuleText, "utils.go")
	assert.Contains(t, result.RuleText, "internal/util")
}

func TestGenerate_EmptyResponse(t *testing.T) {
	tmp := t.TempDir()
	claude := writeMockClaude(t, tmp, `#!/bin/bash
echo ""
`)

	result, err := Generate(context.Background(), Config{
		ClaudePath: claude,
	}, sampleInput())

	require.NoError(t, err)
	assert.True(t, result.Skipped)
	assert.Contains(t, result.Reason, "파싱 실패")
}

func TestGenerate_CommandFailure(t *testing.T) {
	tmp := t.TempDir()
	claude := writeMockClaude(t, tmp, `#!/bin/bash
echo "model not found" >&2
exit 1
`)

	result, err := Generate(context.Background(), Config{
		ClaudePath: claude,
	}, sampleInput())

	require.NoError(t, err)
	assert.True(t, result.Skipped)
	assert.Contains(t, result.Reason, "실행 실패")
	assert.Contains(t, result.Reason, "model not found")
}

func TestGenerate_NoDivergences(t *testing.T) {
	result, err := Generate(context.Background(), Config{}, Input{
		Harness:     "harness",
		StepPrompt:  "step",
		Divergences: nil,
	})

	require.NoError(t, err)
	assert.True(t, result.Skipped)
	assert.Contains(t, result.Reason, "발산 없음")
}

func TestGenerate_FallbackMarkdown(t *testing.T) {
	tmp := t.TempDir()
	// Response has markdown rules but no markers.
	claude := writeMockClaude(t, tmp, `#!/bin/bash
cat <<'RESP'
## 네이밍 규칙

- 모든 헬퍼 파일은 helpers.go로 명명한다
- 테스트 파일은 _test.go 접미사를 사용한다
RESP
`)

	result, err := Generate(context.Background(), Config{
		ClaudePath: claude,
	}, sampleInput())

	require.NoError(t, err)
	assert.False(t, result.Skipped)
	assert.Contains(t, result.RuleText, "helpers.go")
}

func TestGenerate_MultipleDivergenceKinds(t *testing.T) {
	tmp := t.TempDir()
	claude := writeMockClaude(t, tmp, `#!/bin/bash
cat <<'RESP'
---RULES---
- 파일명은 utils.go로 통일
- cmd/ 디렉토리 밖에 파일 생성 금지
---END---
RESP
`)

	input := Input{
		Harness:    "# Harness\n",
		StepPrompt: "# Step\n",
		Divergences: []DivergenceDetail{
			{Kind: diff.KindNaming, Path: "internal/util", Detail: "naming", DiffText: ""},
			{Kind: diff.KindScopeExcess, Path: "extra.go", Detail: "일부 run에서만 존재", DiffText: ""},
		},
	}

	result, err := Generate(context.Background(), Config{
		ClaudePath: claude,
	}, input)

	require.NoError(t, err)
	assert.False(t, result.Skipped)
	assert.Contains(t, result.RuleText, "utils.go")
	assert.Contains(t, result.RuleText, "생성 금지")
}

// ---------------------------------------------------------------------------
// Apply tests
// ---------------------------------------------------------------------------

func TestApply_AppendsRules(t *testing.T) {
	tmp := t.TempDir()
	stepFile := filepath.Join(tmp, "steps", "1.md")
	writeFile(t, stepFile, "# Step 1\n\n기존 내용\n")

	err := Apply(stepFile, "- 파일명은 utils.go를 사용한다\n- 변수명은 camelCase")
	require.NoError(t, err)

	data, err := os.ReadFile(stepFile)
	require.NoError(t, err)
	content := string(data)

	// Original content preserved.
	assert.Contains(t, content, "# Step 1")
	assert.Contains(t, content, "기존 내용")

	// Rules appended with section header.
	assert.Contains(t, content, "## 아키텍처 제약 (자동 생성)")
	assert.Contains(t, content, "utils.go")
	assert.Contains(t, content, "camelCase")
}

func TestApply_HarnessUntouched(t *testing.T) {
	tmp := t.TempDir()
	harnessFile := filepath.Join(tmp, "harness-for-plan.md")
	stepFile := filepath.Join(tmp, "steps", "1.md")

	harnessContent := "# Global Harness\n\n원본 하네스 규칙\n"
	writeFile(t, harnessFile, harnessContent)
	writeFile(t, stepFile, "# Step 1\n")

	// Apply patch to step file only.
	err := Apply(stepFile, "- 새 규칙")
	require.NoError(t, err)

	// Verify harness is completely unchanged (D-008).
	data, err := os.ReadFile(harnessFile)
	require.NoError(t, err)
	assert.Equal(t, harnessContent, string(data))
}

func TestApply_NonExistentFile(t *testing.T) {
	err := Apply("/nonexistent/path/step.md", "- 규칙")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "읽기 실패")
}

// ---------------------------------------------------------------------------
// parseResponse tests
// ---------------------------------------------------------------------------

func TestParseResponse_ValidMarkers(t *testing.T) {
	raw := `Some intro text.

---RULES---
- 파일명은 snake_case를 사용한다
- 함수명은 camelCase를 사용한다
---END---

Some trailing text.`

	rules, err := parseResponse(raw)
	require.NoError(t, err)
	assert.Contains(t, rules, "snake_case")
	assert.Contains(t, rules, "camelCase")
	assert.NotContains(t, rules, "intro text")
	assert.NotContains(t, rules, "trailing text")
}

func TestParseResponse_EmptyResponse(t *testing.T) {
	_, err := parseResponse("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "빈 응답")
}

func TestParseResponse_EmptyRules(t *testing.T) {
	raw := "---RULES---\n(여기에 규칙 작성)\n---END---"
	_, err := parseResponse(raw)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "비어있음")
}

func TestParseResponse_BlankBetweenMarkers(t *testing.T) {
	raw := "---RULES---\n   \n---END---"
	_, err := parseResponse(raw)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "비어있음")
}

func TestParseResponse_FallbackMarkdown(t *testing.T) {
	raw := "## 제약 사항\n\n- 규칙 A\n- 규칙 B"
	rules, err := parseResponse(raw)
	require.NoError(t, err)
	assert.Contains(t, rules, "규칙 A")
}

func TestParseResponse_Unparseable(t *testing.T) {
	raw := "I don't know how to generate rules for this."
	_, err := parseResponse(raw)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "파싱할 수 없음")
}

// ---------------------------------------------------------------------------
// buildPrompt tests
// ---------------------------------------------------------------------------

func TestBuildPrompt_ContainsAllSections(t *testing.T) {
	input := sampleInput()
	prompt := buildPrompt(input)

	// Harness section.
	assert.Contains(t, prompt, "현재 하네스")
	assert.Contains(t, prompt, "snake_case")

	// Step prompt section.
	assert.Contains(t, prompt, "현재 Step 프롬프트")
	assert.Contains(t, prompt, "초기 스캐폴딩")

	// Divergence section.
	assert.Contains(t, prompt, "감지된 발산")
	assert.Contains(t, prompt, "네이밍 발산")
	assert.Contains(t, prompt, "utils.go vs helpers.go")

	// Diff text in code block.
	assert.Contains(t, prompt, "```diff")
	assert.Contains(t, prompt, "run-1/internal/util/utils.go")

	// Kind-specific instruction.
	assert.Contains(t, prompt, "네이밍 컨벤션")

	// Output format markers.
	assert.Contains(t, prompt, ruleStartMarker)
	assert.Contains(t, prompt, ruleEndMarker)
}

func TestBuildPrompt_MultipleKinds(t *testing.T) {
	input := Input{
		Harness:    "harness",
		StepPrompt: "step",
		Divergences: []DivergenceDetail{
			{Kind: diff.KindNaming, Path: "a.go", Detail: "naming issue"},
			{Kind: diff.KindCodePattern, Path: "b.go", Detail: "pattern issue", DiffText: "some diff"},
			{Kind: diff.KindScopeExcess, Path: "c.go", Detail: "scope issue"},
		},
	}
	prompt := buildPrompt(input)

	assert.Contains(t, prompt, "네이밍 발산")
	assert.Contains(t, prompt, "코드 패턴 발산")
	assert.Contains(t, prompt, "범위 초과")
	assert.Contains(t, prompt, "네이밍 컨벤션")
	assert.Contains(t, prompt, "구현 패턴")
	assert.Contains(t, prompt, "생성해야 할 파일")
}
