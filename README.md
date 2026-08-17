# GOAPI for NUBO

<p align="center">
  <img src="https://img.shields.io/github/license/sirini/goapi?style=flat-square&color=5D6D7E" alt="license">
  <img src="https://img.shields.io/github/stars/sirini/goapi?style=flat-square&color=F4D03F" alt="stars">
  <img src="https://img.shields.io/github/last-commit/sirini/goapi?style=flat-square&color=2ECC71" alt="last commit">
</p>

GOAPI는 [NUBO](https://github.com/sirini/nubo)의 백엔드입니다. GoFiber v3로 HTTP API를 제공하고 MySQL/MariaDB의 회원·게시물·알림 데이터를 처리하며, 이미지 변환과 Resend 메일 발송도 담당합니다.

대부분의 운영자는 이 저장소를 따로 설치할 필요가 없습니다. NUBO 저장소에 Linux용 `goapi-linux`가 포함되어 있으며, NUBO의 `.env`와 `env.sample`을 함께 사용합니다. 이 저장소는 GOAPI를 수정하거나 다른 운영체제·CPU용으로 직접 빌드하려는 경우에 사용하세요.

## 담당 기능

- 회원가입, 로그인, JWT 갱신 및 권한 관리
- 게시판·게시물·댓글·알림·채팅 데이터 처리
- 업로드 이미지 리사이즈 및 메타데이터 처리
- Resend 기반 가입 인증, 비밀번호 초기화, 댓글 알림
- Resend Broadcast 기반 관리자 단체 메일
- 이메일 인증·초대 전용·가입 중지 정책
- 반복 실행 가능한 데이터베이스 스키마 설치

## 실행 구조

| 항목 | 기본값 | 설명 |
| --- | --- | --- |
| API 주소 | `http://127.0.0.1:3006/goapi` | `GOAPI_PORT`, `GOAPI_BASE`로 변경 |
| 설정 파일 | NUBO 디렉터리의 `.env` | Nuxt와 GOAPI가 함께 사용 |
| 설치 템플릿 | NUBO 디렉터리의 `env.sample` | 최초 실행 시 `.env` 생성에 사용 |
| 데이터베이스 | MySQL/MariaDB | 테이블 접두사 지원 |

GOAPI 바이너리는 **NUBO 프로젝트 디렉터리에서 실행해야 합니다.** 다른 디렉터리에서 실행하면 `env.sample`과 `.env`를 찾지 못합니다.

## 가장 쉬운 사용법

NUBO를 내려받아 포함된 바이너리를 실행합니다.

```bash
git clone https://github.com/sirini/nubo.git
cd nubo
chmod +x ./goapi-linux
./goapi-linux
```

`.env`가 없으면 DB 정보와 최초 관리자 계정을 묻고 다음을 자동으로 처리합니다.

1. `env.sample`의 자리표시자를 실제 값으로 치환
2. 권한 `0600`으로 `.env` 생성
3. 데이터베이스와 기본 테이블 생성
4. 관리자 계정 생성
5. API 서버 시작

기존 설치를 업데이트했거나 새로운 테이블이 추가된 버전을 배포했다면 다음 명령을 한 번 실행하세요.

```bash
./goapi-linux install
```

이 명령은 필요한 테이블과 컬럼을 확인해 추가하며 반복 실행할 수 있습니다.

## 직접 빌드하기

### 요구 사항

- Go 1.25 이상
- CGO를 사용할 수 있는 C/C++ 빌드 환경
- `libvips` 개발 패키지
- 실행 시 MySQL 또는 MariaDB

Ubuntu 계열 예시:

```bash
sudo apt update
sudo apt install build-essential libvips-dev
git clone https://github.com/sirini/goapi.git
cd goapi
go test ./...
go build -trimpath -o goapi-linux ./cmd
```

### 배포용 Ubuntu 22.04 호환 바이너리

NUBO에 포함할 공식 x86-64 Linux 바이너리는 호스트 운영체제에서 직접 빌드하지 않고 Docker의 Ubuntu 22.04 환경에서 만듭니다. 이렇게 하면 빌드 호스트가 더 최신 배포판이어도 바이너리가 glibc 2.35보다 새로운 심볼을 요구하지 않습니다.

Docker와 Buildx가 준비된 환경에서 다음 스크립트를 실행하세요.

```bash
./scripts/build-ubuntu22.sh
```

- README의 `git clone` 예시처럼 형제 경로에 `nubo` 디렉터리가 있으면 `../nubo/goapi-linux`를 자동으로 교체합니다.
- NUBO 디렉터리가 없으면 `dist/goapi-linux`에 생성합니다.
- NUBO를 다른 디렉터리 이름이나 위치에 clone했다면 첫 번째 인자로 출력 경로를 지정합니다.

```bash
./scripts/build-ubuntu22.sh /path/to/nubo/goapi-linux
```

빌드 과정은 Ubuntu 22.04 안에서 필요한 공유 라이브러리를 설치하고, `ldd`로 누락된 라이브러리가 없는지와 glibc 2.35보다 새로운 심볼을 요구하지 않는지를 검사합니다. 같은 바이너리를 Ubuntu 24.04 컨테이너에서도 다시 검사합니다. 산출물은 Ubuntu 22.04 이상 x86-64 서버에서 `libvips42` 런타임 라이브러리가 설치된 환경을 기준으로 합니다.

빌드한 파일을 NUBO 디렉터리로 옮긴 뒤 그 위치에서 실행합니다.

```bash
cp ./goapi-linux /var/www/nubo/goapi-linux
cd /var/www/nubo
./goapi-linux install
./goapi-linux
```

Apple Silicon, ARM Linux 등에서는 해당 환경에서 직접 빌드하는 것이 가장 단순합니다. 교차 컴파일은 `libvips`와 CGO용 크로스 툴체인도 함께 준비해야 합니다.

## `.env` 핵심 설정

전체 템플릿은 NUBO 저장소의 [env.sample](https://github.com/sirini/nubo/blob/main/env.sample)에 있습니다.

### 서버와 데이터베이스

```dotenv
GOAPI_BASE=goapi
GOAPI_PORT=3006
GOAPI_DOMAIN=https://example.com
GOAPI_TITLE=My NUBO
GOAPI_VERSION=1.2.1

DB_HOST=localhost
DB_PORT=3306
DB_USER=nubo
DB_PASS=change-me
DB_NAME=nubo
DB_TABLE_PREFIX=nubo_
DB_UNIX_SOCKET=
```

- `GOAPI_DOMAIN`은 브라우저가 접속하는 공개 사이트 주소이며 운영 환경에서는 HTTPS 주소를 사용합니다.
- TCP로 DB에 연결하면 `DB_HOST`와 `DB_PORT`를 사용합니다.
- Unix socket을 지정하면 `DB_UNIX_SOCKET`이 우선합니다.
- `DB_TABLE_PREFIX`는 SQL 식별자의 일부이므로 설치 후 임의로 바꾸지 마세요.

### 보안 키

```dotenv
JWT_SECRET_KEY=자동으로-생성됨
SYNC_SECRET_KEY=자동으로-생성됨
JWT_ACCESS_HOURS=1
JWT_REFRESH_DAYS=30
```

두 비밀키는 최초 설치 과정에서 서로 다른 무작위 값으로 생성됩니다. 외부에 공개하거나 여러 사이트에서 재사용하지 마세요. 기존 사이트에서 키를 변경하면 로그인 세션이나 외부 동기화 연동이 끊길 수 있습니다.

## Resend 메일 설정

GOAPI는 Gmail SMTP를 지원하지 않으며 **Resend만 사용**합니다. 신규 설치로 생성되는 `.env`에는 다음 항목이 포함되고 `GMAIL_*` 항목은 생성되지 않습니다.

```dotenv
RESEND_API_KEY=
RESEND_FROM_EMAIL=
RESEND_FROM_NAME=
RESEND_REPLY_TO_EMAIL=
```

설정 순서:

1. [Resend](https://resend.com)에서 계정을 만듭니다.
2. 발신 도메인을 추가하고 Resend가 제공하는 SPF/DKIM DNS 레코드를 등록합니다.
3. 도메인이 `Verified`인지 확인합니다.
4. API 키를 만들어 `RESEND_API_KEY`에 입력합니다.
5. 인증된 도메인의 주소를 `RESEND_FROM_EMAIL`에 입력하고 GOAPI를 다시 시작합니다.

```dotenv
RESEND_API_KEY=re_xxxxxxxxx
RESEND_FROM_EMAIL=noreply@example.com
RESEND_FROM_NAME=My NUBO
RESEND_REPLY_TO_EMAIL=admin@example.com
```

- `RESEND_FROM_EMAIL`을 비우면 `GOAPI_DOMAIN`의 호스트로 `noreply@도메인`을 구성하지만, 명시적으로 설정하는 편이 문제를 찾기 쉽습니다.
- `RESEND_FROM_NAME`을 비우면 `GOAPI_TITLE`을 사용합니다.
- `RESEND_REPLY_TO_EMAIL`은 선택 사항이며 Gmail 같은 외부 주소도 가능합니다.
- 관리자 단체 메일은 Resend의 연락처·세그먼트·Broadcast API를 사용하므로 API 키에 해당 작업 권한이 필요합니다. 간단하게 운영하려면 Full Access 키를 사용할 수 있습니다.
- `.env`에 키를 넣어도 Resend 도메인 인증이 끝나지 않았거나 발신 주소의 도메인이 다르면 실제 발송은 실패합니다.

회원가입 인증, 비밀번호 초기화, 댓글 알림의 발송 요청은 `mail_delivery` 테이블에도 기록됩니다. 최근 30일 요약과 전체 페이지 목록은 관리자 메일 화면에서 확인할 수 있으며, 이 조회 기능은 Resend API나 웹훅에 접속하지 않습니다. 수신자·유형·제목·제공자 응답 ID·성공 또는 실패 상태만 저장하고 메일 본문과 인증 코드는 저장하지 않습니다. 기존 설치는 새 바이너리를 배포한 뒤 `./goapi-linux install`을 실행해 테이블을 추가하세요.

## 가입 정책

```dotenv
# verified_email | invite_only | disabled
SIGNUP_MODE=verified_email
```

| 값 | 동작 |
| --- | --- |
| `verified_email` | Resend 인증 코드 또는 설정된 소셜 로그인의 인증된 이메일로 가입 완료 |
| `invite_only` | 관리자 발급 초대 링크와 초대 이메일이 일치해야 가입 가능 |
| `disabled` | 신규 가입 요청 차단 |

알 수 없는 값은 안전하게 `verified_email`로 처리합니다. Resend가 없으면 일반 이메일 가입은 완료할 수 없지만 설정된 OAuth를 통한 신규 가입은 가능합니다. `invite_only`와 `disabled`에서는 신규 OAuth 가입도 차단되고 기존 회원 로그인만 유지됩니다. 소규모 비공개 사이트는 `invite_only`를 사용할 수 있습니다.

## 선택 연동

```dotenv
OAUTH_GOOGLE_CLIENT_ID=
OAUTH_GOOGLE_SECRET=
OAUTH_NAVER_CLIENT_ID=
OAUTH_NAVER_SECRET=
OAUTH_KAKAO_CLIENT_ID=
OAUTH_KAKAO_SECRET=
OPENAI_API_KEY=
OPENAI_IMAGE_DESCRIPTION_ENABLED=false
OPENAI_IMAGE_DESCRIPTION_MODEL=gpt-4o-mini
OPENAI_IMAGE_DESCRIPTION_MAX_PER_POST=3
OPENAI_IMAGE_DESCRIPTION_CONCURRENCY=1
```

- OAuth 키는 해당 소셜 로그인을 사용할 때만 필요합니다.
- OAuth 제공자에는 `https://example.com/goapi/...` 형태의 콜백 경로를 정확히 등록해야 합니다.
- OpenAI 키는 자격 증명일 뿐 기능 활성화 동의로 간주하지 않습니다. 이미지 설명은 키와 함께 `OPENAI_IMAGE_DESCRIPTION_ENABLED=true`를 설정해야 호출됩니다.
- 이미지 설명은 기본적으로 게시글당 최대 3개, 서버 전체 동시 1개로 제한됩니다. 모델과 상한은 위 환경 변수로 변경할 수 있으며 API 사용료는 운영자가 부담합니다.

## 개발과 검증

```bash
go test ./...
go vet ./...
go run ./cmd
```

`go run ./cmd` 역시 현재 작업 디렉터리에서 `.env`와 `env.sample`을 찾습니다. GOAPI 저장소에서 직접 실행하려면 NUBO의 설정 파일을 복사하거나, NUBO 디렉터리에서 빌드한 바이너리를 실행하세요. 실제 운영 DB 대신 별도 개발 DB를 사용하는 것을 권장합니다.

이미지 변환 호출부는 `pkg/imageprocessor.Processor` 뒤에 있으며 govips를 사용합니다. Ubuntu 22.04의 libvips 8.12에서 컴파일·실행되는 v2.16으로 고정되어 있습니다. JPEG 입력 및 WebP 다중 변형 성능은 다음 명령으로 확인합니다.

```bash
go test ./pkg/imageprocessor -run '^$' -bench BenchmarkGovipsProcessorVariants -benchmem
```

## 자주 겪는 문제

- `No .env file found`: 바이너리를 NUBO 디렉터리에서 실행했는지 확인합니다.
- DB 접속 실패: `DB_HOST`, `DB_PORT`, socket 경로와 DB 계정 권한을 확인합니다.
- 이미지 처리 실패: 실행 환경에 `libvips`가 설치되어 있는지 확인합니다.
- 메일 설정은 보이지만 발송 실패: Resend 도메인 인증 상태와 `RESEND_FROM_EMAIL` 도메인을 확인합니다.
- 업데이트 후 테이블/컬럼 오류: `./goapi-linux install`을 실행합니다.
- 더 필요한 안내는 [NUBO README](https://github.com/sirini/nubo#readme) 또는 [nubohub.org](https://nubohub.org)를 참고하세요.

## 관련 프로젝트

- [NUBO](https://github.com/sirini/nubo): 웹 프런트엔드와 배포 패키지
- [TSBOARD](https://github.com/sirini/tsboard): NUBO가 계승한 이전 프로젝트

## 라이선스

MIT License
