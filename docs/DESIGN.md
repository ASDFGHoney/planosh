# 설계 문서: planosh — AI 코딩 팀을 위한 결정적 plan 생성기

## 제공 스킬

이 오픈소스 프로젝트는 2개의 Claude Code 스킬을 제공한다:

### 1. `/planosh` — plan.sh 생성

PRD를 입력받아 실행 가능한 plan.sh와 하네스를 생성한다.

```
/planosh path/to/prd.md

입력: PRD (마크다운)
과정: PRD 분석 → 대화형 기술 결정 → Step 분해
출력: plan.sh + .plan/harness-global.md + .plan/harness-step-N.md
```

### 2. `/planosh-calibrate` — 발산 탐지 및 하네스 강화

생성된 plan.sh를 격리된 환경에서 N번 병렬 실행하고, 발산 지점을 찾아 사용자에게 결정을 요청한 뒤, 그 결정을 하네스에 추가한다.

```
/planosh-calibrate [--step=N] [--runs=3]

입력: plan.sh + 기존 하네스
과정: 병렬 실행 → 발산 분석 → 사용자에게 결정 요청 → 하네스 업데이트
출력: 강화된 하네스 + 발산 리포트
```

두 스킬의 관계:

```
/planosh             /planosh-calibrate        /planosh-calibrate
  PRD → plan.sh  →    발산 발견 →               발산 발견 →
       + 하네스 v0     사용자 결정 →              사용자 결정 →
                       하네스 v1                  하네스 v2     → ... 수렴
```

`/planosh`는 한 번 실행한다. `/planosh-calibrate`는 하네스가 충분히 결정적이 될 때까지 반복한다. plan.sh 자체는 수정하지 않고, 하네스만 강화한다.

## 문제 정의

AI 코딩 도구가 팀 전체에 보급되면서, 이제 개발자뿐 아니라 기획자, 디자이너도 코드를 만들 수 있게 되었다. 하지만 결정적인 실행 계획 없이 각자가 AI에게 "알아서 해줘"를 시키면, 팀 전체의 코드베이스가 비결정적으로 변한다.

현장에서 실제로 일어나는 일:

> **개발자**: "그건 claude로 해봐야 알 것 같아요..ㅠㅠ"
>
> — 계획이 없으니 실행 전까지 결과를 예측할 수 없다. "해봐야 안다"가 개발 프로세스가 되어버렸다.

> **기획자**: "저 추가하고 싶은 기능이 있어서 한번 개발하고 PR 올려봤어요. 리뷰해주세용."
> → 파일 변경 40개, +4,567 -12,300
>
> — 의도는 좋지만, 결정적인 계획 없이 AI에게 통째로 맡기면 이런 PR이 나온다. 4,567줄의 코드를 리뷰해야 한다. 하지만 만약 이 기능의 실행 계획이 먼저 리뷰되었고, 그 계획이 결정적이었다면? 코드를 한 줄도 안 봐도 됐다. 계획이 곧 결과를 보장하니까. 문제는 코드가 아니라, 계획 없이 실행한 것이다.

> **3개월 후**: "이 인증 코드, 왜 이 구조로 만들었지?"
> → AI 세션은 이미 사라졌고, 커밋 메시지는 "feat: add auth"뿐이다. 왜 JWT를 골랐는지, 왜 파일을 이렇게 나눴는지, 아무도 모른다.
>
> — 계획이 리뷰되었다면, 모든 기술 결정과 그 이유가 `.plan/`에 남아있다. 코드는 바뀌어도, 결정의 맥락은 사라지지 않는다.

> **기획자가 명세를 보고 AI로 기능을 만들었다. 동작한다.** 일주일 후, 같은 명세를 보고 작은 수정을 하려고 다시 시켰더니, AI가 전혀 다른 구조로 만들어버렸다. 기획자는 코드를 읽을 수 없으니 뭐가 달라졌는지 모른다. 결국 개발자에게 도움을 요청한다. "제가 직접 하는 게 더 빠를 것 같아요."
>
> — 명세는 있었다. 계획도 있었다. 하지만 같은 계획에서 같은 결과가 나오지 않았다. 비개발자가 AI로 코드를 만들 수 있게 된 시대에, 결정적이지 않은 계획은 계획이 아니다.

네 사례의 공통점: **결정적인 실행 계획의 부재**. 계획이 없으면 AI는 매번 다르게 해석하고, 누가 실행하든 예측 불가능한 결과가 나온다. 팀의 모든 구성원이 AI로 개발할 수 있게 된 시대에, "무엇을 어떤 순서로 어떻게 만들지"가 결정되지 않은 채 실행하는 것은 재앙이다.

이에 대한 자연스러운 반응은 "코드베이스 자체에 방어선을 만들자"였다. CLAUDE.md에 컨벤션을 적고, 린터 규칙을 강화하고, pre-commit hook으로 검증을 걸고, 아키텍처 문서를 정비했다. 하지만 이 접근은 근본적으로 한계가 있었다 — **팀의 모든 구성원이 어떤 프롬프트를 넣을지 예측할 수 없기 때문이다.** "로그인 기능 만들어줘"라고 할 수도 있고, "인증 시스템 전체를 리팩토링해줘"라고 할 수도 있다. 어떤 프롬프트에서도 방어할 수 있는 만능 하네스를 코드베이스에 구축하는 것은 불가능에 가까웠다.

문제는 코드베이스의 방어력이 아니라, 실행 자체가 비결정적이라는 것이었다. 방어선을 아무리 높여도, 들어오는 공격(프롬프트)이 매번 다르면 뚫린다. 필요한 것은 방어선이 아니라, **실행 자체를 결정적으로 만드는 것**이었다.

스펙 기반 개발 도구들(GitHub Spec Kit, BMAD, Kiro, Zencoder)은 이 문제를 마크다운 명세로 풀려고 했다. 하지만 마크다운 명세가 결정적으로 코드가 되지는 않는다. "리뷰된 명세"와 "동작하는 구현" 사이의 간극에는 여전히 해석이 필요하고, 바로 거기서 컨텍스트가 유실되고 결정이 흔들리며 비결정성이 들어온다.

근본적인 문제: **명세는 읽기 위한 문서이지, 실행하기 위한 지시가 아니다.** AI 에이전트가 명세를 읽을 때마다 재해석이 일어난다. 다른 세션, 다른 해석, 다른 결과.

## 핵심 가치

plan.sh는 계획과 실행 사이의 간극을 제로로 만든다.

각 단계가 명세를 프롬프트에 직접 내장한 `claude -p` 호출이고, 자동 검증이 뒤따르는 셸 스크립트다. 팀은 코드 리뷰를 읽듯 계획을 읽는다. 각 단계가 Claude에게 무엇을 만들라고 할지, 어떻게 검증할지, 어떤 순서로 진행할지 정확하게 보인다.

"실행해"라고 하면 해석 계층이 없다. 계획이 곧 실행이다.

계획이 결정적이면 실행은 완전히 비동기가 된다. 퇴근할 때 `bash plan.sh` 돌려놓으면 출근하면 되어있다. 지켜볼 필요도, 중간에 개입할 필요도 없다. 계획 리뷰는 동기적(팀이 모여서), 실행은 비동기적(밤새 돌리면 됨). 비개발자도 마찬가지다. 계획을 승인받았으면 실행 버튼만 누르고 퇴근하면 된다.

리뷰 의식(ritual)이 곧 제품이다. PR에 plan.sh를 올리고, 팀이 프롬프트를 논의하고, 머지하고, 실행한다. 이것이 워크플로우다.

### planosh는 프레임워크가 아니라 제안이다

planosh는 완성된 프레임워크를 제공하지 않는다. "계획을 결정적인 셸 스크립트로 만들자"는 개념을 제안하고, 그 개념을 실현하는 패턴과 모범 사례를 커뮤니티가 함께 만들어가는 것이 목표다.

커뮤니티에서 일어나길 바라는 일:

- while 루프와 하네스를 조합해서 특정 Step의 결정률을 100%로 만든 사례를 공유한다
- 엄청 큰 작업(50개 파일, 20 Step)을 plan.sh 한 번으로 해치운 패턴을 공유한다
- Next.js, Rails, Flutter 등 특정 스택에서 동작하는 하네스 템플릿을 만든다
- verify에서 AI 판단 없이 결정적 검증만으로 품질을 보장하는 패턴을 발견한다
- 교정 루프를 돌려서 발견한 하네스 빈틈과 그 해결법을 공유한다

이런 사례가 쌓이면 어느 순간 패턴이 보이고, 패턴이 모이면 프레임워크가 된다. planosh의 역할은 그 씨앗이 되는 것이다. `/planosh`와 `/planosh-calibrate` 스킬은 이 개념의 레퍼런스 구현이지, 유일한 사용법이 아니다.

## 결정성 모델: 3계층 제약 아키텍처

### 왜 마크다운 명세는 비결정적인가

기존 SDD(Spec-Driven Development) 도구들의 흐름:

```
PRD.md → spec.md → plan.md → tasks.md → AI 세션 → 코드
```

각 화살표마다 해석이 개입한다. 같은 spec.md를 두 번 읽으면 두 가지 다른 구현이 나온다. 이것이 "문서를 읽는" 패러다임의 구조적 한계다. 에이전틱 건망증(Agentic Amnesia) 연구에 따르면 35분 이후 AI 세션의 성공률이 급격히 떨어지고, 작업 시간을 2배로 늘리면 실패율은 4배가 된다. 멀티에이전트 시스템은 41-87%가 프로덕션에서 실패하며, 36.9%는 에이전트 간 정렬 실패와 컨텍스트 붕괴에 기인한다.

plan.sh는 이 해석 계층을 제거하지만, `claude -p "프롬프트"`만으로는 LLM 고유의 비결정성이 남는다. 같은 "Google OAuth를 구현하세요"라는 프롬프트에도 파일 구조, 네이밍, 패턴 선택이 매번 달라질 수 있다.

### 해법: `--append-system-prompt`로 해의 공간 압축

`claude -p`의 프롬프트가 **무엇을** 만들지 제약한다면, `--append-system-prompt`는 **어떻게** 만들지 제약한다. 두 축의 제약이 교차하면 해의 공간(solution space)이 극적으로 좁아진다.

```
해의 공간 축소 다이어그램:

  마크다운 명세만          plan.sh (-p만)         plan.sh (-p + 시스템 프롬프트)
  ┌─────────────┐       ┌─────────────┐       ┌─────────────┐
  │ ░░░░░░░░░░░ │       │             │       │             │
  │ ░░░░░░░░░░░ │       │   ░░░░░░░   │       │             │
  │ ░░░░░░░░░░░ │  →    │   ░░░░░░░   │  →    │    ░░░      │
  │ ░░░░░░░░░░░ │       │   ░░░░░░░   │       │             │
  │ ░░░░░░░░░░░ │       │             │       │             │
  └─────────────┘       └─────────────┘       └─────────────┘
  해석 자유도: 높음       WHAT 제약됨            WHAT + HOW 제약됨
```

### 3계층 제약 모델

```
┌─────────────────────────────────────────────────────┐
│ Layer 1: 시스템 프롬프트 (--append-system-prompt)    │
│ HOW — 코딩 컨벤션, 아키텍처 규칙, 금지 패턴          │
│ "어떻게 만들지"를 제약                                │
├─────────────────────────────────────────────────────┤
│ Layer 2: 유저 프롬프트 (-p)                          │
│ WHAT — 만들 것, 하지 않을 것, 선행 조건               │
│ "무엇을 만들지"를 제약                                │
├─────────────────────────────────────────────────────┤
│ Layer 3: 검증 (verify)                              │
│ CHECK — 빌드, 파일 존재, 테스트 통과                  │
│ 결과가 제약에 부합하는지 사후 확인                     │
└─────────────────────────────────────────────────────┘
```

- **Layer 1 (시스템 프롬프트)**: 프로젝트 전체에 적용되는 글로벌 하네스와, 각 Step에 맞춤화된 Step별 하네스로 구성. 시스템 프롬프트는 유저 프롬프트보다 우선순위가 높으므로 LLM이 컨벤션을 "잊거나" 무시할 가능성이 낮다.
- **Layer 2 (유저 프롬프트)**: 기존 plan.sh의 `run_claude` 프롬프트. 해당 Step의 구체적 산출물과 범위 경계를 명시.
- **Layer 3 (검증)**: 사후 게이트. Layer 1-2의 제약에도 불구하고 발생할 수 있는 편차를 잡는 안전망.

### 하네스 구조

plan.sh 생성 시 프롬프트와 함께 하네스 파일을 생성한다:

```
project/
├── plan.sh                    ← 실행 가능한 계획
└── .plan/
    ├── harness-global.md      ← 글로벌 하네스 (모든 Step에 적용)
    └── harness-step-N.md      ← Step별 하네스 (해당 Step에만 적용)
```

**글로벌 하네스** (`.plan/harness-global.md`): 프로젝트 전체에 일관되게 적용되는 규칙

```markdown
# 프로젝트 컨벤션

## 기술 결정

- D-001: Next.js 14 App Router + TypeScript strict mode
- D-002: Prisma ORM, PostgreSQL
- D-003: Google OAuth only (NextAuth.js)

## 코딩 규칙

- 컴포넌트: src/components/{feature}/{ComponentName}.tsx
- API 라우트: src/app/api/{resource}/route.ts
- 서버 액션 금지 — API 라우트만 사용
- 'use client' 최소화 — 서버 컴포넌트 우선
- CSS: Tailwind만 사용, 인라인 style 금지, CSS 모듈 금지
- 에러 처리: try-catch 대신 Result 패턴
- import 순서: react → next → 외부 → 내부 → 타입

## 절대 금지

- any 타입 사용
- console.log (logger 모듈 사용)
- 하드코딩된 문자열 (상수 파일 사용)
- default export (named export만)
- 이 Step의 범위 밖 파일 수정
```

**Step별 하네스** (`.plan/harness-step-2.md`): 해당 Step의 맥락에 맞는 추가 제약

```markdown
# Step 2 하네스: Google OAuth 로그인

## 현재 프로젝트 상태

Step 1 완료 후 존재하는 파일:

- prisma/schema.prisma (빈 스키마, datasource만 정의됨)
- src/app/layout.tsx, src/app/page.tsx
- package.json (next, prisma, @playwright/test 설치됨)

## 이 Step의 아키텍처 제약

- NextAuth 설정: src/lib/auth.ts (단일 파일)
- Prisma Adapter 사용, 커스텀 adapter 금지
- 미들웨어: src/middleware.ts (Next.js 공식 패턴)
- 세션 전략: JWT (database 세션 아님)

## 이 Step에서 생성할 파일 목록 (이 목록 외 파일 생성 금지)

- prisma/schema.prisma (User, Account, Session 모델 추가)
- src/lib/auth.ts
- src/app/api/auth/[...nextauth]/route.ts
- src/app/auth/signin/page.tsx
- src/app/dashboard/page.tsx
- src/middleware.ts
```

### 결정성을 높이는 메커니즘

SDD 연구와 컨텍스트 엔지니어링 연구에서 도출한 plan.sh의 결정성 보강 근거:

1. **해석 계층 제거**: 마크다운 명세가 5단계 변환(Constitution → Specify → Plan → Tasks → Implement)을 거치는 반면, plan.sh는 명세를 프롬프트에 직접 내장하여 변환을 1단계로 압축한다.

2. **시스템 프롬프트의 제약 우선순위**: `--append-system-prompt`는 Claude의 기본 시스템 프롬프트에 추가되어 유저 프롬프트보다 높은 우선순위로 작동한다. "파일 목록 외 생성 금지"같은 규칙이 프롬프트의 모호함을 덮어쓴다.

3. **컨텍스트 엔지니어링 3계층과의 정렬**: 연구에서 정의한 결정적 로딩(deterministic loading) — CLAUDE.md, Rules, Hooks — 에 대응하여, 글로벌 하네스가 CLAUDE.md의 역할을, Step별 하네스가 Rules의 역할을, verify가 Hooks의 역할을 한다.

4. **JSON 형식의 비결정성 감소**: 연구에 따르면 "모델은 마크다운보다 JSON을 부적절하게 수정할 가능성이 낮다." 하네스에서 파일 목록, 의존성 등을 구조화된 형식으로 명시하면 해석 여지가 줄어든다.

5. **이전 Step 상태의 명시적 전달**: Step별 하네스에 "현재 프로젝트 상태" 섹션이 있어 에이전틱 건망증을 방지한다. 에이전트가 이전 Step의 산출물을 추측하지 않고 하네스에서 읽는다.

6. **비목표(non-goals)의 명시**: SDD 연구의 핵심 발견 — "AI는 쓰여 있지 않은 것에서 추론할 수 없다." 글로벌 하네스의 "절대 금지" 섹션과 Step별 하네스의 "이 목록 외 파일 생성 금지"가 암묵적 범위 확장을 차단한다.

7. **사후 검증 게이트**: Layer 1-2의 제약에도 불구하고 발생하는 편차를 verify가 잡는다. 검증 실패 시 즉시 중단하고 `.plan-state`에 기록하여 같은 지점에서 재시도할 수 있다.

8. **교정 루프를 통한 하네스 빈틈 봉쇄**: 위 1-7은 모두 "사전에 좋은 제약을 설계한다"는 전제에 의존한다. 교정 루프(`--calibrate`)는 이 전제를 검증한다 — N번 실행하여 실제 발산 지점을 찾고, 그 지점에 정확히 대응하는 하네스를 추가한다. 이론이 아닌 측정에 기반한 결정성 확보.

## `/planosh-calibrate` — 발산 탐지 루프

### 왜 결정성이 유일한 eval인가

AI가 생성한 코드의 품질을 평가하는 루프를 만든다고 하자. 후보가 되는 eval 방식들을 살펴보면:

**코드 품질 평가 (AI reviewer)**: "좋은 코드인가", "좋은 아키텍처인가", "유지보수하기 쉬운가" — 이런 기준은 AI evaluator에게 맡기기에 너무 추상적이다. 평가 기준 자체가 주관적이고 컨텍스트에 의존하기 때문에, evaluator AI의 판정도 비결정적이 된다. 비결정적인 evaluator로 비결정적인 생성기를 개선하는 루프는 수렴하지 않는다.

**시각적 테스트 (Playwright, Cypress 등)**: 에이전트가 스크린샷을 찍고 "의도한 UI와 일치하는가"를 시각적으로 판단하는 방식. 이것도 비결정적이다 — AI의 시각적 판단 자체가 매번 달라지고, "의도한 UI"의 정의도 모호하다. 설령 특정 케이스에서 동작하더라도 범용 eval로는 사용할 수 없다. 로그인 페이지의 버튼 위치를 판단하는 것과, 대시보드의 데이터 시각화 레이아웃을 판단하는 것은 완전히 다른 난이도다. 모든 Step, 모든 프로젝트에 일관되게 적용할 수 있는 eval이 아니다.

**유닛 테스트/타입 체크**: `npm run build`나 `npm test`는 객관적이지만, "통과했다"는 것이 "의도한 구현이다"를 보장하지 않는다. 빌드가 되고 테스트가 통과해도 파일 구조, 네이밍, 패턴이 매번 다를 수 있다.

**결정성 (병렬 실행 diff)**: 같은 입력에 같은 출력이 나오는가 — 이것은 주관이 아니라 사실이다. diff가 있으면 비결정적이고, 없으면 결정적이다. 모든 Step, 모든 프로젝트, 모든 기술 스택에 동일하게 적용된다. 도메인 지식이 필요 없고, AI의 판단에 의존하지 않으며, 이진 판정이 가능하다.

**이것만이 AI 코드 생성에서 자동화된 개선 루프를 돌릴 수 있는 유일한 객관적 메트릭이다.**

`/planosh-calibrate`는 이 원리 위에 서 있다. "더 나은 코드"를 찾는 루프가 아니라, "발산하는 지점"을 찾는 루프다.

### 왜 필수 기능인가

3계층 제약 모델은 해의 공간을 좁히지만, **얼마나 좁혔는지 측정할 방법이 없으면** 하네스의 품질은 직감에 의존한다. "파일 목록 외 생성 금지"라는 규칙이 실제로 동작하는지, 에이전트가 같은 파일에서 다른 패턴을 쓰진 않는지, 네이밍이 매번 달라지진 않는지 — 실행해봐야 안다.

`/planosh-calibrate`는 plan.sh를 격리된 환경에서 N번 병렬 실행하고, 결과물 간의 발산을 측정한 뒤, **발산 지점마다 사용자에게 결정을 요청하고 그 결정을 하네스에 추가**하는 스킬이다. 목적은 "어떤 결과가 맞는지" 판별하는 것이 아니다. **하네스에 빈틈이 어디에 있는지** 찾는 것이다.

```
하네스에 빈틈이 있으면              빈틈을 발견하면

 같은 프롬프트인데                   "세션 전략이 매번 달랐다
  매번 다른 구현이 나온다             → 하네스에 세션 전략이
                                     명시되어 있지 않았다
 → 어디가 빈틈인지 모름               → 사용자에게 물어보자"

                                    발산 = 하네스의 빈틈 신호
                                    → 빈틈 발견 → 사용자 결정 → 봉쇄 → 재측정
```

발산 자체가 문제가 아니다. 발산은 **하네스가 아직 제약하지 못한 지점을 가리키는 신호**다. 3번 실행해서 세션 전략이 JWT/JWT/database로 갈렸다면, 어느 쪽이 맞는지는 사람이 결정할 일이다. `/planosh-calibrate`의 역할은 "여기서 갈라진다"를 발견하고, 사용자에게 결정을 받아 하네스에 기록하여 더 이상 갈라지지 않게 만드는 것이다.

### 작동 방식

```
/planosh-calibrate [--step=M] [--runs=N]

┌─────────────────────────────────────────────────────────┐
│ Phase 1: 격리 실행                                       │
│                                                         │
│  git worktree로 N개의 격리된 환경 생성                    │
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │ run-1    │  │ run-2    │  │ run-3    │  (병렬)       │
│  │ Step M   │  │ Step M   │  │ Step M   │              │
│  │ 실행     │  │ 실행     │  │ 실행     │              │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘              │
│       │             │             │                     │
├───────▼─────────────▼─────────────▼─────────────────────┤
│ Phase 2: 발산 분석                                       │
│                                                         │
│  diff run-1 vs run-2 vs run-3                           │
│                                                         │
│  발산 유형 분류:                                         │
│  ├── 구조 발산: 파일/디렉토리 구조가 다름                  │
│  ├── 패턴 발산: 같은 파일인데 구현 패턴이 다름              │
│  ├── 네이밍 발산: 변수명/함수명/컴포넌트명이 다름           │
│  └── 범위 발산: 요청하지 않은 기능이 추가됨                 │
│                                                         │
├─────────────────────────────────────────────────────────┤
│ Phase 3: 사용자 결정 요청                                 │
│                                                         │
│  각 발산 지점을 사용자에게 제시:                           │
│                                                         │
│  "세션 전략이 갈렸습니다: JWT vs database                  │
│   어느 쪽을 사용할까요?"                                  │
│                                                         │
│  "auth 설정 파일 구조가 갈렸습니다:                       │
│   A) auth.ts 단일 파일  B) auth.ts + auth-options.ts     │
│   어느 쪽을 사용할까요?"                                  │
│                                                         │
│  → 사용자의 결정을 수집                                   │
│                                                         │
├─────────────────────────────────────────────────────────┤
│ Phase 4: 하네스 업데이트                                  │
│                                                         │
│  사용자 결정을 harness-step-M.md에 규칙으로 추가:          │
│                                                         │
│  구조 결정 → "생성할 파일 목록" 화이트리스트               │
│  패턴 결정 → 명시적 아키텍처 제약                          │
│  네이밍 결정 → 네이밍 규칙                                │
│  범위 결정 → "절대 금지" 항목                             │
│                                                         │
│  → .plan/divergence-report.md 생성                      │
│  → 다시 /planosh-calibrate 실행하여 수렴 확인             │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### 발산 리포트 예시

`--calibrate` 실행 후 `.plan/divergence-report.md`에 생성:

```markdown
# 발산 리포트 — Step 2: Google OAuth 로그인

실행 횟수: 3 | 날짜: 2026-04-07

## 수렴율: 67% (3회 중 2회 일치)

## 구조 발산

| 파일 경로                               | run-1 | run-2 | run-3 |
| --------------------------------------- | ----- | ----- | ----- |
| src/lib/auth.ts                         | ✅    | ✅    | ✅    |
| src/app/api/auth/[...nextauth]/route.ts | ✅    | ✅    | ✅    |
| src/middleware.ts                       | ✅    | ✅    | ✅    |
| src/lib/auth-options.ts                 | ❌    | ❌    | ✅    |

→ 제안: run-3이 auth 설정을 별도 파일로 분리함.
harness-step-2.md에 "NextAuth 설정은 src/lib/auth.ts 단일 파일" 명시.

## 패턴 발산

| 항목           | run-1        | run-2        | run-3          |
| -------------- | ------------ | ------------ | -------------- |
| 세션 전략      | JWT          | JWT          | database       |
| adapter import | named import | named import | default import |
| 에러 페이지    | 없음         | 커스텀       | 없음           |

→ 제안: 세션 전략이 결정되지 않았음.
harness-step-2.md에 "세션 전략: JWT (database 세션 아님)" 추가.

→ 제안: run-2가 요청하지 않은 에러 페이지를 생성함.
harness-step-2.md "절대 금지"에 "커스텀 에러 페이지 생성" 추가.

## 네이밍 발산

| 항목             | run-1       | run-2      | run-3   |
| ---------------- | ----------- | ---------- | ------- |
| auth config 변수 | authOptions | authConfig | options |

→ 제안: harness-global.md에 "NextAuth 설정 변수명: authOptions" 추가.

## 발산 요약

- 구조: 1건 (파일 분리 차이)
- 패턴: 3건 (세션 전략, import 방식, 범위 초과)
- 네이밍: 1건 (변수명)
- 범위: 1건 (에러 페이지)

## 하네스 개선 제안 (6건)

위 제안 모두 적용 시 예상 수렴율: ~95%
```

### Step 단위 교정

`--step=M` 옵션으로 특정 Step만 교정할 수 있다. 전체 plan.sh를 처음부터 N번 돌리는 것은 비용이 크므로, 발산이 관찰된 Step만 집중 교정한다.

```bash
# Step 2만 3번 병렬 실행하여 발산 측정
/planosh-calibrate --step=2 --runs=3

# 하네스 강화 후 다시 교정하여 수렴 확인
/planosh-calibrate --step=2 --runs=3
```

이때 `--step=M`은 이전 Step들이 이미 완료된 상태의 worktree에서 시작한다. Step 1까지 완료된 커밋을 base로 사용하여 Step 2만 N번 실행한다.

### 교정 루프의 본질: 발산 탐지기

GCC 오라클 기법은 "정답을 아는 시스템"으로 AI의 오류를 검증한다. 교정 루프는 다르다. 정답을 찾으려는 것이 아니다.

교정 루프는 **발산 탐지기**다:

- 3번 실행해서 세션 전략이 JWT/JWT/database로 갈렸다 → 하네스에 세션 전략이 없었다는 신호
- 2번은 auth.ts 하나, 1번은 auth.ts + auth-options.ts → 파일 구조가 제약되지 않았다는 신호
- 변수명이 authOptions/authConfig/options로 갈렸다 → 네이밍 컨벤션이 누락됐다는 신호

발산이 발견되면 사람이 결정한다 ("JWT를 쓰자", "파일은 auth.ts 하나로"). 그 결정을 하네스에 기록하면, 해당 지점에서 더 이상 갈라지지 않는다. 교정 루프가 하는 일은 **결정이 필요한 지점을 빠짐없이 드러내는 것**이다.

### 비용과 수렴 속도

- 1회 교정: Step 1개 × 3회 실행 = API 호출 3번
- 일반적으로 2-3회 교정 루프면 수렴 (경험적 추정)
- 총 비용: Step당 6-9회 API 호출
- 이 비용은 "잘못된 구현을 디버깅하는 시간"보다 저렴하다

### `.plan/` 디렉토리 최종 구조

```
.plan/
├── harness-global.md           ← 글로벌 하네스 (프로젝트 컨벤션)
├── harness-step-1.md           ← Step 1 하네스
├── harness-step-2.md           ← Step 2 하네스 (교정 후 강화됨)
├── harness-step-3.md           ← Step 3 하네스
├── divergence-report.md        ← 최근 교정의 발산 리포트
└── calibration-history/        ← 교정 이력
    ├── step-2-run-1.log        ← 교정 실행 로그
    └── step-2-convergence.md   ← 수렴 추이 (67% → 95% → 100%)
```

## 제약 조건

- Claude Code 스킬 (독립 실행 CLI가 아님)
- v0: `/planosh`는 순차 실행만 (병렬은 `/planosh-calibrate`에서만 사용)
- v0: 화려한 UI 없이 터미널 출력만
- 대상 사용자: bash를 읽을 수 있는 개발팀 리드
- plan.sh는 사람이 읽을 수 있고, git diff가 가능하며, 편집 가능해야 함

## planosh가 하지 않는 것

planosh는 **PRD가 이미 존재한다고 가정**한다. PRD 작성, 아이디어 정리, 요구사항 도출은 planosh의 범위가 아니다.

PRD 작성에는 [gstack](https://github.com/ASDFGHoney/gstack)의 `/office-hours`(아이디어 검증, 브레인스토밍)와 `/plan-ceo-review`(제품 방향 리뷰) 스킬을 추천한다.

```
gstack                          planosh
─────────────────────           ─────────────────────
아이디어 → PRD 작성              PRD → plan.sh 생성
/office-hours                   /planosh
/plan-ceo-review                /planosh-calibrate
/plan-eng-review

"무엇을 만들지" 결정              "어떻게 실행할지" 결정
```

planosh는 PRD를 받아서 결정적인 실행 계획으로 변환하는 도구다. PRD의 품질이 plan.sh의 품질을 결정하므로, 좋은 PRD를 먼저 준비하는 것이 중요하다.

## 입력

`/planosh path/to/prd.md` 또는 `/planosh` (스킬이 PRD 경로를 프롬프트로 요청)

### 대화형 기술 결정 단계

스킬이 PRD를 읽은 후, 다음을 사용자와 대화형으로 결정:

- **기술 스택**: 프레임워크, DB, 인증, 스타일링, 테스트 프레임워크
- **아키텍처**: 디렉토리 구조, API 패턴, 데이터 모델 개요
- **범위 경계**: MVP에 포함/불포함, 명시적 비목표(non-goals)
- **검증 전략**: 각 Step의 검증 방법 (빌드, 테스트, curl)

사용자 승인 후 plan.sh 생성. 승인 없이 생성하지 않음.

## 오류 처리

- Step 실패 시 즉시 중단 (`set -euo pipefail`)
- 마지막 성공 Step 번호를 `.plan-state` 파일에 기록
- 실패 시 안내: `plan.sh --from N 으로 재시작하세요`
- 실패한 Step의 커밋은 dev 브랜치에 남아있음 (롤백하지 않음 — 디버깅 맥락 보존)
- `--from N` 실행 시 `.plan-state` 자동 읽기로 N 제안

## 전제 조건

1. 마크다운 명세 → 코드 변환 시 컨텍스트 유실은 개발팀의 실제 고통점이다
2. 셸 스크립트는 개발자 팀에게 적절한 리뷰 형식이다
3. Claude Code 스킬은 대상 사용자(개발팀 리드)에게 유효한 배포 수단이다
4. 핵심 가치는 "리뷰 가능한 실행 계획"이지, 인프라 자동화가 아니다

## 교차 모델 관점

독립적인 Claude 서브에이전트의 콜드 리딩 결과:

- **아직 고려되지 않은 가장 멋진 버전:** plan.sh를 git diff 가능한 실행 DAG로 만들고, `plan-status.json` 사이드카로 실시간 관측성을 부여. 엔지니어가 개별 Step의 프롬프트를 수정하는 plan.sh PR.
- **핵심 인사이트:** "리뷰 의식이 곧 제품이다. plan.sh PR이 곧 제품이다."
- **50% 기존 도구:** `just` (casey/just) 커맨드 러너 — 이미 "읽기 쉬운 실행 가능" 형식. 나머지 50%에 해당하는 구현: PRD→계획 생성기, 검증 하네스, 상태/재개 레이어, git 안무.
- **주말 빌드 우선순위:** 생성기 스킬 → 러너 (--dry, --from) → 실제 PRD로 E2E 테스트 → 패키지 + 데모 GIF. v0에서 병렬 실행은 건너뛰기.

## 검토한 접근법

### 접근법 A: 최소 구현 (채택)

스킬 2개:

- `.claude/skills/planosh.md` — PRD에서 plan.sh + 하네스를 생성하는 스킬
- `.claude/skills/planosh-calibrate.md` — plan.sh를 병렬 실행하여 발산을 탐지하고 하네스를 강화하는 스킬

생성 결과물:

- `plan.sh` — `--dry`, `--from N`, 검증, 체크포인트를 갖춘 실행 가능한 계획
- `.plan/harness-global.md` — 프로젝트 전체 컨벤션 하네스
- `.plan/harness-step-N.md` — 각 Step의 맥락 하네스
- `.plan/divergence-report.md` — 교정 시 생성되는 발산 리포트

범위:

**`/planosh`**:

- PRD 읽기 → 기술 결정 (대화형) → Step 분해 → plan.sh + 하네스 생성
- 각 Step: `claude -p "명세" --append-system-prompt "하네스"` + 검증 + git 커밋
- `--dry` 모드 (실행 없이 프롬프트만 출력)
- `--from N` 모드 (Step N부터 재개)
- dev 브랜치에 자동 커밋
- 실행 전 plan.sh 수동 리뷰

**`/planosh-calibrate`**:

- 격리된 worktree에서 plan.sh의 특정 Step을 N번 병렬 실행
- 결과물 간 발산 분석 (구조/패턴/네이밍/범위)
- 발산 지점마다 사용자에게 결정 요청
- 사용자 결정을 하네스 규칙으로 추가
- 발산 리포트 생성

제외: plan-status.json, 팀 리뷰 UI

공수: M (2-4일, 반복 테스트 포함)
리스크: 낮음

### 구체적인 plan.sh 예시 (3-Step 스켈레톤, git 전략 등은 예시일 뿐 강제하지 않음)

```bash
#!/bin/bash
# 계획: 팀 회고 앱 — 스프린트 1
# 생성: 2026-04-07 by /planosh
# PRD: docs/PRD.md
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_ROOT"

DRY_RUN=false; START_FROM=1
for arg in "$@"; do
  case $arg in
    --dry) DRY_RUN=true ;;
    --from=*) START_FROM="${arg#*=}" ;;
  esac
done

# ── 하네스 경로 ──
HARNESS_DIR="$PROJECT_ROOT/.plan"
GLOBAL_HARNESS="$HARNESS_DIR/harness-global.md"

step() { local n=$1 name=$2
  [ "$n" -lt "$START_FROM" ] && echo "⏭️ Step $n: $name (건너뜀)" && return 0
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
    echo "[DRY] 프롬프트:"; echo "$prompt" | head -5; echo "..."
    echo "[DRY] 하네스: $GLOBAL_HARNESS + $step_harness"
    return 0
  fi

  claude -p "$prompt" \
    --append-system-prompt "$harness" \
    --dangerously-skip-permissions
}
verify() { $DRY_RUN && echo "[DRY] 검증: $1" && return 0; echo "🔍 $1"; eval "$2" || { echo "❌ $1"; echo "$CURRENT_STEP" > .plan-state; exit 1; }; echo "✅ $1"; }
checkpoint() {
  $DRY_RUN && return 0
  local step_harness="$HARNESS_DIR/harness-step-${CURRENT_STEP}.md"
  # Step별 하네스에 파일 목록이 있으면 선택적 git add
  if [ -f "$step_harness" ] && grep -q '생성할 파일 목록' "$step_harness"; then
    # 목록 외 변경 파일 경고
    local unexpected=$(git diff --name-only | grep -v -f <(grep '^\- ' "$step_harness" | sed 's/^- //') 2>/dev/null || true)
    [ -n "$unexpected" ] && echo "⚠️ 범위 외 변경 감지: $unexpected"
  fi
  git add -A && git commit -m "$1"
}

[ "$START_FROM" -eq 1 ] && ! $DRY_RUN && git checkout -b dev-$(date +%Y%m%d) main 2>/dev/null || true

# ── Step 1: 프로젝트 스캐폴딩 ──
# 하네스: .plan/harness-global.md + .plan/harness-step-1.md
CURRENT_STEP=1; step 1 "프로젝트 스캐폴딩"
run_claude "
Next.js 14 프로젝트를 생성하세요.
## 만들 것
- npx create-next-app@latest . --ts --tailwind --app --src-dir
- Prisma + NextAuth + Playwright 설치
- .env.example 생성
## 하지 않을 것
- 어떤 페이지나 기능도 구현하지 않음. 스캐폴딩만.
"
verify "빌드 성공" "npm run build"
verify "prisma 존재" "[ -f prisma/schema.prisma ]"
checkpoint "chore: project scaffolding"

# ── Step 2: Google OAuth ──
# 하네스: .plan/harness-global.md + .plan/harness-step-2.md
# Step별 하네스에 Step 1 산출물 목록, NextAuth 아키텍처 제약,
# 생성 허용 파일 목록이 포함되어 해의 공간을 압축
CURRENT_STEP=2; step 2 "Google OAuth 로그인"
run_claude "
Google OAuth 로그인을 구현하세요.
## 만들 것
- User/Account/Session 모델 (prisma)
- NextAuth Google Provider + Prisma Adapter
- /auth/signin 페이지 (Google 로그인 버튼)
- /dashboard 페이지 (빈 셸, 환영 텍스트)
- 미인증 → /auth/signin 리다이렉트 미들웨어
## 하지 않을 것
- 이메일/비밀번호 가입 (D-002)
- 프로필 편집, 팀 기능
"
verify "빌드 성공" "npm run build"
verify "로그인 페이지 존재" "[ -f src/app/auth/signin/page.tsx ]"
checkpoint "feat: Google OAuth login"

# ── Step 3: 팀 생성 ──
# 하네스: .plan/harness-global.md + .plan/harness-step-3.md
# Step별 하네스에 Step 1-2 산출물 목록, Team 모델 스키마 제약 포함
CURRENT_STEP=3; step 3 "팀 생성"
run_claude "
팀 생성 기능을 구현하세요.
## 만들 것
- Team, TeamMember 모델 (prisma)
- POST /api/teams API
- /teams/new 페이지 (팀 이름 입력)
- /teams/:id 페이지 (팀 대시보드)
- /dashboard 수정 (팀 목록 + 생성 버튼)
## 하지 않을 것
- 팀원 초대 (스프린트 2)
- 팀 설정 변경
"
verify "빌드 성공" "npm run build"
verify "Team 모델 존재" "grep -q 'model Team' prisma/schema.prisma"
checkpoint "feat: team creation"

echo ""; echo "🎉 계획 완료! 브랜치: $(git branch --show-current)"
echo "다음: gh pr create --base main --head $(git branch --show-current)"
```

### 프롬프트 + 하네스 분리 전략

프롬프트(-p)와 하네스(--append-system-prompt)의 역할을 명확히 분리:

**프롬프트 (-p)** — WHAT: 이 Step에서 만들 것

- "만들 것" — 이 Step의 구체적 산출물만
- "하지 않을 것" — 이 Step에서 특별히 주의할 범위 외 항목만

**글로벌 하네스** — HOW (불변): 프로젝트 전체 규칙

- 기술 결정 (D-001, D-002 ...)
- 코딩 컨벤션 (네이밍, import 순서, 패턴)
- 절대 금지 패턴

**Step별 하네스** — HOW (가변): 이 Step의 맥락

- 이전 Step 산출물 목록 (에이전틱 건망증 방지)
- 이 Step의 아키텍처 제약
- 생성 허용 파일 목록 (화이트리스트)

이 분리의 장점:

1. 프롬프트가 짧아진다 — 기술 결정, 선행 조건을 하네스로 이동
2. 글로벌 하네스는 모든 Step에서 재사용 — 일관성 보장
3. Step별 하네스만 교체하면 같은 프롬프트로 다른 맥락에 적용 가능
4. `--dry` 모드에서 하네스 파일 경로를 출력하여 리뷰 가능

### 검증 전략

plan.sh의 verify는 exit code로 성공/실패를 판정할 수 있는 검증이라면 자유롭게 사용한다:

- **빌드**: `npm run build`, `cargo build` 등
- **파일 존재**: `[ -f path ]`
- **API 응답**: `curl -sf http://localhost:3000/api/health`
- **DB 스키마**: `grep -q 'model X' prisma/schema.prisma`
- **유닛 테스트**: `npm test`, `pytest`
- **E2E 테스트**: Playwright, Cypress 등 — 특정 요소의 존재, 네비게이션 동작 등 결정적으로 판정 가능한 시나리오
- **타입 체크**: `npx tsc --noEmit`
- **린트**: `npm run lint`
- **검증 생략**: 타입 정의만 추가하는 등 독립 검증이 무의미한 경우

verify에 넣을지 판단하는 기준은 하나다: **같은 입력에 항상 같은 판정(pass/fail)을 내리는가.** `npm run build`는 된다. `playwright test --grep "로그인 버튼 존재"`도 된다. 단, "스크린샷이 예쁜가"를 AI에게 판정시키는 것은 안 된다 — 그것은 verify가 아니라 사람의 리뷰 영역이다.

### Git 전략

planosh는 git 전략을 강제하지 않는다. 브랜치 네이밍, 커밋 메시지 컨벤션, checkpoint 방식 등은 사용자가 plan.sh를 생성할 때 자유롭게 결정한다.

planosh가 git과 관련하여 유일하게 강제하는 것은 `/planosh-calibrate`에서 **worktree로 격리된 환경을 만들어 병렬 실행**하는 부분뿐이다.

## 미해결 질문

1. ~~**네이밍:** "planosh"로 확정~~
2. **plan.sh 이식성:** plan.sh가 다른 LLM CLI에서도 동작해야 하는가 (`claude -p`만이 아닌)? v0는 Claude 전용이지만, 형식 자체는 어댑터 친화적일 수 있음.
3. **검증 깊이:** 검증을 어디까지 할 것인가? 빌드만? 빌드 + 테스트? 빌드 + 테스트 + curl? 스킬이 사용자에게 물어야 하는가?
4. **Step 세분화:** Step의 적절한 크기는? 스킬에 휴리스틱이 필요 (30분, 5-10개 파일, 단일 책임).
5. **교정 루프 수렴 비보장:** 하네스에 규칙을 추가하면 프롬프트 맥락이 바뀌어 기존에 수렴했던 항목이 새로 발산할 수 있다. "2-3회 교정 루프면 수렴"은 경험적 추정이며, 수렴이 보장되지 않는 루프다. 규칙 추가 후 기존 수렴 항목의 회귀 테스트가 필요할 수 있음. 실제 실험으로 검증 필요.

## 성공 기준

- PRD를 가진 개발자가 `/planosh`를 실행하면 10분 이내에 plan.sh를 얻을 수 있다
- 생성된 plan.sh는 읽기 쉽다 (다른 개발자가 각 단계를 이해할 수 있다)
- `bash plan.sh` 실행 시 각 Step의 verify 명령이 exit 0을 반환한다
- `bash plan.sh --from N`이 실패 후 Step N부터 성공적으로 재개한다
- `/planosh-calibrate --step=N --runs=3` 실행 시 발산 지점이 사용자에게 제시되고 하네스가 강화된다
- 교정 루프 2-3회 반복 후 동일 Step의 발산율이 10% 이하로 감소한다
- plan.sh와 `.plan/` 하네스를 git에 커밋하고 PR로 리뷰할 수 있다

## 배포 계획

- Claude Code 스킬 2개: `.claude/skills/planosh.md` + `.claude/skills/planosh-calibrate.md`를 프로젝트에 복사하여 설치
- GitHub에 오픈소스 레포 + README + 데모
- v0: 수동 복사. v1: gstack 방식의 스킬 설치 고려

## 다음 단계

1. **planosh 스킬 2개 구현** — `/planosh` (생성)과 `/planosh-calibrate` (교정)
2. **오픈소스 레포 생성** — 깔끔한 README, 설치 안내, 스킬 파일
3. **데모 제작** — 실제 작은 프로젝트(인증 + CRUD REST API)에 대해 plan.sh 생성
4. **종단간 테스트** — 생성된 plan.sh를 실행하고 무엇이 깨지는지 확인
5. **30초 데모 GIF 촬영** — PRD 입력 → plan.sh 생성 → 리뷰 → 실행 → 동작하는 앱

## 사고방식에 대한 관찰

- 이 세션에서 대규모 리서치 탐구(SDD, BMAD, Carlini의 컴파일러, 컨텍스트 엔지니어링)로 시작해서, 복잡한 멀티파일 하네스 시스템을 만들다가, "이건 너무 복잡해"라고 말하고 하나의 우아한 개념으로 전부 압축했다. 그것이 취향(taste) — 뺄 때를 아는 것.
- "이 마크다운 문서가 코드로 변환되는게 생각보다 결정적이지 않아" — 이것이 전체 세션에서 가장 명확한 문제 정의다. "더 나은 명세가 필요하다"가 아니라 "명세→코드 변환이 비결정적이다"라고 프레이밍했다. 그것이 핵심 통찰.
- 10배 비전으로 "더 빠른 실행"이나 "더 많은 자동화"가 아닌 "팀 리뷰 워크플로우"를 선택했다. 도구만 생각하는 게 아니라 팀 프로세스를 생각하고 있다. 엔지니어링이 아닌 제품적 사고.
- 접근법 A vs B vs C가 제시됐을 때 주저 없이 가장 작은 범위를 선택했다. 데모를 위해 만드는 게 아니라, 아이디어를 검증하기 위해 만들고 있다.
