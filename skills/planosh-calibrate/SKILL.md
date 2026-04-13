---
name: planosh-calibrate
description: plan.sh의 각 Step을 순차적으로 교정하여 하네스를 강화. "calibrate", "하네스 강화", "/planosh-calibrate" 등의 요청에 사용.
---

# /planosh-calibrate — Step별 순차 수렴

plan.sh의 각 Step을 순서대로 N번 병렬 실행하여, 해당 Step이 수렴할 때까지 하네스를 강화한 뒤 다음 Step으로 넘어간다.

```
/planosh-calibrate [--from=M] [--runs=N]

입력: plan.sh + 기존 하네스
과정: Step 1 교정 → 수렴 → Step 2 교정 → 수렴 → ... → 전체 완료
출력: 강화된 하네스 + .plan/{plan-name}/calibration-report.md
```

기본값: `--runs=3`, `--from=1`.

## 전제 조건

- `plan.sh`가 현재 프로젝트에 존재해야 한다
- plan.sh 내부의 `PLAN_NAME` 변수를 읽어 `.plan/{plan-name}/` 디렉토리를 특정한다
- `--from=M`(M≥2)을 사용할 경우, Step M-1까지 완료된 커밋이 현재 브랜치에 있어야 한다
- v0 제약: 포트나 DB를 사용하는 verify가 있으면 병렬 실행 시 충돌할 수 있다. 빌드/파일 검증만 있는 plan.sh에서 가장 안전하다.

## Phase 0: 사전 검증 + Step 파싱

### 0-1. plan.sh 검증

```bash
[ -f plan.sh ] || { echo "plan.sh를 찾을 수 없습니다."; exit 1; }
PLAN_NAME=$(grep '^PLAN_NAME=' plan.sh | head -1 | cut -d'"' -f2)
[ -z "$PLAN_NAME" ] && { echo "plan.sh에 PLAN_NAME이 정의되지 않았습니다."; exit 1; }
[ -d ".plan/$PLAN_NAME" ] || { echo ".plan/$PLAN_NAME/ 디렉토리를 찾을 수 없습니다."; exit 1; }
```

### 0-2. Step 목록 추출

plan.sh에서 `CURRENT_STEP=N; step N "이름"` 패턴을 파싱하여 전체 Step 목록을 추출한다.

파싱 결과를 사용자에게 보여준다:

```
plan.sh에서 N개 Step을 감지했습니다:
  Step 1: 프로젝트 스캐폴딩
  Step 2: DB 스키마 + API
  Step 3: 인증
  ...

Step 1부터 순차 교정을 시작합니다.
```

### 0-3. 골든 베이스 초기화

`.plan/$PLAN_NAME/testbed/golden`에 현재 repo의 clone을 생성한다. 골든 베이스는 수렴된 각 Step의 결과를 누적하여, 다음 Step 교정의 출발점이 된다.

```bash
TESTBED_DIR=".plan/$PLAN_NAME/testbed"
REPO_ROOT=$(git rev-parse --show-toplevel)
rm -rf "$TESTBED_DIR"
mkdir -p "$TESTBED_DIR"
git clone --depth 1 "file://$REPO_ROOT" "$TESTBED_DIR/golden"
```

`--from=M`(M≥2)인 경우, 현재 브랜치에 Step 1..M-1의 커밋이 있으므로 golden base에도 그 상태가 반영된다.

`testbed/`는 `.gitignore`에 추가:
```bash
grep -q 'testbed/' ".plan/$PLAN_NAME/.gitignore" 2>/dev/null || echo "testbed/" >> ".plan/$PLAN_NAME/.gitignore"
```

## Phase 1: Step M 교정 (각 Step에 대해 반복)

Step `--from`부터 마지막 Step까지 순차적으로 진행한다.

```
1-1. testbed 생성 (golden → N개 clone)  <--+
  |                                         |
1-2. 병렬 실행 (Step M만)                   |
  |                                         |
1-3. 결과 수집                              |
  |                                         |
1-4. 발산 분석                              |
  |                                         |
1-5. 수렴 판정                              |
  |- 수렴 → Phase 2로                       |
  |- 발산 → 1-6으로                         |
  |                                         |
1-6. 사용자 결정 요청                       |
  |                                         |
1-7. 하네스 업데이트                        |
  |                                         |
1-8. 재검증 ------------------------------------+
```

### 1-1. testbed 생성

golden base에서 N개 clone을 생성한다.

```bash
for i in $(seq 1 $RUNS); do
  cp -r "$TESTBED_DIR/golden" "$TESTBED_DIR/run-$i"
done
```

`cp -r` 사용: golden base는 이미 로컬 clone이므로, 다시 git clone보다 빠르다.

plan.sh가 `--from`과 `--to`를 지원하므로 clone을 수정할 필요 없이 `--from=M --to=M`으로 Step M만 실행한다. golden base에 Step 1..M-1이 이미 커밋되어 있으므로 이전 Step은 skip된다.

### 1-2. 병렬 실행

Claude Code의 Agent 도구로 각 run을 병렬 실행한다:

```
각 run에 대해 Agent를 spawn:
  - $TESTBED_DIR/run-$i 디렉토리에서
  - bash plan.sh --from=M --to=M 실행
  - 실행 결과를 기록
```

실패한 run은 제외하고 나머지로 분석 진행. 모든 run 실패 시 에러 보고 후 중단.

### 1-3. 결과 수집

각 clone에서 Step M의 변경분을 수집한다.

```bash
for i in $(seq 1 $RUNS); do
  RUN_DIR="$TESTBED_DIR/run-$i"
  git -C "$RUN_DIR" diff --name-only HEAD > "$TESTBED_DIR/step-$M-run-$i-files.txt"
  git -C "$RUN_DIR" diff HEAD > "$TESTBED_DIR/step-$M-run-$i-diff.patch"
done
```

### 1-4. 발산 분석

#### 구조 발산 (자동 탐지)

각 run의 파일 목록을 비교한다. 모든 run에 존재하는 파일과, 일부 run에만 존재하는 파일을 분류한다.

```
Step M 파일 존재 매트릭스:
| 파일 경로            | run-1 | run-2 | run-3 |
| -------------------- | ----- | ----- | ----- |
| src/lib/auth.ts      | ✅    | ✅    | ✅    |  ← 수렴
| src/lib/auth-opts.ts | ❌    | ❌    | ✅    |  ← 발산
```

#### 내용 발산 (AI 분류)

모든 run에 존재하는 파일에 대해, run 간 diff를 비교하고 발산 유형을 분류한다:

- **패턴 발산**: 같은 기능인데 다른 구현 (예: JWT vs database 세션)
- **네이밍 발산**: 같은 개념인데 다른 이름 (예: authOptions vs authConfig)
- **범위 발산**: 요청하지 않은 기능 추가

**사용자에게 AI 분류 결과를 확인받는다.** 분류가 틀렸으면 사용자가 수정한다.

### 1-5. 수렴 판정

**수렴**: 구조 발산 0건, 내용 발산 0건.

```
✅ Step M: {이름} — 수렴 ({RUNS}회 실행 모두 동일)
```

→ Phase 2로 이동.

**발산**: 1건 이상.

```
Step M: {이름} — {N}건 발산
하나씩 결정합니다.
```

→ 1-6으로 이동.

### 1-6. 사용자 결정 요청

각 발산을 사용자에게 제시하고 결정을 받는다.

```
[Step M] 발산 #1: 세션 전략
  run-1: JWT
  run-2: JWT
  run-3: database
  
  A) JWT (2/3 runs 선택)
  B) database (1/3 runs 선택)
  
어느 쪽을 사용할까요?
```

범위 발산의 경우:

```
[Step M] 발산 #2: 커스텀 에러 페이지
  run-1: 없음
  run-2: 생성됨
  run-3: 없음
  
  이 기능은 요청하지 않았습니다.
  A) 금지 (harness에 "절대 금지" 추가)
  B) 허용
```

### 1-7. 하네스 업데이트

사용자 결정을 하네스 규칙으로 변환한다.

| 발산 유형 | 하네스 위치 | 규칙 형태 |
|----------|-----------|----------|
| 구조 발산 | harness-step-M.md | "생성할 파일 목록"에 추가/제거 |
| 패턴 발산 | harness-step-M.md | "아키텍처 제약"에 규칙 추가 |
| 네이밍 발산 | harness-global.md | "코딩 규칙"에 컨벤션 추가 |
| 범위 발산 | harness-step-M.md 또는 harness-global.md | "절대 금지"에 항목 추가 |

주의:
- 기존 규칙을 삭제하지 않는다. 추가만 한다.
- 기존 규칙과 모순되면 사용자에게 알리고 결정받는다.

### 1-8. 재검증

하네스 업데이트 후 사용자에게 확인:

```
Step M 하네스를 업데이트했습니다. ({N}건 규칙 추가)
재실행하여 수렴을 확인할까요? (Y/건너뛰기)
```

- **Y**: run-* 디렉토리 정리 후 1-1로 돌아가 재실행
- **건너뛰기**: Phase 2로 이동 (수렴 미확인 상태로 다음 Step 진행)

## Phase 2: 골든 베이스 업데이트

Step M이 수렴(또는 하네스 업데이트 완료)하면, 수렴된 결과를 golden base에 반영한다.

```bash
# run-1의 Step M 변경분을 golden base에 적용
cd "$TESTBED_DIR/run-1"
git diff HEAD > /tmp/step-$M.patch
cd "$TESTBED_DIR/golden"
git apply /tmp/step-$M.patch
git add -A && git commit -m "calibrate: Step $M 수렴"
```

수렴 시 모든 run이 동일하므로 run-1을 대표로 사용한다.
발산 후 건너뛰기한 경우, 사용자 결정에 가장 부합하는 run을 선택한다.

run-* 디렉토리 정리:
```bash
for i in $(seq 1 $RUNS); do rm -rf "$TESTBED_DIR/run-$i"; done
```

이제 golden base에는 Step 1..M까지의 수렴 결과가 누적되어 있다.

**다음 Step이 있으면**: Step M+1로 Phase 1을 반복.

```
--- Step M 교정 완료. Step {M+1}: {이름} 으로 넘어갑니다. ---
```

**마지막 Step이면**: Phase 3으로.

## Phase 3: 최종 리포트 + cleanup

### 교정 리포트 생성

`.plan/$PLAN_NAME/calibration-report.md`에 전체 교정 결과를 기록:

```markdown
# 교정 리포트 — {PLAN_NAME}

실행 횟수: {N}회/Step | 날짜: {날짜}

## Step별 교정 결과

### Step 1: {이름}
- 수렴: ✅ (첫 실행에서 수렴)
- 발산: 0건
- 하네스 변경: 없음

### Step 2: {이름}
- 수렴: ✅ (재실행 1회 후 수렴)
- 발산: 2건 (패턴 1, 네이밍 1)
- 사용자 결정:
  - 세션 전략 → JWT
  - 설정 변수명 → authConfig
- 하네스 변경:
  - harness-step-2.md: 아키텍처 제약 1건 추가
  - harness-global.md: 코딩 규칙 1건 추가

### Step 3: {이름}
...

## 전체 요약

| Step | 초기 수렴 | 재실행 | 발산 | 결정 | 하네스 변경 |
|------|----------|--------|------|------|-----------|
| 1    | ✅       | 0      | 0    | 0    | 0         |
| 2    | ❌       | 1      | 2    | 2    | 2         |
| 3    | ✅       | 0      | 0    | 0    | 0         |

전체 하네스 변경: {N}건
```

교정 이력은 `.plan/$PLAN_NAME/calibration-history/`에 누적한다.

### testbed 정리

```bash
rm -rf "$TESTBED_DIR"
```

### 결과 안내

```
교정 완료:
  교정 Step: {from}~{last} ({N}개 Step)
  발산 총합: {N}건 발견, {M}건 해결
  하네스 변경: {파일 목록}
  리포트: .plan/{plan-name}/calibration-report.md

다음 단계:
  1. 변경된 하네스를 확인하세요
  2. bash plan.sh --dry 로 프롬프트를 확인하세요
  3. bash plan.sh 로 실제 실행하세요
```

모든 Step이 첫 실행에서 수렴한 경우:

```
전 Step 수렴! 하네스가 충분히 결정적입니다.
추가 교정이 필요하지 않습니다.
```
