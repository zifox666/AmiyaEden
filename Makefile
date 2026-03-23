.PHONY: dev

SHELL := /bin/bash

dev:
	@command -v air >/dev/null 2>&1 || { echo "air is not installed. Install it before running 'make dev'."; exit 1; }
	@command -v pnpm >/dev/null 2>&1 || { echo "pnpm is not installed. Install it before running 'make dev'."; exit 1; }
	@set -u; \
	trap 'status=$$?; kill $$frontend_pid $$backend_pid 2>/dev/null || true; wait $$frontend_pid $$backend_pid 2>/dev/null || true; exit $$status' INT TERM EXIT; \
	(cd static && pnpm dev) & frontend_pid=$$!; \
	air -c .air.toml & backend_pid=$$!; \
	wait -n $$frontend_pid $$backend_pid; \
	status=$$?; \
	kill $$frontend_pid $$backend_pid 2>/dev/null || true; \
	wait $$frontend_pid $$backend_pid 2>/dev/null || true; \
	exit $$status
