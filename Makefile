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

ANSIBLE_DIR=init/ansible

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

install-rpi3:
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "gomovies_bin=$$(pwd)/$(OUTPUT_DIR)/arm/$(OUTPUT)" \
		$(if $(config), --extra-vars "install_config=true") \
		$(if $(systemd), --extra-vars "install_systemd=true") \
		$(ANSIBLE_DIR)/install.yaml

service-start:
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "service_state=started" \
		$(ANSIBLE_DIR)/service.yaml

service-stop:
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "service_state=stopped" \
		$(ANSIBLE_DIR)/service.yaml

service-restart:
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "service_state=restarted" \
		$(ANSIBLE_DIR)/service.yaml
