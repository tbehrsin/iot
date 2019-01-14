
.PHONY: clean push-iot-dns push-iot-server debian

all: \
	bin/iot-gateway \
	bin/iot-server \
	bin/iot-dns

clean:
	rm -f bin/*

deps:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/cespare/reflex
	cd src/gateway; GOPATH=$(shell pwd) dep ensure
	mkdir -p src/gateway/vendor/github.com/behrsin
	ln -sf $(shell pwd)/../go-v8 src/gateway/vendor/github.com/behrsin/go-v8
	src/gateway/vendor/github.com/behrsin/go-v8/install-v8.sh
	test -d src/api && (cd src/api; GOPATH=$(shell pwd) dep ensure && GOPATH=$(shell pwd) go install ./vendor/github.com/golang/protobuf/protoc-gen-go/) || true
	test -d src/cli && (cd src/cli; GOPATH=$(shell pwd) dep ensure) || true
	test -d src/server && (cd src/server; GOPATH=$(shell pwd) dep ensure) || true
	test -d src/dns && (cd src/dns; GOPATH=$(shell pwd) dep ensure) || true
	test -d src/db && (cd src/db; GOPATH=$(shell pwd) dep ensure) || true

PATH := $(shell pwd)/bin:$(PATH)

src/api/protocol/%.pb.go: src/api/protocol/%.proto
	protoc --go_out=source_relative:. $<

bin/iot-gateway: $(patsubst %.proto, %.pb.go, $(shell test -d src/gateway && find src/gateway src/api ../go-v8 -name "*.go" -or -name "*.proto" -or -name "*.cc" -or -name "*.h"))
	GOPATH=$(shell pwd) go build -o bin/iot-gateway gateway

bin/iot: $(patsubst %.proto, %.pb.go, $(shell test -d src/cli && find src/cli src/api -name "*.go" -or -name "*.proto" -or -name "*.cc" -or -name "*.h"))
	GOPATH=$(shell pwd) go build -o bin/iot cli

bin/iot-server: $(shell test -d src/server && find src/server src/db -name "*.go" ! -path "src/server/vendor/*")
	GOPATH=$(shell pwd) go build -o bin/iot-server server

bin/iot-dns: $(shell test -d src/dns && find src/dns src/db -name "*.go" ! -path "src/dns/vendor/*")
	GOPATH=$(shell pwd) go build -o bin/iot-dns dns

bin/iot-zigbee: $(shell test -d src/zigbee && find src/gateway src/zigbee -name "*.go")
	GOPATH=$(shell pwd) go build -o bin/iot-zigbee zigbee

zigbee: bin/iot-zigbee
	bin/iot-zigbee

iot: bin/iot-gateway
	@cd apps; DATABASE_FILE=$(shell pwd)/z3js.db BLUETOOTH_EMULATION=true ../bin/iot-gateway -profile

iot-debug:
	@cd apps; DATABASE_FILE=$(shell pwd)/z3js.db GOPATH=$(shell pwd) BLUETOOTH_EMULATION=true dlv -l localhost:2345 --headless debug gateway -- -profile

dns: bin/iot-dns
	PORT=8053 bin/iot-dns

server: bin/iot-server
	INSECURE_PORT=8080 PORT=8443 bin/iot-server

push-iot-gateway:
	$(eval TMP := $(shell mktemp -d))
	@rsync -ar --exclude-from=.syncignore --delete . $(TMP)
	@cd $(TMP)/src/gateway; GOPATH=$(TMP) dep ensure
	@cd $(TMP); GOPATH=$(TMP) GOOS=linux GOARCH=arm $(TMP)/src/gateway/vendor/github.com/behrsin/go-v8/install-v8.sh
	@cd $(TMP); rsync -ar --exclude=/pkg/ --delete . iot-gateway:/app
	@rm -Rf $(TMP)
	@ssh iot-gateway "cd /app && make bin/iot-gateway && sudo systemctl restart iot-gateway && sudo journalctl -xfu iot-gateway"

push-iot-dns:
	rm -f bin/iot-dns
	GOOS=linux GOARCH=amd64 make bin/iot-dns
	rm -Rf /tmp/iot-dns
	mkdir -p /tmp/iot-dns
	cp Dockerfile.dns bin/iot-dns /tmp/iot-dns
	docker build -t gcr.io/behrsin-iot/iot-dns:latest -f Dockerfile.dns /tmp/iot-dns/
	docker push gcr.io/behrsin-iot/iot-dns:latest
	cat src/dns/iot-dns.service | ssh iot-ns1.behrsin.com "sudo cp /dev/stdin /etc/systemd/system/iot-dns.service && sudo systemctl enable iot-dns && sudo systemctl restart iot-dns"
	cat src/dns/iot-dns.service | ssh iot-ns2.behrsin.com "sudo cp /dev/stdin /etc/systemd/system/iot-dns.service && sudo systemctl enable iot-dns && sudo systemctl restart iot-dns && journalctl -xfu iot-dns"

push-iot-server:
	rm -f bin/iot-server
	GOOS=linux GOARCH=amd64 make bin/iot-server
	rm -Rf /tmp/iot-server
	mkdir -p /tmp/iot-server
	cp Dockerfile.server bin/iot-server /tmp/iot-server
	docker build -t gcr.io/behrsin-iot/iot-server:latest -f Dockerfile.server /tmp/iot-server/
	docker push gcr.io/behrsin-iot/iot-server:latest
	# cat src/turn/iot-turn.rules.sh | ssh iot-ns1.behrsin.com "sudo cp /dev/stdin /usr/sbin/iot-turn.rules.sh && sudo chmod +x /usr/sbin/iot-turn.rules.sh"
	# cat src/turn/iot-turn.rules.sh | ssh iot-ns2.behrsin.com "sudo cp /dev/stdin /usr/sbin/iot-turn.rules.sh && sudo chmod +x /usr/sbin/iot-turn.rules.sh"
	cat src/server/iot-server.service | ssh iot-ns2.behrsin.com "sudo cp /dev/stdin /etc/systemd/system/iot-server.service && sudo systemctl enable iot-server && sudo systemctl restart iot-server"
	cat src/server/iot-server.service | ssh iot-ns1.behrsin.com "sudo cp /dev/stdin /etc/systemd/system/iot-server.service && sudo systemctl enable iot-server && sudo systemctl restart iot-server && journalctl -xfu iot-server"

debian:
	cd debian; docker-compose -p iot up --build --force-recreate

watch: deps
	GOPATH=$(shell pwd) reflex \
		-r '\.(go|c|cpp|h|js|lua)$\' \
		-G 'src/client/' \
		-G 'ios/' \
		-G 'android/' \
		-G 'app.json' \
		-G 'src/gateway/vendor/' \
		-G 'src/server/vendor/' \
		-G 'src/dns/vendor/' \
		-G 'pkg/' \
		-s -- sh -c 'make -s all && bin/iot-gateway'
