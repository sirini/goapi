docker run --rm -v "$PWD":/app -w /app ubuntu:22.04 /bin/bash -c "
  # 1. 필수 패키지 설치 (CGO를 위한 gcc 및 다운로드를 위한 wget 포함)
  apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y wget ca-certificates build-essential libvips-dev && \
  
  # 2. 프로젝트 요구사항에 맞는 Go 버전 다운로드 및 설치
  wget https://go.dev/dl/go1.26.4.linux-amd64.tar.gz && \
  rm -rf /usr/local/go && tar -C /usr/local -xzf go1.26.4.linux-amd64.tar.gz && \
  export PATH=\$PATH:/usr/local/go/bin && \
  
  # 3. 설치된 Go 버전 확인 및 빌드 진행
  go version && \
  go mod tidy && \
  go get -u ./... && \
  go build -o ../nubo.git/goapi-linux cmd/main.go

  # 정리
  rm -rf go1.26.4.linux-amd64.tar.gz
"
