diff 엔진을 구현하라. N개 run의 결과를 비교하여 발산을 감지하고 분류한다.

## 만들 것
- `internal/diff/diff.go` — 2단계 비교:
  - Stage 1: 파일 목록 비교 (.planoshignore 제외). 다르면 발산
  - Stage 2: 동일 파일의 내용 diff (whitespace-only 차이 무시). 실질적 차이만 발산
  - 수렴 판정: 모든 run이 Stage 1 + Stage 2 동일 → 수렴
  - 비교는 파일 시스템 기반이다 (git diff가 아님). testbed run-*에서 git commit이 발생해도 영향받지 않아야 함
  - `.planoshignore` 패턴은 파라미터로 받는다 (`internal/testbed`의 ignore 파서가 파싱한 결과를 전달받음)
- `internal/diff/classify.go` — 발산 분류 알고리즘:
  - 네이밍: 같은 위치에 다른 파일명/변수명
  - 구조: 파일 개수 또는 디렉토리 구조 차이
  - 코드 패턴: 같은 파일, 다른 구현 (import 순서, 함수 분해)
  - 범위 초과: 일부 run에서만 존재하는 파일
  - 비결정적 부산물: .planoshignore 매칭 파일 (자동 무시)
- `internal/diff/diff_test.go` — 테스트: 수렴, 네이밍 발산, 구조 발산, 코드 패턴 발산, 범위 초과, whitespace 무시 (D-004, D-011)

## 하지 않을 것
- 하네스 패치 생성 (Step 6에서)
- plan.sh 실행 (Step 5에서)
- 리포트 생성 (Step 7에서)
