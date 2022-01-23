VERSION := $(git describe --tags --abbrev=0)
LDFLAGS := "-X \"main.Version=$(VERSION)\""

default: build

.PHONY: build
build:
	go mod download
	go build -ldflags=$(LDFLAGS) -v github.com/eciavatta/caronte/cmd/caronte

run:
	go run github.com/eciavatta/caronte/cmd/caronte

.PHONY: setcap
setcap: caronte
	sudo setcap cap_net_raw,cap_net_admin=eip ./caronte

clean:
	rm -rf caronte

remove_pcaps:
	rm -rf backend/pcaps/*.pcap backend/pcaps/processing/*.pcap

.PHONY: test
test:
	go test -v -race github.com/eciavatta/caronte/...

coverage: test
	go test -v -coverprofile=coverage.txt github.com/eciavatta/caronte/...

build_deps:
	docker build -t ghcr.io/eciavatta/test-environment:latest -f .github/docker/Dockerfile-environment .

run_deps:
	docker run -d --name caronte-mongo -p 127.0.0.1:27017:27017 mongo:4.4
	docker run -d --name caronte-test-environment -p 127.0.0.1:2222:22 -p 127.0.0.1:8080:8080 ghcr.io/eciavatta/test-environment:latest

start_deps:
	docker start caronte-mongo
	docker start caronte-test-environment

stop_deps:
	docker stop caronte-mongo || exit 0
	docker stop caronte-test-environment || exit 0

destroy_deps:
	(docker stop caronte-mongo && docker rm caronte-mongo) || exit 0
	(docker stop caronte-test-environment && docker rm caronte-test-environment) || exit 0
