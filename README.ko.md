<p align="center">
  <h1 align="center">planosh</h1>
  <p align="center">
    <strong>결정성 하나. 나머지는 전부 거기서 나온다.</strong>
  </p>
  <p align="center">
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
    <a href="https://github.com/ASDFGHoney/planosh/stargazers"><img src="https://img.shields.io/github/stars/ASDFGHoney/planosh?style=social" alt="GitHub Stars"></a>
    <a href="https://github.com/ASDFGHoney/planosh/discussions"><img src="https://img.shields.io/github/discussions/ASDFGHoney/planosh" alt="Discussions"></a>
    <a href="README.md">English</a>
  </p>
</p>

---

**AI 코딩 팀이 진짜로 깨지는 지점은 '플랜과 코드 사이의 갭'이다. SDD 도구들은 전부 그 갭을 막는다고 주장하지만, 실제로 막는 도구는 하나도 없다.**

마크다운 명세는 코드가 되지 않는다. 코드에 대한 *해석*이 될 뿐이고, AI 세션마다 다른 해석이 나온다. 더 나쁜 건, **AI는 해석의 여지를 마주칠 때마다 '더 많이 만드는 쪽'으로 달린다**는 것이다. 로그인 하나 시켰더니 MFA, 이메일 인증, rate limiting, 추상 레이어 3개가 같이 딸려나온다. **40 파일 변경, +4,567 -12,300짜리 PR이 그렇게 태어난다.**

SDD 도구들은 이걸 알고 있다. 드러내지 않을 뿐이다. 과잉 구현은 바이브 코딩 데모를 바이럴로 만드는 기능이다 &mdash; "문장 하나로 이걸 다 만들었어요." 실행할 때마다 결과가 갈라진다거나, 생성된 코드의 절반은 요청한 적도 없다는 사실을 보여주면 마법이 깨진다. 그래서 갭은 그럴듯한 UI 뒤에 숨겨진 채로 남고, 명세는 여전히 프로즈(prose)이고, 같은 입력이 다른 출력으로 흩어지는 구조는 그대로다.

실제로 프로덕션에 코드를 내보내는 팀에서 이런 프레임워크를 도입해본 적이 있다면, 이미 안다. 데모는 인상적이다. 그리고 실제 코드에 돌리고, PR을 열어보면 깨닫는다.

다음 물결은 "하네스"를 들고 온다. 룰 파일, 시스템 프롬프트, 코딩 컨벤션 &mdash; 범용 하네스를 에이전트에 달면 결정성이 따라온다는 식이다. **따라오지 않는다.** 하네스가 제약하는 건 *스타일*이다: 네이밍, 폴더 구조, 린트 룰. 하네스가 제약하지 못하는 건 *결정*이다: 어떤 파일을 만들지, 어떤 라이브러리를 쓸지, 기능을 함수 몇 개로 어떻게 쪼갤지. 이 결정들은 플랜 레이어에 존재하고, 범용 하네스는 거기에 대해 아무 의견이 없다. 같은 프롬프트에 같은 하네스를 달고 두 번 돌려보라 &mdash; 스타일 룰은 둘 다 통과하는데 구조적으로 완전히 다른 코드가 나온다. **하네스는 초록불인데 출력은 발산했다.**

범용 하네스는 틀린 레이어를 푸는 것이다. 결정성은 코드가 어떻게 *생겼는지*를 제약해서 나오지 않는다. 뭘 만들고, 어떤 순서로, 어떤 경계로 만드는지를 제약해야 나온다. 그건 하네스가 아니라 플랜이다.

---

planosh의 답은 하나다. **결정성.**

같은 플랜이 같은 코드로 수렴하게 만드는 것. 이 원칙 하나에서 나머지가 전부 따라 나온다:

- **해석의 여지가 없으면 과잉 구현도 일어나지 않는다.** 과잉 구현은 결정성이 낮을 때 드러나는 증상이지, 따로 풀어야 할 문제가 아니다.
- **결과가 수렴하면 무인 실행이 도박이 아니게 된다.** `plan.sh`를 걸어두고 퇴근할 수 있다.
- **플랜이 결과를 예측하면 팀은 코드 대신 플랜을 리뷰한다.** 4,000줄 PR을 열어보지 않아도 된다.
- **비개발자도 플랜을 리뷰할 수 있으면, 팀 전원이 AI로 개발할 수 있다.** 결정성이 코드를 읽을 수 없는 사람들까지 개발 프로세스 안으로 데려온다.

결정성은 측정 가능하다 &mdash; 같은 플랜을 N번 병렬로 돌려서 결과 diff를 본다. 발산하면 그 지점이 아직 결정적이지 않은 곳이다. 발산을 찾고, 봉쇄하고, 다시 측정한다. 이 루프가 planosh의 전부다.

---

**증거.** `plan.sh` 하나를 **16시간 동안 무인 실행**해서, 1년 넘게 운영되던 프로덕션 앱을 Flutter에서 React Native로 마이그레이션했다. 중간에 키보드를 만진 사람은 없었다. 결과물이 프로덕션 레디는 아니었다 &mdash; 빈틈이 있었다. 하지만 같은 마이그레이션을 Claude Code 배치 모드와 Speckit 기반 명세로 돌린 결과물과 비교하면, 완성도 자체가 비교 불가능한 수준이었다. 그리고 이건 **calibrate 없이** 나온 결과다 &mdash; 병렬 실행도, 발산 탐지도, 하네스 강화도 하지 않았다.

플랜을 쓰는 데는 2일이 걸렸다. **그 비율이 planosh의 전부다.** 실행을 보모하는 데 시간을 쓰지 말고, 플랜에 투자하라. 교정 루프를 돌렸다면 남은 빈틈은 실행 전에 이미 봉쇄되었을 것이다.

## 기존 접근법 vs planosh

| 기존 접근법                                            | planosh                                                            |
| ------------------------------------------------------ | ------------------------------------------------------------------ |
| AI가 명세를 해석하며 **과잉 구현한다**                 | 플랜이 결정적이라 해석 자체가 일어나지 않는다                      |
| 범용 하네스는 스타일만 제약하고 결정은 제약하지 못한다 | **플랜 전용 하네스가 뭘 만들지, 어떤 순서로 만들지를 제약한다**    |
| "동작하는가"를 측정한다                                | **"수렴하는가"를 측정한다** (같은 플랜 &rarr; 같은 코드, N회 실행) |
| 4,000줄 PR을 리뷰한다                                  | 플랜을 리뷰한다. 코드는 건너뛴다.                                  |
| 무인 실행은 운에 의존한다                              | **결정성이 보장하니 16시간 무인 실행이 가능하다**                  |
| 결정성은 주장이다                                      | 결정성은 측정해서 줄이는 숫자다                                    |

## 이렇게 생겼다

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

`run_claude`는 프롬프트(`-p`)와 하네스(`--append-system-prompt`)를 결합하여 Claude를 호출한다. `verify`는 exit code로 성공/실패를 판정한다. 이게 전부다.

## 어떻게 작동하는가

planosh는 3개 레이어로 AI 실행을 제약한다:

| Layer | 역할 | 제약 |
|-------|------|------|
| **Layer 1: 시스템 프롬프트** (`--append-system-prompt`) | HOW — 코딩 컨벤션, 아키텍처 규칙, 금지 패턴 | "어떻게 만들지"를 제약 |
| **Layer 2: 유저 프롬프트** (`-p`) | WHAT — 만들 것, 하지 않을 것, 선행 조건 | "무엇을 만들지"를 제약 |
| **Layer 3: 검증** (`verify`) | CHECK — 빌드, 파일 존재, 테스트 통과 | 결과가 제약에 부합하는지 사후 확인 |

WHAT(프롬프트)과 HOW(하네스)가 동시에 제약되면 해의 공간이 극적으로 좁아진다. 마크다운 명세를 AI가 읽는 5단계 변환(spec &rarr; plan &rarr; tasks &rarr; sessions &rarr; code)이 1단계(prompt &rarr; code)로 압축된다.

상세 설계는 [설계 문서](docs/DESIGN.md)를 참고.

## 시작하기

planosh는 설치하는 패키지가 아니라 개념이다. 시작하려면:

**1. plan.sh를 직접 작성** &mdash; `claude -p` 호출이 있는 셸 스크립트면 된다.

**2. 또는 레퍼런스 스킬 사용** (Claude Code 플러그인):

```bash
# Claude Code 플러그인으로 설치
claude plugin add ASDFGHoney/planosh

# PRD에서 plan 생성
/planosh path/to/prd.md

# 결정성 교정 (병렬 실행, 발산 탐지)
/planosh-calibrate --runs=3
```

**3. 실행:**

```bash
bash plan.sh
```

## 기여하기

planosh는 완성된 프레임워크가 아니라 제안이다. 패턴과 모범 사례를 함께 만들어간다. **코드를 기여할 필요 없다 &mdash; plan을 공유하면 된다.**

다만 두 가지 방향에서 추가 개발이 이어진다:

- **plan.sh 가독성** &mdash; plan.sh는 이제 `steps.json`과 `steps/*.md`를 읽는 범용 러너다. 프롬프트가 읽기 쉬운 마크다운 파일로 분리되어 bash 문자열 인라인 문제를 해결했다.
- **`planosh run`** &mdash; shim-git 기반 격리 testbed N개를 띄우고 `plan.sh`를 병렬 실행하는 CLI. 교정(`--mode calibrate`)과 내부 병렬 실행(`--mode split`)을 실제로 가능하게 만드는 실행 환경이다. [#1](https://github.com/ASDFGHoney/planosh/issues/1)과 [#2](https://github.com/ASDFGHoney/planosh/issues/2) 참고.

둘 다 기여를 받고 있다.

### plan.sh 공유

이 레포의 `.plan/` 디렉토리가 커뮤니티의 best practice 컬렉션이다. PR로 기여:

```
.plan/
  your-plan-name/
    plan.sh
    steps.json
    steps/
      step-1.md
      step-2.md
    harness-for-plan.md
    README.md
```

README에 포함할 내용:

| 항목          | 내용                                                      |
| ------------- | --------------------------------------------------------- |
| **프로젝트**  | 무엇을 만들었는지                                         |
| **Steps**     | 몇 개, 각 Step이 하는 일                                  |
| **결정률**    | N번 실행 중 동일 결과 비율 (예: 3/3)                      |
| **핵심 발견** | 어떤 하네스/패턴이 효과적이었는지, 어떤 발산을 발견했는지 |

### 발산 보고

"같은 plan.sh를 돌렸는데 다른 결과가 나왔다"는 하네스 개선의 가장 좋은 출발점이다. [Issues](https://github.com/ASDFGHoney/planosh/issues)에 보고.

### 패턴 논의

발견한 패턴, 실험 결과, 아이디어를 [Discussions](https://github.com/ASDFGHoney/planosh/discussions)에서 공유.

### 우리가 바라는 일

- 하네스 + verify 루프로 **결정률 100%**를 달성한 사례
- 50개 파일, 20 Step 규모를 `plan.sh` 한 번으로 해치운 패턴
- Next.js, Rails, Flutter 등 특정 스택의 **하네스 템플릿**
- AI 판단 없이 결정적 검증만으로 품질을 보장하는 패턴
- 교정 루프에서 발견한 하네스 빈틈과 해결법

사례가 쌓이면 패턴이 보이고, 패턴이 모이면 프레임워크가 된다. planosh는 그 씨앗이다.

## In the Wild

planosh로 무언가를 만들었다면? [PR을 열어](https://github.com/ASDFGHoney/planosh/pulls) 여기에 추가해주세요.

<!--
- [프로젝트명](링크) -- 간략 설명, 결정률
-->

_당신의 프로젝트가 첫 번째가 될 수 있다._

## 모범사례

`best-practices/` 디렉토리는 planosh 자체의 **실행 패턴 모범사례**를 모아둔다. (`.plan/`가 커뮤니티의 **도메인 plan 컬렉션**이라면, 이쪽은 plan.sh를 어떻게 설계/실행할지에 관한 메타 패턴이다.)

_아직 항목이 없다. PR로 기여해주세요._

## 더 읽기

- [설계 문서](docs/DESIGN.md) &mdash; 문제 정의, 3계층 제약 모델, 교정 루프, plan.sh 전체 예시
- [예시 PRD](docs/) &mdash; 레트로 웹앱, C 컴파일러, 마크다운 슬라이드
- [모범사례](best-practices/) &mdash; 실행 패턴 참조 구현

## 라이선스

[MIT](LICENSE)
