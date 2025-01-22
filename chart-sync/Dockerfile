FROM golang:1.22 AS build-env

RUN apt update && \
	apt install git gcc musl-dev make -y && \
	go install github.com/google/wire/cmd/wire@latest

WORKDIR /go/src/github.com/devtron-labs/chart-sync

ADD . /go/src/github.com/devtron-labs/chart-sync

RUN GOOS=linux make

FROM ubuntu

RUN apt update && \
	apt install ca-certificates -y && \
	apt clean autoclean && \
	apt autoremove -y && rm -rf /var/lib/apt/lists/* && \
	useradd -ms /bin/bash devtron

COPY --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/chart-sync/chart-sync .

USER devtron

CMD ["./chart-sync"]
