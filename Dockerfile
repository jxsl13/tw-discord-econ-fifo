FROM golang:alpine as build


LABEL maintainer "github.com/jxsl13"

RUN apk --update add git openssh && \
    rm -rf /var/lib/apt/lists/* && \
    rm /var/cache/apk/*

WORKDIR /build
COPY . ./
COPY go.* ./


RUN CGO_ENABLED=0 go build -a -ldflags '-w -extldflags "-static"' -o tw-discord-econ-fifo .


FROM alpine:latest as minimal

WORKDIR /app
COPY --from=build /build/tw-discord-econ-fifo .
VOLUME ["/data", "/app/.env"]
ENTRYPOINT ["/app/tw-discord-econ-fifo"]