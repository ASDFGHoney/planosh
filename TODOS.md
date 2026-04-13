# TODOS

## TODO-001: --append-system-prompt lost-in-the-middle 검증

- **What:** 하네스가 기존 시스템 프롬프트 뒤에 append될 때 실제로 우선순위가 높은지 실험으로 검증
- **Why:** 3계층 모델의 Layer 1 (시스템 프롬프트)이 설계 의도대로 작동하는지 근거 확보. 외부 리뷰에서 lost-in-the-middle 위험 지적됨.
- **Pros:** 근거 기반 설계. calibrate 결과가 하네스 효과를 간접적으로 보여줌.
- **Cons:** calibrate가 이미 하네스 효과를 측정하므로 별도 검증이 중복될 수 있음.
- **Context:** `/planosh-calibrate`를 실제 실행하면 하네스 유무에 따른 수렴률 차이로 간접 검증 가능. 하지만 "하네스가 진짜 시스템 프롬프트처럼 작동하는가"는 별도 실험이 필요할 수 있음.
- **Depends on:** /planosh-calibrate 스킬 구현 후

## TODO-002: 커뮤니티 초기 기여자 확보 전략 (GTM)

- **What:** 오픈소스 출시 후 첫 외부 기여(PR 또는 Discussion)를 이끌어내는 전략
- **Why:** 커뮤니티 성장 플라이휠의 첫 바퀴. 스킬과 README만으로 자발적 기여자가 올지 불확실.
- **Pros:** 전략이 있으면 launch 후 방치 방지. 측정 가능한 목표 설정 가능.
- **Cons:** 제품이 좋으면 사람은 온다는 관점도 유효.
- **Context:** 외부 리뷰에서 "닭과 달걀 문제"로 지적됨. 가능한 전략: 자체 dogfooding 사례를 .plan/에 먼저 채우기, Twitter/HN에 데모 GIF 공유, AI 개발 커뮤니티에 직접 소개.
- **Depends on:** 스킬 구현 + seed 예제 완성 후

## TODO-003: /planosh-patterns — GitHub 레포에서 best practice 검색 및 패턴 제안

- **What:** 커뮤니티가 공유한 examples, discussions, issues에서 사용자의 스택/기능에 맞는 패턴을 검색하여 하네스 규칙, verify 패턴, 발산 봉쇄법을 제안하는 스킬
- **Why:** 커뮤니티 사례가 쌓이면, 새 plan.sh를 만들 때 검증된 하네스 규칙을 처음부터 적용하여 calibrate 반복 횟수를 줄일 수 있다. 커뮤니티 기여 → 검색 → 재사용 플라이휠의 핵심.
- **Pros:** calibrate 전에 이미 검증된 규칙으로 하네스를 보강하면 초기 수렴률이 높아진다. 커뮤니티 기여에 대한 인센티브가 된다.
- **Cons:** examples가 충분히 쌓이기 전까지는 검색 결과가 빈약할 수 있다. gh CLI 의존성.
- **Scope:**
  - `gh` CLI로 레포의 .plan/ 하네스 코드, Discussions 패턴 논의, Issues 발산 보고를 검색
  - 쿼리 유형: 스택 기반 ("nextjs auth"), 발산 기반 ("네이밍 발산"), 패턴 기반 ("verify 패턴"), 탐색 ("수렴률 높은 예제")
  - 현재 프로젝트에 plan.sh가 있으면 추천 규칙을 하네스에 바로 적용 가능
  - `/planosh` Phase 2에서 선택적으로 패턴 검색을 제안하는 연동
  - 추천 라벨 체계: `divergence:naming`, `stack:nextjs`, `convergence:100` 등
- **Depends on:** .plan/ 커뮤니티 기여가 일정 수준 이상 쌓인 후

## TODO-004: `planosh` CLI — 플랫폼별 스킬 설치 + 프로젝트별 커스텀 + 선택적 업데이트

- **Status:** 🚧 진행 중 — 요구사항 정리 완료 (`docs/CLI-REQUIREMENTS.md`), TODO-006/008/009와 통합 설계
- **What:** CLI 도구(`planosh`)를 만들어서 (1) 원하는 플랫폼(Claude Code, Cursor, Windsurf 등)의 스킬 형태로 설치, (2) 프로젝트별로 스킬을 커스텀, (3) 오픈소스 업스트림 변경사항을 plan.sh 방식으로 선택적 적용
- **Why:** 현재 planosh는 Claude Code 플러그인으로만 배포된다. 플랫폼마다 스킬/룰 포맷이 다르고(SKILL.md, .cursorrules, .windsurfrules 등), 프로젝트마다 하네스 전략이 다르다. 설치 후 커스텀한 스킬에 업스트림 변경이 오면 전부 덮어쓸 수도 없고 무시할 수도 없다 — 선택적 머지가 필요하다.
- **Scope:**
  - **설치:** `planosh install --platform=claude-code` → 해당 플랫폼 포맷으로 스킬 파일 생성. 플랫폼 어댑터 패턴.
  - **커스텀:** 설치된 스킬은 프로젝트 로컬에 복사되어 자유롭게 수정 가능. 원본 버전을 `.planosh/upstream/`에 보존하여 diff 기준점 유지.
  - **업데이트:** `planosh update` → 업스트림 변경사항을 3-way diff로 보여주고, 변경 단위(섹션/Phase/규칙)별로 적용 여부를 사용자가 결정. plan.sh의 "결정성" 원칙 그대로 — AI가 머지하지 않고, 사용자가 각 변경의 적용 여부를 결정한다.
  - **지원 플랫폼 (초기):** Claude Code (SKILL.md), Cursor (.cursorrules), Windsurf (.windsurfrules)
  - **버전 추적:** `.planosh/manifest.json`에 설치된 스킬별 업스트림 버전 + 커스텀 여부 기록
- **Pros:** 플랫폼 종속 탈피. 프로젝트별 커스텀이 업데이트에 안전. planosh 철학(결정성, 사용자 결정)이 배포/업데이트에도 일관되게 적용.
- **Cons:** 플랫폼 어댑터 유지보수 부담. 3-way diff 구현 복잡도. 플랫폼별 스킬 포맷이 바뀌면 어댑터도 갱신 필요.
- **Depends on:** 스킬이 안정화된 후 (v0.3.0+). 플랫폼별 스킬 포맷 리서치 선행.

## TODO-005: Testbed/Calibrate 실행 진행과정 가시성

- **Phase 1 완료** (2026-04-13): 에이전트 4개 분해 + --testbed 플래그
  - `skills/planosh/SKILL.md`: plan.sh 템플릿에 `--testbed` 플래그 추가 (Haiku 모델 사용)
  - `skills/planosh-calibrate/SKILL.md`: orchestrator + testbed-run + diff-collector + harness-patch 에이전트 분해 아키텍처로 재작성
  - 에이전트 프롬프트 템플릿 3개 정의 (구조화된 보고 형식 포함)
  - 병렬 스폰 패턴 (Agent × N, run_in_background) + 진행 상황 표시 포맷
- **What:** calibrate나 testbed 실행 중 진행 상황을 실시간으로 보여주는 메커니즘. 코드 변경사항 통계(`+1234 -234`), 현재 Step, 수렴/발산 상태 등을 스트리밍으로 표시.
- **Why:** 현재 calibrate는 에이전트를 스폰해서 진행하는데, 내부에서 뭘 하고 있는지 전혀 보이지 않음. 사용자가 "돌아가고 있긴 한 건가?" 상태로 대기해야 하는 것은 UX 치명적.
- **표시 항목 (안):**
  - 현재 실행 중인 Step 번호/이름
  - 코드 변경 통계: `+1234 -234` (git diff --stat 스타일, 실시간 증가)
  - 하네스 규칙 추가/수정 카운트
  - 수렴률 변화: `72% → 85% → 100%`
  - 소요 시간 / 반복 횟수
- **Phase 1 (CLI 이전) — 에이전트 4개 분해:**

  ### 1. `planosh-testbed-run` — 단일 testbed 실행기
  - **역할:** testbed/run-i 에서 plan.sh --from=M --to=M **--testbed** 실행
  - **왜 분리:** 병렬 스폰 단위. N개가 동시에 떠야 하므로 독립 에이전트 필수
  - **입력:** testbed 경로, plan.sh 경로, Step 번호
  - **출력:** 실행 결과 + 변경 통계
  - **중간 보고 (SendMessage → orchestrator):**
    - "run-2: claude 호출 시작"
    - "run-2: +347 -12 (src/lib/ 4파일)"
    - "run-2: verify 2/3 통과"
    - "run-2: checkpoint 완료"
  - 실패 시 에러 컨텍스트 전달

  ### plan.sh `--testbed` 모드
  - **What:** plan.sh에 `--testbed` 플래그를 추가하여 `run_claude`의 `claude -p` 호출 시 `--model haiku` 사용
  - **Why:** calibrate의 목적은 발산 패턴 탐지이지 최고 품질 코드 생성이 아님. N회 × M스텝 × claude 호출이 전부 Opus로 돌면 비용이 과도. Haiku로도 발산 패턴은 동일하게 드러남.
  - **근거:** 하이쿠에서 수렴하면 → Opus는 당연히 수렴. 하이쿠에서 발산하면 → 하네스 부족 신호. 탐지 목적에 정확히 부합.
  - **구현 (안):**
    ```bash
    # plan.sh 상단
    TESTBED=false
    # 플래그 파싱에 추가
    --testbed) TESTBED=true ;;
    
    # run_claude 함수 내
    MODEL_FLAG=""
    $TESTBED && MODEL_FLAG="--model haiku"
    
    claude -p "$prompt" \
      $MODEL_FLAG \
      --append-system-prompt "$harness" \
      --dangerously-skip-permissions
    ```
  - **testbed-run 에이전트가 하는 일:** `bash plan.sh --from=M --to=M --testbed` 호출. 에이전트 자체 모델은 기본(Opus/Sonnet) 유지 — 판단력이 아니라 실행+보고 역할이므로 에이전트 모델은 무관.

  ### 2. `planosh-diff-collector` — 결과 수집 + 발산 분류
  - **역할:** 각 run의 diff를 수집하고 발산을 AI 분류
  - **왜 분리:** 비교 로직이 무거움. run이 많을수록 diff 양이 커서 orchestrator 컨텍스트 오염 방지
  - **입력:** testbed/run-1..N 경로
  - **출력:** 파일 존재 매트릭스 + 발산 분류 테이블 (패턴/네이밍/범위) + 수렴 판정
  - **중간 보고:**
    - "3개 run 수집 완료, 파일 12개 비교 중"
    - "발산 2건 감지: 패턴 1, 네이밍 1"

  ### 3. `planosh-harness-patch` — 하네스 규칙 작성기
  - **역할:** 사용자 결정을 하네스 규칙으로 변환하여 파일 수정
  - **왜 분리:** 하네스 수정은 정밀해야 하고, orchestrator 컨텍스트가 비교 결과로 이미 커져있을 때 깨끗한 컨텍스트에서 수행
  - **입력:** 사용자 결정 목록 + 현재 하네스 파일들
  - **출력:** 수정된 하네스 파일들 + 변경 diff 리포트
  - 기존 규칙과 충돌 검사 포함

  ### 4. `planosh-calibrate-orchestrator` — 전체 조율 (기존 calibrate 스킬 리팩토링)
  - **역할:** 위 3개 에이전트를 스폰하고 조율. 사용자 대화 담당.
  - **직접 수행:** Phase 0 (plan.sh 파싱, golden 초기화), testbed 생성, 사용자 결정 대화, golden 업데이트, 리포트 생성
  - **에이전트 스폰:**
    - Phase 1-2: `planosh-testbed-run` × N 병렬 스폰 → 각 run에서 SendMessage로 progress 수신
    - Phase 1-3~4: `planosh-diff-collector` 스폰
    - Phase 1-7: `planosh-harness-patch` 스폰
  - **사용자에게 종합 표시:**
    ```
    ┌─────────────────────────────────┐
    │ Step 3 교정 (3회 실행)          │
    │ run-1: ✅ +347 -12  2m 13s      │
    │ run-2: ⏳ +198 -8   (진행 중)   │
    │ run-3: ⏳ claude 호출 중         │
    └─────────────────────────────────┘
    ```

  ### 별도 에이전트로 안 만드는 것
  - **golden-updater:** patch apply + git commit 한 줄이라 에이전트 오버헤드가 더 큼
  - **report-generator:** orchestrator가 이미 모든 데이터를 갖고 있으므로 직접 작성이 효율적
  - **재검증 전용:** testbed-run을 재사용하면 됨

- **Phase 2 (CLI):** `planosh` CLI에서 터미널 TUI로 실시간 progress bar, diff stat 스트리밍 제공. `planosh calibrate --progress` 같은 플래그.
  - CLI TUI 프레임워크 선택: ink (React for CLI)? blessed? 직접 ANSI?
- **Pros:** 사용자 신뢰도 증가. 문제 발생 시 어디서 멈췄는지 즉시 파악 가능. calibrate 도중 방향 수정 가능.
- **Cons:** 에이전트 스킬 방식은 SendMessage 오버헤드. CLI TUI는 구현 복잡도.
- **Depends on:** TODO-004 (CLI)와 병렬 진행 가능. Phase 1은 지금 바로 시작 가능.

## TODO-006: `planosh run` — 세션 내 N-plan.sh 병렬 실행과 shim-git testbed

> GitHub Issue #1에서 이관 (2026-04-13)

- **Status:** 🚧 진행 중 — TODO-004 CLI에 통합. testbed 위치 `~/.planosh/testbed/`로 확정 (`docs/CLI-REQUIREMENTS.md`)
- **What:** Claude 세션 안에서 N개 plan.sh를 결정적으로 병렬 실행할 수 있는 CLI 환경(`planosh run`) + 각 testbed에서 진짜 git 대신 shim-git으로 격리하는 구조.
- **Why:** Claude의 bash tool call은 stateless라 CWD 미보존·타임아웃·병렬 상태 공유 불가. 드라이버 역할을 Claude가 아닌 CLI가 해야 한다. 또한 병렬 worker가 같은 repo에서 git을 쓰면 index lock·ref 경합·원격 오염이 발생. calibrate에 진짜 git은 필요 없고 파일트리 diff면 충분 — shim-git이 이 깨달음의 구현체.
- **구조:**
  - **계층 1 — `planosh run` CLI:** 단일 명령으로 N-병렬 plan.sh 실행. testbed 디렉토리 생성, shim PATH 주입, bash spawn, 로그 수집, 결과 집계. 출력은 구조화된 JSON 요약.
  - **계층 2 — shim-git:** 각 testbed의 PATH 앞에 git shim 삽입. `git add/commit` → hardlink 스냅샷, `git reset --hard` → 스냅샷 rsync, `git push/fetch` → no-op (원격 오염 차단), 미지원 옵션 → 조기 실패.
- **모드 분리:**
  - `--mode calibrate`: shim-git 기반 발산 측정 (비교 목적)
  - `--mode race`: real git + real remote 경쟁 착륙 (push 목적)
  - `--mode split` (향후): shim-git + POSIX claim 협력 분할 실행
- **열린 질문:** shim-git API 경계(최소 호환 범위), CLI 구현 언어(bash vs Rust/Go), 스냅샷 COW 안전성(hardlink vs reflink vs cp -a), 장시간 실행 시 상태 폴링 패턴, 결정성 자체 검증(shim self-calibration)
- **산출물:**
  - [ ] `planosh run` CLI 스펙 초안
  - [ ] shim-git 최소 API 정의 (지원 / no-op / 조기실패 3분류)
  - [ ] shim-git self-calibration 테스트
  - [ ] 프로토타입: 2-워커 calibrate, 단일 호스트
  - [ ] 하네스 문서에 "shim-safe git subset" 섹션 추가
- **Depends on:** DESIGN.md 원칙 1(가독성) + 원칙 3(결정성). TODO-004(CLI)와 연관.

## TODO-007: plan.sh 내부 병렬 실행 — 낙관적 실행 + 수렴 재시도

> GitHub Issue #2에서 이관 (2026-04-13)

- **What:** 단일 plan.sh 안의 독립적인 Step들을 병렬로 실행하여 총 실행 시간을 줄이되, 결정성을 깨지 않는 메커니즘.
- **Why:** 실전 dogfooding에서 16시간(Flutter→RN 마이그레이션) 사례 발생. 많은 Step이 실제로 독립적(DAG)인데 순차 실행이 강제됨. 피드백 루프 속도를 올려야 calibration 사이클이 빨라진다.
- **방향 — 낙관적 실행 + 수렴 재시도:**
  1. **격리 스냅샷 실행** — 병렬 Step을 각각 격리된 환경에서 실행
  2. **낙관적 병합** — 완료 후 순서대로 본 트리에 합침
  3. **충돌 시 재실행** — 지는 Step을 이긴 Step 결과 위에서 재실행 (컨텍스트 주입)
  4. **stuck 감지** — 같은 Step이 N회 연속 재시도 실패 시 순차로 강등 + 리포트
- **plan.sh에서의 표현:**
  ```bash
  planosh parallel --name "auth" \
    --step "Step 2: auth 백엔드" \
    --step "Step 3: 설정 페이지 UI" \
    --max-retries 3 \
    --on-stuck "serial"
  ```
- **중첩 병렬 문제 (calibrate × split):** calibrate N=3, split M=4이면 동시 12개 AI 세션 + 재시도. 자원 폭발, 격리 레이어 합성, 수렴 재시도가 calibrate 발산 시그널을 오염시키는 문제가 핵심. 재시도 경로의 결정성 보장과 calibrate의 발산 분류("경로 의존 발산" vs "plan 발산" 분리)가 필요.
- **열린 질문:** 격리 메커니즘(shim-git 재사용 vs worktree), 재실행 결정성, verify 경계(개별 vs 그룹), checkpoint 단위, stuck 파라미터, DRY 모드 시각화, 중첩 병렬 결정성 보장
- **산출물:**
  - [ ] `planosh parallel` CLI 스펙 초안
  - [ ] 격리 메커니즘 결정 — TODO-006과 합의
  - [ ] 수렴 재시도 루프 + stuck 감지 구현
  - [ ] 병렬 그룹 포함 plan.sh 예제 1개 (16시간 → 단축 측정)
  - [ ] calibrate 모드에서 병렬 그룹 결정성 수렴 검증
  - [ ] 중첩 병렬 시나리오 동작 검증
  - [ ] 재시도 경로가 calibrate 발산 시그널에 미치는 영향 측정
- **Depends on:** TODO-006 (`planosh run` + shim-git). DESIGN.md 규칙 2(수렴 루프).

## TODO-008: CLI 구조-중립 설계 — 프로젝트 레이아웃에 의존하지 않는 `.plan/` 발견 규칙

- **Status:** 🚧 진행 중 — TODO-004 CLI의 기반 모듈로 통합 (`docs/CLI-REQUIREMENTS.md`)
- **What:** CLI가 single repo, monorepo, submodule, codespace 등 어떤 코드 관리 방식에서도 동일하게 작동하도록 `.plan/` 위치 발견 규칙과 PROJECT_ROOT 계산 방식을 확정.
- **Why:** 팀마다 코드 관리 방식이 전부 다르다. monorepo에서 `apps/web/.plan/`에 plan을 두는 팀, submodule 안에 `.plan/`을 넣는 팀, codespace에서 clone 후 harness로 감싸는 팀 등. CLI가 특정 구조를 가정하면 그 순간 범용성이 깨진다.
- **현재 문제:**
  - `PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"` — 2단계 상위로 올라가는 하드코딩. `.plan/`이 repo root에 있다는 가정.
  - verify 커맨드의 실행 디렉토리가 암묵적으로 repo root.
  - monorepo에서 package 단위 plan을 지원하지 못함.
- **설계 원칙: ".plan/의 부모가 곧 PROJECT_ROOT"**
  ```
  planosh CLI
    ├── 발견 계층 (Discovery)
    │   ├── .plan/ 위치를 cwd → git root 방향으로 탐색
    │   ├── 없으면 cwd 기준으로 생성
    │   └── PROJECT_ROOT = .plan/의 직접 부모 (항상 동적 계산)
    │
    ├── 실행 계층 (Execution)
    │   ├── cwd = PROJECT_ROOT로 고정
    │   ├── verify 커맨드는 PROJECT_ROOT에서 실행
    │   └── claude -p도 PROJECT_ROOT에서 실행
    │
    └── 격리 계층 (Isolation)
        ├── plan 산출물은 전부 .plan/{name}/ 안에
        ├── .plan-state도 .plan/{name}/ 안에
        └── repo의 다른 파일에 절대 기록하지 않음
  ```
- **plan.sh 변경점:**
  ```bash
  # AS-IS: 구조 가정 (2단계 상위)
  PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

  # TO-BE: .plan/의 부모가 곧 root
  PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
  ```
- **시나리오별 동작:**
  - monorepo `apps/web/.plan/auth-plan/` → PROJECT_ROOT = `apps/web/`
  - single repo `.plan/mvp/` → PROJECT_ROOT = repo root
  - submodule `libs/shared/.plan/refactor/` → PROJECT_ROOT = `libs/shared/`
  - codespace `/workspaces/myapp/.plan/sprint-3/` → 그냥 작동
- **Pros:** 구조 중립. 팀이 기존 방식을 바꿀 필요 없음. plan.sh 변경은 1줄.
- **Cons:** monorepo에서 여러 `.plan/`이 중첩될 경우 탐색 우선순위 정의 필요.
- **산출물:**
  - [ ] plan.sh 템플릿의 PROJECT_ROOT 계산 변경
  - [ ] CLI의 `.plan/` 발견 알고리즘 (cwd → 상위 탐색, 중첩 시 가장 가까운 것 우선)
  - [ ] 시나리오별 동작 검증 (monorepo, submodule, codespace)
  - [ ] `/planosh` 스킬의 plan.sh 생성 로직에 반영
- **Depends on:** TODO-004 (CLI). TODO-006과 연관 (testbed 격리 시에도 동일 규칙 적용).

## TODO-009: testbed 자기참조 문제 — testbed 경로를 plan 디렉토리 밖으로 분리

- **Status:** 🚧 진행 중 — `~/.planosh/testbed/`로 확정. TODO-004 CLI에서 해결 (`docs/CLI-REQUIREMENTS.md`)
- **What:** calibrate의 testbed 디렉토리가 `$PLAN_DIR/testbed` (= `.plan/<name>/testbed/`)에 생성되는데, plan 파일을 golden에 `cp -r`로 복사할 때 testbed 자체가 재귀적으로 포함되는 자기참조 문제.
- **Why:** `.plan/<name>/` 안에 testbed가 있으면 plan 파일을 golden에 복사할 때 `cp -r`이 testbed까지 딸려 보낸다. `.gitignore`에 testbed를 추가해도 `cp -r`은 gitignore를 무시하므로 해결 안 됨. 현재 calibrate SKILL.md에 `echo "testbed/" >> "$PLAN_DIR/.gitignore"` 코드가 있지만 git clone 단계에만 효과가 있고, 이후 cp -r 단계에서는 무력.
- **현상:** golden clone 내부의 `.plan/<name>/`에 `testbed/` 디렉토리가 잔류. 재귀 복사 중 깊은 경로에서 에러 발생.
- **해결 방안 비교:**
  1. `rsync --exclude=testbed` — 최소 변경이지만 증상 치료. 제외 대상이 늘면 exclude 누적.
  2. 선택적 cp (개별 파일 복사) — 외부 의존 없지만 plan 구조 변경 시 복사 목록 수동 갱신 필요.
  3. **testbed 경로를 plan 밖으로 분리** (추천) — `TESTBED_DIR=".testbed/$PLAN_NAME"` 등으로 변경. 자기참조 원천 차단. cp/rsync 어떤 방식이든 문제 없음.
- **변경 범위:** `skills/planosh-calibrate/SKILL.md`의 `TESTBED_DIR` 정의 + 참조 경로 전체. `.testbed/`를 root `.gitignore`에 추가.
- **Depends on:** 없음 (독립 수정 가능). TODO-006 (shim-git testbed)과 연관 — testbed 경로 규칙 통일 필요.
