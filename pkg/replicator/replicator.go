package replicator

import (
	"context"
	"time"
)

type Replicator interface {
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
