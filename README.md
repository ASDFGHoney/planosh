# planosh

> [한국어 버전은 아래에 있습니다](#planosh-1)

In the age of AI coding tools for everyone — developers, PMs, designers all generate code with AI. But without a deterministic execution plan, just telling AI to "figure it out" leads to this:

> **Developer**: "I'd have to run it through Claude to find out..."

> **PM**: Submitted a PR. Please review. → 40 files changed, **+4,567 -12,300**
>
> If the execution plan had been reviewed first, and that plan was deterministic? No one would need to read a single line of code.

> **A PM reads the spec and builds a feature with AI. It works.** A week later, they ask AI to make a small change to the same spec — and it rebuilds everything in a completely different structure. The PM can't read code, so they don't know what changed. They end up asking a developer for help.
>
> The spec existed. But the same plan didn't produce the same result.

**A plan that isn't deterministic isn't a plan.**

## plan.sh

planosh proposes making AI execution plans into **deterministic shell scripts**.

Each step is a `claude -p` call with an embedded prompt, a harness (system prompt) constraining HOW, and a verify step that validates results — all in a shell script. The plan is the execution.

```bash
# -- Step 2: Google OAuth --
CURRENT_STEP=2; step 2 "Google OAuth login"
run_claude "
Implement Google OAuth login.
## Build
- User/Account/Session models (prisma)
- NextAuth Google Provider + Prisma Adapter
- /auth/signin page (Google login button)
## Don't build
- Email/password signup
- Profile editing, team features
"
verify "Build succeeds" "npm run build"
verify "Login page exists" "[ -f src/app/auth/signin/page.tsx ]"
checkpoint "feat: Google OAuth login"
```

`run_claude` combines the prompt (-p) with a harness (--append-system-prompt) to invoke Claude. verify judges pass/fail by exit code. That's it.

## Why is it deterministic?

```
+-------------------------------------------------+
| Layer 1: System Prompt (--append-system-prompt)  |
| HOW -- coding conventions, architecture rules,   |
| forbidden patterns                               |
+-------------------------------------------------+
| Layer 2: User Prompt (-p)                        |
| WHAT -- what to build, what not to build,        |
| preconditions                                    |
+-------------------------------------------------+
| Layer 3: Verification (verify)                   |
| CHECK -- build, file existence, test pass        |
+-------------------------------------------------+
```

When WHAT (prompt) and HOW (harness) are constrained simultaneously, the solution space narrows dramatically. The 5-stage transformation where AI reads a markdown spec (spec -> plan -> tasks -> sessions -> code) compresses into 1 stage (prompt -> code).

**When the plan is deterministic, execution becomes fully asynchronous.** Run `bash plan.sh` before leaving work, and it's done by morning.

For the detailed design of the 3-layer constraint model, harness structure, and calibration loops, see the [design document](docs/DESIGN.md).

## planosh is a proposal, not a framework

planosh doesn't provide a finished framework. It **proposes the concept of making AI execution plans into deterministic shell scripts**, with the goal of building patterns and best practices together as a community.

What we hope to see:

- Share cases where combining while loops and harnesses achieved **100% determinism** for a specific step
- Share patterns that knocked out 50-file, 20-step projects in a single plan.sh run
- Create **harness templates** for specific stacks like Next.js, Rails, Flutter
- Discover patterns that guarantee quality through deterministic verification alone, without AI judgment in verify
- Share harness gaps found through calibration loops and their solutions

As cases accumulate, patterns emerge. As patterns collect, they become a framework. planosh's role is to be that seed.

## Contributing

### Sharing plan.sh files

Contribute your plan.sh and harness files via PR to the `.plan/` directory. The `.plan/` in this repo is the best practice collection.

```
.plan/
+-- your-plan-name/
    +-- plan.sh
    +-- harness-global.md
    +-- harness-step-N.md
    +-- README.md          <- Use the template below
```

Contribution README template:

```markdown
## Project
(What you built)

## Steps
(How many, what each step does)

## Determinism rate
(Identical results ratio across N runs)

## Key findings
(Which harness/patterns were effective, what divergence you found and how you contained it)
```

### Pattern discussions

Share your discovered patterns, experiment results, and ideas in [GitHub Discussions](../../discussions).

### Divergence reports

Report divergence cases in [Issues](../../issues). "I ran the same plan.sh and got this difference" is the best starting point for harness improvement.

## Reference implementation

planosh provides two Claude Code skills as reference implementations. plan.sh can be written by hand without them.

- **`/planosh`** -- Takes a PRD, makes technical decisions interactively, and generates plan.sh + harness
- **`/planosh-calibrate`** -- Runs plan.sh N times in parallel in isolated environments, finds divergence points, takes user decisions, and strengthens the harness

Installation: Copy skill files to `.claude/skills/`

```bash
cp planosh.md .claude/skills/
cp planosh-calibrate.md .claude/skills/
```

## Further reading

- [Design document](docs/DESIGN.md) -- Problem definition, 3-layer constraint model, calibration loops, full plan.sh example

---

# planosh

AI 코딩 도구가 팀 전체에 보급된 시대. 개발자, 기획자, 디자이너 모두가 AI로 코드를 만든다. 하지만 결정적인 실행 계획 없이 AI에게 "알아서 해줘"를 시키면, 이런 일이 벌어진다:

> **개발자**: "그건 claude로 해봐야 알 것 같아요..ㅠㅠ"

> **기획자**: PR 올렸어요. 리뷰해주세용. → 파일 변경 40개, **+4,567 -12,300**
>
> 만약 실행 계획이 먼저 리뷰되었고, 그 계획이 결정적이었다면? 코드를 한 줄도 안 봐도 됐다.

> **기획자가 명세를 보고 AI로 기능을 만들었다. 동작한다.** 일주일 후, 같은 명세로 작은 수정을 하려고 다시 시켰더니, AI가 전혀 다른 구조로 만들어버렸다. 기획자는 코드를 읽을 수 없으니 뭐가 달라졌는지 모른다. 결국 개발자에게 도움을 요청한다.
>
> 명세는 있었다. 하지만 같은 계획에서 같은 결과가 나오지 않았다.

**결정적이지 않은 계획은 계획이 아니다.**

## plan.sh

planosh는 AI 실행 계획을 **결정적인 셸 스크립트**로 만들자는 제안이다.

각 Step이 프롬프트를 직접 내장한 `claude -p` 호출이고, 하네스(시스템 프롬프트)가 HOW를 제약하고, verify가 결과를 검증하는 셸 스크립트. 계획이 곧 실행이다.

```bash
# ── Step 2: Google OAuth ──
CURRENT_STEP=2; step 2 "Google OAuth 로그인"
run_claude "
Google OAuth 로그인을 구현하세요.
## 만들 것
- User/Account/Session 모델 (prisma)
- NextAuth Google Provider + Prisma Adapter
- /auth/signin 페이지 (Google 로그인 버튼)
## 하지 않을 것
- 이메일/비밀번호 가입
- 프로필 편집, 팀 기능
"
verify "빌드 성공" "npm run build"
verify "로그인 페이지 존재" "[ -f src/app/auth/signin/page.tsx ]"
checkpoint "feat: Google OAuth login"
```

`run_claude`는 프롬프트(-p)와 하네스(--append-system-prompt)를 결합하여 Claude를 호출한다. verify는 exit code로 성공/실패를 판정한다. 이게 전부다.

## 왜 결정적인가

```
┌─────────────────────────────────────────────────────┐
│ Layer 1: 시스템 프롬프트 (--append-system-prompt)    │
│ HOW — 코딩 컨벤션, 아키텍처 규칙, 금지 패턴          │
├─────────────────────────────────────────────────────┤
│ Layer 2: 유저 프롬프트 (-p)                          │
│ WHAT — 만들 것, 하지 않을 것, 선행 조건               │
├─────────────────────────────────────────────────────┤
│ Layer 3: 검증 (verify)                              │
│ CHECK — 빌드, 파일 존재, 테스트 통과                  │
└─────────────────────────────────────────────────────┘
```

WHAT(프롬프트)과 HOW(하네스)가 동시에 제약되면 해의 공간이 극적으로 좁아진다. 마크다운 명세를 AI가 읽는 5단계 변환(spec → plan → tasks → sessions → code)이 1단계(prompt → code)로 압축된다.

**계획이 결정적이면 실행은 완전히 비동기가 된다.** 퇴근할 때 `bash plan.sh` 돌려놓으면 출근하면 되어있다.

3계층 제약 모델의 상세 설계, 하네스 구조, 교정 루프 등은 [설계 문서](docs/DESIGN.md)를 참고.

## planosh는 프레임워크가 아니라 제안이다

planosh는 완성된 프레임워크를 제공하지 않는다. **"AI 실행 계획을 결정적인 셸 스크립트로 만들자"는 개념을 제안**하고, 그 개념을 실현하는 패턴과 모범 사례를 커뮤니티가 함께 만들어가는 것이 목표다.

우리가 바라는 일:

- while 루프와 하네스를 조합해서 특정 Step의 **결정률 100%**를 달성한 사례를 공유한다
- 50개 파일, 20 Step 규모의 작업을 plan.sh 한 번으로 해치운 패턴을 공유한다
- Next.js, Rails, Flutter 등 특정 스택에서 동작하는 **하네스 템플릿**을 만든다
- verify에서 AI 판단 없이 결정적 검증만으로 품질을 보장하는 패턴을 발견한다
- 교정 루프를 돌려서 발견한 하네스 빈틈과 그 해결법을 공유한다

사례가 쌓이면 패턴이 보이고, 패턴이 모이면 프레임워크가 된다. planosh의 역할은 그 씨앗이 되는 것이다.

## 기여하기

### plan.sh 공유

`.plan/` 디렉토리에 당신의 plan.sh와 하네스를 PR로 기여해주세요. 이 레포의 `.plan/`이 곧 best practice 컬렉션이다.

```
.plan/
└── your-plan-name/
    ├── plan.sh
    ├── harness-global.md
    ├── harness-step-N.md
    └── README.md          ← 아래 템플릿 사용
```

기여 README 템플릿:

```markdown
## 프로젝트
(무엇을 만들었는지)

## Steps
(몇 개, 각 Step이 하는 일)

## 결정률
(N번 실행 중 동일 결과 비율)

## 핵심 발견
(어떤 하네스/패턴이 효과적이었는지, 어떤 발산을 발견하고 어떻게 봉쇄했는지)
```

### 패턴 논의

[GitHub Discussions](../../discussions)에서 발견한 패턴, 실험 결과, 아이디어를 공유해주세요.

### 발산 보고

[Issues](../../issues)에서 발산 사례를 보고해주세요. "같은 plan.sh를 돌렸는데 이런 차이가 나왔다"는 하네스 개선의 가장 좋은 출발점이다.

## 레퍼런스 구현

planosh는 두 개의 Claude Code 스킬을 레퍼런스 구현으로 제공한다. 스킬 없이도 plan.sh는 손으로 쓸 수 있다.

- **`/planosh`** — PRD를 입력받아 대화형으로 기술 결정을 하고, plan.sh + 하네스를 생성
- **`/planosh-calibrate`** — plan.sh를 격리된 환경에서 N번 병렬 실행하여 발산 지점을 찾고, 사용자 결정을 받아 하네스를 강화

설치: `.claude/skills/`에 스킬 파일을 복사

```bash
cp planosh.md .claude/skills/
cp planosh-calibrate.md .claude/skills/
```

## 더 읽기

- [설계 문서](docs/DESIGN.md) — 문제 정의, 3계층 제약 모델, 교정 루프, plan.sh 전체 예시
