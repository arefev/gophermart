include .env

GOLANGCI_LINT_CACHE?=/tmp/gophermart-golangci-lint-cache
USER=CURRENT_UID=$$(id -u):0
DOCKER_PROJECT_NAME=gophermart
DATABASE_DSN="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_LOCAL_PORT}/${DB_NAME}?sslmode=disable"


gofmt:
	gofmt -s -w ./
.PHONY: gofmt


containers:
	$(USER) docker-compose --project-name $(DOCKER_PROJECT_NAME) up -d
.PHONY: containers


server: server-run
.PHONY: server


server-run: server-build
	./cmd/gophermart/server -d=${DATABASE_DSN} -a="${SERVER_ADDRESS}" -l="${LOG_LEVEL}" -s="${TOKEN_SECRET}"
.PHONY: server-run


server-build:
	go build -o ./cmd/gophermart/server ./cmd/gophermart/
.PHONY: server-build


migrate-up:
	migrate -path ./db/migrations -database ${DATABASE_DSN} up
.PHONY: migrate-up


migrate-down:
	migrate -path ./db/migrations -database ${DATABASE_DSN} down
.PHONY: migrate-down


migrate-create:
	migrate create -ext sql -dir ./db/migrations $(name)
.PHONY: migrate-create


golangci-lint-run: _golangci-lint-rm-unformatted-report
.PHONY: golangci-lint-run


_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint
.PHONY: _golangci-lint-reports-mkdir


_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.62.0 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json
.PHONY: _golangci-lint-run


_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json
.PHONY: _golangci-lint-format-report


_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json
.PHONY: _golangci-lint-rm-unformatted-report


golangci-lint-clean:
	sudo rm -rf ./golangci-lint 
.PHONY: golangci-lint-clean