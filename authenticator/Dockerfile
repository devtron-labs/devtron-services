FROM golang:1.24-alpine3.21 AS build-env
USER root
RUN apk add --no-cache git gcc musl-dev
RUN apk add --update make
RUN go install github.com/google/wire/cmd/wire@latest
WORKDIR /go/src/github.com/devtron-labs/authenticator
ADD . /go/src/github.com/devtron-labs/authenticator/
RUN GOOS=linux make

FROM alpine:3.21.2@sha256:56fa17d2a7e7f168a043a2712e63aed1f8543aeafdcee47c58dcffe38ed51099
RUN apk add --no-cache ca-certificates
RUN apk add git --no-cache
COPY --from=build-env  /go/src/github.com/devtron-labs/authenticator/authenticator .

RUN adduser -D devtron
RUN chown -R devtron:devtron ./authenticator
USER devtron

CMD ["./authenticator"]
