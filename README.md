# GOAPI for NUBO

<p align="center">
  <img src="https://img.shields.io/github/license/sirini/nubo?style=flat-square&color=5D6D7E" alt="license">
  <img src="https://img.shields.io/github/stars/sirini/nubo?style=flat-square&color=F4D03F" alt="stars">
  <img src="https://img.shields.io/github/last-commit/sirini/nubo?style=flat-square&color=2ECC71" alt="last commit">
</p>

GOAPI는 **NUBO의 고성능 백엔드** 구현체입니다. `GoFiber v3` 기반으로 개발하였습니다.

### 🛠 NUBO 프로젝트 기술 스택

| Category     | Tools                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| :----------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Frontend | ![Nuxt](https://img.shields.io/badge/Nuxt_4-00DC82?style=flat-square&logo=nuxtdotjs&logoColor=white) ![Vue](https://img.shields.io/badge/Vue_3-4FC08D?style=flat-square&logo=vuedotjs&logoColor=white) ![Pinia](https://img.shields.io/badge/Pinia-FFE148?style=flat-square&logo=pinia&logoColor=black) ![Tailwind](https://img.shields.io/badge/Tailwind-38B2AC?style=flat-square&logo=tailwindcss&logoColor=white) ![Shadcn](https://img.shields.io/badge/Shadcn_Vue-000000?style=flat-square&logo=shadcnui&logoColor=white) |
| **Backend**  | ![Go](https://img.shields.io/badge/Go_Fiber_v3-00ADD8?style=flat-square&logo=go&logoColor=white)                                                                                                                                                                                                                                                                                                                                                                                                                               |
| Database | ![MySQL](https://img.shields.io/badge/MySQL-4479A1?style=flat-square&logo=mysql&logoColor=white) ![MariaDB](https://img.shields.io/badge/MariaDB-003545?style=flat-square&logo=mariadb&logoColor=white)                                                                                                                                                                                                                                                                                                                        |

- 본래 `TSBOARD` 프로젝트에서 `Bun` 런타임 기반 웹프레임워크인 `ElysiaJS` 를 대체하기 위해 `Go` 언어로 개발한 백엔드입니다. JS/TS 런타임 기반으로 구현한 백엔드 대비하여 동시성 기반의 보다 빠르고 효율적인 연산을 제공하면서도 더 적은 메모리만 사용합니다.
- 26년 부터는 `NUBO` 프로젝트에서 사용하는 백엔드 프로젝트가 되었습니다. `TSBOARD`의 백엔드에서 한층 더 개선되었고, `NUBO`에서 제공하는 추가 기능들을 모두 지원하도록 같이 업데이트 됩니다.
- `GoFiber v3` 기반으로 설계되어 있습니다. 최신 `NUBO` 프로젝트에서는 64bit Linux용 바이너리가 포함됩니다.

## GOAPI 사용 준비

- 기본으로 제공하는 `./goapi-linux` 바이너리는 64bit Linux OS에서, `libvips-dev` 라이브러리가 설치된 환경을 가정하고 컴파일 되었습니다.
- 만약 사용하시는 서버의 CPU가 Intel/AMD가 아닌 Arm 계열일 경우(예를 들어 Mac), 이 프로젝트를 `git clone`으로 내려받아서 본인의 서버에 맞게 새로 컴파일 하여 바이너리 파일을 만들어서 사용하실 수 있습니다.
- `./goapi-linux` 바이너리 (혹은 여러분이 직접 컴파일하신 `./goapi-mac` 등)는 `NUBO` 폴더 아래에 두고 실행하셔야 합니다. (예: `/var/www/nubo.git/goapi-linux`)
- `./goapi-linux` 바이너리는 실행 시점에 동일 경로에 `.env` 파일이 있는지 검사하고, 없다면 설치를 진행합니다.
- `Go` 언어로 작성되어 있으므로 직접 코드를 수정하거나 컴파일하기 위해서는 서버에 `Go` 개발 환경이 준비되어 있어야 합니다.

## GOAPI 최대로 활용하기

- `./goapi-linux` 바이너리 파일은 실행 전에 항상 같은 경로에서 `.env` 파일을 확인합니다.
- `.env` 파일에는 아래의 내용들이 있는데, 소셜 로그인이나 메일 발송 등에 사용되는 키 값들을 보관하는 용도입니다.
- (아래 예시)

```conf
# Resend로 메일 발송에 필요한 API Key
# 주의) localhost 에서는 동작하지 않으며, 실제 도메인 인증을 받은 서버에서만 동작합니다
# 참고) Resend API Key가 지정되어 있으면 Gmail 메일 대신 Resend를 사용합니다
RESEND_API_KEY=

# 구글 앱비밀번호 for GMAIL 발송
# 참고) https://velog.io/@seul06/nodemailer
GMAIL_ID=
GMAIL_APP_PASSWORD=

# 구글 OAuth 클라이언트
OAUTH_GOOGLE_CLIENT_ID=
OAUTH_GOOGLE_SECRET=

# 네이버 OAuth 클라이언트
OAUTH_NAVER_CLIENT_ID=
OAUTH_NAVER_SECRET=

# 카카오 OAuth 클라이언트
OAUTH_KAKAO_CLIENT_ID=
OAUTH_KAKAO_SECRET=

# OpenAI API Key
OPENAI_API_KEY=
```

- GOAPI 백엔드에서 메일을 발송하거나, 소셜 로그인 처리를 원할하게 하려면 가급적 공란을 모두 채워주세요.
- OpenAI API를 제외한 나머지 값들은 모두 무료로 어느 정도 이용이 가능합니다.
  - OpenAI API가 있으면 업로드된 이미지에 대해서 설명글을 생성하여 추후 검색이 가능하게 만들어줍니다.
  - 이밖에도 OpenAI API 기반으로 NUBO에서 다양한 (부가)기능들을 제공할 예정입니다.
  - OpenAI API는 정말로 없어도 됩니다. 메일 발송 등과는 달리 필수적이진 않습니다.

> GOAPI가 제대로 동작하도록 하는 Nginx 설정은 아래 NUBO 프로젝트의 README.md를 참조하세요.

## 참고 리포지토리

- NUBO: https://github.com/sirini/nubo
- TSBOARD: https://github.com/sirini/tsboard
