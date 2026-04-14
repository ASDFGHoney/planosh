Go 프로젝트를 초기화하고 Cobra CLI 스캐폴딩을 구성하라.

## 만들 것
- `go.mod` — 모듈 경로 `github.com/ASDFGHoney/planosh`, Go 1.22+ (D-009)
- `cmd/planosh/main.go` — 엔트리포인트, root command 실행 (D-012)
- `cmd/planosh/root.go` — Cobra root command 정의 (D-002)
- `cmd/planosh/calibrate.go` — `calibrate` 서브커맨드 스캐폴딩. 플래그: `--plan`, `--runs`, `--keep-testbed`, `--max-retries`, `--model`, `--patch-model`, `--concurrency`, `--timeout`, `--dry`. RunE placeholder: `return fmt.Errorf("not implemented: calibrate")` (Step 7에서 교체) (D-002)
- `cmd/planosh/version.go` — `version` 서브커맨드. 빌드 시 ldflags로 주입되는 버전 변수 (D-002)
- 의존성 추가 후 `go mod tidy` 실행: lipgloss (D-010), testify (D-011), golang.org/x/sync (D-005)

## 하지 않을 것
- internal/ 패키지 구현 (Step 2~6에서)
- calibrate 실행 로직 구현 (Step 7에서)
- 배포 설정 (Step 9에서)
