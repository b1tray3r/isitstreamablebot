.PHONY: setup

setup:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/air-verse/air@latest
	go install github.com/rubenv/sql-migrate/...@latest