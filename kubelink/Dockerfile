FROM golang:1.22 AS build-env

RUN apt update && \
	apt install git gcc musl-dev make -y && \
	go install github.com/google/wire/cmd/wire@latest

WORKDIR /go/src/github.com/devtron-labs/kubelink
ADD . /go/src/github.com/devtron-labs/kubelink/
RUN GOOS=linux make

FROM ubuntu:22.04@sha256:1b8d8ff4777f36f19bfe73ee4df61e3a0b789caeff29caa019539ec7c9a57f95
RUN apt update && \
	apt install ca-certificates -y && \
	apt clean autoclean && \
	apt autoremove -y && rm -rf /var/lib/apt/lists/* && \
	useradd -ms /bin/bash devtron 
	
COPY --chown=devtron:devtron --from=build-env  /go/src/github.com/devtron-labs/kubelink/kubelink .

USER devtron

CMD ["./kubelink"]
