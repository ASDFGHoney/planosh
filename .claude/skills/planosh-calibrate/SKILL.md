---
description: plan.sh를 병렬 실행하여 발산 지점을 찾고 하네스를 강화. "calibrate", "발산 측정", "하네스 강화", "/planosh-calibrate" 등의 요청에 사용.
---

# /planosh-calibrate — 발산 탐지 및 하네스 강화

plan.sh를 격리된 환경에서 N번 병렬 실행하고, 발산 지점을 찾아 사용자에게 결정을 요청한 뒤, 그 결정을 하네스에 추가한다.

```
/planosh-calibrate [--step=M] [--runs=N]

입력: plan.sh + 기존 하네스
과정: 병렬 실행 → 발산 분석 → 사용자에게 결정 요청 → 하네스 업데이트
출력: 강화된 하네스 + .plan/{plan-name}/divergence-report.md
```

기본값: `--runs=3`. `--step`을 생략하면 전체 plan.sh를 실행한다.

## 전제 조건

- `plan.sh`가 현재 프로젝트에 존재해야 한다
- plan.sh 내부의 `PLAN_NAME` 변수를 읽어 `.plan/{plan-name}/` 디렉토리를 특정한다
- `--step=M`을 사용할 경우, Step M-1까지 완료된 커밋이 현재 브랜치에 있어야 한다
- v0 제약: 빌드/파일 검증만 있는 plan.sh에서만 병렬 실행이 안전하다. 포트나 DB를 사용하는 verify가 있으면 병렬 실행 시 충돌할 수 있다.

## Phase 1: 격리 환경 생성 및 병렬 실행

### 1-1. 사전 검증

plan.sh에서 `PLAN_NAME`을 읽고, 해당 하네스 디렉토리가 존재하는지 확인한다.

```bash
[ -f plan.sh ] || { echo "plan.sh를 찾을 수 없습니다."; exit 1; }
PLAN_NAME=$(grep '^PLAN_NAME=' plan.sh | head -1 | cut -d'"' -f2)
[ -z "$PLAN_NAME" ] && { echo "plan.sh에 PLAN_NAME이 정의되지 않았습니다."; exit 1; }
[ -d ".plan/$PLAN_NAME" ] || { echo ".plan/$PLAN_NAME/ 디렉토리를 찾을 수 없습니다."; exit 1; }
```

`--step=M` 옵션이 있으면, plan.sh에서 해당 Step이 존재하는지 확인한다.

### 1-2. worktree 생성

N개의 격리된 git worktree를 생성한다.

```bash
BASE_BRANCH=$(git branch --show-current)
for i in $(seq 1 $RUNS); do
  WORKTREE_DIR=".calibrate/run-$i"
  git worktree add "$WORKTREE_DIR" -b "calibrate-run-$i" HEAD
done
```

각 worktree에 `.plan/$PLAN_NAME/` 디렉토리를 복사한다 (worktree는 tracked 파일만 공유하므로).

### 1-3. 병렬 실행

각 worktree에서 plan.sh를 실행한다. `--step=M`이 지정되면 해당 Step만 실행한다.

Claude Code의 Agent 도구를 사용하여 각 run을 병렬로 실행한다:

```
각 run에 대해 Agent를 spawn:
  - worktree 디렉토리로 이동
  - bash plan.sh (또는 bash plan.sh --from=M 으로 특정 Step만)
  - 실행 결과를 run-N.log에 기록
```

실패한 run이 있으면 해당 run을 제외하고 나머지로 분석을 진행한다.
모든 run이 실패하면 에러를 보고하고 중단한다.

### 1-4. 결과 수집

각 worktree에서 실행 후 변경된 파일 목록과 내용을 수집한다.

```bash
for i in $(seq 1 $RUNS); do
  WORKTREE_DIR=".calibrate/run-$i"
  # 변경된 파일 목록
  git -C "$WORKTREE_DIR" diff --name-only HEAD > ".calibrate/run-$i-files.txt"
  # 각 파일의 내용
  git -C "$WORKTREE_DIR" diff HEAD > ".calibrate/run-$i-diff.patch"
done
```

## Phase 2: 발산 분석

### 2-1. 구조 발산 (자동 탐지)

각 run의 파일 목록을 비교한다. 모든 run에 존재하는 파일과, 일부 run에만 존재하는 파일을 분류한다.

```
파일 존재 매트릭스:
| 파일 경로            | run-1 | run-2 | run-3 |
| -------------------- | ----- | ----- | ----- |
| src/lib/auth.ts      | ✅    | ✅    | ✅    |  ← 수렴
| src/lib/auth-opts.ts | ❌    | ❌    | ✅    |  ← 발산
```

모든 run에 동일하게 존재하는 파일은 수렴, 그렇지 않은 파일은 구조 발산으로 분류한다.

### 2-2. 내용 발산 (AI 분류)

모든 run에 존재하는 파일에 대해, run 간 diff를 Claude에게 보여주고 발산 유형을 분류하게 한다.

분류할 발산 유형:

- **패턴 발산**: 같은 기능인데 다른 구현 패턴 (예: JWT vs database 세션)
- **네이밍 발산**: 같은 개념인데 다른 이름 (예: authOptions vs authConfig)
- **범위 발산**: 요청하지 않은 기능이 추가됨 (예: 커스텀 에러 페이지)

각 발산에 대해:
- 어떤 run들이 어떤 선택을 했는지 표로 정리
- 다수결이 있으면 기본 추천으로 제시

**사용자에게 AI 분류 결과를 확인받는다.** 분류가 틀렸으면 사용자가 수정한다.

## Phase 3: 사용자 결정 요청

각 발산 지점을 사용자에게 하나씩 제시하고 결정을 받는다.

발산 제시 형식:

```
발산 #1: 세션 전략
  run-1: JWT
  run-2: JWT
  run-3: database
  
  A) JWT (2/3 runs 선택)
  B) database (1/3 runs 선택)
  
어느 쪽을 사용할까요?
```

범위 발산의 경우:

```
발산 #2: 커스텀 에러 페이지
  run-1: 없음
  run-2: 생성됨
  run-3: 없음
  
  이 기능은 요청하지 않았습니다.
  A) 금지 (harness에 "절대 금지" 추가)
  B) 허용 (이 기능을 포함하기로 결정)
```

## Phase 4: 하네스 업데이트

사용자 결정을 하네스 규칙으로 변환하여 추가한다.

### 결정 → 규칙 매핑

| 발산 유형 | 하네스 위치 | 규칙 형태 |
|----------|-----------|----------|
| 구조 발산 | harness-step-M.md | "생성할 파일 목록"에 파일 추가/제거 |
| 패턴 발산 | harness-step-M.md | "이 Step의 아키텍처 제약"에 규칙 추가 |
| 네이밍 발산 | harness-global.md | "코딩 규칙"에 네이밍 컨벤션 추가 |
| 범위 발산 | harness-step-M.md 또는 harness-global.md | "절대 금지"에 항목 추가 |

네이밍 발산은 프로젝트 전체에 영향을 미치므로 글로벌 하네스에 추가한다.
나머지는 해당 Step의 하네스에 추가한다.

### 하네스 수정 시 주의

- 기존 규칙을 삭제하지 않는다. 추가만 한다.
- 추가하는 규칙이 기존 규칙과 모순되면 사용자에게 알리고 어느 쪽을 유지할지 결정받는다.
- 규칙 추가 후 기존에 수렴했던 항목이 새로 발산할 수 있다 (프롬프트 맥락 변경). 이 리스크를 사용자에게 안내한다.

## Phase 5: 발산 리포트 생성

`.plan/$PLAN_NAME/divergence-report.md`에 교정 결과를 기록한다.

```markdown
# 발산 리포트 — Step M: {Step 이름}

실행 횟수: {N} | 날짜: {날짜}

## 수렴율: {X}% ({N}회 중 {Y}회 일치)

## 구조 발산

{파일 존재 매트릭스}

## 패턴 발산

{패턴 비교 테이블}

## 네이밍 발산

{네이밍 비교 테이블}

## 범위 발산

{범위 초과 항목}

## 발산 요약

- 구조: {N}건
- 패턴: {N}건
- 네이밍: {N}건
- 범위: {N}건

## 사용자 결정 ({N}건)

{각 결정 내용}

## 하네스 변경 ({N}건)

{추가된 규칙 목록}
```

교정 이력은 `.plan/$PLAN_NAME/calibration-history/`에 누적한다:

```
.plan/{plan-name}/calibration-history/
├── step-M-run-1.log
├── step-M-run-2.log
├── step-M-run-3.log
└── step-M-convergence.md    ← 수렴 추이 (67% → 95% → ...)
```

## Phase 6: cleanup 및 안내

### worktree 정리

```bash
for i in $(seq 1 $RUNS); do
  git worktree remove ".calibrate/run-$i" --force 2>/dev/null || true
  git branch -D "calibrate-run-$i" 2>/dev/null || true
done
rmdir .calibrate 2>/dev/null || true
```

### 결과 안내

```
교정 완료:
  발산: {N}건 발견, {M}건 해결
  수렴율: {이전}% → {현재}% (예상)
  하네스 변경: {파일 목록}
  리포트: .plan/{plan-name}/divergence-report.md

다음 단계:
  1. 변경된 하네스를 확인하세요
  2. /planosh-calibrate --step=M 으로 다시 교정하여 수렴을 확인하세요
  3. 수렴율 100%에 가까워지면 교정을 종료하세요
```

수렴율이 이미 100%면:

```
수렴율 100%! Step M의 하네스가 충분히 결정적입니다.
추가 교정이 필요하지 않습니다.
```
