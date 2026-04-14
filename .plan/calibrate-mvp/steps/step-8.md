E2E 테스트를 작성하라. mock claude -p로 전체 calibrate 흐름을 검증한다.

## 만들 것
- `internal/e2e/calibrate_test.go` — E2E 테스트 파일 (D-004, D-012). mock claude -p 스크립트 + 전체 calibrate 흐름 검증:
  - Happy path: 3 runs, 모든 step 수렴 → 100% 점수
  - 발산 → 패치 → 수렴: 발산 감지 후 패치 적용, 재실행 시 수렴
  - Stuck: max-retries 초과 → 경고 + 계속 진행
  - 실행 실패: plan.sh 실패 → 에러 처리
  - 테스트용 fixtures: 최소 steps.json, step 프롬프트, plan.sh mock
- mock 스크립트: claude -p 대신 미리 정의된 출력 반환 (D-004)

## 하지 않을 것
- 실제 claude CLI 호출
- 배포 설정 (Step 9에서)
- 성능 최적화
