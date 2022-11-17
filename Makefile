run-db:
	docker-compose up

help:
	go run main.go --help

generate:
	go generate ./...