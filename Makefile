
BUILDDIR          := ${CURDIR}/build
TARBUILDDIR       := ${BUILDDIR}/tar
ARCH              := $(shell go env GOHOSTARCH)
OS                := $(shell go env GOHOSTOS)
GOVER             := $(shell go version | awk '{print $$3}' | tr -d '.')
APP_NAME          := static-server
APP_VER           := $(shell git describe --always --dirty --tags|sed 's/^v//')
VERSION_VAR       := main.ServerVersion
GOTEST_FLAGS      := -cpu=1,2
GOBUILD_DEPFLAGS  := -tags netgo
GOBUILD_LDFLAGS   ?=
GOBUILD_FLAGS     := ${GOBUILD_DEPFLAGS} -ldflags "${GOBUILD_LDFLAGS} -X ${VERSION_VAR}=${APP_VER}"
CC_BUILD_TARGETS   = static-server
CC_BUILD_ARCHES    = darwin/arm64 freebsd/amd64 linux/amd64 linux/arm64 windows/amd64
CC_OUTPUT_TPL     := ${BUILDDIR}/bin/{{.Dir}}.{{.OS}}-{{.Arch}}

define HELP_OUTPUT
Available targets:
  help                this help
  clean               clean up
  all                 build binaries and man pages
  test                run tests
  cover               run tests with cover output
  build-setup         fetch dependencies
  build               build all binaries
  man                 build all man pages
  tar                 build release tarball
  cross-tar           cross compile and build release tarballs
endef
export HELP_OUTPUT

.PHONY: help clean build test cover man man-copy all tar cross-tar

help:
	@echo "$$HELP_OUTPUT"

clean:
	@rm -rf "${BUILDDIR}"

.PHONY: setup
setup:

build-setup:
	@[ -d "${BUILDDIR}/bin" ] || mkdir -p "${BUILDDIR}/bin"
	@[ -d "${BUILDDIR}/man" ] || mkdir -p "${BUILDDIR}/man"

build: build-setup
	@echo "Building..."
	@(for x in ${CC_BUILD_TARGETS}; do \
		printf -- "...%s...\n" "$${x}"; \
		go build ${GOBUILD_FLAGS} -o "${BUILDDIR}/bin/$${x}" ./cmd/$${x}; \
	done)

test:
	@echo "Running tests..."
	@go test ${GOTEST_FLAGS} ./...

generate:
	@echo "Running generate..."
	@go generate ./...

cover:
	@echo "Running tests with coverage..."
	@go test -cover ${GOTEST_FLAGS} ./...

${BUILDDIR}/man/%: man/%.mdoc
	@cat $< | sed -E "s#.Os (.*) VERSION#.Os \1 ${APP_VER}#" > $@

man: build-setup $(patsubst man/%.mdoc,${BUILDDIR}/man/%,$(wildcard man/*.1.mdoc))

tar: all
	@echo "Creating tar archive..."
	@mkdir -p ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/bin
	@mkdir -p ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/man
	@cp ${BUILDDIR}/bin/* ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/bin/
	@cp ${BUILDDIR}/man/*.[1-9] ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/man/
	@tar -C ${TARBUILDDIR} -czf ${TARBUILDDIR}/${APP_NAME}-${APP_VER}.${GOVER}.${OS}-${ARCH}.tar.gz ${APP_NAME}-${APP_VER}

cross-tar: man setup
	@echo "Building (cross-compile: ${CC_BUILD_ARCHES})..."
	@(for x in ${CC_BUILD_TARGETS}; do \
		for y in $(subst /,-,${CC_BUILD_ARCHES}); do \
			printf -- "--> %15s: %s\n" "$${y}" "$${x}"; \
			GOOS="$${y%%-*}"; \
			GOARCH="$${y##*-}"; \
			EXT=""; \
			if echo "$${y}" | grep -q 'windows-'; then EXT=".exe"; fi; \
			env GOOS=$${GOOS} GOARCH=$${GOARCH} go build ${GOBUILD_FLAGS} -o "${BUILDDIR}/bin/$${x}.$${GOOS}-$${GOARCH}$${EXT}" ./cmd/$${x}; \
		done; \
	done)

	@echo "Creating tar archives..."
	@(for x in $(subst /,-,${CC_BUILD_ARCHES}); do \
		printf -- "--> %15s\n" "$${x}"; \
		EXT=""; \
		if echo "$${x}" | grep -q 'windows-'; then EXT=".exe"; fi; \
		XDIR="${GOVER}.$${x}"; \
		ODIR="${TARBUILDDIR}/$${XDIR}/${APP_NAME}-${APP_VER}"; \
		mkdir -p "$${ODIR}/bin"; \
		mkdir -p "$${ODIR}/man"; \
		for t in ${CC_BUILD_TARGETS}; do \
			cp ${BUILDDIR}/bin/$${t}.$${x}$${EXT} $${ODIR}/bin/$${t}$${EXT}; \
		done; \
		cp ${BUILDDIR}/man/*.[1-9] $${ODIR}/man/; \
		tar -C ${TARBUILDDIR}/$${XDIR} -czf ${TARBUILDDIR}/${APP_NAME}-${APP_VER}.$${XDIR}.tar.gz ${APP_NAME}-${APP_VER}; \
		rm -rf "${TARBUILDDIR}/$${XDIR}/"; \
	done)
	@echo "done!"

all: build man
