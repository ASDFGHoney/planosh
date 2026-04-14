calibrate 커맨드에서 모든 모듈을 연결하고, 리포트 생성 모듈을 구현하라.

## 만들 것
- `cmd/planosh/calibrate.go` 수정 — Step 1의 `return fmt.Errorf("not implemented: calibrate")` placeholder를 실제 로직으로 교체:
  - Phase 0: .plan/{name}/ 존재 확인 (discover), steps.json 파싱 (step), PROJECT_ROOT 계산
  - Phase 1: step별 calibrate 루프:
    - testbed 생성 (testbed.Create) → run 복사 (testbed.CopyRuns)
    - N개 run에서 plan.sh 병렬 실행 (runner)
    - 실행 결과 diff (diff)
    - 수렴 → golden 업데이트 (testbed.UpdateGolden) → 다음 step
    - 발산 → 자동 패치 (patch) → run 재복사 (testbed.ResetRun) → 재실행 (max-retries까지)
    - stuck → 경고 출력 + 계속 진행
  - Phase 2: golden의 .plan/{name}/ 변경사항을 원본 레포에 rsync. 코드는 미반영
  - testbed 정리 (--keep-testbed 아니면)
  - lipgloss로 step별 진행 상황 출력 (D-010)
- `internal/report/report.go` — calibration-report.md 생성 (D-012):
  - step별 수렴 상태 (수렴/발산/stuck/failed)
  - 적용된 패치 목록
  - 결정성 점수 + 해석 (100%=완전 수렴, 80%+=권장, <50%=하네스 부족)
- `internal/report/report_test.go` — 테스트: 리포트 생성, 점수 계산 (D-004, D-011)

## 하지 않을 것
- E2E 테스트 (Step 8에서)
- 배포 설정 (Step 9에서)
- `planosh run` / `planosh bisect` 구현 (Phase 2/3)
