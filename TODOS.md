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
