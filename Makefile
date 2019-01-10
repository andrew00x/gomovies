GO_VERSION=1.10.2
ARCH=$(shell go env GOARCH)

PKG=github.com/andrew00x/gomovies
DEPS=github.com/godbus/dbus \
	github.com/stretchr/testify \
	github.com/andrew00x/omxcontrol

CMD_DIR=cmd
SRC_DIRS=$(CMD_DIR) pkg
OUTPUT_DIR=bin
OUTPUT=gomovies

.DEFAULT_GOAL=build

clean:
	@go clean
	@rm -rf $(OUTPUT_DIR)

deps:
	@go get $(DEPS)

test: clean deps
	@go test -v $(addprefix ./, $(addsuffix /..., $(SRC_DIRS)))

build: clean deps
	@go build -v -installsuffix "static" -o $(OUTPUT_DIR)/$(ARCH)/$(OUTPUT) $(addprefix ./, $(addsuffix /..., $(CMD_DIR)))

install: test deps
	@go install -v -installsuffix "static" $(addprefix ./, $(addsuffix /..., $(CMD_DIR)))

build-rpi3: clean
	@docker run \
		--rm \
		-e GOOS=linux \
		-e GOARCH=arm \
		-e GOARM=7 \
		-v "$$(pwd):/go/src/$(PKG)" \
		-w "/go/src/$(PKG)" \
		golang:$(GO_VERSION) \
		/bin/bash -c "go get $(DEPS) && go build -v -installsuffix "static" -o $(OUTPUT_DIR)/arm/$(OUTPUT) $(addprefix ./, $(addsuffix /..., $(CMD_DIR)))"
