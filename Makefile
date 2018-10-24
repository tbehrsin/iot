
.PHONY: clean

all: \
	bin/iot-gateway

clean:
	rm -f bin/*

deps:
	@go get github.com/cespare/reflex
	@cd src/iot; GOPATH=$(shell pwd) dep ensure
	@GOPATH=$(shell pwd) bash -xe src/iot/install-v8.sh

bin/iot-gateway: src/iot/*.go src/iot/*/*.go
	GOPATH=$(shell pwd) go build -o bin/iot-gateway iot

watch: deps
	@cd src/iot; \
	GOPATH=$(shell pwd)/../.. reflex \
		-r '\.(go|c|cpp|h|js|lua)$\' \
		-G 'vendor/' \
		-s -- sh -c 'cd ../..; make -s all && bin/iot-gateway'
