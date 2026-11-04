.DEFAULT_GOAL := build

# Keep the GoFrame CLI aligned with the framework version declared in go.mod.
.PHONY: up
up: cli.install
	@"$(GF_BIN)" up -a

# Build binary using configuration from hack/config.yaml.
.PHONY: build
build: cli.install
	@"$(GF_BIN)" build -ew

# Parse api and generate controller/sdk.
.PHONY: ctrl
ctrl: cli.install
	@"$(GF_BIN)" gen ctrl

# Generate Go files for DAO/DO/Entity.
.PHONY: dao
dao: cli.install
	@test -n "$$INDU_SCOPE_DATABASE_DSN" || (echo "INDU_SCOPE_DATABASE_DSN is required" >&2; exit 1)
	@"$(GF_BIN)" gen dao -l "$$INDU_SCOPE_DATABASE_DSN" -p internal -c -s -gt -a -x schema_migrations

# Parse current project go files and generate enums go file.
.PHONY: enums
enums: cli.install
	@"$(GF_BIN)" gen enums

# Generate Go files for Service.
.PHONY: service
service: cli.install
	@"$(GF_BIN)" gen service


# Build docker image.
.PHONY: image
image: cli.install
	$(eval _TAG  = $(shell git rev-parse --short HEAD))
ifneq (, $(shell git status --porcelain 2>/dev/null))
	$(eval _TAG  = $(_TAG).dirty)
endif
	$(eval _TAG  = $(if ${TAG},  ${TAG}, $(_TAG)))
	$(eval _PUSH = $(if ${PUSH}, ${PUSH}, ))
	@"$(GF_BIN)" docker ${_PUSH} -tn $(DOCKER_NAME):${_TAG};


# Build docker image and automatically push to docker repo.
.PHONY: image.push
image.push: cli.install
	@$(MAKE) image PUSH=-p;


# Deploy image and yaml to current kubectl environment.
.PHONY: deploy
deploy: cli.install
	$(eval _TAG = $(if ${TAG},  ${TAG}, develop))

	@set -e; \
	mkdir -p $(ROOT_DIR)/temp/kustomize;\
	cd $(ROOT_DIR)/manifest/deploy/kustomize/overlays/${_ENV};\
	kustomize build > $(ROOT_DIR)/temp/kustomize.yaml;\
	kubectl   apply -f $(ROOT_DIR)/temp/kustomize.yaml; \
	if [ $(DEPLOY_NAME) != "" ]; then \
		kubectl patch -n $(NAMESPACE) deployment/$(DEPLOY_NAME) -p "{\"spec\":{\"template\":{\"metadata\":{\"labels\":{\"date\":\"$(shell date +%s)\"}}}}}"; \
	fi;


# Parsing protobuf files and generating go files.
.PHONY: pb
pb: cli.install
	@"$(GF_BIN)" gen pb

# Generate protobuf files for database tables.
.PHONY: pbentity
pbentity: cli.install
	@"$(GF_BIN)" gen pbentity
