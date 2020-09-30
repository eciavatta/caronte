# BUILD STAGE
FROM ubuntu:20.04 AS BUILDSTAGE

# Install tools and libraries
RUN apt-get update && \
	DEBIAN_FRONTEND=noninteractive apt-get install -qq golang-1.14 pkg-config libpcap-dev libhyperscan-dev yarnpkg

COPY . /caronte

WORKDIR /caronte

RUN ln -sf ../lib/go-1.14/bin/go /usr/bin/go && \
    go mod download && \
    go build && \
    cd frontend && \
	yarnpkg install && \
	yarnpkg build --production=true && \
	cd - && \
	mkdir -p /caronte-build/frontend && \
	cp -r caronte pcaps/ scripts/ shared/ test_data/ /caronte-build && \
	cp -r frontend/build/ /caronte-build/frontend


# LAST STAGE
FROM ubuntu:20.04

COPY --from=BUILDSTAGE /caronte-build /caronte

RUN apt-get update && \
	DEBIAN_FRONTEND=noninteractive apt-get install -qq golang-1.14 pkg-config libpcap-dev libhyperscan-dev && \
	rm -rf /var/lib/apt/lists/*

ENV GIN_MODE release

WORKDIR /caronte

CMD ./caronte
