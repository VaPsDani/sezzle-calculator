BACKEND  := backend
FRONTEND := frontend
COVERAGE := docs/coverage

.DEFAULT_GOAL := help
.PHONY: help test-backend test-frontend coverage run-backend run-frontend docker-up

help:
	@echo "test-backend   run the Go test suite"
	@echo "test-frontend  run the Vitest suite"
	@echo "coverage       write both coverage reports into $(COVERAGE)"
	@echo "run-backend    start the API on :8080"
	@echo "run-frontend   start the Vite dev server on :5173"
	@echo "docker-up      build and start the whole stack"

test-backend:
	cd $(BACKEND) && go test ./...

test-frontend:
	cd $(FRONTEND) && npm test

# The frontend report lands in docs/coverage/frontend on its own: the
# reportsDirectory in vite.config.ts already points there.
coverage:
	mkdir -p $(COVERAGE)
	cd $(BACKEND) && go test ./... -covermode=atomic -coverprofile=coverage.out
	cd $(BACKEND) && go tool cover -func=coverage.out
	cd $(BACKEND) && go tool cover -html=coverage.out -o ../$(COVERAGE)/backend.html
	cd $(FRONTEND) && npm run test:coverage

run-backend:
	cd $(BACKEND) && go run ./cmd/server

run-frontend:
	cd $(FRONTEND) && npm run dev

docker-up:
	docker compose up --build
