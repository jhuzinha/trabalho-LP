.PHONY: run mock python compare test race fmt

run:
	go run ./cmd/scraper

mock:
	go run ./cmd/mockserver

python:
	python3 python/comparador_python.py --mock-interno

compare:
	./scripts/comparar_go_python.sh

test:
	go test ./...

race:
	go test -race ./...

fmt:
	gofmt -w .
