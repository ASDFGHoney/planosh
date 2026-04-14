testbed 내부 모듈을 구현하라. testbed는 calibrate 전용이며 독립 명령으로 노출하지 않는다.

## 만들 것
- `internal/testbed/testbed.go` — testbed 생명주기 관리 (D-006):
  - `Create`: `git clone --local` → `~/.planosh/testbed/{repo}--{plan}/golden/`
  - `CopyRuns`: golden → run-1, run-2, ..., run-N 복사. `rsync -a --delete --exclude-from=.planoshignore`
  - `ResetRun`: 특정 run을 golden에서 다시 복사 (step 재실행 시)
  - `UpdateGolden`: run-1의 코드 변경을 golden에 반영 (step 수렴 후)
  - `Cleanup`: testbed 삭제 (`--keep-testbed` 시 보존)
  - lockfile: `.lock` 파일에 PID 기록. 동시 실행 시 경고. stale PID 감지
- `internal/testbed/ignore.go` — `.planoshignore` 파서:
  - 프로젝트 루트에 없으면 내장 기본값 사용
  - 기본값: `node_modules/`, `.next/`, `.nuxt/`, `dist/`, `build/`, `*.lock`, `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `.claude/`
- `internal/testbed/testbed_test.go` — 테스트: Create/CopyRuns/ResetRun/UpdateGolden/Cleanup, lockfile, .planoshignore (D-004, D-011)

## 하지 않을 것
- submodule/LFS 지원 (v2)
- `planosh testbed create/clean/list` 명령 노출 (calibrate 내부 전용)
- diff 비교 로직 (Step 4에서)
- discover 모듈 사용 — testbed는 PROJECT_ROOT를 직접 파라미터로 받는다. .plan/ 발견은 calibrate 커맨드(Step 7)에서 discover를 호출하여 결과를 testbed에 전달
