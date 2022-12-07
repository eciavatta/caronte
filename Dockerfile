# Build backend with go
FROM golang:1.19 AS BACKEND_BUILDER

# Install tools and libraries
RUN apt-get update && \
	DEBIAN_FRONTEND=noninteractive apt-get install -qq \
	git \
	pkg-config \
	libpcap-dev \
	libhyperscan-dev

WORKDIR /caronte

COPY . .

RUN export VERSION=$(git describe --tags --abbrev=0) && \
    go mod download && \
    go build -ldflags "-X main.Version=$VERSION" -v github.com/eciavatta/caronte/cmd/caronte && \
	mkdir -p build/pcaps/processing build/pcaps/connections build/shared && \
	cp -r caronte build/

# Build web via yarn
FROM node:16 as WEB_BUILDER

WORKDIR /caronte-web

COPY ./web ./

RUN yarn install && yarn build --production=true


# LAST STAGE
FROM ubuntu:20.04

COPY --from=BACKEND_BUILDER /caronte/build /opt/caronte

COPY --from=WEB_BUILDER /caronte-web/build /opt/caronte/web/build

RUN apt-get update && \
	DEBIAN_FRONTEND=noninteractive apt-get install -qq \
	libpcap-dev \
	libhyperscan-dev && \
	rm -rf /var/lib/apt/lists/*

ENV GIN_MODE=release

ENV MONGO_HOST=127.0.0.1

ENV MONGO_PORT=27017

WORKDIR /opt/caronte

ENTRYPOINT sleep 3 && ./caronte -mongo-host ${MONGO_HOST} -mongo-port ${MONGO_PORT} -assembly_memuse_log
