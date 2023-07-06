VERSION=0.1.5-dev
GOVARS = -X main.Version=$(VERSION)
SYSTEM = ${GOOS}_${GOARCH}

build:
	rm dist/stackup
	go build -trimpath -ldflags "-s -w -X main.Version=0.1.5-dev" -o dist ./cmd/stackup

build-dist:
	go build -trimpath -ldflags "-s -w $(GOVARS)" -o build/bin/stackup-$(VERSION)-$(SYSTEM) ./cmd/stackup

stackup:
	go build -trimpath -ldflags "-s -w $(GOVARS)" -o dist ./cmd/stackup

build-dist-all:
	go run tools/build-all.go

package-setup:
	if [ ! -d "build/archives" ]; then\
		mkdir -p build/archives;\
	fi

package: build-dist package-setup

	mkdir -p build/stackup-$(VERSION)-$(SYSTEM);\
	cp README.md build/stackup-$(VERSION)-$(SYSTEM)
	if [ "${GOOS}" = "windows" ]; then\
		cp build/bin/stackup-$(VERSION)-$(SYSTEM) build/stackup-$(VERSION)-$(SYSTEM)/stackup.exe;\
		cd build;\
		zip -r -q -T archives/stackup-$(VERSION)-$(SYSTEM).zip stackup-$(VERSION)-$(SYSTEM);\
	else\
		cp build/bin/stackup-$(VERSION)-$(SYSTEM) build/stackup-$(VERSION)-$(SYSTEM)/stackup;\
		cd build;\
		tar -czf archives/stackup-$(VERSION)-$(SYSTEM).tar.gz stackup-$(VERSION)-$(SYSTEM);\
	fi

clean:
	rm -rf build

lint:
	golangci-lint run cmd/stackup
