# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26.4
FROM golang:${GO_VERSION}-bookworm AS go-toolchain

FROM node:24-bookworm-slim AS sharp-libvips

ARG SHARP_LIBVIPS_VERSION=1.3.2
ARG SHARP_LIBVIPS_RUNTIME_SHA256=8cf0eafeaca832b68942fe1a770fb5f3b490504d3a9f2e3f56ee8784c9d65c45
ARG SHARP_LIBVIPS_DEV_SHA256=b125613cd2c9c860fe49d8744aab59f8926f4589bb36f3a2ab3a095382adc755

WORKDIR /packages
RUN npm pack "@img/sharp-libvips-linux-x64@${SHARP_LIBVIPS_VERSION}" \
    && npm pack "@img/sharp-libvips-dev@${SHARP_LIBVIPS_VERSION}" \
    && echo "${SHARP_LIBVIPS_RUNTIME_SHA256}  img-sharp-libvips-linux-x64-${SHARP_LIBVIPS_VERSION}.tgz" | sha256sum --check --strict \
    && echo "${SHARP_LIBVIPS_DEV_SHA256}  img-sharp-libvips-dev-${SHARP_LIBVIPS_VERSION}.tgz" | sha256sum --check --strict \
    && mkdir -p /opt/sharp-libvips/include /opt/sharp-libvips/lib /opt/sharp-libvips/licenses \
    && tar -xzf "img-sharp-libvips-linux-x64-${SHARP_LIBVIPS_VERSION}.tgz" --strip-components=2 -C /opt/sharp-libvips/lib package/lib \
    && tar -xzf "img-sharp-libvips-dev-${SHARP_LIBVIPS_VERSION}.tgz" --strip-components=2 -C /opt/sharp-libvips/include package/include \
    && tar -xOf "img-sharp-libvips-linux-x64-${SHARP_LIBVIPS_VERSION}.tgz" package/README.md > /opt/sharp-libvips/licenses/README.md \
    && tar -xOf "img-sharp-libvips-linux-x64-${SHARP_LIBVIPS_VERSION}.tgz" package/package.json > /opt/sharp-libvips/licenses/package.json \
    && tar -xOf "img-sharp-libvips-linux-x64-${SHARP_LIBVIPS_VERSION}.tgz" package/versions.json > /opt/sharp-libvips/licenses/versions.json \
    && cp /usr/share/common-licenses/Apache-2.0 /opt/sharp-libvips/licenses/Apache-2.0.txt \
    && cp /usr/share/common-licenses/LGPL-3 /opt/sharp-libvips/licenses/LGPL-3.txt

# sharp 공식 빌드와 동일한 소스·도구를 사용하되 최소 CPU만 x86-64로 낮춥니다.
FROM rockylinux/rockylinux:8-ubi-init AS sharp-libvips-compat

ARG SHARP_LIBVIPS_COMMIT=4da6d14c0d59866adfb9d8cf52bcaa53846dc4f6
ARG SHARP_LIBVIPS_SOURCE_SHA256=4f438a108427ed9054c62c134c559af10522815ea42892a9ca31f655d97fc806
ARG RUST_TOOLCHAIN=nightly-2026-08-18
ARG CARGO_C_VERSION=0.10.24+cargo-0.98.0
ARG MESON_VERSION=1.12.0
ARG NINJA_VERSION=1.13.0

ENV RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo \
    PATH=/usr/local/cargo/bin:/opt/rh/gcc-toolset-15/root/usr/bin:${PATH}

RUN dnf update --assumeyes \
    && dnf install --assumeyes epel-release \
    && dnf config-manager --set-enabled powertools \
    && dnf install --assumeyes --setopt=tsflags=nodocs \
        autoconf automake bzip2 cmake curl-devel findutils \
        gcc-toolset-15-gcc gcc-toolset-15-gcc-c++ gettext git gperf jq make \
        nasm openssl-devel patch pkgconf python3.12-pip tar xz \
    && curl --proto '=https' --tlsv1.2 --silent --show-error --fail https://sh.rustup.rs \
        | sh -s -- -y --no-modify-path --profile minimal --default-toolchain "${RUST_TOOLCHAIN}" \
    && cargo install cargo-c --locked --version "${CARGO_C_VERSION}" \
    && pip3.12 install "meson==${MESON_VERSION}" "ninja==${NINJA_VERSION}"

ENV PKG_CONFIG="x86_64-linux-gnu-pkgconf --static" \
    PLATFORM=linux-x64 \
    FLAGS=-march=x86-64 \
    RUSTFLAGS="-Zlocation-detail=none -Zfmt-debug=none" \
    MESON=--cross-file=/root/meson.ini

WORKDIR /packaging
RUN curl --proto '=https' --tlsv1.2 --location --silent --show-error --fail \
        "https://github.com/lovell/sharp-libvips/archive/${SHARP_LIBVIPS_COMMIT}.tar.gz" \
        --output /tmp/sharp-libvips.tar.gz \
    && echo "${SHARP_LIBVIPS_SOURCE_SHA256}  /tmp/sharp-libvips.tar.gz" | sha256sum --check --strict \
    && tar -xzf /tmp/sharp-libvips.tar.gz --strip-components=1 \
    && test "$(grep '^VERSION_VIPS=' versions.properties)" = "VERSION_VIPS=8.18.3" \
    && cp platforms/linux-x64/Toolchain.cmake /root/ \
    && cp platforms/linux-x64/meson.ini /root/ \
    && install -m 0755 platforms/linux-x64/x86_64-linux-gnu-pkgconf /usr/bin/x86_64-linux-gnu-pkgconf \
    && ./build/posix.sh \
    && tar -xzf sharp-libvips-linux-x64.tar.gz -C /opt \
    && cp THIRD-PARTY-NOTICES.md /opt/THIRD-PARTY-NOTICES.md \
    && printf '{\n  "source": "https://github.com/lovell/sharp-libvips",\n  "commit": "%s",\n  "sourceSha256": "%s",\n  "compilerFlags": "-march=x86-64"\n}\n' \
        "${SHARP_LIBVIPS_COMMIT}" "${SHARP_LIBVIPS_SOURCE_SHA256}" > /opt/compat-build.json

FROM ubuntu:22.04 AS build

ENV DEBIAN_FRONTEND=noninteractive
ENV PATH=/usr/local/go/bin:${PATH}
ENV PKG_CONFIG_PATH=/opt/sharp-libvips/lib/pkgconfig
ENV LD_LIBRARY_PATH=/opt/sharp-libvips/lib
# govips에 남아 있는 미사용 ImageMagick 호환 코드를 링크 결과에서 제거합니다.
ENV CGO_CFLAGS="-ffunction-sections -fdata-sections"
ENV CGO_LDFLAGS="-Wl,--gc-sections"

RUN apt-get update \
    && apt-get install --yes --no-install-recommends \
        build-essential \
        ca-certificates \
        patchelf \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

COPY --from=go-toolchain /usr/local/go /usr/local/go
COPY --from=sharp-libvips /opt/sharp-libvips /opt/sharp-libvips
COPY build/sharp-libvips/vips.pc /opt/sharp-libvips/lib/pkgconfig/vips.pc
COPY build/sharp-libvips/runtime_probe.c /src/build/sharp-libvips/runtime_probe.c

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN install -d /out \
    && gcc -O2 -o /out/image-runtime-probe build/sharp-libvips/runtime_probe.c \
        $(pkg-config --cflags --libs vips) \
    && patchelf --set-rpath '$ORIGIN/../lib' /out/image-runtime-probe
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go test -c -o /out/imageprocessor.test ./pkg/imageprocessor \
    && patchelf --set-rpath '$ORIGIN/../lib' /out/imageprocessor.test \
    && /out/imageprocessor.test
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/goapi-linux ./cmd \
    && patchelf --set-rpath '$ORIGIN/../lib' /out/goapi-linux

FROM ubuntu:22.04 AS verify

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install --yes --no-install-recommends binutils \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/goapi-linux /opt/nubo/bin/goapi-linux
COPY --from=build /out/imageprocessor.test /opt/nubo/bin/imageprocessor.test
COPY --from=build /out/image-runtime-probe /opt/nubo/bin/image-runtime-probe
COPY --from=sharp-libvips-compat /opt/lib/libvips-cpp.so.8.18.3 /opt/nubo/lib/libvips-cpp.so.8.18.3
COPY --from=sharp-libvips /opt/sharp-libvips/lib/libvips-cpp.so.8.18.3 /opt/nubo/lib/glibc-hwcaps/x86-64-v2/libvips-cpp.so.8.18.3
COPY --from=sharp-libvips /opt/sharp-libvips/licenses /opt/nubo/licenses/sharp-libvips
COPY --from=sharp-libvips-compat /opt/THIRD-PARTY-NOTICES.md /opt/nubo/licenses/sharp-libvips/THIRD-PARTY-NOTICES.md
COPY --from=sharp-libvips-compat /opt/compat-build.json /opt/nubo/licenses/sharp-libvips/compat-build.json
RUN ldd /opt/nubo/bin/goapi-linux \
    && ! ldd /opt/nubo/bin/goapi-linux | grep --quiet "not found" \
    && ldd /opt/nubo/bin/goapi-linux | grep --quiet "/opt/nubo/bin/../lib/glibc-hwcaps/x86-64-v2/libvips-cpp.so.8.18.3" \
    && readelf --dynamic /opt/nubo/bin/goapi-linux | grep --quiet '\$ORIGIN/../lib' \
    && ! readelf --version-info /opt/nubo/bin/goapi-linux | grep --extended-regexp "GLIBC_2\.(3[6-9]|[4-9][0-9])" \
    && /opt/nubo/bin/imageprocessor.test \
    && /opt/nubo/bin/image-runtime-probe

FROM ubuntu:24.04 AS verify-newer

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install --yes --no-install-recommends binutils \
    && rm -rf /var/lib/apt/lists/*

COPY --from=verify /opt/nubo /opt/nubo
RUN ldd /opt/nubo/bin/goapi-linux \
    && ! ldd /opt/nubo/bin/goapi-linux | grep --quiet "not found" \
    && ldd /opt/nubo/bin/goapi-linux | grep --quiet "/opt/nubo/bin/../lib/glibc-hwcaps/x86-64-v2/libvips-cpp.so.8.18.3" \
    && /opt/nubo/bin/imageprocessor.test \
    && /opt/nubo/bin/image-runtime-probe

FROM ubuntu:22.04 AS verify-compat

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install --yes --no-install-recommends qemu-user \
    && rm -rf /var/lib/apt/lists/*

COPY --from=verify-newer /opt/nubo /opt/nubo
RUN qemu-x86_64 -cpu qemu64 /opt/nubo/bin/image-runtime-probe

FROM scratch AS artifact
COPY --from=verify-compat /opt/nubo/bin/goapi-linux /goapi-linux
COPY --from=verify-compat /opt/nubo/lib /lib
COPY --from=verify-compat /opt/nubo/licenses /licenses
