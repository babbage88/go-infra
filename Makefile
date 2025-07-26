GHCR_REPO:=ghcr.io/babbage88/go-infra:
GHCR_REPO_TEST:=jtrahan88/goinfra-test:
GOINFRA_SRC_DIR:=$$HOME/projects/go-infra
GOOSEY_ENV_FILE:=.env
GOOSEY_PROJ_DIR:=../infra-db
ENV_FILE:=.env
BUILDER := infrabuilder
CUR_DUR := $(shell pwd)
mig:=$(shell date '+%m%d%Y.%H%M%S')
SHELL := /bin/bash
SPEC_JSON_SRC_FILE := spec/swagger.local-https.json
SPEC_YAML_SRC_FILE := spec/swagger.local-https.json
tag := $(shell cat version.yaml | yq -r .version)

check-swagger:
	which swagger || (GO111MODULE=off go get -u github.com/go-swagger/go-swagger/cmd/swagger)

swagger:
	swagger generate spec -o ./swagger.yaml --scan-models && swagger generate spec -o swagger.json --scan-models

dev-swagger: check-swagger
	swagger generate spec -o ./dev-swagger.yaml --scan-models && swagger generate spec -o dev-swagger.json --scan-models
	swagger mixin spec/swagger.dev.json dev-swagger.json --output swagger.json --format=json
	swagger mixin spec/swagger.dev.yaml dev-swagger.yaml --output swagger.yaml --format=yaml
	rm dev-swagger.json && rm dev-swagger.yaml

local-swagger: check-swagger
	swagger generate spec -o ./local-swagger.yaml --scan-models && swagger generate spec --scan-models -o local-swagger.json --scan-models
	swagger mixin spec/swagger.local.json local-swagger.json --output swagger.json --format=json
	swagger mixin spec/swagger.local.yaml local-swagger.yaml --output swagger.yaml --format=yaml
	rm local-swagger.json && rm local-swagger.yaml

local-swagger-https: check-swagger
	swagger generate spec -o ./local-swagger.yaml --scan-models && swagger generate spec --scan-models -o local-swagger.json --scan-models
	swagger mixin $(SPEC_JSON_SRC_FILE) local-swagger.json --output swagger.json --format=json
	swagger mixin $(SPEC_YAML_SRC_FILE) local-swagger.yaml --output swagger.yaml --format=yaml
	rm local-swagger.json && rm local-swagger.yaml

k3local-swagger: check-swagger
	swagger generate spec -o ./k3local-swagger.yaml --scan-models && swagger generate spec -o k3local-swagger.json --scan-models
	swagger mixin spec/swagger.localdev.json k3local-swagger.json --output swagger.json --format=json
	swagger mixin spec/swagger.localdev.yaml k3local-swagger.yaml --output swagger.yaml --format=yaml
	rm k3local-swagger.json && rm k3local-swagger.yaml

run-local: local-swagger
	go run . --local-development

run-local-with-https: local-swagger-https
	go run . --local-development --use-https

embed-swagger:
	swagger generate spec -o ./embed/swagger.yaml --scan-models && swagger generate spec > ./embed/swagger.json

serve-swagger: check-swagger
	swagger serve -F=swagger swagger.yaml --no-open --port 4443


check-builder:
	@if ! docker buildx inspect $(BUILDER) > /dev/null 2>&1; then \
		echo "Builder $(BUILDER) does not exist. Creating..."; \
		docker buildx create --name $(BUILDER) --bootstrap; \
	fi

create-builder: check-builder

buildandpush: dev-swagger
	docker buildx use infrabuilder
	docker buildx build --platform linux/amd64,linux/arm64 -t $(GHCR_REPO)$(tag) . --push


buildandpush-arm64: dev-swagger
	docker buildx use infrabuilder
	docker buildx build --platform linux/arm64 -t $(GHCR_REPO)$(tag) . --push

buildandpush-amd64: dev-swagger
	docker buildx use infrabuilder
	docker buildx build --platform linux/amd64 -t $(GHCR_REPO)$(tag) . --push

buildandpushlocalk3: k3local-swagger
	docker buildx use infrabuilder
	docker buildx build --platform linux/amd64,linux/arm64 -t $(GHCR_REPO_TEST)$(tag) . --push

deploydev: buildandpushdev
	kubectl apply -f deployment/kubernetes/go-infra.yaml
	kubectl rollout restart deployment go-infra

deploylocalk3: buildandpushlocalk3
	kubelocal apply -f deployment/kubernetes/localdev/go-infra-dev.yaml
	kubelocal rollout restart deployment go-infra

new-sqlmigration:
	@set -o allexport && source .env && set +o allexport  && \
		goose create -s $(mig) sql

apply-migration:
	@set -o allexport && source .env && set +o allexport && \
		goose up -v

build-goosey:
	@echo building goosey binary in $(GOOSEY_PROJ_DIR)
	@cd $(GOOSEY_PROJ_DIR) && set -o allexport && source .env && set +o allexport && \
		go build -v -o goosey . && cd $(CUR_DUR)
	@echo copying goosey binary from $(GOOSEY_PROJ_DIR) to current directory
	@cp $(GOOSEY_PROJ_DIR)/goosey ./goosey

fetch-tags:
	@{ \
	  branch=$$(git rev-parse --abbrev-ref HEAD); \
	  if [ "$$branch" != "$(MAIN_BRANCH)" ]; then \
	    echo "Error: You must be on the $(MAIN_BRANCH) branch. Current branch is '$$branch'."; \
	    exit 1; \
	  fi; \
	  git fetch origin $(MAIN_BRANCH); \
	  UPSTREAM=origin/$(MAIN_BRANCH); \
	  LOCAL=$$(git rev-parse @); \
	  REMOTE=$$(git rev-parse "$$UPSTREAM"); \
	  BASE=$$(git merge-base @ "$$UPSTREAM"); \
	  if [ "$$LOCAL" != "$$REMOTE" ]; then \
	    echo "Error: Your local $(MAIN_BRANCH) branch is not up-to-date with remote. Please pull the latest changes."; \
	    exit 1; \
	  fi; \
	  git fetch --tags; \
	}

release: fetch-tags
	@{ \
	  echo "Latest tag: $(LATEST_TAG)"; \
	  new_tag=$$(go run . utils version-bumper --latest-version "$(LATEST_TAG)" --increment-type=$(VERSION_TYPE)); \
	  echo "Creating new tag: $$new_tag"; \
	  git tag -a $$new_tag -m $$new_tag && git push --tags; \
	}
