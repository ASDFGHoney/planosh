# PRD: Rust C Compiler (rcc)

Rust로 작성하는 from-scratch C 컴파일러. GCC 호환, Linux 커널 컴파일이 최종 목표.

> 참고: https://www.anthropic.com/engineering/building-c-compiler

## 배경

Anthropic의 Nicholas Carlini가 16개의 병렬 Claude 에이전트로 C 컴파일러를 만든 프로젝트를 재현한다. 원본은 2,000 세션에 걸쳐 100,000줄의 Rust 코드를 생산했다. 이 PRD는 동일한 컴파일러를 단계적으로 구축하기 위한 명세다.

## 제품 정의

### 한 줄 설명

Rust 표준 라이브러리만으로 구현하는 C11 컴파일러. x86-64, ARM64, RISC-V 백엔드를 지원하며, Linux 6.9 커널을 컴파일할 수 있다.

### 핵심 제약

- **언어**: Rust (stable)
- **의존성**: Rust 표준 라이브러리만 사용. 외부 crate 없음 (clean-room)
- **타깃 C 표준**: C11 (GNU 확장 포함 — 리눅스 커널이 요구)
- **GCC 호환**: `-O0` ~ `-O2` 플래그, 주요 경고/에러 플래그 호환

## 아키텍처

```
Source (.c)
  │
  ▼
┌─────────────┐
│  Lexer       │  토큰화
└──────┬──────┘
       ▼
┌─────────────┐
│  Preprocessor│  #include, #define, #ifdef, 매크로 확장
└──────┬──────┘
       ▼
┌─────────────┐
│  Parser      │  AST 생성 (C11 + GNU 확장)
└──────┬──────┘
       ▼
┌─────────────┐
│  Semantic    │  타입 검사, 심볼 테이블, 암묵적 변환
│  Analysis    │
└──────┬──────┘
       ▼
┌─────────────┐
│  IR Gen      │  SSA 형태의 중간 표현 생성
└──────┬──────┘
       ▼
┌─────────────┐
│  Optimizer   │  상수 전파, DCE, 인라이닝, 레지스터 할당 등
└──────┬──────┘
       ▼
┌─────────────┐
│  Codegen     │  타깃별 어셈블리 출력
│  (Backend)   │  x86-64 / ARM64 / RISC-V
└──────┬──────┘
       ▼
  Assembly (.s)
       │
       ▼  (외부 어셈블러/링커 — GNU as, ld)
  Binary (ELF)
```

### SSA IR

- 모든 최적화는 SSA IR 위에서 수행
- 백엔드는 IR → 타깃 어셈블리로 lowering
- IR은 자체 텍스트 형식으로 덤프 가능 (`--dump-ir`)

## 기능 범위

### MVP (v0.1) — 단일 파일 컴파일

- C11 기본 문법 파싱 (변수, 함수, 제어문, 구조체, 포인터)
- x86-64 코드 생성
- 단순 산술/비교 연산
- 함수 호출 (System V ABI)
- `printf` 등 libc 함수 링크
- 검증: 간단한 C 프로그램 컴파일 → 실행 → 기대 출력 비교

### v0.2 — 프리프로세서 + 타입 시스템

- `#include`, `#define`, `#ifdef`, `#pragma`
- typedef, enum, union
- 배열, 다차원 배열
- 문자열 리터럴, 이스케이프 시퀀스
- 암묵적 타입 변환 (integer promotion 등)
- 검증: c-testsuite 기본 테스트 통과

### v0.3 — 고급 C 기능

- 가변 인자 함수 (`va_list`, `va_arg`)
- 비트필드
- 복합 리터럴
- 지정 초기화자 (designated initializer)
- `_Generic`, `_Static_assert`
- inline 함수, `restrict`
- 검증: GCC torture test suite 50%+ 통과

### v0.4 — GNU 확장 + 리눅스 커널 호환

- GNU 확장: `__attribute__`, `__builtin_*`, statement expression, `typeof`
- 인라인 어셈블리 (`asm volatile`)
- 링커 스크립트 호환 심볼 (`__section`, `__used`, `__aligned`)
- `-fno-strict-aliasing`, `-fno-common` 등 GCC 호환 플래그
- 검증: SQLite, Redis, Lua 컴파일 + 자체 테스트 통과

### v0.5 — SSA 최적화

- 상수 전파 (constant propagation)
- 죽은 코드 제거 (dead code elimination)
- 공통 부분식 제거 (CSE)
- 함수 인라이닝
- 루프 불변 코드 이동 (LICM)
- 레지스터 할당 (linear scan 또는 graph coloring)
- 검증: 벤치마크에서 최적화 전 대비 개선 측정

### v0.6 — 멀티 백엔드

- ARM64 (AArch64) 코드 생성
- RISC-V 64 코드 생성
- 검증: 동일 테스트 스위트가 3개 아키텍처에서 통과

### v0.7 — Linux 커널 컴파일

- Linux 6.9 커널 빌드 성공 (x86-64)
- QEMU에서 부팅 확인
- ARM64, RISC-V에서도 커널 빌드
- 검증: QEMU 부팅 → 쉘 프롬프트 도달

## 비목표 (v1 이후로 미룸)

- 자체 어셈블러/링커 (GNU as/ld 사용)
- 16-bit x86 real mode 코드 생성 (GCC fallback 허용)
- C++ 지원
- Windows/macOS 타깃
- LTO (Link-Time Optimization)
- `-O3` 수준 고급 최적화
- 자체 표준 라이브러리 (libc)

## 성공 기준

| 지표 | 목표 |
|------|------|
| GCC torture test suite pass rate | >= 99% |
| 컴파일 가능한 프로젝트 | Linux 6.9, QEMU, FFmpeg, SQLite, postgres, redis |
| 지원 아키텍처 | x86-64, ARM64, RISC-V |
| 외부 의존성 | Rust std만 |
| 재현 가능한 빌드 | `cargo build --release` 한 번으로 완료 |

## 디렉토리 구조 (예상)

```
rcc/
├── Cargo.toml
├── src/
│   ├── main.rs              # CLI 진입점
│   ├── lexer/               # 토큰화
│   ├── preprocessor/        # 프리프로세서
│   ├── parser/              # AST 생성
│   ├── sema/                # 의미 분석
│   ├── ir/                  # SSA IR 정의 + 생성
│   ├── opt/                 # 최적화 패스
│   ├── codegen/
│   │   ├── x86_64/
│   │   ├── aarch64/
│   │   └── riscv64/
│   └── common/              # 공유 유틸리티
└── tests/
    ├── unit/                # 단위 테스트
    ├── fixtures/            # 테스트용 .c 파일
    └── integration/         # 통합 테스트 (컴파일→실행→비교)
```

## 참고

- 원본 프로젝트: Anthropic의 Nicholas Carlini가 Opus 4.6 에이전트 팀으로 구현
- 2,000 Claude Code 세션, $20,000 API 비용, 2주 소요
- 100,000줄 Rust 코드
