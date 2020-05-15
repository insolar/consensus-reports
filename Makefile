all: report metricreplicator

report:
	pkger -o ./pkg/report
	go build -o bin/report cmd/report/main.go

metricreplicator:
	go build -o bin/metricreplicator cmd/metricreplicator/main.go