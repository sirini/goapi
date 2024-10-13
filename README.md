# GOAPI for TSBOARD

<p align="center">
    <img src="https://img.shields.io/badge/go-00ADD8.svg?&style=for-the-badge&&logoColor=white"/>
    <img src="https://img.shields.io/badge/MySQL-4479A1.svg?&style=for-the-badge&&logoColor=white"/>
    <img src="https://img.shields.io/badge/TSBOARD-000000.svg?&style=for-the-badge&&logoColor=3178C6"/>
</p>

## GOAPI는 무엇인가요?

- 짧은 설명: GOAPI는 **TSBOARD의 고성능 백엔드** 구현체로, `Go`언어(`Fiber`)로 작성되었습니다.
- 조금 길게: 타입스크립트로 작성된 오픈소스 커뮤니티 빌더이자 게시판인 **TSBOARD** 프로젝트가 있습니다.
  해당 프로젝트는 타입스크립트 단일 언어로 프론트엔드와 백엔드 모두 개발되어 있습니다.
  백엔드 코드의 동작을 위해 현재는 JS/TS 런타임 엔진이자 툴킷으로 유명한 `Bun`(<https://bun.sh>)을 사용중입니다.
  그러나, 아래와 같은 이유로 백엔드 코드를 타입스크립트(in `Bun`)에서 지금 보고 계신 **GOAPI로 변경할 예정**입니다.
  - `Bun`은 훌륭한 TS/JS 런타임 엔진이지만, **가상 서버 환경에서는 제대로 동작하지 않는 제약**이 있습니다.
  - 비단 `Bun` 뿐만 아니라, `Deno`/`Node`로 구성한 API 서버는 `pm2` 등을 이용한 부하 분산이 추가로 요구됩니다.
  - 마지막으로 `Bun`의 런타임 성능은 우수하지만, 더 집중된 부하를 견딜 수 있게 TSBOARD를 개선하고 싶었습니다.

> 백엔드가 Bun 기반의 타입스크립트 코드에서 Go언어로 작성된 바이너리로 교체되더라도, 기존에 사용하던 기능들은 모두 그대로 사용하실 수 있습니다.

## TSBOARD를 사용하려면 GOAPI도 필요한가요?

- TSBOARD는 v0.9.8 현재 TS/JS 런타임인 `Bun` 기반으로 백엔드 코드들을 동작시키고 있습니다.
- GOAPI로의 전환 시기는 아직 미정이지만, 그 전에는 TSBOARD 프로젝트에 포함된 server 코드들만으로 충분합니다.
- GOAPI로의 전환 준비가 완료되면, TSBOARD 프로젝트에서 **백엔드 바이너리가 기본으로 포함**되어 배포됩니다.
  - 전환 이후 TSBOARD에서 타입스크립트로 작성된 기존 코드들은 제거됩니다.
  - `Bun` 런타임에 의존적인 API 요청/응답 코드들도 모두 재작성됩니다.
  - 프론트엔드쪽 코드는 백엔드 변경의 영향을 최소한으로 받도록 할 예정입니다.

> 백엔드용으로 미리 컴파일된 바이너리를 그대로 쓰셔도 되며, 혹시 원하실 경우 이 곳 GOAPI 프로젝트를 clone 하셔서 본인의 커뮤니티/사이트 용도에 맞게 수정 후 다시 컴파일하여 사용하실 수 있습니다.

## TSBOARD에서 백엔드를 서비스하려면 이제 어떻게 해야 하나요?

- GOAPI 전환 전에는 기존처럼 `tsboard.git` 폴더로 이동 후 `bun server/index.ts` 를 통해 실행 할 수 있습니다.
- GOAPI 전환 후에는 아래와 같은 절차대로 백엔드를 실행 하실 수 있습니다.
  - `tsboard.git/server/` 폴더에서 본인의 서버 OS에 맞는 바이너리 실행
    - 리눅스의 경우 `server/tsboard-goapi-linux` 로 실행
    - 윈도우의 경우 `server/tsboard-goapi-win.exe` 로 실행
    - 맥의 경우 `server/tsboard-goapi-mac` 으로 실행
  - (필요한 경우) 해당 파일에 실행 권한 부여 (리눅스에서는 `chmod +x ./tsboard-goapi-linux`)

> TSBOARD의 백엔드를 완전히 GOAPI로 교체하는 동안, TSBOARD 프로젝트 자체적인 버전업은 계속될 예정입니다. 교체 완료 시점에 TSBOARD 공식 홈페이지를 통해서 상세한 안내를 드리겠습니다.

---

1. TSBOARD 공식 홈페이지 <https://tsboard.dev>
2. TSBOARD GitHub <https://github.com/sirini/tsboard>
3. GeekNews 소개글 <https://news.hada.io/topic?id=14914>
