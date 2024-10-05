FROM golang:1.21-alpine3.19  AS build-env

RUN echo $GOPATH

RUN apk add --no-cache git gcc musl-dev
RUN apk add --update make
WORKDIR /go/src/github.com/devtron-labs/kubewatch
ADD . /go/src/github.com/devtron-labs/kubewatch
RUN GOOS=linux make

FROM alpine:3.19

RUN apk add --update ca-certificates

RUN adduser -D devtron

COPY --from=build-env  /go/src/github.com/devtron-labs/kubewatch/kubewatch .

RUN chown devtron:devtron ./kubewatch

RUN chmod +x ./kubewatch

USER devtron

ENTRYPOINT ["./kubewatch"]