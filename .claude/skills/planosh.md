# /planosh — PRD를 결정적인 plan.sh로 변환

PRD를 입력받아 대화형으로 기술 결정을 하고, 실행 가능한 plan.sh + 하네스를 생성한다.

```
/planosh path/to/prd.md

입력: PRD (마크다운)
과정: PRD 분석 → 대화형 기술 결정 → Step 분해 → plan.sh + 하네스 생성
출력: plan.sh + .plan/{plan-name}/harness-global.md + .plan/{plan-name}/harness-step-N.md
```

## Phase 1: PRD 읽기

사용자가 `/planosh path/to/prd.md`로 호출하면 해당 파일을 읽는다.
경로 없이 `/planosh`만 호출하면 PRD 경로를 물어본다.

PRD를 읽은 뒤, 핵심을 3줄로 요약하여 사용자에게 확인한다:

```
PRD 요약:
- 제품: (한 줄)
- 핵심 기능: (한 줄)
- 대상 사용자: (한 줄)

이 PRD로 plan.sh를 생성할까요?
```

사용자가 확인하면 plan 이름을 결정한다. 이 이름이 `.plan/{name}/` 디렉토리명과 `PLAN_NAME` 변수가 된다. PRD 제목이나 기능명에서 slug를 추출하여 기본값으로 제안한다 (예: "팀 회고 앱 — 스프린트 1" → `retro-sprint1`).

하나의 프로젝트에 여러 plan이 공존할 수 있으므로, 이미 `.plan/` 안에 다른 plan 폴더가 있으면 목록을 보여준다.

Phase 2로 진행한다.

## Phase 2: 대화형 기술 결정

PRD에서 추론 가능한 것은 미리 채우고, 결정이 필요한 것만 사용자에게 물어본다.
한번에 물어보지 말고, 카테고리별로 나눠서 질문한다.

### 2-1. 기술 스택

다음을 사용자와 결정한다. PRD에 힌트가 있으면 기본값으로 제시한다.

- **프레임워크**: Next.js, Remix, SvelteKit, Rails, Django, FastAPI, ...
- **언어**: TypeScript, Python, Ruby, Go, ...
- **DB**: PostgreSQL, MySQL, SQLite, MongoDB, ...
- **ORM**: Prisma, Drizzle, TypeORM, SQLAlchemy, ActiveRecord, ...
- **인증**: NextAuth, Clerk, Supabase Auth, 직접 구현, ...
- **스타일링**: Tailwind, CSS Modules, styled-components, ...
- **테스트**: Playwright, Vitest, Jest, Pytest, ...

각 결정에 ID를 부여한다: D-001, D-002, ...

### 2-2. 아키텍처

- **디렉토리 구조**: 프레임워크 표준 구조를 기본값으로 제안
- **API 패턴**: REST, GraphQL, tRPC, Server Actions, ...
- **데이터 모델 개요**: 핵심 모델 3-5개와 관계를 ASCII 다이어그램으로

### 2-3. 범위 경계

- **MVP에 포함**: PRD에서 도출한 핵심 기능 목록
- **MVP에 불포함 (비목표)**: 명시적으로 제외할 것. PRD에 언급되었더라도 v1으로 미룰 것.

### 2-4. 검증 전략

각 Step에서 사용할 검증 방법을 결정한다.

verify에 넣을지 판단하는 기준: **같은 입력에 항상 같은 판정(pass/fail)을 내리는가.**

사용 가능한 검증:
- `npm run build` / `cargo build` (빌드)
- `[ -f path ]` (파일 존재)
- `npx tsc --noEmit` (타입 체크)
- `npm test` / `pytest` (유닛 테스트)
- `npm run lint` (린트)
- `grep -q 'pattern' file` (파일 내용)
- `curl -sf http://localhost:PORT/path` (API 응답 — 서버가 필요하므로 주의)

사용하지 않는 검증:
- AI에게 "스크린샷이 예쁜가" 판정시키는 것
- 비결정적 판정이 포함된 것

## Phase 3: Step 분해

PRD의 기능을 Step으로 분해한다.

### Step 크기 휴리스틱

- 단일 책임: 하나의 Step은 하나의 기능 단위
- 5-10개 파일 변경이 적정 범위
- Step 간 의존성은 순방향만 (Step N은 Step 1..N-1에만 의존)
- 첫 Step은 항상 스캐폴딩 (프로젝트 초기화)
- 마지막 Step은 마무리 (최종 검증, cleanup)

### Step별 정보

각 Step에 대해 다음을 정의한다:

- **이름**: 한 줄 설명
- **만들 것**: 구체적 산출물 목록
- **하지 않을 것**: 이 Step에서 명시적으로 제외할 것
- **생성할 파일 목록**: 이 Step에서 생성/수정할 파일의 정확한 경로
- **검증**: verify에 사용할 명령어
- **커밋 메시지**: checkpoint에 사용할 메시지

### 사용자 승인

Step 분해 결과를 사용자에게 보여주고 승인을 받는다.

```
Plan 구조 (N Steps):

Step 1: 프로젝트 스캐폴딩
  만들 것: ...
  검증: npm run build
  커밋: chore: project scaffolding

Step 2: ...
  ...

이 구조로 plan.sh를 생성할까요?
```

사용자가 수정을 요청하면 반영한 후 다시 승인을 받는다.
승인 없이 파일을 생성하지 않는다.

## Phase 4: 파일 생성

승인을 받으면 다음 파일들을 생성한다.

### 4-1. `.plan/{plan-name}/harness-global.md`

```markdown
# 프로젝트 컨벤션

## 기술 결정

- D-001: {결정 내용}
- D-002: {결정 내용}
...

## 코딩 규칙

{Phase 2에서 결정한 코딩 컨벤션. 프레임워크에 맞는 표준 규칙.}

## 절대 금지

- any 타입 사용 (TypeScript인 경우)
- 하드코딩된 시크릿
- 이 Step의 범위 밖 파일 수정
{프레임워크에 맞는 금지 패턴 추가}
```

### 4-2. `.plan/{plan-name}/harness-step-N.md`

각 Step마다 하나씩 생성한다.

```markdown
# Step N 하네스: {Step 이름}

## 현재 프로젝트 상태

{이전 Step들이 완료된 후 존재하는 파일/상태 목록}

## 이 Step의 아키텍처 제약

{이 Step에 특화된 아키텍처 결정}

## 이 Step에서 생성할 파일 목록 (이 목록 외 파일 생성 금지)

- path/to/file1
- path/to/file2
...
```

Step 1의 하네스에는 "현재 프로젝트 상태"가 빈 프로젝트이므로 해당 섹션을 생략하거나 "빈 프로젝트"로 명시한다.

### 4-3. `plan.sh`

아래 템플릿을 기반으로 생성한다. Step별 내용을 채운다.

```bash
#!/bin/bash
# 계획: {PRD 제목}
# 생성: {날짜} by /planosh
# PRD: {PRD 파일 경로}
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

PLAN_NAME="{plan-name}"

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
  # Step별 하네스에 파일 목록이 있으면 범위 외 변경 경고
  if [ -f "$step_harness" ] && grep -q '생성할 파일 목록' "$step_harness"; then
    local unexpected=$(git diff --name-only | grep -v -f <(grep '^\- ' "$step_harness" | sed 's/^- //') 2>/dev/null || true)
    [ -n "$unexpected" ] && echo "⚠️  범위 외 변경 감지: $unexpected"
  fi
  git add -A && git commit -m "$1"
}

# ── 브랜치 생성 ──
[ "$START_FROM" -eq 1 ] && ! $DRY_RUN && git checkout -b plan-$(date +%Y%m%d) main 2>/dev/null || true

{각 Step의 코드 블록이 여기에 들어간다}

# ── 완료 ──
echo ""; echo "🎉 계획 완료! 브랜치: $(git branch --show-current)"
echo "다음: gh pr create --base main --head $(git branch --show-current)"
rm -f ".plan-state-$PLAN_NAME"
```

각 Step은 다음 패턴으로 생성한다:

```bash
# ── Step N: {이름} ──
# 하네스: .plan/{plan-name}/harness-global.md + .plan/{plan-name}/harness-step-N.md
CURRENT_STEP=N; step N "{이름}"
run_claude "
{만들 것과 하지 않을 것을 포함한 프롬프트}
"
verify "{검증 이름}" "{검증 명령}"
checkpoint "{커밋 메시지}"
```

### 프롬프트 작성 원칙

plan.sh에 들어가는 각 Step의 프롬프트(-p)에는 WHAT만 넣는다:

- "만들 것" — 이 Step의 구체적 산출물
- "하지 않을 것" — 이 Step의 범위 외 항목

HOW는 하네스에 넣는다:
- 기술 결정, 코딩 컨벤션 → harness-global.md
- 이전 Step 상태, 아키텍처 제약, 파일 화이트리스트 → harness-step-N.md

프롬프트가 짧을수록 좋다. 기술 결정이나 코딩 규칙을 프롬프트에 반복하지 않는다.

## Phase 5: 최종 안내

파일 생성이 완료되면 사용자에게 안내한다:

```
생성 완료:
  plan.sh                                    ← 실행 계획
  .plan/{plan-name}/harness-global.md        ← 글로벌 하네스
  .plan/{plan-name}/harness-step-1.md        ← Step 1 하네스
  .plan/{plan-name}/harness-step-2.md        ← Step 2 하네스
  ...

다음 단계:
  1. plan.sh를 읽고 리뷰하세요
  2. bash plan.sh --dry 로 프롬프트를 미리 확인하세요
  3. 리뷰 완료 후 bash plan.sh 로 실행하세요
  4. /planosh-calibrate --step=N 으로 발산을 측정하고 하네스를 강화하세요
```
