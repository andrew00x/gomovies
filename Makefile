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

deps:                 ## Get all gomovies dependencies
	@go get $(DEPS)

test: clean deps      ## Run tests
	@go test -v $(addprefix ./, $(addsuffix /..., $(SRC_DIRS)))

build: clean deps     ## Build gomovies locally
	@go build -v -installsuffix "static" -o $(OUTPUT_DIR)/$(ARCH)/$(OUTPUT) $(addprefix ./, $(addsuffix /..., $(CMD_DIR)))

install: test deps    ## Install gomovies locally
	@go install -v -installsuffix "static" $(addprefix ./, $(addsuffix /..., $(CMD_DIR)))

build-rpi3: clean     ## Build gomovies for Rapberry PI 3 architecture
                      ## Note: docker is required
	@docker run \
		--rm \
		-e GOOS=linux \
		-e GOARCH=$(ARCH) \
		-e GOARM=7 \
		-v "$$(pwd):/go/src/$(PKG)" \
		-w "/go/src/$(PKG)" \
		golang:$(GO_VERSION) \
		/bin/bash -c "go get $(DEPS) && go build -v -installsuffix "static" -o $(OUTPUT_DIR)/$(ARCH)/$(OUTPUT) $(addprefix ./, $(addsuffix /..., $(CMD_DIR)))"

clean:                ## Remove build results
	@go clean
	@rm -rf $(OUTPUT_DIR)

install-rpi3-bin:     ## Install gomovies on Raspberry PI
                      ## It will be installed in directory /home/pi/bin/ on Raspberry
                      ## Note: Need to run build first, make build-rpi3
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "gomovies_bin=$$(pwd)/$(OUTPUT_DIR)/$(ARCH)/$(OUTPUT)" \
		--tags "binaries" \
		$(ANSIBLE_DIR)/install.yaml

install-rpi3-config:  ## Install gomovies configuration on Raspberry PI
                      ## Note: existed configuration will be overwritten
                      ## Update file init/ansible/roles/install.config/files/config.json before run this command
                      ## New configuration file will be coppied to directory /home/pi/.gomovies/ on Raspberry
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--tags "configuration" \
		$(ANSIBLE_DIR)/install.yaml

install-rpi3-systemd: ## Install systemd service to manage gomovies on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "gomovies_bin=$$(pwd)/$(OUTPUT_DIR)/$(ARCH)/$(OUTPUT)" \
		--tags "systemd" \
		$(ANSIBLE_DIR)/install.yaml

install-rpi3-webui:   ## Install web UI on Raspberry PI, WEB_UI_TAR must point to tar.gz that contains UI build,
                      ## see https://github.com/andrew00x/gomovies-react
                      ## EXAMPLE:
                      ## make install-rpi3-webui WEB_UI_TAR=~/src/js/gomovies-react/build.tar.gz
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		$(if $(WEB_UI_TAR), --extra-vars "web_ui_tar=$(WEB_UI_TAR)") \
		--tags "web_ui" \
		$(ANSIBLE_DIR)/install.yaml

install-rpi3-all:     ## Install all in one on Rapberry, see install-rpi3-webui about WEB_UI_TAR variable
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "gomovies_bin=$$(pwd)/$(OUTPUT_DIR)/$(ARCH)/$(OUTPUT) $(if $(WEB_UI_TAR), web_ui_tar=$(WEB_UI_TAR))" \
		$(ANSIBLE_DIR)/install.yaml

service-start:        ## Start gomovies service on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "service_state=started" \
		$(ANSIBLE_DIR)/service.yaml

service-stop:         ## Stop gomovies service on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "service_state=stopped" \
		$(ANSIBLE_DIR)/service.yaml

service-restart:      ## Restart gomovies service on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "service_state=restarted" \
		$(ANSIBLE_DIR)/service.yaml

help:                 ## Print this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/##//'