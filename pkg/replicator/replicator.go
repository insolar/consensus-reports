package replicator

import (
	"context"
	"time"
)

// Replicator is a tool for getting metrics from one source and uploading them into another source.
type Replicator interface {
	// MakeConfigFile saves OutputConfig json data to file.
	MakeConfigFile(ctx context.Context, cfg OutputConfig, filename string) error
	GrabRecords(ctx context.Context, quantiles []string, periods []PeriodInfo) (files, charts []string, err error)
	GrabRecordsByPeriod(ctx context.Context, quantiles []string, period PeriodInfo) (string, error)
	UploadFiles(ctx context.Context, cfg LoaderConfig, files []string) error
}

type PeriodInfo struct {
	Start       time.Time
	End         time.Time
	Interval    time.Duration
	Properties  []PeriodProperty
	Network     []PeriodProperty
	Description string
}

type OutputConfig struct {
	Charts    []string `json:"charts"`
	Quantiles []string `json:"quantiles"`
}

type PeriodProperty struct {
	Name  string
	Value string
}

type LoaderConfig struct {
	URL           string
	User          string
	Password      string
	RemoteDirName string
	Timeout       time.Duration
}

const DefaultConfigFilename = "config.json"
