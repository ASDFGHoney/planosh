#!/bin/bash
# 계획: Hello World HTTP 서버
# 생성: 2026-04-07 by /planosh
# PRD: (데모 — PRD 없음)
#
# 사용법:
#   bash plan.sh          전체 실행
#   bash plan.sh --dry    실행 없이 프롬프트만 출력
#   bash plan.sh --from=N Step N부터 재개
#
# 주의: --dangerously-skip-permissions를 사용합니다.
# 반드시 plan.sh를 리뷰한 후 실행하세요.
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_ROOT"

PLAN_NAME="hello-world"

DRY_RUN=false; START_FROM=1
# .plan-state가 있으면 마지막 실패 Step을 기본값으로 사용
[ -f ".plan-state-$PLAN_NAME" ] && START_FROM=$(cat ".plan-state-$PLAN_NAME")
for arg in "$@"; do
  case $arg in
    --dry) DRY_RUN=true ;;
    --from=*) START_FROM="${arg#*=}" ;;
  esac
done

# ── 하네스 경로 ──
HARNESS_DIR="$PROJECT_ROOT/.plan/$PLAN_NAME"
GLOBAL_HARNESS="$HARNESS_DIR/harness-global.md"

step() {
  local n=$1 name=$2
  [ "$n" -lt "$START_FROM" ] && echo "⏭️  Step $n: $name (건너뜀)" && return 0
  echo ""; echo "━━ Step $n: $name ━━"
}

run_claude() {
  local prompt=$1
  local step_harness="$HARNESS_DIR/harness-step-${CURRENT_STEP}.md"

  # 하네스 구성: 글로벌 + Step별
  local harness=""
  [ -f "$GLOBAL_HARNESS" ] && harness="$(cat "$GLOBAL_HARNESS")"
  [ -f "$step_harness" ] && harness="$harness"$'\n\n'"$(cat "$step_harness")"

  if $DRY_RUN; then
    echo "[DRY] 프롬프트:"; echo "$prompt"; echo ""
    echo "[DRY] 하네스: $GLOBAL_HARNESS + $step_harness"
    return 0
  fi

  claude -p "$prompt" \
    --append-system-prompt "$harness" \
    --dangerously-skip-permissions
}

verify() {
  $DRY_RUN && echo "[DRY] 검증: $1" && return 0
  echo "🔍 $1"
  eval "$2" || { echo "❌ $1"; echo "$CURRENT_STEP" > ".plan-state-$PLAN_NAME"; exit 1; }
  echo "✅ $1"
}

checkpoint() {
  $DRY_RUN && return 0
  local step_harness="$HARNESS_DIR/harness-step-${CURRENT_STEP}.md"
  if [ -f "$step_harness" ] && grep -q '생성할 파일 목록' "$step_harness"; then
    local unexpected=$(git diff --name-only | grep -v -f <(grep '^\- ' "$step_harness" | sed 's/^- //') 2>/dev/null || true)
    [ -n "$unexpected" ] && echo "⚠️  범위 외 변경 감지: $unexpected"
  fi
  git add -A && git commit -m "$1"
}

# ── 브랜치 생성 ──
[ "$START_FROM" -eq 1 ] && ! $DRY_RUN && git checkout -b plan-$(date +%Y%m%d) main 2>/dev/null || true

# ── Step 1: 프로젝트 초기화 ──
# 하네스: .plan/hello-world/harness-global.md + .plan/hello-world/harness-step-1.md
CURRENT_STEP=1; step 1 "프로젝트 초기화"
run_claude "
Node.js + TypeScript 프로젝트를 초기화하세요.
## 만들 것
- package.json (name: hello-planosh, type: module)
- tsconfig.json (strict, ES2022, NodeNext)
- src/index.ts (console.error('hello') 한 줄만)
## 하지 않을 것
- 외부 패키지 설치
- 어떤 기능도 구현하지 않음. 초기화만.
"
verify "package.json 존재" "[ -f package.json ]"
verify "tsconfig.json 존재" "[ -f tsconfig.json ]"
verify "src/index.ts 존재" "[ -f src/index.ts ]"
checkpoint "chore: project init"

# ── Step 2: HTTP 서버 ──
# 하네스: .plan/hello-world/harness-global.md + .plan/hello-world/harness-step-2.md
CURRENT_STEP=2; step 2 "HTTP 서버"
run_claude "
node:http로 간단한 HTTP 서버를 구현하세요.
## 만들 것
- GET / → { \"message\": \"hello, planosh\" } (JSON, 200)
- GET /health → { \"status\": \"ok\" } (JSON, 200)
- 그 외 → 404
- 포트: 환경변수 PORT 또는 3000
## 하지 않을 것
- express, fastify 등 외부 프레임워크 사용
- 테스트 작성
- 미들웨어, 로깅, CORS 등 부가 기능
"
verify "src/index.ts 존재" "[ -f src/index.ts ]"
verify "src/server.ts 존재" "[ -f src/server.ts ]"
checkpoint "feat: http server"

# ── 완료 ──
echo ""; echo "🎉 계획 완료! 브랜치: $(git branch --show-current)"
echo "다음: gh pr create --base main --head $(git branch --show-current)"
rm -f ".plan-state-$PLAN_NAME"
