# 프로젝트 컨벤션

## 기술 결정 (locked — 변경 불가)

- D-001: 언어 = Go (싱글 바이너리, 런타임 의존 없음)
- D-002: CLI 프레임워크 = Cobra (서브커맨드 구조)
- D-003: 배포 = goreleaser + GitHub Actions + install.sh (curl | bash)
- D-004: 테스트 전략 = 단위 테스트 + E2E (mock claude -p). 라이브러리 = testify
- D-005: 외부 의존 = rsync, git, bash, python3(plan.sh 파싱), claude CLI. Phase 1에서는 plan.sh를 bash로 직접 호출
- D-006: testbed 위치 = `~/.planosh/testbed/{repo}--{plan}/`. 프로젝트 repo 안에 두지 않음
- D-007: .plan/ 발견 규칙 = cwd에서 상위 탐색 → .plan/ 디렉토리의 부모가 PROJECT_ROOT. git root까지 없으면 cwd. 중첩 시 가장 가까운 .plan/ 우선
- D-008: 패치 전략 = 자동 패치는 해당 step의 프롬프트(steps/{M}.md)에만 적용. 글로벌 하네스(harness-for-plan.md) 미수정. 이전 step regression 원천 차단
- D-009: Go 모듈 경로 = `github.com/ASDFGHoney/planosh`
- D-010: CLI 출력 스타일 = charmbracelet/lipgloss (색상 + 스타일링)
- D-011: 테스트 라이브러리 = testify (assert/require/mock)
- D-012: Go 패키지 구조 = `cmd/planosh/` + `internal/` (testbed, discover, step, runner, diff, patch, report, e2e)
- D-013: `.planoshignore` 기본값 = `node_modules/`, `.next/`, `.nuxt/`, `dist/`, `build/`, `*.lock`, `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `.claude/` (PRD 확정, 변경 불가)

## AI 재량 (discretion — 발산 허용)

다음 영역은 AI가 자유롭게 판단한다:
- 함수 시그니처 설계 (인터페이스 분리 판단). 단, 이전 step에서 이미 생성된 모듈의 인터페이스는 변경하지 않고 그대로 사용한다
- 에러 메시지 문구
- 테스트 헬퍼 함수 내부 구현
- lipgloss 색상/스타일 구체값

## 코딩 규칙

- Go 모듈 경로는 `github.com/ASDFGHoney/planosh` (D-009)
- 모든 public 패키지는 `internal/` 하위에 둔다 (D-012)
- CLI 엔트리포인트는 `cmd/planosh/` (D-012)
- 테스트 파일은 해당 패키지 안에 `*_test.go` (D-004)
- testify의 `assert`/`require`를 사용한다 (D-011)
- CLI 출력은 lipgloss로 스타일링한다 (D-010)
- 외부 명령(rsync, git, bash, claude) 호출은 `os/exec`으로 한다 (D-005)
- 동시성은 Go channel 또는 `golang.org/x/sync/errgroup` 기반 semaphore (D-005)
- testbed 경로는 `~/.planosh/testbed/` 하위 (D-006)
- `.plan/` 발견은 cwd에서 상위 탐색 (D-007)

## 절대 금지

- 프로젝트 repo 안에 testbed 생성 (D-006)
- 글로벌 하네스(harness-for-plan.md) 자동 수정 (D-008)
- Docker/컨테이너 의존 (D-001: 런타임 의존 없음)
- `internal/` 밖에 라이브러리 패키지 생성 (D-012)
- 하드코딩된 시크릿
- 이 Step의 범위 밖 파일 수정
