DASHBOARDS_DIR := ./dashboards

# 本物の Docker Compose を優先し、なければ podman-compose を使う
COMPOSE ?= $(shell \
	if command -v docker >/dev/null 2>&1 \
		&& ! docker --version 2>/dev/null | grep -qi podman \
		&& docker compose version >/dev/null 2>&1; then echo docker compose; \
	elif command -v podman-compose >/dev/null 2>&1; then echo podman-compose; \
	fi)

.PHONY: generate
generate:
	go run ./cmd/generate/ $(DASHBOARDS_DIR)/

.PHONY: build
build:
	go build ./...

.PHONY: check
check:
	@set -e; tmp=$$(mktemp -d); \
	trap 'rm -rf "$$tmp"' EXIT; \
	go run ./cmd/generate/ "$$tmp"; \
	diff -ruN "$(DASHBOARDS_DIR)" "$$tmp"

.PHONY: compose-check
compose-check:
	@test -n "$(COMPOSE)" || { \
		echo "No usable compose implementation found."; \
		echo "Run mise install (or ./setup-linux.sh on Linux) to install podman-compose."; \
		echo "Note: docker-compose v1 and the podman-docker shim are ignored on purpose."; \
		exit 1; \
	} >&2

.PHONY: dev
dev: compose-check generate
	$(COMPOSE) up -d
	@echo "Grafana: http://localhost:3000"

.PHONY: dev-stop
dev-stop: compose-check
	$(COMPOSE) down
