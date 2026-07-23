# Monorepo root helpers. Component builds stay in server/, sdk-*/, samples-*/.

ROOT_DIR := $(abspath .)

.PHONY: help copyright copyright-check copyright-replace

help: ## Show targets
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-22s %s\n", $$1, $$2}'

copyright: ## Add missing license headers (per-directory templates)
	cd server && GOWORK=off go run ./cmd/tools/copyright -rootDir "$(ROOT_DIR)"

copyright-check: ## Verify license headers are present
	cd server && GOWORK=off go run ./cmd/tools/copyright -rootDir "$(ROOT_DIR)" -verifyOnly

copyright-replace: ## Replace existing headers with Super Durable per-directory templates (destructive)
	cd server && GOWORK=off go run ./cmd/tools/copyright -rootDir "$(ROOT_DIR)" -replace
