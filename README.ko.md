<p align="center">
  <h1 align="center">planosh</h1>
  <p align="center">
    <strong>AI 코딩 팀을 위한 결정적 실행 계획.</strong>
  </p>
  <p align="center">
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
    <a href="https://github.com/ASDFGHoney/planosh/stargazers"><img src="https://img.shields.io/github/stars/ASDFGHoney/planosh?style=social" alt="GitHub Stars"></a>
    <a href="https://github.com/ASDFGHoney/planosh/discussions"><img src="https://img.shields.io/github/discussions/ASDFGHoney/planosh" alt="Discussions"></a>
    <a href="README.md">English</a>
  </p>
</p>

---

> **개발자**: "그건 claude로 해봐야 알 것 같아요..ㅠㅠ"
>
> **기획자**: PR 올렸어요. 리뷰해주세용. &rarr; 파일 변경 40개, **+4,567 -12,300**
>
> 만약 실행 계획이 먼저 리뷰되었고, 그 계획이 결정적이었다면? 코드를 한 줄도 안 봐도 됐다.

**결정적이지 않은 계획은 계획이 아니다.**

실제로 `plan.sh` 하나를 **16시간 동안** 돌린 적이 있다 &mdash; 사람 개입 제로 &mdash; 첫 실행에 성공했다. 과하게 리니어하게 설계했고 검증이 너무 많아서 오래 걸렸을 뿐이다. 하지만 됐다. 모든 스텝이 통과했다. 아무도 키보드를 만지지 않았다.

이것이 약속이다: **plan을 한 번 쓰고, 자리를 뜨고, 돌아오면 코드가 되어 있다.**

## 왜 planosh인가

<table>
<tr>
<td width="25%" align="center"><strong>결정적</strong></td>
<td width="25%" align="center"><strong>리뷰 가능</strong></td>
<td width="25%" align="center"><strong>비동기</strong></td>
<td width="25%" align="center"><strong>검증 가능</strong></td>
</tr>
<tr>
<td>같은 plan, 같은 결과. 매번.</td>
<td>AI가 생성한 4,567줄이 아니라 plan을 리뷰하세요.</td>
<td>퇴근 전 <code>bash plan.sh</code>. 출근하면 끝.</td>
<td>모든 스텝에 검증. AI 판단이 아닌 shell exit code.</td>
</tr>
</table>

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

WHAT(프롬프트)과 HOW(하네스)가 동시에 제약되면 해의 공간이 극적으로 좁아진다. 마크다운 명세를 AI가 읽는 5단계 변환(spec &rarr; plan &rarr; tasks &rarr; sessions &rarr; code)이 1단계(prompt &rarr; code)로 압축된다.

상세 설계는 [설계 문서](docs/DESIGN.md)를 참고.

## 시작하기

planosh는 설치하는 패키지가 아니라 개념이다. 시작하려면:

**1. plan.sh를 직접 작성** &mdash; `claude -p` 호출이 있는 셸 스크립트면 된다.

**2. 또는 레퍼런스 스킬 사용** (Claude Code):

```bash
# 스킬을 프로젝트에 복사
cp -r skills/planosh/ .claude/skills/
cp -r skills/planosh-calibrate/ .claude/skills/

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

### plan.sh 공유

이 레포의 `.plan/` 디렉토리가 커뮤니티의 best practice 컬렉션이다. PR로 기여:

```
.plan/
└── your-plan-name/
    ├── plan.sh
    ├── harness-global.md
    ├── harness-step-N.md (선택)
    └── README.md
```

README에 포함할 내용:

| 항목 | 내용 |
|---|---|
| **프로젝트** | 무엇을 만들었는지 |
| **Steps** | 몇 개, 각 Step이 하는 일 |
| **결정률** | N번 실행 중 동일 결과 비율 (예: 3/3) |
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

*당신의 프로젝트가 첫 번째가 될 수 있다.*

## 더 읽기

- [설계 문서](docs/DESIGN.md) &mdash; 문제 정의, 3계층 제약 모델, 교정 루프, plan.sh 전체 예시
- [예시 PRD](docs/) &mdash; 레트로 웹앱, C 컴파일러, 마크다운 슬라이드

## 라이선스

[MIT](LICENSE)
