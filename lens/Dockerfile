FROM golang:1.25-alpine3.21 AS build-env

RUN apk add --no-cache git gcc musl-dev && \
	apk add --update make

WORKDIR /go/src/github.com/devtron-labs/lens

ADD . /go/src/github.com/devtron-labs/lens

RUN go install github.com/google/wire/cmd/wire@latest && \
	GOOS=linux make

FROM alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

RUN apk upgrade --no-cache && \
    apk add --no-cache ca-certificates && \
	adduser -D devtron

COPY --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/lens/lens /go/src/github.com/devtron-labs/lens/scripts/ .

USER devtron

CMD ["./lens"]
