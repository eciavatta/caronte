FROM ubuntu:20.04

# Install tools and libraries
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -qq git golang-1.14 pkg-config libpcap-dev libhyperscan-dev yarnpkg

RUN ln -sf ../lib/go-1.14/bin/go /usr/bin/go

ENV GIN_MODE release

COPY . /caronte

WORKDIR /caronte

RUN go mod download && go build

RUN cd frontend && yarnpkg install && yarnpkg build

CMD ./caronte
