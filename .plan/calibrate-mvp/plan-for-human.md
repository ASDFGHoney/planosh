# planosh CLI — calibrate MVP — 사람 작업 체크리스트

## 사후 작업 (plan.sh 완료 후)

- [ ] GitHub repo에 `GITHUB_TOKEN` secret 설정 (goreleaser 배포용)
- [ ] 첫 릴리스 태그 푸시: `git tag v0.1.0 && git push --tags`
- [ ] `/planosh-calibrate` 스킬에서 CLI 호출 방식으로 전환 확인
- [ ] `curl -fsSL .../install.sh | bash` 설치 테스트
