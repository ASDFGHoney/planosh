# Step 2 하네스: HTTP 서버

## 현재 프로젝트 상태

Step 1 완료 후 존재하는 파일:

- package.json (type: "module")
- tsconfig.json (strict: true)
- src/index.ts (빈 파일)

## 이 Step의 아키텍처 제약

- node:http 내장 모듈만 사용
- 포트: 환경변수 PORT 또는 기본값 3000
- GET / → { message: "hello, planosh" } (JSON)
- GET /health → { status: "ok" } (JSON)
- 그 외 경로 → 404

## 이 Step에서 생성할 파일 목록 (이 목록 외 파일 생성 금지)

- src/index.ts (HTTP 서버 구현으로 교체)
- src/server.ts (서버 로직)
