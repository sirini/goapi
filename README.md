# GOAPI for NUBO

<p align="center">
  <img src="https://img.shields.io/github/v/tag/sirini/goapi?style=flat-square&color=E07A5F" alt="version">
  <img src="https://img.shields.io/github/license/sirini/goapi?style=flat-square&color=5D6D7E" alt="license">
  <img src="https://img.shields.io/github/stars/sirini/goapi?style=flat-square&color=F4D03F" alt="stars">
  <img src="https://img.shields.io/github/last-commit/sirini/goapi?style=flat-square&color=2ECC71" alt="last commit">
</p>

GOAPI는 [NUBO](https://github.com/sirini/nubo)의 백엔드입니다. GoFiber v3로 HTTP API를 제공하고 MySQL/MariaDB의 회원·게시물·알림 데이터를 처리하며, 이미지 변환과 Resend 메일 발송도 담당합니다.

> 문서 기준: 2026-08-25 · 최신 통합 버전: NUBO/GOAPI 1.2.29

GOAPI는 NUBO와 별도 제품으로 배포하지 않습니다. v1.2.26부터 두 저장소의 공개 버전을 동일하게 맞추고, NUBO 릴리스 manifest가 실제 구성 요소 commit과 API contract를 고정합니다. 운영자는 GOAPI 버전을 따로 선택하거나 교체하지 않습니다.

Ubuntu 22.04 이상 x86-64 서버에서 NUBO를 운영한다면 **이 저장소를 따로 clone하거나 Go를 설치해
빌드하지 마세요.** NUBO 저장소의 `npm run server:install`이 Nuxt와 GOAPI, libvips, `nuboctl`,
systemd unit을 하나의 검증된 릴리스로 설치합니다. 이 저장소는 GOAPI를 수정하거나 macOS·다른 Linux
배포판·다른 CPU에서 소스를 직접 빌드해 시험하려는 경우에 사용합니다.

## 담당 기능

- 회원가입, 로그인, JWT 갱신 및 권한 관리
- 게시판·게시물·댓글·알림·채팅 데이터 처리
- 업로드 이미지 리사이즈 및 메타데이터 처리
- 게시물 권한과 파일 소유 관계를 다시 검사하는 단기 원본 이미지 스트리밍과 HTTP byte range 처리
- Resend 기반 가입 인증, 비밀번호 초기화, 댓글 알림
- Resend Broadcast 기반 관리자 단체 메일
- 이메일 인증·초대 전용·가입 중지 정책
- 반복 실행 가능한 데이터베이스 스키마 설치

## 실행 구조

| 항목 | 기본값 | 설명 |
| --- | --- | --- |
| API 주소 | `http://127.0.0.1:3006/goapi` | `GOAPI_PORT`, `GOAPI_BASE`로 변경 |
| 수신 주소 | `0.0.0.0` | `GOAPI_HOST`; prebuilt는 `127.0.0.1` 권장 |
| 설정 파일 | `.env` 또는 `NUBO_ENV_FILE` 경로 | Nuxt와 GOAPI가 함께 사용 |
| 업로드 루트 | `NUBO_UPLOAD_DIR` 또는 `./upload` | DB/URL의 `/upload/...` 경로는 그대로 유지 |
| 설치 템플릿 | NUBO 디렉터리의 `env.sample` | 최초 실행 시 `.env` 생성에 사용 |
| 데이터베이스 | MySQL/MariaDB | 테이블 접두사 지원 |

기존 소스 설치는 GOAPI 바이너리를 **NUBO 프로젝트 디렉터리에서 실행해야 합니다.** Prebuilt 설치는 `NUBO_ENV_FILE=/etc/nubo/nubo.env`처럼 외부 설정 파일을 명시하여 릴리스 디렉터리와 설정을 분리할 수 있습니다. 최초 대화형 설치에는 여전히 현재 디렉터리의 `env.sample`이 필요합니다.

`NUBO_UPLOAD_DIR`에는 `/var/lib/nubo/upload`나 기존 `/var/www/example.org/upload`처럼 절대 경로를 지정할 수 있습니다. 값을 생략하면 기존처럼 실행 작업 디렉터리의 `./upload`를 사용하며, 기존 심볼릭 링크도 계속 동작합니다. 파일시스템 위치와 무관하게 DB와 HTTP 경로는 `/upload/...` 형태를 유지합니다.

## Ubuntu 서버에서 NUBO 운영

공식 prebuilt 운영 범위는 Ubuntu 22.04 이상 x86-64입니다. NUBO 저장소만 내려받아 설치하세요.
`npm install`, GOAPI 저장소, Go toolchain, `libvips-dev`는 필요하지 않습니다.

```bash
git clone --depth=1 https://github.com/sirini/nubo.git
cd nubo
npm run server:install
sudo /opt/nubo/current/nuboctl activate-nginx
```

설치 과정은 통합 릴리스와 SHA-256을 검증하고 외부 환경 파일, 데이터베이스와 최초 관리자, 업로드 경로,
systemd 서비스를 준비합니다. 설치 후 두 프로세스는 하나의 대표 service로 관리합니다.

```bash
sudo systemctl status nubo nubo-goapi nubo-web
sudo systemctl restart nubo
sudo journalctl -u nubo-goapi -u nubo-web -f
```

DB와 업로드를 외부에 백업한 뒤 NUBO 저장소에서 공식 설치를 업데이트합니다.

```bash
nuboctl update --dry-run
nuboctl update
nuboctl status
```

`nuboctl update`가 checkout의 안전한 fast-forward, 통합 asset 검증, 필요한 DB migration, 원자적 전환과 readiness 확인을 한 번에 수행합니다. GOAPI 변경이 포함되면 실행 전에 DB·업로드 외부 백업을 확인합니다.

기존 소스·PM2 설치를 systemd 기반 prebuilt로 전환하는 `server:adopt` 절차까지 포함한 운영 안내는
[NUBO README](https://github.com/sirini/nubo#readme)를 따르세요. GOAPI만 따로 교체하거나
`./goapi-linux install`을 수동 실행하는 방식은 공식 서버 업데이트 절차가 아닙니다.

## 소스에서 직접 빌드하기

### 요구 사항

- Go 1.25 이상
- CGO를 사용할 수 있는 C/C++ 빌드 환경
- libvips 8.14 이상과 pkg-config
- 실행 시 MySQL 또는 MariaDB

직접 빌드한 실행 파일은 현재 OS와 CPU에서만 사용할 수 있습니다. 공식 NUBO 릴리스용 Linux 바이너리를
만드는 방법과는 다르며, 개발·시험 또는 자신이 직접 검증하는 커스텀 배포를 위한 경로입니다.

### Ubuntu와 다른 Linux

Ubuntu에서 소스를 시험하는 예시입니다. 다른 배포판은 C 컴파일러, pkg-config, libvips 개발 패키지의
이름을 해당 패키지 관리자에 맞게 바꾸세요.

```bash
sudo apt update
sudo apt install build-essential pkg-config libvips-dev
git clone https://github.com/sirini/goapi.git
cd goapi
go test ./...
go build -trimpath -o goapi-local ./cmd
```

위 패키지는 GOAPI 소스를 호스트에서 직접 빌드하는 개발자에게만 필요합니다. NUBO 공식 릴리스 사용자는 `libvips-dev`나 `libvips42`를 설치하지 않습니다.

ARM Linux에서도 같은 네이티브 빌드 방식을 시험할 수 있습니다. 다만 공식 prebuilt, `nuboctl` 운영 자동화,
릴리스 QA의 지원 범위에는 포함되지 않습니다. CGO와 libvips 때문에 단순한 `GOOS`/`GOARCH` 교차
컴파일보다 대상 환경에서 직접 빌드하는 편이 안전합니다.

### macOS

Homebrew와 Xcode Command Line Tools가 준비된 환경에서 다음처럼 빌드할 수 있습니다.

```bash
xcode-select --install
brew install go pkg-config vips mysql
brew services start mysql
git clone https://github.com/sirini/goapi.git
cd goapi
export CGO_CFLAGS_ALLOW="-Xpreprocessor"
go test ./...
go build -trimpath -o goapi-local ./cmd
```

NUBO 화면까지 함께 시험하려면 NUBO도 clone하고 Node.js 22 이상에서 `npm install`을 실행합니다.
빌드한 `goapi-local`을 NUBO 디렉터리로 옮겨 그곳에서 실행하면 `env.sample`과 `.env`, `upload`의 기존
상대 경로 계약을 그대로 사용할 수 있습니다.

```bash
git clone https://github.com/sirini/nubo.git ../nubo
cp goapi-local ../nubo/
cd ../nubo
brew install node@22
export PATH="$(brew --prefix node@22)/bin:$PATH"
npm install
./goapi-local
# 다른 터미널에서: npm run dev
```

최초 실행의 질문에 DB 정보와 관리자 계정을 입력하면 `.env`, 데이터베이스와 기본 테이블, 관리자 계정을
준비하고 API를 시작합니다. 운영 데이터를 복사하지 말고 별도 개발 DB를 사용하세요.

### Windows

Windows에서는 WSL2에 Ubuntu 22.04 이상을 설치하는 방법을 권장합니다. WSL2 안에서는 앞의 공식
`server:install` 또는 Linux 소스 빌드 절차를 그대로 사용하고 Windows 브라우저에서 localhost로
접속할 수 있습니다. systemd 기반 공식 설치를 사용할 때는 WSL의 systemd가 활성화되어 있어야 합니다.

네이티브 Windows는 govips가 권장 빌드 환경이나 정기 CI를 제공하지 않으며 NUBO도 검증하지 않습니다.
직접 시도하려면 Go 1.25 이상, CGO를 지원하는 C toolchain, pkg-config와 호환되는 libvips 개발 파일을
모두 준비해야 하지만, 재현 가능한 공식 설치 절차나 prebuilt 실행 파일은 현재 제공하지 않습니다.

### 배포용 Ubuntu 22.04 호환 묶음

NUBO 공식 통합 릴리스에 넣을 x86-64 Linux 바이너리는 호스트 운영체제에서 직접 빌드하지 않고 Docker의 Ubuntu 22.04 환경에서 만듭니다. sharp-libvips 기반 libvips를 함께 넣으므로 서버에 libvips 패키지를 별도로 설치할 필요가 없습니다.

Docker와 Buildx가 준비된 환경에서 다음 스크립트를 실행하세요.

```bash
./scripts/build-ubuntu22.sh
```

- 기본 출력은 이 저장소의 `dist/nubo-runtime/bin/goapi`와 `dist/nubo-runtime/lib/`입니다. NUBO 소스 저장소를 자동으로 수정하지 않습니다.
- NUBO 통합 릴리스 빌드는 첫 번째부터 세 번째 인자로 staging의 바이너리·라이브러리·라이선스 경로를 지정합니다.
- x86-64 기본 호환판은 `lib/`, sharp 공식 x86-64-v2판은 `lib/glibc-hwcaps/x86-64-v2/`에 생성됩니다.
- 라이선스와 호환판 빌드 출처는 `licenses/sharp-libvips/`에 생성됩니다. 두 번째와 세 번째 인자로 각 경로를 바꿀 수 있습니다.

```bash
./scripts/build-ubuntu22.sh /path/to/release/bin/goapi \
  /path/to/release/lib \
  /path/to/release/licenses/sharp-libvips
```

빌드 과정은 공식 npm 패키지와 sharp-libvips 소스의 고정 버전·SHA-256을 확인하고, GOAPI에 상대 경로(`$ORIGIN/../lib`)를 기록합니다. glibc는 CPU가 x86-64-v2를 만족하면 공식 최적화판을, 그렇지 않으면 `-march=x86-64` 호환판을 자동 선택합니다. 시스템 libvips가 없는 Ubuntu 22.04와 24.04에서 두 변형을 시험하고, SSE4가 없는 QEMU `qemu64` CPU에서도 JPEG→WebP 변환을 검증합니다.

빌드한 파일을 NUBO 디렉터리로 옮긴 뒤 그 위치에서 실행합니다.

```bash
cp ./dist/nubo-runtime/bin/goapi /var/www/nubo/bin/goapi
cp -a ./dist/nubo-runtime/lib /var/www/nubo/lib
cd /var/www/nubo
./bin/goapi install
./bin/goapi
```

Apple Silicon, ARM Linux 등에서는 위와 같이 해당 환경에서 직접 빌드하는 것이 가장 단순합니다. 교차 컴파일은 `libvips`와 CGO용 크로스 툴체인도 함께 준비해야 합니다.

## `.env` 핵심 설정

전체 템플릿은 NUBO 저장소의 [env.sample](https://github.com/sirini/nubo/blob/main/env.sample)에 있습니다.

기본값은 현재 디렉터리의 `.env`이며 기존 설치 방식은 그대로 동작합니다. 외부 설정을 사용할 때는 실행 전에 경로를 명시합니다.

```bash
NUBO_ENV_FILE=/etc/nubo/nubo.env ./goapi-linux
```

GOAPI는 프로세스 환경, 지정한 파일, 코드 기본값 순서로 설정을 선택합니다. 따라서 systemd 같은 프로세스 관리자가 직접 제공한 값이 파일보다 우선합니다. 같은 파일을 prebuilt Nuxt의 `node --env-file=/etc/nubo/nubo.env`에도 전달할 수 있도록 `NUXT_*` 값은 `${...}` 참조 없이 최종값으로 작성해야 합니다.

### 서버와 데이터베이스

```dotenv
GOAPI_BASE=goapi
GOAPI_PORT=3006
GOAPI_DOMAIN=https://example.com
GOAPI_TITLE=My NUBO
GOAPI_VERSION=1.2.29

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

### 원본 이미지 스트리밍

`/board/original`은 게시물 보기 레벨·포인트 잔액, 비밀글, 삭제글, 작성자 차단과 file–board 소유 관계를 검사한 뒤 2분짜리 전송 토큰만 발급합니다. `/board/original/transfer`는 토큰을 소비해 원본을 인라인으로 전송하고 byte range를 지원합니다. 실제 파일시스템 경로는 게시글 JSON이나 브라우저 URL에 노출하지 않으며, 게시물 열람 때 처리한 보기 포인트를 다시 차감하지 않습니다.

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

회원가입 인증, 비밀번호 초기화, 댓글 알림의 발송 요청은 `mail_delivery` 테이블에도 기록됩니다. 최근 30일 요약과 전체 페이지 목록은 관리자 메일 화면에서 확인할 수 있으며, 이 조회 기능은 Resend API나 웹훅에 접속하지 않습니다. 수신자·유형·제목·제공자 응답 ID·성공 또는 실패 상태만 저장하고 메일 본문과 인증 코드는 저장하지 않습니다. 공식 서버 설치는 NUBO의 `nuboctl update`가 필요한 migration을 적용합니다. 직접 빌드한 개발 환경에서는 새 실행 파일로 `./goapi-local install`을 실행해 테이블을 추가할 수 있습니다.

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
# Android 앱이 발급받은 Google ID 토큰의 audience(Web application client ID). 미설정 시 위 client ID 사용
OAUTH_GOOGLE_ANDROID_CLIENT_ID=
OAUTH_NAVER_CLIENT_ID=
OAUTH_NAVER_SECRET=
OAUTH_KAKAO_CLIENT_ID=
OAUTH_KAKAO_SECRET=
OPENAI_API_KEY=
OPENAI_IMAGE_DESCRIPTION_ENABLED=false
OPENAI_IMAGE_DESCRIPTION_MODEL=gpt-5.6-luna
OPENAI_IMAGE_DESCRIPTION_MAX_PER_POST=3
OPENAI_IMAGE_DESCRIPTION_CONCURRENCY=1
FIREBASE_PROJECT_ID=
FIREBASE_CREDENTIALS_FILE=
```

- OAuth 키는 해당 소셜 로그인을 사용할 때만 필요합니다.
- OAuth 제공자에는 `https://example.com/goapi/...` 형태의 콜백 경로를 정확히 등록해야 합니다.
- OpenAI 키는 자격 증명일 뿐 기능 활성화 동의로 간주하지 않습니다. 이미지 설명은 키와 함께 `OPENAI_IMAGE_DESCRIPTION_ENABLED=true`를 설정해야 호출됩니다.
- 이미지 설명은 기본적으로 게시글당 최대 3개, 서버 전체 동시 1개로 제한됩니다. 모델과 상한은 위 환경 변수로 변경할 수 있으며 API 사용료는 운영자가 부담합니다.

### Android 실시간 푸시 알림

Android 앱의 댓글·좋아요·1:1 대화 알림을 실시간으로 보내려면 Firebase 프로젝트에서 서비스 계정 JSON을 발급하고 서버 외부의 읽기 제한된 경로에 저장합니다. 저장소나 웹 공개 디렉터리에는 자격 증명을 두지 마세요.

```dotenv
FIREBASE_PROJECT_ID=my-firebase-project
FIREBASE_CREDENTIALS_FILE=/etc/nubo/firebase-service-account.json
```

두 값을 비우면 푸시 발송만 안전하게 비활성화되며 기존 알림 목록은 계속 동작합니다. 앱은 Firebase 프로젝트 설정이 없을 때 주기 조회 방식으로 자동 대체합니다. `push_device` 테이블이 없는 기존 설치는 새 실행 파일로 `install` 명령을 한 번 실행해 재실행 가능한 스키마 업데이트를 적용하세요.

## 개발과 검증

```bash
go test ./...
go vet ./...
go run ./cmd
```

`go run ./cmd` 역시 기본적으로 현재 작업 디렉터리에서 `.env`와 `env.sample`을 찾습니다. GOAPI 저장소에서 직접 실행하려면 NUBO의 설정 파일을 복사하거나 `NUBO_ENV_FILE`에 절대 경로를 지정하세요. 실제 운영 DB 대신 별도 개발 DB를 사용하는 것을 권장합니다.

이미지 변환 호출부는 `pkg/imageprocessor.Processor` 뒤에 있으며 govips v2.18을 사용합니다. 공식 릴리스는 sharp-libvips 1.3.2의 libvips 8.18.3과 이미지 코덱을 CPU 호환판·최적화판으로 함께 배포합니다. JPEG 입력 및 WebP 다중 변형 성능은 다음 명령으로 확인합니다.

```bash
go test ./pkg/imageprocessor -run '^$' -bench BenchmarkGovipsProcessorVariants -benchmem
```

## 자주 겪는 문제

- `load environment file`: 기본 `.env`가 있는 디렉터리에서 실행했는지, 또는 `NUBO_ENV_FILE` 경로와 읽기 권한이 올바른지 확인합니다.
- DB 접속 실패: `DB_HOST`, `DB_PORT`, socket 경로와 DB 계정 권한을 확인합니다.
- 이미지 처리 실패: 공식 릴리스의 `lib/libvips-cpp.so.8.18.3`과 `lib/glibc-hwcaps/x86-64-v2/libvips-cpp.so.8.18.3` 파일을 확인합니다.
- 메일 설정은 보이지만 발송 실패: Resend 도메인 인증 상태와 `RESEND_FROM_EMAIL` 도메인을 확인합니다.
- 업데이트 후 테이블/컬럼 오류: 공식 서버는 NUBO의 `nuboctl update` 로그와 `nubo-goapi` journal을 확인합니다. 직접 빌드한 개발 환경만 `./goapi-local install`을 실행합니다.
- 더 필요한 안내는 [NUBO README](https://github.com/sirini/nubo#readme) 또는 [nubohub.org](https://nubohub.org)를 참고하세요.

## 관련 프로젝트

- [NUBO](https://github.com/sirini/nubo): 웹 프런트엔드와 배포 패키지
- [TSBOARD](https://github.com/sirini/tsboard): NUBO가 계승한 이전 프로젝트

## 라이선스

MIT License
