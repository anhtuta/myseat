.PHONY: be fe

be:
	cd backend && go run ./cmd/api

fe:
	cd frontend && pnpm dev
