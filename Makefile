# Makefile

LDFLAGS     ?= "-s -w"
GOBUILD_ENV = GO111MODULE=on CGO_ENABLED=0
GOX         = go run github.com/mitchellh/gox
TARGETS     := darwin/amd64 linux/amd64 windows/amd64
DIST_DIRS   := find * -type d -exec

.PHONY: build cross-build compress test lint down tidy clean

all: build

build:
	$(GOBUILD_ENV) go build -o bin/kubectl-resource-versions -a -ldflags $(LDFLAGS)

cross-build:
	rm -rf _bin
	$(GOBUILD_ENV) $(GOX) -ldflags $(LDFLAGS) -parallel=3 -output="_bin/kubectl-resource-versions-{{.OS}}-{{.Arch}}/kubectl-resource-versions" -osarch='$(TARGETS)' .

compress:
	( \
		cd _bin && \
		$(DIST_DIRS) cp ../LICENSE {} \; && \
		$(DIST_DIRS) cp ../README.md {} \; && \
		$(DIST_DIRS) tar -zcf {}.tar.gz {} \; && \
		$(DIST_DIRS) zip -r {}.zip {} \; && \
		sha256sum *.tar.gz *.zip > sha256sums.txt \
	)

test:
	go test -v ./...

lint:
	golangci-lint run --enable=golint --disable=errcheck --disable=typecheck && goimports -l -w . && go fmt ./... && go vet ./...

down:
	go list ./... && go mod verify

tidy:
	rm -f go.sum && go mod tidy -v

clean:
	rm -rf bin _bin
