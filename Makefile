run-db:
	docker-compose up

help:
	go run main.go --help

run:
	go run main.go -c config

generate:
	go generate ./...

test:
	go test -cover ./...