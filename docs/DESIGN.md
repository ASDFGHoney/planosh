# planosh — AI 코딩 팀을 위한 결정적 실행 프레임워크

## 문제 정의

AI 코딩 도구가 팀 전체에 보급되면서, 이제 개발자뿐 아니라 기획자, 디자이너도 코드를 만들 수 있게 되었다. 하지만 결정적인 실행 계획 없이 각자가 AI에게 "알아서 해줘"를 시키면, 팀 전체의 코드베이스가 비결정적으로 변한다.

현장에서 실제로 일어나는 일:

> **개발자**: "그건 claude로 해봐야 알 것 같아요..ㅠㅠ"
>
> — 계획이 없으니 실행 전까지 결과를 예측할 수 없다. "해봐야 안다"가 개발 프로세스가 되어버렸다.

> **기획자**: "저 추가하고 싶은 기능이 있어서 한번 개발하고 PR 올려봤어요. 리뷰해주세용."
> → 파일 변경 40개, +4,567 -12,300
>
> — 의도는 좋지만, 결정적인 계획 없이 AI에게 통째로 맡기면 이런 PR이 나온다. 만약 이 기능의 실행 계획이 먼저 리뷰되었고, 그 계획이 결정적이었다면? 코드를 한 줄도 안 봐도 됐다. 계획이 곧 결과를 보장하니까.

> **3개월 후**: "이 인증 코드, 왜 이 구조로 만들었지?"
> → AI 세션은 이미 사라졌고, 커밋 메시지는 "feat: add auth"뿐이다. 왜 JWT를 골랐는지, 왜 파일을 이렇게 나눴는지, 아무도 모른다.
>
> — 계획이 리뷰되었다면, 모든 기술 결정과 그 이유가 `.plan/`에 남아있다.

> **기획자가 명세를 보고 AI로 기능을 만들었다. 동작한다.** 일주일 후, 같은 명세로 작은 수정을 시켰더니, AI가 전혀 다른 구조로 만들어버렸다. 기획자는 코드를 읽을 수 없으니 뭐가 달라졌는지 모른다.
>
> — 명세는 있었다. 하지만 같은 명세에서 같은 결과가 나오지 않았다. **결정적이지 않은 계획은 계획이 아니다.**

네 사례의 공통점: **결정적인 실행 계획의 부재**. 팀의 모든 구성원이 AI로 개발할 수 있게 된 시대에, "무엇을 어떤 순서로 어떻게 만들지"가 결정되지 않은 채 실행하는 것은 재앙이다.

스펙 기반 개발 도구들은 이 문제를 마크다운 명세로 풀려고 했다. 하지만 마크다운 명세가 결정적으로 코드가 되지는 않는다. "리뷰된 명세"와 "동작하는 구현" 사이의 간극에는 여전히 해석이 필요하고, 바로 거기서 비결정성이 들어온다.

근본적인 문제: **명세는 읽기 위한 문서이지, 실행하기 위한 지시가 아니다.** AI 에이전트가 명세를 읽을 때마다 재해석이 일어난다. 다른 세션, 다른 해석, 다른 결과.

### plan.sh는 첫 번째 답이었다

이 문제에 대한 첫 번째 답은 단순했다: **계획을 실행 가능한 셸 스크립트로 만들자.** `claude -p "프롬프트"` 호출과 `verify` 검증을 bash 스크립트에 직접 넣으면, 계획과 실행 사이의 간극이 사라진다.

이것은 동작했다. 하지만 plan.sh 자체에 문제가 있었다:

1. **가독성** — 이것이 가장 치명적인 문제다. planosh의 전제는 "계획이 결정적이면 코드 리뷰가 필요 없다 — 계획 리뷰만 하면 된다"이다. 그런데 계획이 bash 스크립트라면? 팀원이 계획을 리뷰하려면 bash를 읽어야 한다. 프롬프트 텍스트가 heredoc에 묻히고, boilerplate(`step()`, `run_claude()`, `verify()`, `checkpoint()`)가 내용보다 많다. **"코드 대신 계획을 리뷰하자"고 했는데, 그 계획이 코드다.** 모순이다.

2. **순차 실행만** — Step 1, 2, 3을 나열하고 순서대로 실행한다. 실패하면 멈춘다. 재시도도 없고, rollback도 없고, 실패에서 배우지도 않는다.

3. **claude 종속** — `claude -p`가 하드코딩되어 있다. codex, aider, 다른 AI 도구를 쓸 수 없다.

4. **옵션 폭발** — `--append-system-prompt`, `--dangerously-skip-permissions`, `--model`, `--effort`, `--allowedTools` 등의 옵션을 bash에서 직접 조합해야 한다.

5. **git 복잡성** — 브랜치 생성, scope 체크, commit, worktree 관리가 전부 bash로 작성되어 있다.

6. **상태 관리 부재** — `.plan-state`에 step 번호만 기록한다. 실패 컨텍스트, 진행 이력, regression 정보가 없다.

**계획이 곧 실행이라는 핵심 가치는 맞았다. 하지만 plan.sh가 혼자 지고 있는 책임이 너무 많았다.**

## 설계 원점: autonomous.sh

planosh 프레임워크의 설계는 이론이 아니라 실전에서 나왔다.

프로젝트 하나를 plan.sh로 작성했을 때, 자연스럽게 나온 구조는 순차 실행이 아니었다. 수렴 루프였다.

```bash
# autonomous.sh — 실제로 16시간 무인 실행에 성공한 스크립트의 구조

while [ "$SESSION" -lt "$MAX_SESSIONS" ]; do
  # ① 다음 할 일 발견 (verify가 NEXT:를 반환)
  NEXT=$(verify --skip-tsc | grep '^NEXT:')

  # ② stuck 감지 (같은 작업 N번 연속 실패 → 중단)
  if [ "$FEATURE" = "$LAST_FEATURE" ]; then
    STUCK_COUNT=$((STUCK_COUNT + 1))
    [ "$STUCK_COUNT" -ge "$THRESHOLD" ] && break
  fi

  # ③ 프롬프트 조립 (태스크 + 진행상태 + 실패 컨텍스트)
  PROMPT="... ${FEATURE_JSON} ... ${REGRESSION_CONTEXT} ..."

  # ④ AI 실행
  claude -p "$PROMPT" --permission-mode bypassPermissions

  # ⑤ regression 체크 → 실패 시 rollback + 컨텍스트 저장 + 재시도
  if verify --regression | grep -q 'REGRESSIONS FOUND'; then
    git reset --hard "$BEFORE_SHA"
    REGRESSION_CONTEXT="$DETAILS"  # 다음 시도에 주입
    continue
  fi

  # ⑥ 성공 → commit + OTA 배포
  checkpoint "$FEATURE"
  ota_publish "$FEATURE"
done
```

이 스크립트를 `bash autonomous.sh` 한 번 실행하고 퇴근했다. 16시간 후 출근하니 모든 feature가 완료되어 있었다. 중간에 키보드를 만진 사람은 아무도 없었다.

**순차 실행(step 1→2→3)은 이 루프의 특수한 경우다.** task_source가 정적 목록이고, on_fail이 stop인 수렴 루프.

planosh 프레임워크는 이 발견에서 시작한다: **수렴 루프가 기본 실행 모델이고, 순차 실행은 그 위의 설정이다.**

## 핵심 가치

1. **계획이 코드를 대체한다**: 계획이 결정적이면, 사람은 생성된 코드를 읽을 필요가 없다. +4,567줄의 PR이 올라와도 계획이 리뷰되었다면 승인하면 된다. **계획 리뷰가 코드 리뷰를 대체하는 것 — 이것이 planosh의 존재 이유다.**
2. **그러므로 계획은 사람이 읽기 쉬워야 한다**: 계획이 코드 리뷰를 대체하려면, 계획 자체가 코드보다 읽기 쉬워야 한다. 기획자도, 디자이너도, 주니어 개발자도 계획을 읽고 "이 step에서 무엇을 만들고, 무엇을 검증하는지"를 이해할 수 있어야 한다. bash 스크립트는 이 기준을 충족하지 못했다. planfile은 이것을 위해 존재한다.
3. **결정성**: 같은 계획, 같은 결과. 결정적이지 않은 계획은 계획이 아니다.
4. **수렴**: 실패에서 배운다. 실패 컨텍스트를 다음 시도에 주입하고, regression이 없을 때까지 루프를 돈다.
5. **선언적 계획, 프로그래밍 가능한 실행**: 계획(planfile)은 사람이 읽고 리뷰한다. 실행(런타임)은 프레임워크가 처리한다. 이 둘을 분리한다.

## 아키텍처

planosh는 세 개의 독립된 계층으로 구성된다:

```
┌─────────────────────────────────────────────┐
│  planfile (선언적 계획)                       │
│  "무엇을, 어떤 순서로, 어떤 제약 하에"         │
│  → YAML. 사람이 읽고 리뷰한다.                │
├─────────────────────────────────────────────┤
│  런타임 (실행 엔진)                           │
│  "while 루프, retry, rollback, git, 상태"     │
│  → planfile을 읽고 실행한다.                  │
├─────────────────────────────────────────────┤
│  백엔드 어댑터 (AI 호출)                      │
│  "claude, codex, aider, custom"              │
│  → 런타임이 호출하는 플러그인.                 │
└─────────────────────────────────────────────┘
```

**planfile에는 `claude`라는 단어가 없다.** 프롬프트와 하네스와 검증만 있다. 어떤 AI로 실행할지는 런타임 설정이다.

### 7 프리미티브

autonomous.sh에서 추출한, 모든 AI 자율 실행에 공통되는 7가지 프리미티브:

| # | 프리미티브 | 정의 | 예시 |
|---|-----------|------|------|
| 1 | **task_source** | 다음 할 일을 어디서 가져오는가 | 정적 step 목록, verify 스크립트, feature JSON |
| 2 | **prompt_builder** | 프롬프트를 어떻게 조립하는가 | 태스크 + 하네스 + 진행상태 + 실패 컨텍스트 |
| 3 | **backend** | AI를 어떻게 호출하는가 | claude -p, codex, aider, custom script |
| 4 | **gate** | 성공/실패를 어떻게 판단하는가 | verify (exit code) + regression check |
| 5 | **on_fail** | 실패하면 어떻게 하는가 | stop, rollback+retry, retry+context |
| 6 | **on_success** | 성공하면 어떻게 하는가 | commit, commit+deploy, commit+OTA |
| 7 | **loop_control** | 언제 멈추는가 | max_sessions, stuck_threshold, all_done |

모든 planfile은 이 7가지의 조합이다. 아래 두 모드는 같은 프리미티브의 다른 설정값일 뿐이다.

### 두 가지 실행 모드

**Sequential** — 사전 정의된 step 목록을 순서대로 실행. 새 프로젝트를 처음부터 만들 때 적합.

```
task_source: steps (정적 목록)
on_fail: stop
loop: step 목록 끝까지
```

**Convergence** — 동적으로 다음 할 일을 발견하고, 실패 시 rollback+재시도. 마이그레이션, 리팩토링, 대규모 기능 추가에 적합.

```
task_source: command (verify 스크립트 등)
on_fail: rollback + retry (실패 컨텍스트 주입)
loop: 할 일이 없거나 stuck될 때까지
```

Sequential은 Convergence의 특수한 경우다: task_source가 정적이고, on_fail이 stop이면 Sequential이다.

## planfile 포맷

planfile은 **사람이 최종적으로 리뷰하는 문서**다.

계획이 결정적이면 코드는 항상 동일하게 나온다. 따라서 사람이 봐야 할 것은 생성된 코드가 아니라 이 planfile이다. planfile이 곧 PR이다. 팀이 모여서 "이 step의 프롬프트가 맞는가, 검증이 충분한가"를 논의하고, 머지하고, 실행한다.

이것이 planfile이 YAML인 이유다:

- **bash가 아니다** — planfile을 리뷰하는 사람이 bash를 알 필요가 없다. boilerplate가 없다. 프롬프트와 검증만 보인다.
- **프롬프트가 자연어로 보인다** — YAML의 `|` (literal block)로 프롬프트가 그대로 노출된다. heredoc에 묻히지 않는다.
- **구조가 한눈에 들어온다** — step 이름, 프롬프트, 검증, 커밋 메시지가 들여쓰기로 구분된다. 위에서 아래로 읽으면 계획 전체가 보인다.
- **diff가 깔끔하다** — git diff에서 "프롬프트의 어느 줄이 바뀌었는지"가 명확하게 보인다. bash에서는 boilerplate 변경과 내용 변경이 섞인다.

### Sequential 예시

```yaml
# planfile.yaml
plan: retro-webapp
mode: sequential

harness:
  global: .plan/retro-webapp/harness-global.md

git:
  branch: auto        # plan-{name}-{date} 자동 생성
  base: main
  scope: strict        # 하네스의 파일 목록 외 변경 감지 → 경고

steps:
  - name: 프로젝트 스캐폴딩
    harness: .plan/retro-webapp/harness-step-1.md
    prompt: |
      Next.js 프로젝트를 초기화하라.

      만들 것:
      - npx create-next-app@latest . --typescript --tailwind --app --src-dir
      - better-sqlite3, drizzle-orm 설치
      - vitest, @playwright/test 설치

      하지 않을 것:
      - 페이지 UI 구현
      - API 라우트 구현
    verify:
      - 빌드 성공: npm run build
    commit: "chore: project scaffolding"

  - name: 데이터 모델 + 세션 API
    harness: .plan/retro-webapp/harness-step-2.md
    prompt: |
      세션 생성 기능을 구현하라.

      만들 것:
      - src/db/schema.ts: sessions, cards, votes 테이블
      - src/app/api/sessions/route.ts: POST
      - src/app/page.tsx: 홈 페이지

      하지 않을 것:
      - 보드 UI
      - 실시간 기능
    verify:
      - 빌드 성공: npm run build
      - 타입 체크: npx tsc --noEmit
    commit: "feat: session creation API and home page"
```

**207줄 bash → 40줄 YAML.** boilerplate가 전부 사라지고 **사람이 판단해야 할 것만 남는다**: 각 step에서 무엇을 만들고, 무엇을 만들지 않고, 어떻게 검증하는가. 이것이 계획 리뷰에서 사람이 봐야 할 전부다.

### Convergence 예시

```yaml
# planfile.yaml
plan: expo-migration
mode: convergence

harness:
  global: .plan/expo-migration/harness-global.md

task_source:
  command: node scripts/verify.mts --skip-tsc
  pattern: "^NEXT: (.*)"
  context_file: claude-progress.txt
  feature_dir: features/

gate:
  verify: "{task_source.command} --feature {task.id}"
  regression: "{task_source.command} --regression"

on_fail:
  action: rollback+retry
  inject_context: true      # 실패 이유를 다음 프롬프트에 주입

loop:
  max_sessions: 200
  stuck_threshold: 3         # 같은 task 3회 연속 실패 → 중단

on_success:
  commit: "feat: {task.id}"
  deploy:
    command: "npx eas-cli update --branch autonomous --message '{task.id}'"

prompt: |
  # 자율 마이그레이션 세션

  ## 현재 진행 상태
  {context}

  ## 이번 세션 작업: {task.id}
  {task.json}

  {regression_warning}

  ## 규칙
  1. 1 세션 = 1 기능
  2. 오라클 = Flutter: ../graphic-app/ 의 동일 기능이 정답
  3. 구현 후 verify 실행하여 PASS 확인
  4. regression 확인 후 커밋
```

같은 프레임워크, 같은 런타임, 다른 설정.

## 런타임

런타임은 planfile을 읽고 실행하는 엔진이다.

### CLI

```bash
planosh run                        # planfile.yaml 자동 탐지 + 실행
planosh run plan.yaml              # 특정 planfile 실행
planosh run --dry                  # 프롬프트만 출력, AI 미호출
planosh run --from=3               # Step 3부터 재개 (sequential)
planosh run --backend=codex        # 백엔드 오버라이드
planosh calibrate --step=2 --runs=3  # 발산 탐지
```

### 수렴 루프

런타임의 핵심은 하나의 while 루프다:

```
initialize:
  git_setup(planfile.git)
  state = load_state() or empty

loop:
  while true:
    # 1. task_source
    task = get_next_task(planfile, state)
    if task is null: break  # 완료

    # 2. loop_control — stuck 감지
    if task == state.last_task:
      state.stuck_count++
      if state.stuck_count >= planfile.loop.stuck_threshold:
        break  # stuck
    else:
      state.stuck_count = 0

    # 3. prompt_builder
    prompt = build_prompt(
      planfile.prompt or task.prompt,
      task,
      state.context,
      state.failure_context
    )
    harness = load_harness(planfile.harness.global, task.harness)

    # 4. backend
    snapshot = git_snapshot()
    result = backend.run(prompt, harness, planfile.backend)

    # 5. gate
    if not gate.verify(task):
      on_fail(snapshot, state)
      continue

    if planfile.gate.regression:
      if not gate.regression():
        on_fail(snapshot, state)
        continue

    # 6. on_success
    git_commit(task.commit)
    if planfile.on_success.deploy:
      deploy(planfile.on_success.deploy)
    state.update(task)

finalize:
  save_state(state)
  report(state)
```

Sequential 모드에서 이 루프는 step 목록을 순서대로 순회하고, 실패 시 즉시 종료한다. Convergence 모드에서는 task_source가 동적으로 다음 작업을 결정하고, 실패 시 rollback + 재시도한다. **같은 루프, 다른 설정.**

### 실패 처리

plan.sh의 실패 처리는 `exit 1`이었다. autonomous.sh에서 검증된 패턴은 다르다:

```
on_fail(snapshot, state):
  if planfile.on_fail.action == "stop":
    save_state(current_step)
    exit 1

  if planfile.on_fail.action == "rollback+retry":
    git_reset(snapshot.sha)
    if planfile.on_fail.inject_context:
      state.failure_context = extract_failure_details()
    # continue → 루프가 같은 task를 재시도
```

핵심은 `inject_context`다. 실패 이유를 다음 시도의 프롬프트에 주입한다:

```
## ⚠ 이전 시도 실패

이전 시도에서 이 기능을 구현하다가 기존 feature를 깨뜨렸다.
롤백되었으므로 아래 항목들을 깨뜨리지 않도록 주의하라.

{failure_details}
```

AI가 같은 실수를 반복하지 않게 하는, 실전에서 검증된 패턴이다.

### Git 추상화

planfile의 `git` 섹션으로 모든 git 복잡성을 런타임이 처리한다:

```yaml
git:
  branch: auto           # plan-{name}-{date} 자동 생성
  base: main             # base 브랜치
  scope: strict          # 하네스 파일 목록 외 변경 감지
  stash_dirty: true      # 세션 시작 전 dirty state 자동 stash
  worktree: true         # calibrate 시 자동 worktree 관리
```

planfile 작성자는 git을 신경쓰지 않는다. 런타임이 브랜치 생성, stash/pop, scope 체크, commit, rollback을 전부 처리한다.

## 백엔드 어댑터

백엔드 어댑터는 `(prompt, harness, options) → result`를 구현하는 인터페이스다.

```yaml
# planfile 또는 글로벌 설정에서 지정
backend:
  provider: claude
  options:
    model: opus
    effort: max
    permissions: bypass
    allowed_tools: [Bash, Edit, Write, Read, Glob, Grep]
```

내장 어댑터:

| 어댑터 | 매핑 |
|--------|------|
| `claude` | `claude -p "$prompt" --append-system-prompt "$harness" --model $model --dangerously-skip-permissions` |
| `codex` | `codex --prompt "$prompt" --context-file "$harness_file"` |
| `aider` | `aider --message "$prompt" --read "$harness_file"` |
| `custom` | 사용자 정의 스크립트. stdin으로 prompt, 환경변수로 harness 경로 전달 |

커스텀 어댑터 예시:

```bash
# backends/my-backend.sh
# 런타임이 이 스크립트를 호출한다
# $PLANOSH_PROMPT: 프롬프트
# $PLANOSH_HARNESS_FILE: 하네스 파일 경로
# $PLANOSH_OPTIONS: JSON으로 된 옵션

my-ai-tool run --prompt "$PLANOSH_PROMPT" --system "$(cat $PLANOSH_HARNESS_FILE)"
```

## 결정성 모델: 3계층 제약 아키텍처

### 왜 마크다운 명세는 비결정적인가

기존 SDD(Spec-Driven Development) 도구들의 흐름:

```
PRD.md → spec.md → plan.md → tasks.md → AI 세션 → 코드
```

각 화살표마다 해석이 개입한다. 같은 spec.md를 두 번 읽으면 두 가지 다른 구현이 나온다. 에이전틱 건망증 연구에 따르면 35분 이후 AI 세션의 성공률이 급격히 떨어지고, 작업 시간을 2배로 늘리면 실패율은 4배가 된다.

planfile은 이 해석 계층을 제거하지만, AI 호출만으로는 LLM 고유의 비결정성이 남는다. 같은 프롬프트에도 파일 구조, 네이밍, 패턴 선택이 매번 달라질 수 있다.

### 해법: 하네스로 해의 공간 압축

프롬프트가 **무엇을** 만들지 제약한다면, 하네스는 **어떻게** 만들지 제약한다. 두 축의 제약이 교차하면 해의 공간이 극적으로 좁아진다.

```
  마크다운 명세만          프롬프트만             프롬프트 + 하네스
  ┌─────────────┐       ┌─────────────┐       ┌─────────────┐
  │ ░░░░░░░░░░░ │       │             │       │             │
  │ ░░░░░░░░░░░ │       │   ░░░░░░░   │       │             │
  │ ░░░░░░░░░░░ │  →    │   ░░░░░░░   │  →    │    ░░░      │
  │ ░░░░░░░░░░░ │       │   ░░░░░░░   │       │             │
  │ ░░░░░░░░░░░ │       │             │       │             │
  └─────────────┘       └─────────────┘       └─────────────┘
  해석 자유도: 높음       WHAT 제약됨            WHAT + HOW 제약됨
```

### 3계층 모델

```
┌─────────────────────────────────────────────────────┐
│ Layer 1: 하네스 (시스템 프롬프트)                      │
│ HOW — 코딩 컨벤션, 아키텍처 규칙, 금지 패턴            │
│ "어떻게 만들지"를 제약                                │
├─────────────────────────────────────────────────────┤
│ Layer 2: 프롬프트                                    │
│ WHAT — 만들 것, 하지 않을 것, 선행 조건               │
│ "무엇을 만들지"를 제약                                │
├─────────────────────────────────────────────────────┤
│ Layer 3: 검증 (gate)                                │
│ CHECK — 빌드, 파일 존재, 테스트 통과, regression      │
│ 결과가 제약에 부합하는지 사후 확인                     │
└─────────────────────────────────────────────────────┘
```

- **Layer 1 (하네스)**: 글로벌 하네스(모든 step에 적용)와 step별 하네스로 구성. 시스템 프롬프트는 유저 프롬프트보다 우선순위가 높으므로 컨벤션을 "잊을" 가능성이 낮다.
- **Layer 2 (프롬프트)**: planfile의 prompt 필드. 해당 step의 구체적 산출물과 범위 경계를 명시.
- **Layer 3 (검증)**: verify + regression check. Layer 1-2의 제약에도 불구하고 발생하는 편차를 잡는 안전망.

### 하네스 구조

```
.plan/
└── {plan-name}/
    ├── harness-global.md           ← 모든 step에 적용
    ├── harness-step-1.md           ← step별 맥락 제약
    ├── harness-step-2.md
    └── ...
```

**글로벌 하네스** — 프로젝트 전체에 일관되게 적용:

```markdown
# 프로젝트 컨벤션

## 기술 결정
- D-001: Next.js 14 App Router + TypeScript strict mode
- D-002: Prisma ORM, PostgreSQL
- D-003: Google OAuth only (NextAuth.js)

## 코딩 규칙
- 컴포넌트: src/components/{feature}/{ComponentName}.tsx
- 'use client' 최소화 — 서버 컴포넌트 우선
- CSS: Tailwind만 사용, 인라인 style 금지

## 절대 금지
- any 타입 사용
- 하드코딩된 시크릿
- 이 Step의 범위 밖 파일 수정
```

**Step별 하네스** — 해당 step의 맥락에 맞는 추가 제약:

```markdown
# Step 2 하네스: Google OAuth 로그인

## 현재 프로젝트 상태
Step 1 완료 후 존재하는 파일:
- prisma/schema.prisma (빈 스키마)
- src/app/layout.tsx, src/app/page.tsx
- package.json (next, prisma 설치됨)

## 이 Step의 아키텍처 제약
- NextAuth 설정: src/lib/auth.ts (단일 파일)
- 세션 전략: JWT (database 세션 아님)

## 생성할 파일 목록 (이 목록 외 파일 생성 금지)
- prisma/schema.prisma
- src/lib/auth.ts
- src/app/api/auth/[...nextauth]/route.ts
- src/middleware.ts
```

### 프롬프트/하네스 분리 원칙

이 분리는 가독성을 위한 것이다.

planfile을 리뷰하는 사람이 봐야 할 것은 **"이 step에서 무엇을 만드는가"**다. 코딩 컨벤션, import 순서, 타입스크립트 strict mode 같은 기술적 제약은 리뷰어의 관심사가 아니다. 그것은 한번 합의하면 모든 step에 적용되는 것이고, 리뷰할 때마다 읽어야 할 것이 아니다.

**프롬프트** (planfile에 보임) — WHAT:
- "만들 것" — 이 step의 구체적 산출물
- "하지 않을 것" — 이 step의 범위 외 항목

**하네스** (별도 파일, 한번 리뷰) — HOW:
- 기술 결정, 코딩 컨벤션 → harness-global.md
- 이전 step 상태, 아키텍처 제약, 파일 화이트리스트 → harness-step-N.md

프롬프트가 짧을수록 planfile이 읽기 쉽다. 기술 결정이나 코딩 규칙을 프롬프트에 반복하면 리뷰어가 매 step마다 같은 내용을 건너뛰어야 한다.

## 교정 (Calibration)

### 왜 결정성이 유일한 eval인가

AI가 생성한 코드의 품질을 평가하는 방법들:

- **코드 품질 평가**: "좋은 코드인가" — 주관적이고 컨텍스트 의존적. AI evaluator의 판정도 비결정적. 비결정적인 evaluator로 비결정적인 생성기를 개선하는 루프는 수렴하지 않는다.
- **시각적 테스트**: "의도한 UI와 일치하는가" — AI의 시각적 판단 자체가 매번 다르다. 범용 eval로 사용 불가.
- **유닛 테스트/타입 체크**: 객관적이지만, "통과했다"가 "의도한 구현이다"를 보장하지 않는다.
- **결정성 (병렬 실행 diff)**: 같은 입력에 같은 출력이 나오는가. 주관이 아니라 사실. diff가 있으면 비결정적이고, 없으면 결정적이다. 모든 step, 모든 프로젝트에 동일하게 적용된다.

**결정성만이 AI 코드 생성에서 자동화된 개선 루프를 돌릴 수 있는 유일한 객관적 메트릭이다.**

### 교정 루프

```bash
planosh calibrate --step=2 --runs=3
```

```
┌───────────────────────────────────────────────────┐
│ Phase 1: 격리 실행                                 │
│  git worktree로 N개의 격리된 환경 생성              │
│  ┌────────┐  ┌────────┐  ┌────────┐               │
│  │ run-1  │  │ run-2  │  │ run-3  │  (병렬)       │
│  └───┬────┘  └───┬────┘  └───┬────┘               │
├──────▼───────────▼───────────▼────────────────────┤
│ Phase 2: 발산 분석                                 │
│  diff run-1 vs run-2 vs run-3                     │
│  발산 유형: 구조 / 패턴 / 네이밍 / 범위             │
├───────────────────────────────────────────────────┤
│ Phase 3: 사용자 결정 요청                          │
│  "세션 전략이 갈렸습니다: JWT vs database.          │
│   어느 쪽을 사용할까요?"                           │
├───────────────────────────────────────────────────┤
│ Phase 4: 하네스 업데이트                           │
│  사용자 결정을 harness-step-N.md에 규칙으로 추가    │
│  → divergence-report.md 생성                      │
│  → 다시 calibrate 실행하여 수렴 확인               │
└───────────────────────────────────────────────────┘
```

발산 자체가 문제가 아니다. 발산은 **하네스가 아직 제약하지 못한 지점을 가리키는 신호**다. 교정 루프의 역할은 "여기서 갈라진다"를 발견하고, 사용자에게 결정을 받아 하네스에 기록하여 더 이상 갈라지지 않게 만드는 것이다.

### `.plan/` 디렉토리 최종 구조

```
.plan/
└── {plan-name}/
    ├── harness-global.md
    ├── harness-step-*.md
    ├── divergence-report.md        ← 교정 시 생성
    └── calibration-history/        ← 교정 이력
        └── step-2-convergence.md   ← 수렴 추이 (67% → 95% → 100%)
```

## planosh가 하지 않는 것

planosh는 **PRD가 이미 존재한다고 가정**한다. PRD 작성, 아이디어 정리, 요구사항 도출은 planosh의 범위가 아니다.

```
아이디어 → PRD 작성     |     PRD → 결정적 실행
━━━━━━━━━━━━━━━━━━━━━  |  ━━━━━━━━━━━━━━━━━━━━━
planosh 범위 밖          |  planosh 범위
```

## 검증 전략

verify에 넣을지 판단하는 기준은 하나다: **같은 입력에 항상 같은 판정(pass/fail)을 내리는가.**

사용 가능한 검증:
- `npm run build`, `cargo build` (빌드)
- `[ -f path ]` (파일 존재)
- `npx tsc --noEmit` (타입 체크)
- `npm test`, `pytest` (유닛 테스트)
- `npm run lint` (린트)
- `grep -q 'pattern' file` (파일 내용)
- regression check (기존 feature 깨지지 않았는지)

사용하지 않는 검증:
- AI에게 "스크린샷이 예쁜가" 판정시키는 것
- 비결정적 판정이 포함된 것

## 미해결 질문

1. **구현 언어**: 런타임을 무엇으로 만드는가? Shell(무의존성), Node.js(생태계), Go(싱글 바이너리)
2. **planfile 파싱**: YAML의 어디까지를 스펙으로 정의하는가? 템플릿 변수(`{task.id}`)의 문법
3. **상태 파일 포맷**: `.plan-state`를 어떤 형식으로 확장하는가 (JSON? YAML?)
4. **calibrate와 convergence의 관계**: convergence 모드에서 calibrate의 역할은?
5. **하네스 상속**: 글로벌 설정의 하네스와 planfile의 하네스가 충돌하면?
6. **append-system-prompt 검증**: 하네스가 실제로 시스템 프롬프트처럼 우선순위가 높은지 실험 필요

## 성공 기준

- `planosh run`으로 planfile.yaml을 읽어서 실행할 수 있다
- `planosh run --dry`로 프롬프트만 출력할 수 있다
- Sequential 모드에서 verify 실패 시 중단하고 `--from=N`으로 재개할 수 있다
- Convergence 모드에서 실패 시 rollback + 재시도하고, stuck 감지가 동작한다
- `planosh run --backend=codex`로 백엔드를 바꿀 수 있다
- `planosh calibrate --step=N --runs=3`으로 발산을 탐지할 수 있다
- 기존 autonomous.sh와 동등한 기능을 planfile.yaml + planosh run으로 재현할 수 있다
- planfile.yaml과 `.plan/` 하네스를 git에 커밋하고 PR로 리뷰할 수 있다

## 설계 원칙

1. **planfile은 사람을 위한 문서다.** planfile이 코드 리뷰를 대체한다. 따라서 planfile의 모든 설계 결정은 "사람이 읽기 쉬운가"를 기준으로 판단한다. 기계가 파싱하기 편한 형식보다 사람이 한눈에 이해하는 형식을 택한다. planfile에 기계만 이해하는 필드가 보이면 설계가 잘못된 것이다.
2. **증명된 것만 넣는다.** autonomous.sh에서 실전으로 검증된 패턴만 프레임워크에 포함한다. 이론적으로 좋아 보이지만 써보지 않은 기능은 넣지 않는다.
3. **선언적 계획, 명령적 실행.** 사람은 YAML을 읽고, 기계는 while 루프를 돌린다. 이 경계를 흐리지 않는다.
4. **설정으로 해결한다.** Sequential과 Convergence를 별도 시스템으로 만들지 않는다. 같은 런타임의 다른 설정으로 표현한다.
5. **plan.sh는 여전히 동작한다.** planfile.yaml은 plan.sh를 대체하지만, 기존 plan.sh를 버리지 않는다. 프레임워크 없이 bash만으로 돌리고 싶은 사람은 계속 plan.sh를 쓸 수 있다.
