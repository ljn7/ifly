SHELL := /bin/bash

VERSION := $(shell cat VERSION)
MODULE  := github.com/ljn7/ifly/cli
OWNER   := ljn7
REPO    := ifly
LDFLAGS := -s -w \
  -X main.version=$(VERSION) \
  -X $(MODULE)/cmd.cliVersion=$(VERSION) \
  -X $(MODULE)/cmd.repoOwner=$(OWNER) \
  -X $(MODULE)/cmd.repoName=$(REPO)

TARGETS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

.PHONY: embed build test lint release install clean

# Populate cli/plugin/ with the authoritative plugin tree so go:embed
# can pick it up. Portable across Linux/macOS/Git Bash (uses `cp -r`
# rather than `rsync`, which isn't installed on some Windows setups).
embed:
	@mkdir -p cli/plugin
	@rm -rf cli/plugin/.claude-plugin cli/plugin/hooks cli/plugin/skills cli/plugin/commands
	@rm -f  cli/plugin/defaults.yaml cli/plugin/.ifly.yaml.example cli/plugin/VERSION cli/plugin/LICENSE
	@cp -r .claude-plugin cli/plugin/
	@cp -r hooks          cli/plugin/
	@cp -r skills         cli/plugin/
	@cp -r commands       cli/plugin/
	@cp    defaults.yaml       cli/plugin/
	@cp    .ifly.yaml.example  cli/plugin/
	@cp    VERSION             cli/plugin/
	@cp    LICENSE             cli/plugin/

build: embed
	cd cli && CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ../dist/ifly .

test: embed
	cd cli && go test ./...
	bash tests/test_guard.sh
	bash tests/test_parse_defaults.sh
	bash tests/test_path_resolve.sh
	bash tests/test_split_command.sh
	bash tests/test_config.sh
	bash tests/test_session_start.sh

lint:
	shellcheck hooks/guard hooks/session-start hooks/lib/*.sh tests/*.sh tests/lib/*.sh
	cd cli && golangci-lint run ./...

release: embed
	@rm -rf dist && mkdir -p dist
	@for t in $(TARGETS); do \
	  os=$${t%-*}; arch=$${t#*-}; \
	  ext=""; \
	  if [[ $$os == windows ]]; then ext=".exe"; fi; \
	  out="dist/ifly-$$os-$$arch$$ext"; \
	  echo "building $$out"; \
	  cd cli && CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o "../$$out" . && cd ..; \
	  if command -v sha256sum >/dev/null; then \
	    sha256sum "$$out" > "$$out.sha256"; \
	  else \
	    shasum -a 256 "$$out" > "$$out.sha256"; \
	  fi; \
	done
	@tar czf dist/ifly-plugin-v$(VERSION).tar.gz \
	  .claude-plugin hooks skills commands defaults.yaml .ifly.yaml.example VERSION LICENSE README.md

install: build
	install -m 0755 dist/ifly $$(go env GOBIN 2>/dev/null || echo $$(go env GOPATH)/bin)/ifly

clean:
	rm -rf dist cli/plugin
