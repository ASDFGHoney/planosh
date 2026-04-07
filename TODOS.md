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
