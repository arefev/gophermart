include .env

USER=CURRENT_UID=$$(id -u):0
DOCKER_PROJECT_NAME=gophermart
DATABASE_DSN="host=${DB_HOST} user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} sslmode=disable"

.PHONY: gofmt containers server server-run

gofmt:
	gofmt -s -w ./

containers:
	$(USER) docker-compose --project-name $(DOCKER_PROJECT_NAME) up -d

server: server-run

server-run: server-build
	./cmd/gophermart/server -d=${DATABASE_DSN} -k="${SECRET_KEY}"

server-build:
	go build -o ./cmd/gophermart/server ./cmd/gophermart/

.PHONY: golangci-lint-run
golangci-lint-run: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.57.2 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: golangci-lint-clean
golangci-lint-clean:
	sudo rm -rf ./golangci-lint 