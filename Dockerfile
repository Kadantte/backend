FROM docker.io/golang:1.23.4-alpine AS builder

WORKDIR /build/src

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target="/root/.cache/go-build" \
    --mount=type=bind,target=. \
    go build -v -ldflags "-s -w" -trimpath -o /build/morbo

FROM alpine:3.20

LABEL org.opencontainers.image.source=https://github.com/morbo-org/Morbo

EXPOSE 80

ARG USER=morbo
ENV HOME=/home/$USER

RUN adduser -D $USER

COPY --from=builder /build/morbo /usr/local/bin

USER $USER
WORKDIR $HOME

RUN mkdir -p ~/.local/share/morbo

ENTRYPOINT [ "morbo" ]
