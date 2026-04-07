# Step 1 하네스: 프로젝트 초기화

## 현재 프로젝트 상태

빈 프로젝트. plan.sh와 .plan/ 디렉토리만 존재.

## 이 Step의 아키텍처 제약

- package.json의 type: "module" (ESM)
- tsconfig.json: strict: true, target: ES2022, module: NodeNext
- src/ 디렉토리에 소스 코드 배치

## 이 Step에서 생성할 파일 목록 (이 목록 외 파일 생성 금지)

- package.json
- tsconfig.json
- src/index.ts (빈 파일, console.error("hello") 한 줄만)
