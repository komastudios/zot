# ---
# Stage 1: Install certs, build binary, create default config file
# ---
FROM --platform=$BUILDPLATFORM ghcr.io/project-zot/golang:1.22 AS builder

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y npm && rm -rf /var/lib/apt/lists/*

ARG TARGETOS
ARG TARGETARCH
ARG COMMIT=""
ARG ZUI_VERSION=""
ARG ZUI_REPO_OWNER="komastudios"
ARG ZUI_REPO_NAME="zui"

RUN mkdir -p /go/src/github.com/project-zot/zot
WORKDIR /go/src/github.com/project-zot/zot
COPY . .
RUN make COMMIT=$COMMIT OS=$TARGETOS ARCH=$TARGETARCH ZUI_VERSION=$ZUI_VERSION ZUI_REPO_OWNER=$ZUI_REPO_OWNER ZUI_REPO_NAME=$ZUI_REPO_NAME clean binary
RUN echo '{\n\
    "storage": {\n\
        "rootDirectory": "/var/lib/registry"\n\
    },\n\
    "http": {\n\
        "address": "0.0.0.0",\n\
        "port": "5000"\n\
    },\n\
    "log": {\n\
        "level": "debug"\n\
    }\n\
}\n' > config.json && cat config.json

# ---
# Stage 2: Final image with nothing but certs, binary, and default config file
# ---
FROM gcr.io/distroless/base-debian12 AS final
ARG TARGETOS
ARG TARGETARCH
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/project-zot/zot/bin/zot-$TARGETOS-$TARGETARCH /usr/bin/zot
COPY --from=builder /go/src/github.com/project-zot/zot/config.json /etc/zot/config.json
ENTRYPOINT ["/usr/bin/zot"]
EXPOSE 5000
VOLUME ["/var/lib/registry"]
CMD ["serve", "/etc/zot/config.json"]
