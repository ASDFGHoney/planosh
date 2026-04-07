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
- **Context:** 외부 리뷰에서 "닭과 달걀 문제"로 지적됨. 가능한 전략: 자체 dogfooding 사례를 examples/에 먼저 채우기, Twitter/HN에 데모 GIF 공유, AI 개발 커뮤니티에 직접 소개.
- **Depends on:** 스킬 구현 + seed 예제 완성 후

## TODO-003: /planosh-patterns — GitHub 레포에서 best practice 검색 및 패턴 제안

- **What:** 커뮤니티가 공유한 examples, discussions, issues에서 사용자의 스택/기능에 맞는 패턴을 검색하여 하네스 규칙, verify 패턴, 발산 봉쇄법을 제안하는 스킬
- **Why:** 커뮤니티 사례가 쌓이면, 새 plan.sh를 만들 때 검증된 하네스 규칙을 처음부터 적용하여 calibrate 반복 횟수를 줄일 수 있다. 커뮤니티 기여 → 검색 → 재사용 플라이휠의 핵심.
- **Pros:** calibrate 전에 이미 검증된 규칙으로 하네스를 보강하면 초기 수렴률이 높아진다. 커뮤니티 기여에 대한 인센티브가 된다.
- **Cons:** examples가 충분히 쌓이기 전까지는 검색 결과가 빈약할 수 있다. gh CLI 의존성.
- **Scope:**
  - `gh` CLI로 레포의 examples/ 하네스 코드, Discussions 패턴 논의, Issues 발산 보고를 검색
  - 쿼리 유형: 스택 기반 ("nextjs auth"), 발산 기반 ("네이밍 발산"), 패턴 기반 ("verify 패턴"), 탐색 ("수렴률 높은 예제")
  - 현재 프로젝트에 plan.sh가 있으면 추천 규칙을 하네스에 바로 적용 가능
  - `/planosh` Phase 2에서 선택적으로 패턴 검색을 제안하는 연동
  - 추천 라벨 체계: `divergence:naming`, `stack:nextjs`, `convergence:100` 등
- **Depends on:** examples/ 커뮤니티 기여가 일정 수준 이상 쌓인 후
