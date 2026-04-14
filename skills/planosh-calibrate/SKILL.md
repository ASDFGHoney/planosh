---
name: planosh-calibrate
description: plan.sh의 각 Step을 순차적으로 교정하여 하네스를 강화. "calibrate", "하네스 강화", "/planosh-calibrate" 등의 요청에 사용.
---

# /planosh-calibrate — Step별 순차 수렴

`planosh calibrate` CLI를 호출하여 plan.sh의 각 Step을 순서대로 N번 실행하고, 수렴할 때까지 하네스를 강화한다.

```
/planosh-calibrate [plan-name] [--runs=N]

입력: .plan/{plan-name}/plan.sh + 하네스
출력: 강화된 하네스 + .plan/{plan-name}/calibration-report.md
```

이 스킬의 역할은 **CLI 실행 전/후**로 제한된다. 교정 로직 자체(testbed 생성, 병렬 실행, diff 분석, 패치, 재시도, 리포트)는 전부 CLI 내부에서 수행한다.

---

## 실행

단일 Bash 호출로 끝난다:

```bash
planosh calibrate {plan-name} [flags]
```

### 플래그

| 플래그 | 기본값 | 설명 |
|--------|--------|------|
| `[plan-name]` 또는 `--plan` | (자동 감지) | `.plan/` 안에 하나뿐이면 생략 가능 |
| `--runs` | 3 | Step당 병렬 실행 횟수 |
| `--max-retries` | 2 | 발산 Step당 패치 재시도 횟수 |
| `--concurrency` | 1 | 동시에 실행할 run 수 |
| `--timeout` | 30m | Step 실행당 타임아웃 |
| `--model` | (기본) | Step 실행용 claude 모델 |
| `--patch-model` | (기본) | 패치 생성용 claude 모델 |
| `--keep-testbed` | false | 교정 후 testbed 보존 |
| `--dry` | false | 실행 없이 plan 검증만 |

---

## 실행 전 검증 (이 스킬이 수행)

CLI 호출 전에 다음을 확인한다. 실패 시 사용자에게 안내하고 중단.

1. **plan-name 모호성**: 사용자가 plan-name을 주지 않았고 `.plan/`에 plan이 여러 개면 어느 것인지 묻는다.
2. **plan.sh의 `--testbed` 지원**: `grep -q 'testbed' .plan/{plan-name}/plan.sh`. 없으면 `/planosh`로 재생성하거나 수동 추가를 안내.
3. **git 커밋 상태**: `.plan/{plan-name}/` 전체가 git에 추적되어 있어야 한다 (testbed golden은 `file://` clone에서 복사하므로 미커밋 파일은 누락됨).

---

## 실행 후 (이 스킬이 수행)

CLI가 종료되면 `calibration-report.md`를 읽고 사용자에게 요약한다:

- 결정성 점수 (%) + 해석
- 수렴/stuck/실패 Step 수
- STUCK Step이 있으면 해당 Step의 발산 유형과 사용자가 취할 수 있는 다음 행동 제시

---

## 제약

- plan.sh가 포트·DB·외부 서비스를 사용하는 `verify()`를 포함하면 `--concurrency > 1`에서 충돌할 수 있다. 빌드/파일 검증 위주일 때 안전.
- 교정된 하네스는 golden에서 원본 `.plan/{plan-name}/`으로 rsync되어 반영된다. 커밋은 사용자가 직접 수행.
