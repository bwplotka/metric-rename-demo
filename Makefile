GOMODS := $(abspath $(shell find . -name "go.mod" | grep -v .bingo | xargs dirname))

DOCKER_IMAGE=quay.io/bwplotka/my-app
DOCKER_TAG=latest
DOCKER_PUSH="no"

# --- deps ---
GOFUMPT = goimports
$(GOIMPORTS):
	@go install golang.org/x/tools/cmd/goimports@latest

GOFUMPT = gofumpt
$(GOFUMPT):
	@go install mvdan.cc/gofumpt@latest

BUF = buf
$(BUF):
	@go install github.com/bufbuild/buf/cmd/buf@v1.39.0

MDOX = mdox
$(MDOX):
	@go install github.com/bwplotka/mdox@latest

CARGO_HOME ?= ${HOME}/.cargo
WEAVER_VERSION = v0.13.2
#WEAVER = $(CARGO_HOME)/bin/weaver-$(WEAVER_VERSION)
# This demo uses an experimental build with --simple flag.
# See https://github.com/open-telemetry/weaver/compare/main...bwplotka:simplepoc?expand=1
WEAVER=../otel-weaver/target/debug/weaver
$(WEAVER):
	@echo "Installing $(WEAVER)"
	@curl --proto '=https' --tlsv1.2 -LsSf https://github.com/open-telemetry/weaver/releases/download/$(WEAVER_VERSION)/weaver-installer.sh | sh
	cp $(CARGO_HOME)/bin/weaver $(WEAVER)
	rm $(CARGO_HOME)/bin/weaver

# ------

.PHONY: help
help: ## Display this help and any documented user-facing targets. Other undocumented targets may be present in the Makefile.
help:
	@awk 'BEGIN {FS = ": ##"; printf "Usage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_\.\-\/%]+: ##/ { printf "  %-45s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: docker
docker:
	@export DOCKER_IMAGE=$(DOCKER_IMAGE) DOCKER_TAG=$(DOCKER_TAG) && bash scripts/build_docker.sh $(DOCKER_PUSH)

.PHONY: test
test: docker
	@for gomod in $(GOMODS); do \
		cd $$gomod && go test -v ./...; \
    done

GO_FILES = $(shell find . -path ./vendor -prune -o -name '*.go' -print)

.PHONY: format
format: $(GOFUMPT) $(GOIMPORTS) $(MDOX)
	@echo ">> formating imports"
	@$(GOIMPORTS) -w $(GO_FILES)
	@echo ">> gofumpt-ing the code; golangci-lint requires this"
	@$(GOFUMPT) -extra -w $(GO_FILES)
	@echo ">> format documentation"
	@$(MDOX) fmt --soft-wraps ./*.md

.PHONY: gen # Generate artefacts e.g. metric definitions from my-org semconv.
gen: $(WEAVER)
	@rm -rf ./my-org/my-app/semconv.gen
	@echo ">> weaver generate 1.0.0 artefacts (just to show what was in the past)"
	@$(WEAVER) registry generate \
		--simple --debug \
		--registry=./my-org/semconv/registry.1.0.0/ \
		--templates=./prometheus/weaver_templates/client_golang \
		--param="schema_url=https://bwplotka.dev/semconv/1.0.0" \
		--future \
		go \
		./my-org/my-app/semconv.gen/1.0.0
	@echo ">> weaver generate 1.1.0 (latest) artefacts"
	@$(WEAVER) registry generate \
		--simple --debug \
		--registry=./my-org/semconv/registry \
		--templates=./prometheus/weaver_templates/client_golang \
		--param="schema_url=https://bwplotka.dev/semconv/1.1.0" \
		--future \
		go \
		./my-org/my-app/semconv.gen/
	@echo ">> weaver generate 1.0.0 -> 1.1.0 diff"
	@$(WEAVER) registry diff \
		--simple --debug \
		--baseline-registry=./my-org/semconv/registry.1.0.0 \
		--registry=./my-org/semconv/registry \
		--diff-format=json \
		--output=./my-org/semconv
#	@echo ">> weaver generate $(SEMCONV_VERSION1) -> $(SEMCONV_VERSION2) relabelling rules for Prometheus"
#	@$(WEAVER) registry diff \
#		--simple --debug \
#		--baseline-registry=./my-org/semconv/$(SEMCONV_VERSION1) \
#		--registry=./my-org/semconv/$(SEMCONV_VERSION2) \
#	    --diff-template=./prometheus/weaver_templates/prometheus --diff-format=yaml \
#	    --output=./my-org/semconv/$(SEMCONV_VERSION2)/.gen
