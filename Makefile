
.PHONY: clean push-iot-dns push-iot-server debian

all: \
	bin/iot-gateway \
	bin/iot-server \
	bin/iot-dns

clean:
	rm -f bin/*

deps:
	@go get -u github.com/golang/dep/cmd/dep
	@go get -u github.com/cespare/reflex
	@cd src/iot; GOPATH=$(shell pwd) dep ensure
	@GOPATH=$(shell pwd) bash src/iot/install-v8.sh
	@test -d src/server && (cd src/server; GOPATH=$(shell pwd) dep ensure) || true
	@test -d src/dns && (cd src/dns; GOPATH=$(shell pwd) dep ensure) || true
	@test -d src/db && (cd src/db; GOPATH=$(shell pwd) dep ensure) || true

bin/iot-gateway: src/iot/*.go src/iot/*/*.go
	GOPATH=$(shell pwd) go build -o bin/iot-gateway iot

bin/iot-server: $(shell test -d src/server && find src/server src/db -name "*.go" ! -path "src/server/vendor/*")
	GOPATH=$(shell pwd) go build -o bin/iot-server server

bin/iot-dns: $(shell test -d src/dns && find src/dns src/db -name "*.go" ! -path "src/dns/vendor/*")
	GOPATH=$(shell pwd) go build -o bin/iot-dns dns

push-iot-gateway:
	rsync -arv --exclude-from=.syncignore --delete . iot-gateway:/app
	ssh iot-gateway "cd /app && make bin/iot-gateway && sudo systemctl restart iot-gateway && sudo journalctl -xfu iot-gateway"

push-iot-dns:
	GOOS=linux GOARCH=amd64 make bin/iot-dns
	rm -Rf /tmp/iot-dns
	mkdir -p /tmp/iot-dns
	cp Dockerfile.dns bin/iot-dns /tmp/iot-dns
	docker build -t 637256544704.dkr.ecr.eu-west-2.amazonaws.com/iot-dns:latest -f Dockerfile.dns /tmp/iot-dns/
	$(shell aws ecr get-login --profile=behrsin --no-include-email)
	docker push 637256544704.dkr.ecr.eu-west-2.amazonaws.com/iot-dns:latest
	ssh admin@ns1.z3js.net "sudo systemctl restart iot-dns && sudo journalctl -xfu iot-dns"

push-iot-server:
	GOOS=linux GOARCH=amd64 make bin/iot-server
	rm -Rf /tmp/iot-server
	mkdir -p /tmp/iot-server
	cp Dockerfile.server bin/iot-server /tmp/iot-server
	docker build -t 637256544704.dkr.ecr.eu-west-2.amazonaws.com/iot-server:latest -f Dockerfile.server /tmp/iot-server/
	$(shell aws ecr get-login --profile=behrsin --no-include-email)
	docker push 637256544704.dkr.ecr.eu-west-2.amazonaws.com/iot-server:latest
	ssh admin@z3js.net "sudo systemctl restart iot-server && sudo journalctl -xfu iot-server"

debian:
	cd debian; docker-compose -p iot up --build --force-recreate

watch: deps
	GOPATH=$(shell pwd) reflex \
		-r '\.(go|c|cpp|h|js|lua)$\' \
		-G 'src/client/' \
		-G 'ios/' \
		-G 'android/' \
		-G 'app.json' \
		-G 'src/iot/vendor/' \
		-G 'src/server/vendor/' \
		-G 'src/dns/vendor/' \
		-G 'pkg/' \
		-s -- sh -c 'make -s all && bin/iot-gateway'
