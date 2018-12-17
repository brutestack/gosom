BUILD=go build
CLEAN=go clean
INSTALL=go install
BUILDPATH=./_build
PACKAGES=$(shell go list ./... | grep -v /examples/)
EXAMPLES=$(shell find examples/* -maxdepth 0 -type d -exec basename {} \;)

examples: builddir
	for example in $(EXAMPLES); do \
		go build -o "$(BUILDPATH)/$$example" "examples/$$example/$$example.go"; \
	done

all: examples

colors: builddir
	go build -o "$(BUILDPATH)/colors" "examples/colors/colors.go"

fcps: builddir
	go build -o "$(BUILDPATH)/fcps" "examples/fcps/fcps.go"

builddir:
	mkdir -p $(BUILDPATH)

install:
	$(INSTALL) ./$(EXDIR)/...

clean:
	rm -rf $(BUILDPATH)

godep:
	wget -O- https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

dep:
	dep ensure -v

check:
	for pkg in ${PACKAGES}; do \
		go vet $$pkg || exit ; \
		golint $$pkg || exit ; \
	done

test:
	for pkg in ${PACKAGES}; do \
		go test -coverprofile="../../../$$pkg/coverage.txt" -covermode=atomic $$pkg || exit; \
	done

.PHONY: clean examples
