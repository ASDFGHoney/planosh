## 프로젝트
Node.js + TypeScript로 만든 Hello World HTTP 서버. planosh의 최소 동작 예제.

## Steps
2개:
1. 프로젝트 초기화 (package.json, tsconfig.json, src/index.ts)
2. HTTP 서버 구현 (GET /, GET /health)

## 결정률
(아직 측정하지 않음 — `/planosh-calibrate --step=1 --runs=3`으로 측정)

## 핵심 발견
- 글로벌 하네스에 "외부 패키지 설치 금지"를 넣으면 의존성 발산이 차단됨
- Step별 하네스에 "생성할 파일 목록"을 명시하면 구조 발산이 차단됨
- 2-Step 규모에서는 하네스 없이도 수렴할 수 있지만, 하네스의 효과를 검증하는 기준선으로 유용
