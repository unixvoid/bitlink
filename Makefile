GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
OS_PERMS=sudo
GIT_HASH=$(shell git rev-parse HEAD | head -c 10)

all: bitlink

dependencies:
	go get github.com/gorilla/mux
	go get github.com/unixvoid/glogger
	go get gopkg.in/gcfg.v1
	go get gopkg.in/redis.v5
	go get golang.org/x/crypto/sha3

daemon:
	bin/bitlink &

bitlink:
	$(GOC) bitlink.go

run:
	go run \
		bitlink/bitlink.go \
		bitlink/link_compressor.go \
		bitlink/token_generator.go

prep_aci: stat
	mkdir -p stage.tmp/bitlink-layout/rootfs/
	cp bin/bitlink* stage.tmp/bitlink-layout/rootfs/bitlink
	cp config.gcfg stage.tmp/bitlink-layout/rootfs/
	cp deps/manifest.json stage.tmp/bitlink-layout/manifest

build_aci: prep_aci
	# build image
	cd stage.tmp/ && \
		actool build bitlink-layout bitlink-api.aci && \
		mv bitlink-api.aci ../
	@echo "bitlink-api.aci built"

build_travis_aci: prep_aci
	wget https://github.com/appc/spec/releases/download/v0.8.7/appc-v0.8.7.tar.gz
	tar -zxf appc-v0.8.7.tar.gz
	# build image
	cd stage.tmp/ && \
		../appc-v0.8.7/actool build bitlink-layout bitlink-api.aci && \
		mv bitlink-api.aci ../
	rm -rf appc-v0.8.7*
	@echo "bitlink-api.aci built"

stat:
	mkdir -p bin/
	$(CGOR) $(GOC) $(GOFLAGS) -o bin/bitlink-$(GIT_HASH)-linux-amd64 bitlink/*.go

install: stat
	cp bitlink /usr/bin

clean:
	rm -rf bin/
	rm -rf stage.tmp/
	rm -f bitlink-api.aci
