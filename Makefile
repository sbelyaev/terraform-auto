BINDIR      := $(CURDIR)/bin
TARGETS     = linux/amd64 windows/amd64
BINNAME     = terraform

.PHONY: all
all: deps build

.PHONY: deps
deps:
	go get -d -v -t ./...

.PHONY: test
test:
	find cmd/testdata -type f -name 'terraform*' -exec chmod +x {} \;
	go test ./...

.PHONY: build
build: $(BINDIR)/$(BINNAME)

$(BINDIR)/$(BINNAME):
	go build -o $(BINDIR)/$(BINNAME) ./cmd

.PHONY: release
release:
	@for TARGET in $(TARGETS); do \
		echo $$TARGET ;\
			echo go get -d -v -t ./... && \
			export GOOS=$$(echo $$TARGET | cut -d/ -f1) && \
			export GOARCH=$$(echo $$TARGET | cut -d/ -f2) && \
			go build -o $(BINDIR)/$(BINNAME)-$$GOOS-$$GOARCH ./cmd; \
	done

.PHONY: clean
clean:
	rm -f bin/*
