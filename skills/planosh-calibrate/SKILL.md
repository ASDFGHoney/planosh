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

---

## 에이전트 아키텍처

교정 작업을 4개 역할로 분해한다. 이 스킬(orchestrator)이 나머지 3개 에이전트를 스폰하고 조율한다.

```
orchestrator (이 스킬)
  ├── testbed-run × N  (병렬) — 단일 testbed 실행
  ├── diff-collector          — 결과 수집 + 발산 분류
  └── harness-patch           — 하네스 규칙 작성
```

| 에이전트 | 분리 이유 |
|----------|----------|
| testbed-run | 병렬 스폰 단위. N개가 동시에 실행되어야 하므로 독립 에이전트 필수 |
| diff-collector | diff 양이 커서 orchestrator 컨텍스트 오염 방지. 비교 로직에 집중 |
| harness-patch | 하네스 수정은 정밀해야 함. 깨끗한 컨텍스트에서 수행 |

**별도 에이전트로 만들지 않는 것:**
- golden-updater: `git apply` + `git commit` 한 줄이라 오버헤드가 더 큼
- report-generator: orchestrator가 이미 모든 데이터를 보유
- 재검증: testbed-run을 재사용

---

## 에이전트 프롬프트 정의

아래 3개 프롬프트 템플릿을 Agent 도구 스폰 시 사용한다. `{변수}`는 런타임에 실제 값으로 치환한다.

### testbed-run 프롬프트

```
planosh calibrate testbed 실행기.

실행할 작업:
1. 디렉토리 이동 후 plan.sh를 **절대 경로**로 실행한다:
   cd {TESTBED_DIR}/run-{i}
   bash {PLAN_SH_ABS_PATH} --from={M} --to={M} --testbed

2. plan.sh가 내부적으로 claude -p를 호출한다.
   --testbed 플래그에 의해 --model haiku가 사용된다.
   이 에이전트는 실행과 보고만 담당한다.

3. 실행 완료 후 다음을 수집하여 최종 결과로 보고한다:
   a. git -C {TESTBED_DIR}/run-{i} diff --stat HEAD
   b. git -C {TESTBED_DIR}/run-{i} diff --name-only HEAD
   c. git -C {TESTBED_DIR}/run-{i} diff --shortstat HEAD
   d. 성공/실패 여부

4. 보고 형식 (이 형식을 정확히 따를 것):

   === TESTBED-RUN REPORT ===
   RUN: {i}
   STATUS: success 또는 fail
   FILES_CHANGED: {변경 파일 수}
   DIFF_STAT: {+추가 -삭제}
   CHANGED_PATHS:
   {변경 파일 경로, 한 줄에 하나}
   ERROR: {실패 시 에러 메시지, 성공 시 "none"}
   === END REPORT ===

5. 실패해도 에러 컨텍스트를 포함하여 보고한다. 스스로 재시도하지 않는다.
```

### diff-collector 프롬프트

```
planosh calibrate diff 수집기.

작업:
{TESTBED_DIR}의 run-1 ~ run-{N}에서 Step {M}의 diff를 수집하고 발산을 분류한다.

1. 각 run에서 수집:
   cd {TESTBED_DIR}/run-{i}
   git diff --name-only HEAD  →  변경 파일 목록
   git diff HEAD              →  전체 diff

2. 구조 발산 분석:
   모든 run의 파일 목록을 비교한다.
   - 모든 run에 존재 → 수렴
   - 일부 run에만 존재 → 구조 발산

3. 내용 발산 분석:
   모든 run에 존재하는 파일에 대해 run 간 diff 내용을 비교한다.
   발산 유형:
   - 패턴 발산: 같은 기능인데 다른 구현 방식 (예: JWT vs database 세션)
   - 네이밍 발산: 같은 개념인데 다른 이름 (예: authOptions vs authConfig)
   - 범위 발산: step 프롬프트에서 요청하지 않은 기능이 추가됨

4. 보고 형식 (이 형식을 정확히 따를 것):

   === DIFF-COLLECTOR REPORT ===

   ## 파일 존재 매트릭스

   | 파일 경로 | run-1 | run-2 | ... | 판정 |
   |-----------|-------|-------|-----|------|
   | path/file | O     | O     | O   | 수렴 |
   | path/file | O     | X     | O   | 발산 |

   ## 구조 발산

   {구조 발산 목록, 없으면 "없음"}
   - {파일}: run-{a}, run-{b}에 존재, run-{c}에 없음

   ## 내용 발산

   {내용 발산 목록, 없으면 "없음"}
   ### 발산 #{n}: {설명}
   유형: {패턴|네이밍|범위}
   파일: {파일 경로}
   - run-{a}: {요약}
   - run-{b}: {요약}

   ## 수렴 판정

   수렴: {true|false}
   구조 발산: {N}건
   내용 발산: {N}건 (패턴 {X}, 네이밍 {Y}, 범위 {Z})

   === END REPORT ===

5. 분류에 확신이 없으면 "유형: 불확실 — {이유}" 로 표시한다.
   orchestrator가 사용자에게 확인받는다.
```

### harness-patch 프롬프트

```
planosh calibrate 하네스 규칙 작성기.

작업:
사용자 결정을 하네스 규칙으로 변환하여 파일을 수정한다.

사용자 결정:
{DECISIONS_JSON}

변환 규칙:
| 발산 유형 | 기록 위치 | 규칙 형태 |
|----------|-----------|----------|
| 구조 발산 | {PLAN_DIR}/steps/step-{M}.md | ## 생성할 파일 목록 섹션에 허용 파일 나열 |
| 패턴 발산 | {PLAN_DIR}/steps/step-{M}.md | ## 아키텍처 제약 섹션에 규칙 추가 |
| 네이밍 발산 | {PLAN_DIR}/harness-for-plan.md | 코딩 규칙 섹션에 컨벤션 추가 |
| 범위 발산 | {PLAN_DIR}/steps/step-{M}.md 또는 {PLAN_DIR}/harness-for-plan.md | 절대 금지 섹션에 항목 추가 |

파일 수정 규칙:
1. step 프롬프트 파일 ({PLAN_DIR}/steps/step-{M}.md):
   - <!-- calibrate에 의해 추가됨 --> 마커가 있으면 그 아래에 추가
   - 마커가 없으면 파일 끝에 빈 줄 + 마커를 추가한 후 규칙 작성
   - 이미 마커 아래에 기존 calibrate 내용이 있으면 기존 내용에 병합 (기존 규칙 유지 + 신규 추가)
   - 마커 위의 기존 프롬프트(만들 것, 하지 않을 것)는 절대 수정하지 않는다

2. harness-for-plan.md ({PLAN_DIR}/harness-for-plan.md):
   - ## 코딩 규칙 섹션에 네이밍 규칙 추가
   - ## 절대 금지 섹션에 범위 금지 항목 추가
   - 기존 규칙과 모순되는지 검사

3. 보고 형식:

   === HARNESS-PATCH REPORT ===

   CHANGES:
   - {파일 경로}: {변경 요약}

   CONFLICTS:
   - {모순이 있으면 설명, 없으면 "없음"}

   === END REPORT ===
```

---

## 전제 조건

- `.plan/{plan-name}/plan.sh`가 존재해야 한다 (plan-name은 사용자가 지정하거나, `.plan/` 안에 하나만 있으면 자동 감지)
- plan.sh가 `--testbed` 플래그를 지원해야 한다 (v0.3.1+ plan.sh 템플릿)
- `--from=M`(M≥2)을 사용할 경우, Step M-1까지 완료된 커밋이 현재 브랜치에 있어야 한다
- v0 제약: 포트나 DB를 사용하는 verify가 있으면 병렬 실행 시 충돌할 수 있다. 빌드/파일 검증만 있는 plan.sh에서 가장 안전하다.

---

## Phase 0: 사전 검증 + Step 파싱 (orchestrator 직접 수행)

### 0-1. plan.sh 검증

사용자가 plan-name을 지정하지 않으면 `.plan/` 안에 디렉토리가 하나뿐일 때 자동 감지한다.

```bash
# plan-name 결정: 사용자 지정 or 자동 감지
if [ -z "$PLAN_NAME" ]; then
  PLAN_DIRS=($(ls -d .plan/*/  2>/dev/null))
  [ ${#PLAN_DIRS[@]} -eq 1 ] && PLAN_NAME=$(basename "${PLAN_DIRS[0]}")
  [ -z "$PLAN_NAME" ] && { echo ".plan/ 안에 plan이 여러 개입니다. --plan=이름을 지정하세요."; exit 1; }
fi
PLAN_DIR=".plan/$PLAN_NAME"
PLAN_SH="$PLAN_DIR/plan.sh"
[ -f "$PLAN_SH" ] || { echo "$PLAN_SH를 찾을 수 없습니다."; exit 1; }
[ -d "$PLAN_DIR" ] || { echo "$PLAN_DIR/ 디렉토리를 찾을 수 없습니다."; exit 1; }
```

### 0-2. Step 목록 추출

`.plan/$PLAN_NAME/steps.json`에서 Step 목록을 읽는다.

```bash
STEPS_FILE="$PLAN_DIR/steps.json"
[ -f "$STEPS_FILE" ] || { echo "$STEPS_FILE을 찾을 수 없습니다."; exit 1; }
```

파싱 결과를 사용자에게 보여준다:

```
.plan/{plan-name}/steps.json에서 N개 Step을 감지했습니다:
  Step 1: 프로젝트 스캐폴딩
  Step 2: DB 스키마 + API
  Step 3: 인증
  ...

Step {from}부터 순차 교정을 시작합니다.
```

### 0-3. --testbed 플래그 확인

plan.sh가 `--testbed` 플래그를 지원하는지 확인한다:

```bash
grep -q 'testbed' "$PLAN_SH"
```

미지원 시 사용자에게 안내:

```
plan.sh에 --testbed 플래그가 없습니다.
/planosh로 plan.sh를 재생성하거나, 수동으로 --testbed 지원을 추가하세요.

--testbed 없이 진행하면 모든 claude -p 호출이 기본 모델(Opus)로 실행됩니다.
calibrate 목적에는 Haiku로 충분하며, 비용이 크게 절감됩니다.

--testbed 없이 계속할까요? (Y/중단)
```

### 0-4. 골든 베이스 초기화

`.plan/$PLAN_NAME/testbed/golden`에 현재 repo의 clone을 생성한다. 골든 베이스는 수렴된 각 Step의 결과를 누적하여, 다음 Step 교정의 출발점이 된다.

```bash
TESTBED_DIR="$PLAN_DIR/testbed"
REPO_ROOT=$(git rev-parse --show-toplevel)
rm -rf "$TESTBED_DIR"
mkdir -p "$TESTBED_DIR"
git clone --depth 1 "file://$REPO_ROOT" "$TESTBED_DIR/golden"
```

`--from=M`(M≥2)인 경우, 현재 브랜치에 Step 1..M-1의 커밋이 있으므로 golden base에도 그 상태가 반영된다.

`testbed/`는 `.gitignore`에 추가:
```bash
grep -q 'testbed/' "$PLAN_DIR/.gitignore" 2>/dev/null || echo "testbed/" >> "$PLAN_DIR/.gitignore"
```

---

## Phase 1: Step M 교정 (각 Step에 대해 반복)

Step `--from`부터 마지막 Step까지 순차적으로 진행한다.

```
1-1. testbed 생성 (orchestrator)      <--+
  |                                       |
1-2. 병렬 실행 (testbed-run × N)        |
  |                                       |
1-3~4. 수집 + 분석 (diff-collector)     |
  |                                       |
1-5. 수렴 판정 (orchestrator)           |
  |- 수렴 → Phase 2로                    |
  |- 발산 → 1-6으로                      |
  |                                       |
1-6. 사용자 결정 (orchestrator)          |
  |                                       |
1-7. 하네스 업데이트 (harness-patch)     |
  |                                       |
1-8. 재검증 ---------------------------------+
```

### 1-1. testbed 생성 (orchestrator 직접)

golden base에서 N개 clone을 생성한다.

```bash
for i in $(seq 1 $RUNS); do
  cp -r "$TESTBED_DIR/golden" "$TESTBED_DIR/run-$i"
done
```

`cp -r` 사용: golden base는 이미 로컬 clone이므로, 다시 git clone보다 빠르다.

### 1-2. 병렬 실행 (testbed-run × N 스폰)

N개의 `testbed-run` 에이전트를 **하나의 메시지에서 동시에** 스폰한다.

**스폰 방법:**

Agent 도구를 N번 호출한다 (하나의 메시지에서 병렬). 각 에이전트에 `run_in_background: true`를 설정한다.

```
Agent({
  description: "testbed run-1 Step M",
  name: "run-1",
  prompt: "{testbed-run 프롬프트, 변수 치환 완료}",
  run_in_background: true
})

Agent({
  description: "testbed run-2 Step M",
  name: "run-2",
  prompt: "{testbed-run 프롬프트, 변수 치환 완료}",
  run_in_background: true
})

... (N개)
```

**변수 치환 규칙:**
- `{TESTBED_DIR}` → 실제 testbed 절대 경로
- `{PLAN_SH_ABS_PATH}` → **testbed run 디렉토리 안의** plan.sh 절대 경로. 예: `{TESTBED_DIR}/run-1/.plan/myplan/plan.sh`. 주 plan.sh가 아니라 testbed에 복사된 plan.sh를 가리켜야 한다 — 그래야 `SCRIPT_DIR`이 testbed 안의 `.plan/`으로 해석되고, `PROJECT_ROOT`가 testbed run 디렉토리로 올바르게 계산된다.
- `{M}` → 현재 Step 번호
- `{i}` → run 번호 (1, 2, ..., N)

**진행 상황 표시:**

에이전트가 완료될 때마다 사용자에게 진행 상황을 표시한다:

```
Step {M} 교정 ({N}회 실행)
  run-1: +347 -12 (8파일)
  run-2: +198 -8 (7파일)
  run-3: (진행 중)
```

모든 run이 완료되면 전체 결과를 표시:

```
Step {M}: 전체 실행 완료
  run-1: +347 -12 (8파일)
  run-2: +198 -8 (7파일)
  run-3: +201 -10 (7파일)
```

실패한 run은 제외하고 나머지로 분석 진행. **전체 실패 시 에러 보고 후 중단.**

### 1-3~4. 결과 수집 + 발산 분석 (diff-collector 스폰)

`diff-collector` 에이전트를 스폰한다 (foreground, 결과 대기):

```
Agent({
  description: "diff-collector Step M",
  name: "diff-collector",
  prompt: "{diff-collector 프롬프트, 변수 치환 완료}"
})
```

**변수 치환 규칙:**
- `{TESTBED_DIR}` → 실제 testbed 절대 경로
- `{N}` → 성공한 run 수 (실패한 run 제외하여 run 번호 목록도 명시)
- `{M}` → 현재 Step 번호

**결과 처리:**

diff-collector의 `DIFF-COLLECTOR REPORT`를 파싱하여 사용자에게 표시한다.

수렴 시:
```
Step {M}: {이름} — 수렴 ({N}회 실행 모두 동일)
```
→ Phase 2로 이동.

발산 시:
```
Step {M}: {이름} — {N}건 발산

{파일 존재 매트릭스}

{내용 발산 목록}
```

**AI 분류 확인:** diff-collector가 분류한 발산 유형을 사용자에게 보여주고 확인받는다.

```
발산 분류 결과:
  #1: 세션 전략 — 패턴 발산
  #2: 설정 변수명 — 네이밍 발산

이 분류가 맞습니까? (Y/수정할 번호)
```

분류가 틀렸으면 사용자가 수정한다.

### 1-5. 수렴 판정 (orchestrator 직접)

**수렴**: 구조 발산 0건, 내용 발산 0건.

```
Step {M}: {이름} — 수렴 ({RUNS}회 실행 모두 동일)
```

→ Phase 2로 이동.

**발산**: 1건 이상.

```
Step {M}: {이름} — {N}건 발산
하나씩 결정합니다.
```

→ 1-6으로 이동.

### 1-6. 사용자 결정 요청 (orchestrator 직접)

각 발산을 사용자에게 제시하고 결정을 받는다.

#### 제시 규칙

1. **발산 원인 한 줄** — Step 프롬프트의 어떤 모호함이 발산을 유발했는지 명시
2. **선택지별 한 줄 설명** — 이름만이 아니라 "뭘 하는 건지" 알 수 있게
3. **추천 + 이유** — 다수결, PRD 맥락, 기술적 트레이드오프 중 가장 강한 근거로 추천

#### 패턴 발산 형식

```
[Step M] 발산 #1: 세션 전략

  원인: Step 프롬프트에 세션 저장 방식 미지정

  A) JWT — 토큰 기반 무상태, 취소 어려움 (2/3 runs)
  B) database — DB 저장, 취소 용이, DB 부하 (1/3 runs)

  추천: A
  이유: PRD에 세션 취소 요구 없음 + 다수 선택.
```

#### 네이밍 발산 형식

```
[Step M] 발산 #2: 설정 변수명

  원인: NextAuth 설정 export명 미지정

  A) authOptions — NextAuth 공식 문서 패턴 (2/3 runs)
  B) authConfig — 일반적 config 네이밍 (1/3 runs)

  추천: A
  이유: 공식 문서와 일치, 검색 시 레퍼런스 찾기 용이.
```

#### 범위 발산 형식

```
[Step M] 발산 #3: 커스텀 에러 페이지

  내용: app/error.tsx — 전역 에러 바운더리 UI 컴포넌트
  PRD 근거: 에러 처리 관련 언급 없음 (1/3 runs만 추가)

  A) 금지 — harness "절대 금지"에 추가
  B) 허용 — Step 범위에 포함

  추천: A
  이유: PRD 범위 외 기능, 결정성 저하.
```

모든 발산에 대해 결정을 받은 후, 결정 목록을 JSON으로 구성한다:

```json
[
  {
    "divergence_id": 1,
    "type": "pattern",
    "description": "세션 전략",
    "decision": "JWT",
    "file": "src/lib/auth.ts",
    "step": 3
  },
  {
    "divergence_id": 2,
    "type": "scope",
    "description": "커스텀 에러 페이지",
    "decision": "금지",
    "file": "src/app/error.tsx",
    "step": 3
  }
]
```

### 1-7. 하네스 업데이트 (harness-patch 스폰)

`harness-patch` 에이전트를 스폰한다 (foreground, 결과 대기):

```
Agent({
  description: "harness-patch Step M",
  name: "harness-patch",
  prompt: "{harness-patch 프롬프트, 변수 치환 완료}"
})
```

**변수 치환 규칙:**
- `{DECISIONS_JSON}` → 1-6에서 구성한 사용자 결정 목록 JSON
- `{PLAN_DIR}` → .plan/{plan-name} 절대 경로
- `{M}` → 현재 Step 번호

**결과 처리:**

harness-patch의 `HARNESS-PATCH REPORT`를 사용자에게 표시한다:

```
하네스 업데이트 완료:
  - steps/step-{M}.md: 아키텍처 제약 1건, 파일 목록 추가
  - harness-for-plan.md: 코딩 규칙 1건 추가
```

CONFLICTS가 있으면:

```
기존 규칙과 충돌이 감지되었습니다:
  - {충돌 설명}
어떻게 처리할까요? (기존 유지/새 규칙 우선/직접 수정)
```

충돌 처리 후 harness-patch에 SendMessage로 수정 지시를 보내거나, orchestrator가 직접 수정한다.

### 1-8. 재검증 (orchestrator 직접)

하네스 업데이트 후 사용자에게 확인:

```
Step {M} 하네스를 업데이트했습니다. ({N}건 규칙 추가)
  수정: {수정된 파일 목록}
재실행하여 수렴을 확인할까요? (Y/건너뛰기)
```

- **Y**: run-* 디렉토리 정리 후 1-1로 돌아가 재실행. 동일한 testbed-run 에이전트를 다시 스폰한다.
- **건너뛰기**: Phase 2로 이동 (수렴 미확인 상태로 다음 Step 진행)

---

## Phase 2: 골든 베이스 업데이트 (orchestrator 직접)

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

---

## Phase 3: 최종 리포트 + cleanup (orchestrator 직접)

### 교정 리포트 생성

`.plan/$PLAN_NAME/calibration-report.md`에 전체 교정 결과를 기록:

```markdown
# 교정 리포트 — {PLAN_NAME}

실행 횟수: {N}회/Step | 모델: haiku (--testbed) | 날짜: {날짜}

## Step별 교정 결과

### Step 1: {이름}
- 수렴: 첫 실행에서 수렴
- 발산: 0건
- 하네스 변경: 없음

### Step 2: {이름}
- 수렴: 재실행 1회 후 수렴
- 발산: 2건 (패턴 1, 네이밍 1)
- 사용자 결정:
  - 세션 전략 → JWT
  - 설정 변수명 → authConfig
- 변경된 파일:
  - steps/step-2.md: 아키텍처 제약 1건 추가
  - harness-for-plan.md: 코딩 규칙 1건 추가

### Step 3: {이름}
...

## 전체 요약

| Step | 초기 수렴 | 재실행 | 발산 | 결정 | 하네스 변경 |
|------|----------|--------|------|------|-----------|
| 1    | Y        | 0      | 0    | 0    | 0         |
| 2    | N        | 1      | 2    | 2    | 2         |
| 3    | Y        | 0      | 0    | 0    | 0         |

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
  모델: haiku (--testbed 모드)
  변경된 파일: {파일 목록}
  리포트: .plan/{plan-name}/calibration-report.md

다음 단계:
  1. 변경된 steps/*.md를 확인하세요 (calibrate 마커 이후 추가된 제약)
  2. bash .plan/{plan-name}/plan.sh --dry 로 프롬프트를 확인하세요
  3. bash .plan/{plan-name}/plan.sh 로 실제 실행하세요
```

모든 Step이 첫 실행에서 수렴한 경우:

```
전 Step 수렴! 프롬프트가 충분히 결정적입니다.
추가 교정이 필요하지 않습니다.
```
