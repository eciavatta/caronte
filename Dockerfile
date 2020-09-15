# BUILD STAGE
FROM ubuntu:20.04 AS BUILDSTAGE

# Install tools and libraries
RUN apt-get update && \
	DEBIAN_FRONTEND=noninteractive apt-get install -qq git golang-1.14 pkg-config libpcap-dev libhyperscan-dev yarnpkg curl

RUN ln -sf ../lib/go-1.14/bin/go /usr/bin/go


COPY . /caronte

WORKDIR /caronte

RUN go mod download && go build

RUN cd frontend && \
	yarnpkg install && \
	yarnpkg build --production=true
RUN curl -sf https://gobinaries.com/tj/node-prune | sh && cd /caronte/frontend && node-prune


# LAST STAGE
FROM ubuntu:20.04

COPY --from=BUILDSTAGE /caronte /caronte

RUN apt-get update && \
	DEBIAN_FRONTEND=noninteractive apt-get install -qq libpcap-dev libhyperscan-dev && \
	rm -rf /var/lib/apt/lists/*

ENV GIN_MODE release

WORKDIR /caronte

CMD ./caronte
