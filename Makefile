VERSION=0.1.5-dev
GOVARS = -X main.Version=$(VERSION)
SYSTEM = ${GOOS}_${GOARCH}

build:
	rm dist/stack-supervisor
	go build -trimpath -ldflags "-s -w -X main.Version=0.1.5-dev" -o dist ./cmd/stack-supervisor

build-dist:
	go build -trimpath -ldflags "-s -w $(GOVARS)" -o build/bin/stack-supervisor-$(VERSION)-$(SYSTEM) ./cmd/stack-supervisor

stack-supervisor:
	go build -trimpath -ldflags "-s -w $(GOVARS)" -o dist ./cmd/stack-supervisor

build-dist-all:
	go run tools/build-all.go

package-setup:
	if [ ! -d "build/archives" ]; then\
		mkdir -p build/archives;\
	fi

package: build-dist package-setup

	mkdir -p build/stack-supervisor-$(VERSION)-$(SYSTEM);\
	cp README.md build/stack-supervisor-$(VERSION)-$(SYSTEM)
	if [ "${GOOS}" = "windows" ]; then\
		cp build/bin/stack-supervisor-$(VERSION)-$(SYSTEM) build/stack-supervisor-$(VERSION)-$(SYSTEM)/stack-supervisor.exe;\
		cd build;\
		zip -r -q -T archives/stack-supervisor-$(VERSION)-$(SYSTEM).zip stack-supervisor-$(VERSION)-$(SYSTEM);\
	else\
		cp build/bin/stack-supervisor-$(VERSION)-$(SYSTEM) build/stack-supervisor-$(VERSION)-$(SYSTEM)/stack-supervisor;\
		cd build;\
		tar -czf archives/stack-supervisor-$(VERSION)-$(SYSTEM).tar.gz stack-supervisor-$(VERSION)-$(SYSTEM);\
	fi

clean:
	rm -rf build

lint:
	golangci-lint run cmd/stack-supervisor
