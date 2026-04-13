병렬 실행 엔진을 구현하라. N개 testbed에서 plan.sh를 병렬로 실행하고 결과를 수집한다.

## 만들 것
- `internal/runner/runner.go` — 병렬 실행 엔진 (D-005):
  - semaphore 기반 동시성 제한 (`--concurrency`, 기본 3)
  - plan.sh 호출: `bash planShPath --from=N --to=N --testbed`
  - 환경변수 주입: `PROJECT_ROOT=runDir`, `PLAN_DIR=runDir/.plan/{name}`. plan.sh는 `PROJECT_ROOT` 환경변수가 있으면 자체 계산 대신 우선 사용한다
  - cmd.Dir = testbed/run-N/ (cwd가 곧 PROJECT_ROOT)
  - testbed run-*에서 plan.sh의 checkpoint()가 git commit을 생성하는 것은 정상 동작. diff 엔진은 git이 아닌 파일 시스템 기반 비교를 사용하므로 영향 없음
  - 결과 수집: exit code, stdout, stderr, 실행 시간
  - 타임아웃: `--timeout` 초과 시 강제 종료
  - 실행 실패 처리:
    - 1개 run 실패, 나머지 성공 → 1회 재시도, 재실패 시 제외
    - 모든 run 실패 → FAILED 마킹, 에러 로그 출력
- `internal/runner/runner_test.go` — 테스트: 병렬 실행, 타임아웃, 실패 처리, concurrency 제한 (D-004, D-011). mock 스크립트로 plan.sh 대체

## 하지 않을 것
- diff 비교 (Step 4에서 구현 완료)
- 자동 패치 (Step 6에서)
- calibrate 오케스트레이션 (Step 7에서)
