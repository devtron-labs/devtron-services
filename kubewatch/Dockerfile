FROM golang:1.25-alpine  AS build-env

RUN echo $GOPATH && \
    apk add --no-cache git gcc musl-dev && \
    apk add --update make

WORKDIR /go/src/github.com/devtron-labs/kubewatch
ADD . /go/src/github.com/devtron-labs/kubewatch
RUN GOOS=linux make

FROM alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

RUN apk upgrade --no-cache && \
    apk add --no-cache ca-certificates && \
    adduser -D devtron

COPY --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/kubewatch/kubewatch .

RUN chmod +x ./kubewatch

USER devtron

ENTRYPOINT ["./kubewatch"]
