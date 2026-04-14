package patch

import (
	"fmt"
	"strings"

	"github.com/ASDFGHoney/planosh/internal/diff"
)

const (
	ruleStartMarker = "---RULES---"
	ruleEndMarker   = "---END---"
)

// kindLabel returns a human-readable Korean label for a divergence kind.
func kindLabel(k diff.DivergenceKind) string {
	switch k {
	case diff.KindNaming:
		return "네이밍 발산"
	case diff.KindStructure:
		return "구조 발산"
	case diff.KindCodePattern:
		return "코드 패턴 발산"
	case diff.KindScopeExcess:
		return "범위 초과"
	default:
		return string(k)
	}
}

// kindInstruction returns kind-specific generation guidance for the prompt.
func kindInstruction(k diff.DivergenceKind) string {
	switch k {
	case diff.KindNaming:
		return "파일명, 함수명, 변수명 등 네이밍 컨벤션을 명확히 지정하는 규칙을 생성하라. 정확한 이름을 고정하라."
	case diff.KindStructure:
		return "디렉토리 구조와 파일 배치를 명확히 지정하는 규칙을 생성하라. 정확한 경로를 고정하라."
	case diff.KindCodePattern:
		return "구현 패턴, 코드 스타일, API 사용 방식을 명확히 지정하는 규칙을 생성하라. 구체적인 코드 패턴을 고정하라."
	case diff.KindScopeExcess:
		return "이 step에서 생성해야 할 파일과 생성하지 말아야 할 파일을 명확히 지정하는 규칙을 생성하라. 범위를 고정하라."
	default:
		return "발산을 해소할 수 있는 구체적인 규칙을 생성하라."
	}
}

// buildPrompt constructs the full prompt for claude -p patch generation.
func buildPrompt(input Input) string {
	var sb strings.Builder

	sb.WriteString("You are a harness calibration assistant. ")
	sb.WriteString("Your task is to generate additional architectural constraint rules for a plan step prompt ")
	sb.WriteString("to eliminate divergence between parallel AI code-generation runs.\n\n")

	sb.WriteString("## 현재 하네스 (harness-for-plan.md)\n\n")
	sb.WriteString(input.Harness)
	sb.WriteString("\n\n")

	sb.WriteString("## 현재 Step 프롬프트\n\n")
	sb.WriteString(input.StepPrompt)
	sb.WriteString("\n\n")

	sb.WriteString("## 감지된 발산\n\n")

	kindSet := make(map[diff.DivergenceKind]bool)
	for _, d := range input.Divergences {
		kindSet[d.Kind] = true
		sb.WriteString(fmt.Sprintf("### %s: %s\n", kindLabel(d.Kind), d.Path))
		sb.WriteString(fmt.Sprintf("상세: %s\n\n", d.Detail))
		if d.DiffText != "" {
			sb.WriteString("```diff\n")
			sb.WriteString(d.DiffText)
			sb.WriteString("\n```\n\n")
		}
	}

	sb.WriteString("## 규칙 생성 지침\n\n")
	for k := range kindSet {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", kindLabel(k), kindInstruction(k)))
	}
	sb.WriteString("\n")

	sb.WriteString("## 출력 형식\n\n")
	sb.WriteString("아래 마커 사이에 마크다운 규칙 블록을 출력하라. 규칙만 출력하고, 설명은 포함하지 마라.\n")
	sb.WriteString("규칙은 `## 아키텍처 제약` 섹션에 추가할 수 있는 형태여야 한다.\n\n")
	sb.WriteString("규칙 요건:\n")
	sb.WriteString("- 구체적이고 실행 가능해야 한다 (모호한 가이드라인 금지)\n")
	sb.WriteString("- 이 step에만 적용되어야 한다\n")
	sb.WriteString("- 기존 하네스 규칙과 호환되어야 한다\n")
	sb.WriteString("- 향후 AI run이 동일한 결과를 생성하도록 강제해야 한다\n\n")
	sb.WriteString(ruleStartMarker + "\n")
	sb.WriteString("(여기에 규칙 작성)\n")
	sb.WriteString(ruleEndMarker + "\n")

	return sb.String()
}

// parseResponse extracts rule text from a claude -p response.
// Returns error if the response is empty or cannot be parsed.
func parseResponse(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("빈 응답")
	}

	startIdx := strings.Index(raw, ruleStartMarker)
	endIdx := strings.Index(raw, ruleEndMarker)

	if startIdx >= 0 && endIdx > startIdx {
		rules := raw[startIdx+len(ruleStartMarker) : endIdx]
		rules = strings.TrimSpace(rules)
		if rules == "" || rules == "(여기에 규칙 작성)" {
			return "", fmt.Errorf("규칙 내용이 비어있음")
		}
		return rules, nil
	}

	// Fallback: use the entire response if it looks like markdown rules.
	if strings.Contains(raw, "##") || strings.Contains(raw, "- ") {
		return raw, nil
	}

	return "", fmt.Errorf("응답에서 규칙을 파싱할 수 없음")
}
