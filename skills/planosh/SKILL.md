---
name: planosh
description: PRD를 결정적인 plan.sh + 하네스로 변환. PRD 입력 → 대화형 기술 결정 → Step 분해 → plan.sh 생성. "plan 만들어줘", "PRD를 plan.sh로", "/planosh" 등의 요청에 사용.
---

# /planosh — PRD를 결정적인 plan.sh로 변환

PRD를 입력받아 대화형으로 기술 결정을 하고, 실행 가능한 plan.sh + 하네스를 생성한다.

```
/planosh path/to/prd.md

입력: PRD (마크다운)
과정: 입구 검증 → PRD 분석 → 기술 결정 인터뷰 → Step 분해 → plan.sh + 하네스 생성
출력: .plan/{plan-name}/plan.sh + steps.json + steps/*.md + harness-for-plan.md
```

---

## Phase 0: 입구 조건 검증

**Phase 0을 통과하지 않으면 Phase 1로 진행하지 않는다.**

planosh는 "뭘 만들 것인가"는 묻지 않는다. "어떻게 만들 것인가"만 다룬다. Phase 0은 이 전제를 강제하는 게이트다.

### 0-1. PRD 검증 (자동 거부)

사용자가 `/planosh path/to/prd.md`로 호출하면 해당 파일을 읽는다.
경로 없이 `/planosh`만 호출하면 PRD 경로를 먼저 물어본다. PRD 없이 진행하지 않는다.

PRD 파일을 읽고 다음을 검증한다:

- [ ] 파일이 존재하는가
- [ ] 마크다운 형식인가
- [ ] 최소 300자 이상인가
- [ ] 기능/요구사항을 서술하는 내용이 있는가

**하나라도 불충족 시 거부한다:**

```
PRD가 충분하지 않습니다.

planosh는 완성된 PRD를 입력으로 받습니다.
PRD에 최소한 다음이 포함되어야 합니다:
- 제품 설명
- 핵심 기능 목록
- 대상 사용자

PRD를 보완한 후 다시 실행하세요.
```

거부 후 진행하지 않는다.

### 0-2. 팀 합의 확인 (사용자에게 질문)

비기술 결정의 완료 여부는 파일로 검증할 수 없다. 사용자에게 명시적으로 확인한다:

```
planosh는 기술 결정만 돕습니다.
다음 항목이 팀 내에서 이미 합의되었는지 확인합니다:

  - [ ] 제품 요구사항이 확정되었는가 (기능 범위, 우선순위)
  - [ ] UX 흐름이 결정되었는가 (와이어프레임, 사용자 여정)
  - [ ] 비기술적 제약이 있다면 PRD에 반영되었는가

위 항목이 확정되지 않은 상태에서 기술 결정을 진행하면,
나중에 PRD 변경 시 plan.sh를 처음부터 다시 만들어야 합니다.

진행하시겠습니까? (Y/아직 준비 안 됨)
```

"아직 준비 안 됨" 선택 시:

```
이해합니다. 팀 합의가 끝나면 다시 /planosh를 실행하세요.
```

진행하지 않는다.

### 0-3. 기존 plan 충돌 확인

`.plan/` 디렉토리에 같은 이름의 plan이 이미 있으면:

```
.plan/{name}/이 이미 존재합니다.

  A. 기존 plan을 덮어쓰기 (이전 harness 삭제)
  B. 다른 이름으로 생성
  C. 중단

선택하세요 (A/B/C):
```

---

## Phase 1: PRD 읽기

Phase 0을 통과한 PRD를 읽고, 핵심을 3줄로 요약하여 사용자에게 확인한다:

```
PRD 요약:
- 제품: (한 줄)
- 핵심 기능: (한 줄)
- 대상 사용자: (한 줄)

이 PRD로 plan.sh를 생성할까요?
```

사용자가 확인하면 plan 이름을 결정한다. 이 이름이 `.plan/{name}/` 디렉토리명과 `PLAN_NAME` 변수가 된다. PRD 제목이나 기능명에서 slug를 추출하여 기본값으로 제안한다 (예: "팀 회고 앱 — 스프린트 1" → `retro-sprint1`).

하나의 프로젝트에 여러 plan이 공존할 수 있으므로, 이미 `.plan/` 안에 다른 plan 폴더가 있으면 목록을 보여준다.

---

## Phase 2: 기술 결정 인터뷰

3단계로 진행한다. 한번에 물어보지 말고, 한 항목씩 순서대로 질문한다.

```
Phase 2-1: PRD 스캔 → 결정 항목 자동 추출
Phase 2-2: 항목별 인터뷰 (Mechanical 자동 / Taste 대화)
Phase 2-3: 결정 확정 → harness-for-plan.md 잠금
```

### 2-1. PRD 스캔 — 결정 항목 추출

PRD를 읽고 **이 프로젝트에서 필요한 기술 결정 목록**을 도출한다. 고정된 질문 리스트가 아니라, PRD 내용에 따라 질문이 달라진다.

#### 기본 항목 (항상 포함)

모든 프로젝트에서 결정해야 하는 것:

- 프레임워크 + 메타 프레임워크
- 언어 + strict mode
- 데이터베이스 + ORM
- 스타일링
- 테스트 전략
- 배포 타겟

#### 도메인 프로브 (PRD 키워드로 활성화)

PRD에서 키워드/기능을 감지하면 해당 도메인의 기술 결정 항목을 추가한다:

| PRD 키워드 | 활성화되는 기술 결정 |
|---|---|
| "로그인", "인증", "회원", "OAuth" | 인증 전략, 세션 관리, OAuth 프로바이더 |
| "실시간", "동기화", "라이브", "채팅" | 실시간 통신 방식 (WebSocket/SSE/Polling) |
| "검색", "필터", "쿼리" | 검색 엔진 (DB-level/FTS/외부) |
| "파일 업로드", "이미지", "미디어" | 스토리지 (S3/R2/local), 이미지 처리 |
| "결제", "구독", "요금" | 결제 프로바이더, 과금 모델 |
| "다국어", "i18n" | 국제화 전략, 라이브러리 |
| "이메일", "알림 발송" | 메일 서비스 |
| "에디터", "마크다운", "WYSIWYG" | 에디터 라이브러리 |
| "API", "외부 연동" | API 패턴 (REST/GraphQL/tRPC) |
| "대시보드", "차트" | 차트 라이브러리, 데이터 집계 |
| "모바일", "반응형" | 네이티브/웹앱/PWA |

이 테이블에 없는 도메인이라도 PRD에서 기술 결정이 필요한 항목이 감지되면 추가한다.

#### Brownfield 자동 감지

기존 코드베이스가 있으면 프로브 전에 자동 감지를 먼저 수행한다:

```
package.json → 프레임워크/언어 감지
tsconfig.json → TypeScript strict 여부
prisma/schema.prisma → ORM = Prisma
tailwind.config.* → 스타일링 = Tailwind
vitest.config.* / jest.config.* → 테스트 프레임워크
fly.toml / vercel.json / render.yaml → 배포 타겟
```

자동 감지된 항목은 확인만 받고 넘어간다. 새로 결정해야 하는 항목만 인터뷰한다.

#### 출력

추출 결과를 사용자에게 보여준다:

```
이 PRD에서 필요한 기술 결정 N개를 추출했습니다.

기본 항목 (6):
  ▪ 프레임워크       — 미정
  ▪ 언어             — 미정
  ...

PRD에서 감지된 항목 (N):
  ▪ 인증 전략         ← "Google 로그인" 언급
  ▪ 실시간 통신       ← "실시간 동기화" 언급
  ...

자동 감지 (N):
  ✓ ORM = Prisma      ← prisma/schema.prisma 존재
  ...

자동 감지 항목을 확인해주세요. 이어서 미정 항목을 하나씩 결정합니다.
```

### 2-2. 항목별 인터뷰

각 미정 항목에 대해 **한 번에 하나씩** 질문한다. 항목의 성격에 따라 두 가지 모드가 있다.

#### Mechanical vs Taste 분류

```
Mechanical (자동 결정):
  하나의 선택이 명백히 우세하거나, 이전 결정이 답을 강제함.
  → AI가 결정 + 근거를 보여주고, 사용자는 확인만.

Taste (사용자 판단 필요):
  합리적인 사람들이 의견 차이를 보일 수 있음.
  → 트레이드오프 테이블을 보여주고 사용자가 선택.
```

판별 기준:

| 조건 | 분류 |
|---|---|
| PRD에 명시적 언급 ("React로 만든다") | Mechanical |
| 기존 코드에서 감지 | Mechanical |
| 상위 결정이 하위를 강제 | Mechanical |
| 선택지 간 유의미한 트레이드오프 | Taste |
| 팀/개인 선호에 의존 | Taste |

이전 결정으로 선택지가 1개로 좁혀지면 Taste → Mechanical로 자동 전환한다.

#### Taste 항목의 인터뷰 형식

트레이드오프 테이블 + 추천을 제시한다:

**D-003: 데이터베이스**

PRD 근거: "실시간 동기화", "다중 사용자 보드" → 동시성 필요

| Option | 장점 | 단점 | Layer |
|--------|------|------|-------|
| A. SQLite | 배포 단순, 설정 0 | 동시 쓰기 제한 | Layer 1 |
| B. PostgreSQL | 동시성, 확장성 | 운영 부담 | Layer 1 |
| C. Supabase | 실시간 내장, Auth | 벤더 종속 | Layer 2 |

> Layer: 1 = 검증된 기존 접근법, 2 = 부상 중, 3 = 실험적

추천: B (PostgreSQL)
이유: 동시성 요구사항 + Layer 1 안정성.

선택하세요 (A/B/C/기타):

PRD에서 해당 결정과 관련된 근거를 반드시 인용한다.
각 옵션에 Layer를 태깅한다 (Layer 1 = 검증됨, Layer 2 = 부상 중, Layer 3 = 실험적).
추천과 이유를 항상 제시한다.
필요하면 추가 질문으로 맥락을 좁힌다.

#### Mechanical 항목의 인터뷰 형식

**D-007: API 패턴**

자동 결정: Server Actions
근거: D-001에서 Next.js 14 App Router 선택 → Server Actions가 표준 패턴.
      PRD에 외부 API 클라이언트 언급 없음.

확인하시겠습니까? (Y/변경):

#### 결정 간 의존성

앞선 결정이 뒤 결정의 선택지를 좁힌다. 이 의존성을 자동으로 반영한다:

```
D-001: Next.js 선택
  → D-005 스타일링: Tailwind 추천도 상승
  → D-007 API 패턴: Server Actions가 Mechanical로 승격
  → D-011 배포: Vercel이 Mechanical로 승격
```

#### NEEDS CLARIFICATION

PRD에 정보가 부족하고, AI도 추천할 수 없고, 사용자도 "모르겠다"고 답하면:

**D-008: 검색 전략**

상태: NEEDS CLARIFICATION

이유: PRD에 "카드 검색" 언급이 있지만, 검색 대상 규모를 판단할 정보 부족.
- 100개 이하 → DB LIKE 쿼리
- 1,000~10,000개 → PostgreSQL FTS
- 10,000개 이상 → 외부 검색 엔진

카드가 최대 몇 개까지 늘어날 것 같으세요?

사용자가 답하면 결정으로 전환한다.
답하지 않으면 보수적 기본값(가장 단순한 옵션)으로 잠정 결정하되, harness에 `# PROVISIONAL` 태그를 붙인다.

### 2-3. 결정 확정

모든 항목이 결정되면 한 번에 요약하고 최종 확인을 받는다:

**기술 결정 요약 (N개)**

| ID | 항목 | 결정 | 분류 | 상태 |
|----|------|------|------|------|
| D-001 | 프레임워크 | Next.js 14 App Router | Taste | 확정 |
| D-002 | 언어 | TypeScript strict | Mechanical | 확정 |
| D-003 | 데이터베이스 | PostgreSQL | Taste | 확정 |
| ... | | | | |
| D-007 | 인증 | NextAuth + Google OAuth | Taste | 확정 |
| ... | | | | |
| D-012 | ORM | Prisma (기존) | 감지 | 확인됨 |

NEEDS CLARIFICATION: 없음 (또는 항목 나열)
PROVISIONAL: 없음 (또는 항목 나열)

이 결정들로 harness-for-plan.md를 생성합니다.
확정하시겠습니까? (Y/수정할 항목 번호):

사용자가 수정을 요청하면 해당 항목만 다시 질문한다.
확정 없이 Phase 3으로 진행하지 않는다.

### 2-4. 아키텍처

기술 결정이 확정된 후, 확정된 결정을 기반으로:

- **디렉토리 구조**: 프레임워크 표준 구조를 기본값으로 제안
- **데이터 모델 개요**: 핵심 모델 3-5개와 관계를 ASCII 다이어그램으로

### 2-5. 범위 경계

- **MVP에 포함**: PRD에서 도출한 핵심 기능 목록
- **MVP에 불포함 (비목표)**: 명시적으로 제외할 것

### 2-6. 검증 전략

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

---

## Phase 3: Step 분해

PRD의 기능을 Step으로 분해한다.

### 핵심 원칙: 에이전트 작업과 사람 작업의 분리

**plan.sh에는 에이전트(claude -p)가 자율적으로 완료할 수 있는 작업만 넣는다.**

사람의 판단이나 행동이 필요한 작업은 plan.sh에 넣지 않고, 별도의 `plan-for-human.md`로 분리한다.

에이전트가 할 수 있는 것:
- 코드 생성, 수정, 삭제
- 빌드, 테스트, 린트 실행
- 파일 구조 생성
- 설정 파일 작성

사람이 해야 하는 것:
- 외부 서비스 설정 (API 키 발급, DB 프로비저닝, 도메인 설정)
- 디자인 리뷰, UX 검증
- 수동 테스트, 브라우저 확인
- 배포, 모니터링 설정
- 비즈니스 판단이 필요한 결정

### plan.sh 분리 규칙

사람 작업이 에이전트 작업 사이에 끼어야 하는 경우, plan.sh를 여러 개로 나눈다:

```
plan-1.sh  (에이전트: 스캐폴딩 ~ API)
  ↓
[사람: API 키 설정, 디자인 리뷰]
  ↓
plan-2.sh  (에이전트: 프론트엔드 ~ 테스트)
```

사람 작업이 모두 시작 전이나 끝 후에만 있으면, plan.sh는 하나로 유지한다.

### Step 크기 휴리스틱

- 단일 책임: 하나의 Step은 하나의 기능 단위
- 5-10개 파일 변경이 적정 범위
- Step 간 의존성은 순방향만 (Step N은 Step 1..N-1에만 의존)
- 첫 Step은 항상 스캐폴딩 (프로젝트 초기화)
- 마지막 Step은 마무리 (최종 검증, cleanup)

### Step별 정보

각 Step에 대해 다음을 정의한다:

- **이름**: 한 줄 설명
- **실행자**: 에이전트 또는 사람
- **만들 것**: 구체적 산출물 목록 (에이전트 Step만). 관련 결정 ID를 괄호로 표기.
- **하지 않을 것**: 이 Step에서 명시적으로 제외할 것 (에이전트 Step만). 관련 결정 ID를 괄호로 표기.
- **검증**: verify에 사용할 명령어 (에이전트 Step만)
- **커밋 메시지**: checkpoint에 사용할 메시지 (에이전트 Step만)
- **할 일**: 사람이 수행할 체크리스트 (사람 Step만)

참고: 생성할 파일 목록, 아키텍처 제약 등 step 수준의 하네스 정보는 초기 생성 시 포함하지 않는다. `/planosh-calibrate`에서 발산이 발견되면 해당 step 프롬프트에 직접 추가된다.

### 사용자 승인

Step 분해 결과를 사용자에게 보여주고 승인을 받는다.

```
Plan 구조:

plan-for-human.md (사전 작업):
  사람: 외부 서비스 API 키 발급
  사람: 디자인 시안 확정

plan.sh (에이전트 Step 1-6):
  Step 1: 프로젝트 스캐폴딩
     검증: npm run build
  Step 2: DB 스키마 + API (D-003, D-004)
     검증: npx tsc --noEmit
  Step 3: 인증 (D-007)
     검증: npm run build
  Step 4: 프론트엔드 기본 UI (D-005)
     검증: npm run build
  Step 5: 실시간 동기화 (D-008)
     검증: npm run build
  Step 6: 테스트 + 최종 정리 (D-006)
     검증: npm run build && npm test

plan-for-human.md (사후 작업):
  사람: 배포 환경 설정
  사람: QA 테스트

이 구조로 생성할까요?
```

사용자가 수정을 요청하면 반영한 후 다시 승인을 받는다.
승인 없이 파일을 생성하지 않는다.

### 실행 옵션

Step 구조가 승인되면, plan.sh의 기본 실행 옵션을 설정한다.

```
실행 옵션을 설정합니다.

모델: plan.sh 실행 시 사용할 Claude 모델
  A. opus   — 최고 품질, 느림
  B. sonnet — 균형 (기본값)
  C. haiku  — 빠름, 저렴

선택하세요 (A/B/C) [기본: B]:
```

사용자 선택 후 effort를 묻는다:

```
추론 노력 수준:
  1. low    — 최소 추론, 단순 작업용
  2. medium — 적절한 균형
  3. high   — 깊은 추론 (기본값)
  4. max    — 최대 추론, 복잡한 작업용
  5. auto   — 모델이 자동 판단

선택하세요 (1-5) [기본: 3]:
```

선택 결과를 plan.sh의 `DEFAULT_MODEL`과 `DEFAULT_EFFORT` 변수에 반영한다.

이 옵션은 plan.sh의 기본값일 뿐이다. 실행 시 `--model=X`, `--effort=X` 플래그로 오버라이드할 수 있다.

---

## Phase 4: 파일 생성

승인을 받으면 다음 파일들을 생성한다.

### 4-0. `.plan/{plan-name}/plan-for-human.md`

사람이 수행해야 할 작업을 시간순으로 정리한다.

```markdown
# {PRD 제목} — 사람 작업 체크리스트

## 사전 작업 (plan.sh 실행 전)

- [ ] {사전 작업 1}
- [ ] {사전 작업 2}

## 중간 작업 (plan-1.sh 실행 후, plan-2.sh 실행 전)

> plan-1.sh가 완료된 후 아래 작업을 수행하고, plan-2.sh를 실행하세요.

- [ ] {중간 작업 1}
- [ ] {중간 작업 2}

## 사후 작업 (모든 plan.sh 완료 후)

- [ ] {사후 작업 1}
- [ ] {사후 작업 2}
```

사람 작업이 없는 섹션은 생략한다. 중간 작업이 없으면 plan.sh를 나누지 않으므로 "중간 작업" 섹션도 생략한다.

### 4-1. `.plan/{plan-name}/harness-for-plan.md`

Phase 2에서 확정된 기술 결정을 두 구간으로 분리하여 기록한다:

```markdown
# 프로젝트 컨벤션

## 기술 결정 (locked — 변경 불가)

- D-001: {결정 내용}
- D-002: {결정 내용}
...

## AI 재량 (discretion — 발산 허용)

다음 영역은 AI가 자유롭게 판단한다:
- 컴포넌트 내부 구현 패턴
- 에러 메시지 문구
- 유틸리티 함수 내부 구현

## 코딩 규칙

{기술 결정에서 파생되는 코딩 컨벤션. 각 규칙에 근거 결정 ID를 괄호로 표기.}

## 절대 금지

{기술 결정에서 파생되는 금지 패턴. 각 금지에 근거 결정 ID를 괄호로 표기.}
- any 타입 사용 (D-002)
- 하드코딩된 시크릿
- 이 Step의 범위 밖 파일 수정
```

PROVISIONAL 항목이 있으면 해당 결정 옆에 `# PROVISIONAL — 추후 변경 가능` 태그를 붙인다.

### 4-2. `.plan/{plan-name}/steps.json`

step 매니페스트. plan.sh가 이 파일을 읽어 실행한다. 사람도 이 파일로 plan의 전체 구조를 한눈에 파악한다.

```json
{
  "plan_name": "{plan-name}",
  "prd": "{PRD 파일 경로}",
  "created": "{날짜}",
  "steps": [
    {
      "id": 1,
      "name": "{Step 이름}",
      "prompt": "step-1.md",
      "verify": [
        { "name": "{검증 이름}", "run": "{검증 명령}" }
      ],
      "commit": "{커밋 메시지}"
    },
    {
      "id": 2,
      "name": "{Step 이름}",
      "prompt": "step-2.md",
      "verify": [
        { "name": "{검증 이름}", "run": "{검증 명령}" },
        { "name": "{검증 이름}", "run": "{검증 명령}" }
      ],
      "commit": "{커밋 메시지}"
    }
  ]
}
```

- `prompt` 값은 `steps/` 디렉토리 기준 상대 경로다.
- `verify` 배열은 순서대로 실행된다. 하나라도 실패하면 중단.
- 사람 step은 steps.json에 포함하지 않는다 (plan-for-human.md에만).

### 4-3. `.plan/{plan-name}/steps/step-N.md`

각 Step의 프롬프트를 개별 마크다운 파일로 생성한다. plan.sh에서 `-p` 인자로 전달되는 내용이다.

```markdown
{한 줄 지시}

## 만들 것
- {산출물 1} (D-00X)
- {산출물 2} (D-00Y)

## 하지 않을 것
- {제외 항목 1} (D-00Z)
```

프롬프트에는 WHAT만 넣는다:
- "만들 것" — 이 Step의 구체적 산출물. 관련 결정 ID를 괄호로 표기.
- "하지 않을 것" — 이 Step의 범위 외 항목. 관련 결정 ID를 괄호로 표기.

HOW(기술 결정, 코딩 컨벤션)는 harness-for-plan.md에 있으므로 프롬프트에 반복하지 않는다.
프롬프트가 짧을수록 좋다.

참고: `/planosh-calibrate`가 발산을 발견하면, 이 파일의 끝에 `## 아키텍처 제약`이나 `## 생성할 파일 목록` 섹션이 자동으로 추가된다. 초기 생성 시에는 "만들 것"과 "하지 않을 것"만 포함한다.

### 4-4. `.plan/{plan-name}/plan.sh` (또는 `plan-1.sh`, `plan-2.sh`, ...)

사람 작업이 에이전트 작업 사이에 끼지 않으면 `plan.sh` 하나로 생성한다.
사람 작업이 중간에 필요하면 `plan-1.sh`, `plan-2.sh`, ... 로 나눈다.

**plan.sh는 `.plan/{plan-name}/` 디렉토리 안에 생성한다.** steps.json, steps/, harness-for-plan.md와 같은 위치에 둔다.

각 plan 파일에는 **에이전트가 자율적으로 완료할 수 있는 Step만** 포함한다. 사람의 판단, 외부 서비스 설정, 수동 테스트 등은 절대 plan.sh에 넣지 않는다.

plan.sh는 **범용 러너**다. step-specific 코드를 포함하지 않는다. steps.json을 읽어 루프로 실행한다.

아래 템플릿을 **그대로** 사용한다. `{plan-name}` 부분만 치환한다.

```bash
#!/bin/bash
# 계획: {PRD 제목}
# 생성: {날짜} by /planosh
# PRD: {PRD 파일 경로}
# 사람 작업: .plan/{plan-name}/plan-for-human.md 참조
#
# 사용법:
#   bash .plan/{plan-name}/plan.sh                  전체 실행
#   bash .plan/{plan-name}/plan.sh --dry            실행 없이 프롬프트만 출력
#   bash .plan/{plan-name}/plan.sh --from=N         Step N부터 재개
#   bash .plan/{plan-name}/plan.sh --to=M           Step M까지만 실행
#   bash .plan/{plan-name}/plan.sh --from=N --to=M  Step N~M만 실행
#   bash .plan/{plan-name}/plan.sh --model=opus    모델 오버라이드
#   bash .plan/{plan-name}/plan.sh --effort=max    effort 오버라이드
#   bash .plan/{plan-name}/plan.sh --testbed       calibrate용 경량 모드 (Haiku, low)
#
# 주의: --dangerously-skip-permissions를 사용합니다.
# 반드시 steps.json + steps/*.md를 리뷰한 후 실행하세요.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

DRY_RUN=false; START_FROM=1; STOP_AFTER=999; TESTBED=false
DEFAULT_MODEL="{chosen-model}"; DEFAULT_EFFORT="{chosen-effort}"
MODEL="$DEFAULT_MODEL"; EFFORT="$DEFAULT_EFFORT"
[ -f "$SCRIPT_DIR/.plan-state" ] && START_FROM=$(cat "$SCRIPT_DIR/.plan-state")
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
  eval "$2" || { echo "FAIL: $1"; echo "$CURRENT_STEP" > "$SCRIPT_DIR/.plan-state"; exit 1; }
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
```

plan.sh가 분리되는 경우(`plan-1.sh`, `plan-2.sh`), steps.json도 `steps-1.json`, `steps-2.json`으로 분리한다. 각 plan 파일의 `STEPS_FILE` 변수가 대응하는 steps JSON을 가리킨다.

---

## Phase 5: 최종 안내

파일 생성이 완료되면 사용자에게 안내한다.

plan.sh가 하나인 경우:

```
생성 완료:
  .plan/{plan-name}/plan-for-human.md          <- 사람 작업 체크리스트
  .plan/{plan-name}/steps.json                 <- step 매니페스트
  .plan/{plan-name}/steps/step-1.md            <- Step 1 프롬프트
  .plan/{plan-name}/steps/step-2.md            <- Step 2 프롬프트
  ...
  .plan/{plan-name}/harness-for-plan.md        <- 글로벌 하네스
  .plan/{plan-name}/plan.sh                    <- 실행 러너

리뷰 순서:
  1. plan-for-human.md를 읽고 사전 작업을 완료하세요
  2. steps.json으로 전체 구조를 확인하세요
  3. steps/*.md로 각 Step 프롬프트를 리뷰하세요
  4. harness-for-plan.md로 기술 결정을 확인하세요
  5. bash .plan/{plan-name}/plan.sh --dry 로 최종 확인하세요
  6. bash .plan/{plan-name}/plan.sh 로 실행하세요
  7. /planosh-calibrate 로 발산을 측정하고 하네스를 강화하세요
```

plan.sh가 분리된 경우:

```
생성 완료:
  .plan/{plan-name}/plan-for-human.md          <- 사람 작업 체크리스트
  .plan/{plan-name}/steps-1.json               <- Phase 1 step 매니페스트
  .plan/{plan-name}/steps-2.json               <- Phase 2 step 매니페스트
  .plan/{plan-name}/steps/step-*.md            <- Step 프롬프트
  .plan/{plan-name}/harness-for-plan.md        <- 글로벌 하네스
  .plan/{plan-name}/plan-1.sh                  <- Phase 1 실행 러너
  .plan/{plan-name}/plan-2.sh                  <- Phase 2 실행 러너

리뷰 순서:
  1. plan-for-human.md를 읽고 사전 작업을 완료하세요
  2. bash .plan/{plan-name}/plan-1.sh --dry -> 리뷰 -> 실행
  3. plan-for-human.md의 중간 작업을 수행하세요
  4. bash .plan/{plan-name}/plan-2.sh --dry -> 리뷰 -> 실행
  5. plan-for-human.md의 사후 작업을 수행하세요
```
