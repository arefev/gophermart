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
	./cmd/gophermart/gophermart \
		-d=${DATABASE_DSN} \
		-a="${SERVER_ADDRESS}:${SERVER_PORT}" \
		-l="${LOG_LEVEL}" \
		-s="${TOKEN_SECRET}" \
		-r="${ACCRUAL_HOST}:${ACCRUAL_PORT}"
.PHONY: server-run


server-build:
	go build -o ./cmd/gophermart/gophermart ./cmd/gophermart/
.PHONY: server-build


server-build-cover:
	go build -cover -o ./cmd/gophermart/gophermart ./cmd/gophermart/
.PHONY: server-build-cover


accrual:
	./cmd/accrual/accrual_linux_amd64 -a=${ACCRUAL_HOST}:${ACCRUAL_PORT}
.PHONY: accrual


migrate-up:
	migrate -path ./db/migrations -database ${DATABASE_DSN} up
.PHONY: migrate-up


migrate-down:
	migrate -path ./db/migrations -database ${DATABASE_DSN} down
.PHONY: migrate-down


migrate-create:
	migrate create -ext sql -dir ./db/migrations $(name)
.PHONY: migrate-create


test: server-build-cover
	go test ./... -cover -coverprofile=coverage.out && \
	go tool cover -html coverage.out -o test.html && \
	go tool cover -func=coverage.out
.PHONY: test


integration-test: test-clear
	gophermarttest \
		-test.v \
		-test.run=^TestGophermart$$ \
		-gophermart-binary-path="./cmd/gophermart/gophermart" \
		-gophermart-database-uri=${DATABASE_DSN} \
		-gophermart-host=${SERVER_ADDRESS} \
		-gophermart-port=${SERVER_PORT} \
		-accrual-binary-path="./cmd/accrual/accrual_linux_amd64" \
		-accrual-database-uri=${DATABASE_DSN} \
		-accrual-host=${ACCRUAL_HOST} \
		-accrual-port=${ACCRUAL_PORT} > ./test_report.txt
.PHONY: integration-test


test-clear: 
	rm -f coverage.out && rm -f test_report.txt && rm -f test.html
.PHONY: test-clear


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
    golangci/golangci-lint:v1.57.2 \
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