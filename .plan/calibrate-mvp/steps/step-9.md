배포 설정을 구성하고 불필요한 파일을 정리하라.

## 만들 것
- `.goreleaser.yaml` — macOS(arm64, amd64) + Linux(arm64, amd64) 바이너리 빌드 설정 (D-003)
- `install.sh` — `curl -fsSL .../install.sh | bash` → `~/.local/bin/planosh` 설치 스크립트 (D-003)
- `.github/workflows/release.yml` — 태그 푸시 시 goreleaser 실행 (D-003)
- `docs/CLI-REQUIREMENTS.md` 삭제 — 디자인 doc으로 대체됨 (PRD Supersedes 섹션). 파일이 존재하지 않으면 무시

## 하지 않을 것
- Homebrew tap 설정 (사용자 증가 후)
- CI/CD 테스트 파이프라인 (별도 이슈)
- `/planosh-calibrate` 스킬 업데이트 (사람이 확인 후 수동)
