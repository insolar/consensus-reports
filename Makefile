all: build test

build: install-deps report metricreplicator

test:
	go test -test.v ./...

lint:
	golangci-lint run

coverage:
	go test -coverprofile=coverage.tmp.out -covermode=count -count=1 -test.v ./...
	cat coverage.tmp.out | grep -v _mock.go > coverage.out
	go tool cover -html=coverage.out -o coverage.html

install-deps:
	go get github.com/markbates/pkger/cmd/pkger

report:
	pkger -o ./pkg/report
	go build -o bin/report cmd/report/main.go

metricreplicator:
	go build -o bin/metricreplicator cmd/metricreplicator/main.go
