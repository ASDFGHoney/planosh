# 프로젝트 컨벤션

## 기술 결정

- D-001: Node.js + TypeScript (strict mode)
- D-002: 외부 프레임워크 없음 — 순수 Node.js HTTP 서버
- D-003: 테스트 없음 (2-Step 데모)

## 코딩 규칙

- ESM (import/export) 사용, CommonJS 금지
- 파일 확장자 .ts
- 세미콜론 사용
- 들여쓰기 2 spaces

## 절대 금지

- any 타입 사용
- console.error 외의 console 메서드 (console.log 포함)
- 이 Step의 범위 밖 파일 수정
- 외부 패키지 설치 (node 내장 모듈만 사용)
