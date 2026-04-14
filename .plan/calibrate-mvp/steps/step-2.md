.plan/ 발견 규칙과 steps.json 파서를 구현하라.

## 만들 것
- `internal/discover/discover.go` — `.plan/` 발견 규칙 구현 (D-007):
  - cwd에서 시작하여 상위로 탐색
  - `.plan/` 디렉토리를 찾으면 그 부모가 PROJECT_ROOT
  - git root까지 없으면 cwd가 PROJECT_ROOT
  - 중첩 시 가장 가까운 `.plan/` 우선
  - `--plan=<name>` 이름으로 `.plan/{name}/` 경로 반환
- `internal/discover/discover_test.go` — 테스트: cwd에 .plan/, 상위에 .plan/, git root, 중첩 시나리오 (D-004, D-011)
- `internal/step/step.go` — steps.json 파서:
  - steps.json 읽기 → Step 목록 반환
  - Step 구조체: ID, Name, Prompt(상대경로), Verify 목록, Commit 메시지
- `internal/step/step_test.go` — 테스트: 정상 파싱, 빈 steps, 잘못된 JSON (D-004, D-011)

## 하지 않을 것
- plan.sh 호출 로직 (Step 5에서)
- testbed 관련 코드 (Step 3에서)
