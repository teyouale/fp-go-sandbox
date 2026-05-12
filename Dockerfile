FROM golang:1.24-alpine

RUN adduser -h /home/sandbox -D sandbox

# Caches and the playground workdir must live OUTSIDE /sandbox — codapi
# bind-mounts the user's code dir onto /sandbox at runtime, which would
# otherwise hide everything we pre-build here.
ENV GOCACHE=/home/sandbox/.cache/go-build \
    GOMODCACHE=/home/sandbox/go/pkg/mod \
    GOFLAGS=-mod=mod \
    CGO_ENABLED=0

USER sandbox
WORKDIR /home/sandbox/playground

# Pre-warm: fetch fp-go (network on), build with all imports so the cache and
# build cache are hot, then lock down. The committed go.mod/go.sum below pin
# the exact versions the sandbox will use at runtime.
RUN go mod init playground && \
    go get github.com/IBM/fp-go/v2/either@latest && \
    go get github.com/IBM/fp-go/v2/option@latest && \
    go get github.com/IBM/fp-go/v2/io@latest && \
    go get github.com/IBM/fp-go/either@latest && \
    printf 'package main\nimport (\n\t_ "fmt"\n\tE "github.com/IBM/fp-go/v2/either"\n\tO "github.com/IBM/fp-go/v2/option"\n\t_ "github.com/IBM/fp-go/v2/io"\n)\nvar _ = E.Right[error](1)\nvar _ = O.Some(1)\nfunc main() {}\n' > warmup.go && \
    go build -o /dev/null . && \
    rm -f warmup.go

# After this point, no network — sandbox must resolve everything from cache.
ENV GOPROXY=off GOFLAGS=-mod=mod
