VERSION=0.1.5-dev
GOVARS = -X main.Version=$(VERSION)
SYSTEM = ${GOOS}_${GOARCH}

build:
	rm dist/{{project.name}}
	go build -trimpath -ldflags "-s -w -X main.Version=0.1.5-dev" -o dist ./cmd/{{project.name}}

build-dist:
	go build -trimpath -ldflags "-s -w $(GOVARS)" -o build/bin/{{project.name}}-$(VERSION)-$(SYSTEM) ./cmd/{{project.name}}

{{project.name}}:
	go build -trimpath -ldflags "-s -w $(GOVARS)" -o dist ./cmd/{{project.name}}

build-dist-all:
	go run tools/build-all.go

package-setup:
	if [ ! -d "build/archives" ]; then\
		mkdir -p build/archives;\
	fi

package: build-dist package-setup

	mkdir -p build/{{project.name}}-$(VERSION)-$(SYSTEM);\
	cp README.md build/{{project.name}}-$(VERSION)-$(SYSTEM)
	if [ "${GOOS}" = "windows" ]; then\
		cp build/bin/{{project.name}}-$(VERSION)-$(SYSTEM) build/{{project.name}}-$(VERSION)-$(SYSTEM)/{{project.name}}.exe;\
		cd build;\
		zip -r -q -T archives/{{project.name}}-$(VERSION)-$(SYSTEM).zip {{project.name}}-$(VERSION)-$(SYSTEM);\
	else\
		cp build/bin/{{project.name}}-$(VERSION)-$(SYSTEM) build/{{project.name}}-$(VERSION)-$(SYSTEM)/{{project.name}};\
		cd build;\
		tar -czf archives/{{project.name}}-$(VERSION)-$(SYSTEM).tar.gz {{project.name}}-$(VERSION)-$(SYSTEM);\
	fi

clean:
	rm -rf build

lint:
	golangci-lint run cmd/{{project.name}}
