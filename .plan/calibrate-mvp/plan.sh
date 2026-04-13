#!/bin/bash
# 계획: planosh CLI — calibrate MVP
# 생성: 2026-04-13 by /planosh
# PRD: ~/.gstack/projects/ASDFGHoney-planosh/graphicchiheon-main-design-20260413-185454.md
# 사람 작업: .plan/calibrate-mvp/plan-for-human.md 참조
#
# 사용법:
#   bash .plan/calibrate-mvp/plan.sh                  전체 실행
#   bash .plan/calibrate-mvp/plan.sh --dry            실행 없이 프롬프트만 출력
#   bash .plan/calibrate-mvp/plan.sh --from=N         Step N부터 재개
#   bash .plan/calibrate-mvp/plan.sh --to=M           Step M까지만 실행
#   bash .plan/calibrate-mvp/plan.sh --from=N --to=M  Step N~M만 실행
#   bash .plan/calibrate-mvp/plan.sh --model=opus    모델 오버라이드
#   bash .plan/calibrate-mvp/plan.sh --effort=max    effort 오버라이드
#   bash .plan/calibrate-mvp/plan.sh --testbed       calibrate용 경량 모드 (Haiku, low)
#
# 주의: --dangerously-skip-permissions를 사용합니다.
# 반드시 steps.json + steps/*.md를 리뷰한 후 실행하세요.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# .plan/{name}/plan.sh → .plan/의 부모가 PROJECT_ROOT
# 환경변수가 주입되면 우선 사용 (calibrate CLI가 testbed 실행 시 주입)
PROJECT_ROOT="${PROJECT_ROOT:-$(cd "$SCRIPT_DIR/../.." && pwd)}"
cd "$PROJECT_ROOT"

# 안전장치: PROJECT_ROOT에 .plan/ 또는 .git이 존재하는지 확인
if [ ! -d "$PROJECT_ROOT/.plan" ] && [ ! -d "$PROJECT_ROOT/.git" ]; then
  echo "ERROR: PROJECT_ROOT($PROJECT_ROOT)에 .plan/ 또는 .git이 없습니다."
  echo "plan.sh가 올바른 위치에서 실행되고 있는지 확인하세요."
  exit 1
fi

DRY_RUN=false; START_FROM=1; STOP_AFTER=999; TESTBED=false
DEFAULT_MODEL="opus"; DEFAULT_EFFORT="max"
MODEL="$DEFAULT_MODEL"; EFFORT="$DEFAULT_EFFORT"
[ -f "$SCRIPT_DIR/.plan-state" ] && START_FROM=$(cat "$SCRIPT_DIR/.plan-state")

# 경로 셀프테스트: SCRIPT_DIR이 .plan/ 안에 있는지 확인
case "$SCRIPT_DIR" in
  */.plan/*) ;; # 정상
  *) echo "WARN: SCRIPT_DIR($SCRIPT_DIR)이 .plan/ 안에 있지 않습니다" ;;
esac
for arg in "$@"; do
  case $arg in
    --dry) DRY_RUN=true ;;
    --from=*) START_FROM="${arg#*=}" ;;
    --to=*) STOP_AFTER="${arg#*=}" ;;
    --model=*) MODEL="${arg#*=}" ;;
    --effort=*) EFFORT="${arg#*=}" ;;
    --testbed) TESTBED=true ;;
  esac
done
$TESTBED && MODEL="haiku" && EFFORT="low"

PLAN_HARNESS="$SCRIPT_DIR/harness-for-plan.md"
STEPS_FILE="$SCRIPT_DIR/steps.json"
STEPS_DIR="$SCRIPT_DIR/steps"

# ── steps.json → bash 변수 로드 ──
eval "$(python3 -c "
import json, shlex
with open('$STEPS_FILE') as f:
    data = json.load(f)
steps = data['steps']
print(f'STEP_COUNT={len(steps)}')
for i, s in enumerate(steps):
    print(f'STEP_{i}_ID={s[\"id\"]}')
    print(f'STEP_{i}_NAME={shlex.quote(s[\"name\"])}')
    print(f'STEP_{i}_PROMPT={shlex.quote(s[\"prompt\"])}')
    print(f'STEP_{i}_COMMIT={shlex.quote(s[\"commit\"])}')
    vlist = s.get('verify', [])
    print(f'STEP_{i}_VERIFY_COUNT={len(vlist)}')
    for j, v in enumerate(vlist):
        print(f'STEP_{i}_VERIFY_{j}_NAME={shlex.quote(v[\"name\"])}')
        print(f'STEP_{i}_VERIFY_{j}_RUN={shlex.quote(v[\"run\"])}')
")"

# ── 공통 함수 ──
run_claude() {
  local prompt_file="$STEPS_DIR/$1"
  local prompt
  prompt=$(cat "$prompt_file")

  local harness=""
  [ -f "$PLAN_HARNESS" ] && harness="$(cat "$PLAN_HARNESS")"

  if $DRY_RUN; then
    echo "[DRY] prompt ($1):"; echo "$prompt"; echo ""
    [ -n "$harness" ] && echo "[DRY] harness: $PLAN_HARNESS"
    echo "[DRY] model: $MODEL, effort: $EFFORT"
    return 0
  fi

  local effort_flag=""
  [ "$EFFORT" != "auto" ] && effort_flag="--effort $EFFORT"

  claude -p "$prompt" \
    --model "$MODEL" \
    $effort_flag \
    --append-system-prompt "$harness" \
    --dangerously-skip-permissions
}

verify() {
  $DRY_RUN && echo "[DRY] verify: $1" && return 0
  echo "verify: $1"
  ( eval "$2" ) || { echo "FAIL: $1"; echo "$CURRENT_STEP" > "$SCRIPT_DIR/.plan-state"; exit 1; }
  echo "PASS: $1"
}

checkpoint() {
  $DRY_RUN && return 0
  local step_file="$STEPS_DIR/$CURRENT_PROMPT"
  if grep -q '생성할 파일 목록' "$step_file" 2>/dev/null; then
    local unexpected
    unexpected=$(git diff --name-only | grep -v -f <(grep '^\- ' "$step_file" | sed 's/^- //') 2>/dev/null || true)
    [ -n "$unexpected" ] && echo "WARN: out-of-scope changes: $unexpected"
  fi
  git add -A && git commit -m "$1"
}

# ── 브랜치 생성 ──
[ "$START_FROM" -eq 1 ] && ! $DRY_RUN && git checkout -b "plan-$(date +%Y%m%d)" main 2>/dev/null || true

# ── Step 루프 ──
for ((i=0; i<STEP_COUNT; i++)); do
  eval "STEP_ID=\$STEP_${i}_ID"
  eval "STEP_NAME=\$STEP_${i}_NAME"
  eval "STEP_PROMPT=\$STEP_${i}_PROMPT"
  eval "STEP_COMMIT=\$STEP_${i}_COMMIT"
  eval "VERIFY_COUNT=\$STEP_${i}_VERIFY_COUNT"

  [ "$STEP_ID" -gt "$STOP_AFTER" ] && echo "" && echo "== Done: Step $STOP_AFTER 완료 ==" && break
  if [ "$STEP_ID" -lt "$START_FROM" ]; then
    echo "Skip Step $STEP_ID: $STEP_NAME"
    continue
  fi

  echo ""; echo "== Step $STEP_ID: $STEP_NAME =="
  CURRENT_STEP=$STEP_ID
  CURRENT_PROMPT=$STEP_PROMPT

  run_claude "$STEP_PROMPT"

  for ((j=0; j<VERIFY_COUNT; j++)); do
    eval "V_NAME=\$STEP_${i}_VERIFY_${j}_NAME"
    eval "V_RUN=\$STEP_${i}_VERIFY_${j}_RUN"
    verify "$V_NAME" "$V_RUN"
  done

  checkpoint "$STEP_COMMIT"
done

# ── 완료 ──
echo ""; echo "Plan complete. Branch: $(git branch --show-current)"
echo "Next: gh pr create --base main --head $(git branch --show-current)"
rm -f "$SCRIPT_DIR/.plan-state"
