자동 패치 모듈을 구현하라. 발산이 감지되면 claude -p를 호출하여 하네스 규칙을 생성하고 step 프롬프트에 적용한다.

## 만들 것
- `internal/patch/patch.go` — 자동 패치 생성 (D-008):
  - claude -p 호출 (`--patch-model`, 기본 sonnet):
    - 입력: 현재 하네스, 현재 step 프롬프트, N개 run의 발산 diff, 발산 유형
    - 출력: step 프롬프트에 추가할 규칙 텍스트
  - 패치 적용: steps/{M}.md 파일 끝에 규칙 추가
  - 글로벌 하네스(harness-for-plan.md)는 절대 수정하지 않음
- `internal/patch/prompt.go` — 패치 프롬프트 템플릿:
  - 발산 유형별 프롬프트 분기 (네이밍, 구조, 코드 패턴, 범위 초과)
  - claude -p 응답 파싱 (규칙 텍스트 추출)
  - 프롬프트 최소 구조:
    - 입력 포맷: (1) 현재 하네스 전문, (2) 현재 step 프롬프트 전문, (3) 발산 diff (파일별), (4) 발산 유형 라벨
    - 출력 포맷: 마크다운 규칙 블록 (`## 아키텍처 제약` 섹션에 추가할 수 있는 형태)
    - 실패 처리: claude -p가 빈 응답이거나 파싱 불가 시 패치를 skip하고 해당 step을 stuck으로 마킹
- `internal/patch/patch_test.go` — 테스트: 패치 생성 (mock claude -p), step 프롬프트 적용, 글로벌 하네스 미수정 확인 (D-004, D-011)

## 하지 않을 것
- 글로벌 하네스 승격 (v2)
- diff 비교 로직 (Step 4에서 구현 완료)
- calibrate 오케스트레이션 (Step 7에서)
