FROM ubuntu:20.04

# Install tools and libraries
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -qq git golang libpcap-dev pkg-config libhyperscan-dev npm

ENV GIN_MODE release

COPY . /caronte

WORKDIR /caronte

RUN go mod download && go build

RUN npm i --global yarn && cd frontend && yarn install && yarn build

CMD ./caronte
