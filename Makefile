GO_VERSION=1.12.5
ARCH=$(shell go env GOARCH)

PKG=github.com/andrew00x/gomovies

CMD_DIR=cmd
SRC_DIRS=$(CMD_DIR) pkg
OUTPUT_DIR=bin
OUTPUT=gomovies

ANSIBLE_DIR=init/ansible

.DEFAULT_GOAL=build

build-rpi3: ARCH=arm
install-rpi3-%: ARCH=arm

test: clean           ## Run tests
	@go test -v $(addprefix ./, $(addsuffix /..., $(SRC_DIRS)))

build: clean          ## Build gomovies locally
	@go build -v -installsuffix "static" -o $(OUTPUT_DIR)/$(ARCH)/$(OUTPUT) $(addprefix ./, $(addsuffix /..., $(CMD_DIR)))

install: test         ## Install gomovies locally
	@go install -v -installsuffix "static" $(addprefix ./, $(addsuffix /..., $(CMD_DIR)))

build-rpi3: clean     ## Build gomovies for Raspberry PI 3 architecture
                      ## Note: docker is required
	@docker run \
		--rm \
		-e GOOS=linux \
		-e GOARCH=$(ARCH) \
		-e GOARM=7 \
		-e GO111MODULE=on \
		-v "$$(pwd):/go/src/$(PKG)" \
		-w "/go/src/$(PKG)" \
		golang:$(GO_VERSION) \
		/bin/bash -c "go build -v -installsuffix "static" -o $(OUTPUT_DIR)/$(ARCH)/$(OUTPUT) $(addprefix ./, $(addsuffix /..., $(CMD_DIR)))"

clean:                ## Remove build results
	@rm -rf $(OUTPUT_DIR)

##

install-rpi3-player:   ## Install video player on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--tags "player" \
		$(ANSIBLE_DIR)/install.gomovies.yaml

install-rpi3-bin:     ## Install gomovies on Raspberry PI
                      ## It will be installed in directory /home/pi/bin/ on Raspberry
                      ## Note: Need to run build first, make build-rpi3
	@if test ! -f $$(pwd)/$(OUTPUT_DIR)/$(ARCH)/$(OUTPUT); then echo "There is no binaries to install, run make build-rpi3 first"; exit 1; fi
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "gomovies_bin=$$(pwd)/$(OUTPUT_DIR)/$(ARCH)/$(OUTPUT)" \
		--tags "binaries" \
		$(ANSIBLE_DIR)/install.gomovies.yaml

install-rpi3-config:  ## Install gomovies configuration on Raspberry PI
                      ## Note: existed configuration will be overwritten
                      ## Update file init/ansible/roles/install.config/files/config.json before run this command
                      ## New configuration file will be coppied to directory GO_MOVIES_HOME on Raspberry
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "tmdb_api_key=$(if $(TMDB_API_KEY),$(TMDB_API_KEY),'')" \
		--tags "configuration" \
		$(ANSIBLE_DIR)/install.gomovies.yaml

install-rpi3-systemd: ## Install systemd service to manage gomovies on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--tags "systemd" \
		$(ANSIBLE_DIR)/install.gomovies.yaml

install-rpi3-all:     ## Install all in one on Raspberry
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "gomovies_bin=$$(pwd)/$(OUTPUT_DIR)/$(ARCH)/$(OUTPUT) "tmdb_api_key=$(if $(TMDB_API_KEY),$(TMDB_API_KEY),'')"" \
		$(ANSIBLE_DIR)/install.gomovies.yaml

##

install-rpi3-torrent:         ## Install torrent client on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--tags "torrent" \
		$(ANSIBLE_DIR)/install.torrent.yaml

install-rpi3-torrent-config:  ## Install torrent client configuration on Raspberry PI
                              ## Note: existed configuration will be overwritten
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "torrent_download_dir=$(if $(TORRENT_DOWNLOAD_DIR),$(TORRENT_DOWNLOAD_DIR),'')" \
		--tags "configuration" \
		$(ANSIBLE_DIR)/install.torrent.yaml

install-rpi3-torrent-systemd: ## Install systemd service to manage torrent client on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--tags "systemd" \
		$(ANSIBLE_DIR)/install.torrent.yaml

install-scripts: ## Install bash scripts to manage application
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		$(ANSIBLE_DIR)/install.scripts.yaml

install-rpi3-torrent-all:    ## Install all torrent client related parts in one on Raspberry
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "torrent_download_dir=$(if $(TORRENT_DOWNLOAD_DIR),$(TORRENT_DOWNLOAD_DIR),'')" \
		$(ANSIBLE_DIR)/install.torrent.yaml



service-start:        ## Start gomovies service on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "service_state=started" \
		$(ANSIBLE_DIR)/service.gomovies.yaml

service-stop:         ## Stop gomovies service on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "service_state=stopped" \
		$(ANSIBLE_DIR)/service.gomovies.yaml

service-restart:      ## Restart gomovies service on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "service_state=restarted" \
		$(ANSIBLE_DIR)/service.gomovies.yaml



torrent-start:        ## Start torrent service on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "service_state=started" \
		$(ANSIBLE_DIR)/service.torrent.yaml

torrent-stop:         ## Stop torrent service on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "service_state=stopped" \
		$(ANSIBLE_DIR)/service.torrent.yaml

torrent-restart:      ## Restart torrent service on Raspberry PI
	@ansible-playbook \
		-i $(ANSIBLE_DIR)/raspberry.ini \
		--extra-vars "service_state=restarted" \
		$(ANSIBLE_DIR)/service.torrent.yaml

##

help:                 ## Print this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/##//'
