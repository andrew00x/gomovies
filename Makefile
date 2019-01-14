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

build-rpi3: ARCH=arm
install-rpi3-%: ARCH=arm

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
		-e GOARCH=$(ARCH) \
		-e GOARM=7 \
		-v "$$(pwd):/go/src/$(PKG)" \
		-w "/go/src/$(PKG)" \
		golang:$(GO_VERSION) \
		/bin/bash -c "go get $(DEPS) && go build -v -installsuffix "static" -o $(OUTPUT_DIR)/$(ARCH)/$(OUTPUT) $(addprefix ./, $(addsuffix /..., $(CMD_DIR)))"

install-rpi3-all:
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "gomovies_bin=$$(pwd)/$(OUTPUT_DIR)/$(ARCH)/$(OUTPUT) $(if $(WEB_UI_TAR), web_ui_tar=$(WEB_UI_TAR))" \
		$(ANSIBLE_DIR)/install.yaml

install-rpi3-bin:
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "gomovies_bin=$$(pwd)/$(OUTPUT_DIR)/$(ARCH)/$(OUTPUT)" \
		--tags "binaries" \
		$(ANSIBLE_DIR)/install.yaml

install-rpi3-config:
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--tags "configuration" \
		$(ANSIBLE_DIR)/install.yaml

install-rpi3-systemd:
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "gomovies_bin=$$(pwd)/$(OUTPUT_DIR)/$(ARCH)/$(OUTPUT)" \
		--tags "systemd" \
		$(ANSIBLE_DIR)/install.yaml

install-rpi3-webui:
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		$(if $(WEB_UI_TAR), --extra-vars "web_ui_tar=$(WEB_UI_TAR)") \
		--tags "web_ui" \
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
