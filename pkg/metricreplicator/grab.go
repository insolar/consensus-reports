package metricreplicator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/insolar/consensus-reports/pkg/replicator"
)

type RecordInfo struct {
	Chart       string  `json:"chart"`
	Formula     string  `json:"original_formula"`
	Description string  `json:"description"`
	Unit        string  `json:"unit"`
	Value       float64 `json:"value"`
	Quantile    string  `json:"quantile,omitempty"`
}

type NetworkProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ResultData struct {
	Warnings    []string          `json:"warnings"`
	Records     []RecordInfo      `json:"records"`
	Network     []NetworkProperty `json:"network"`
	Properties  []NetworkProperty `json:"properties"`
	Description string            `json:"description"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
}

func toNetworkProperties(props []replicator.PeriodProperty) []NetworkProperty {
	networkProps := make([]NetworkProperty, 0, len(props))
	for _, p := range props {
		networkProps = append(networkProps, NetworkProperty{
			Name:  p.Name,
			Value: p.Value,
		})
	}
	return networkProps
}

func (repl Replicator) queryRangeMatrix(ctx context.Context, query string, startTime, endTime time.Time) (float64, []string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	queryRange := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  time.Second * 10, // pulse
	}

	result, warnings, queryErr := repl.APIClient.QueryRange(queryCtx, query, queryRange)
	if queryErr != nil {
		return 0, []string{}, errors.Wrap(queryErr, "failed to query prometheus")
	}

	records, ok := result.(model.Matrix)
	if !ok {
		return 0, []string{}, errors.Errorf("failed to cast result type %T to %T", result, model.Matrix{})
	}

	// Get maximum from all values in time period, because there are outbursts on prometheus graph
	var maxValue float64
	for _, r := range records {
		for _, v := range r.Values {
			if float64(v.Value) > maxValue {
				maxValue = float64(v.Value)
			}
		}
	}

	return maxValue, warnings, nil
}

func (repl Replicator) grabRecord(ctx context.Context, query string, startTime, endTime time.Time, property consensusProperty, quantile string) (RecordInfo, []string, error) {
	value, warnings, grabErr := repl.queryRangeMatrix(ctx, query, startTime, endTime)
	if grabErr != nil {
		return RecordInfo{}, []string{}, errors.Wrap(grabErr, fmt.Sprintf("failed to get result for query: `%s`", query))
	}

	record := RecordInfo{
		Chart:       property.Name,
		Formula:     query,
		Description: property.Description,
		Unit:        property.Unit,
		Value:       value,
		Quantile:    quantile,
	}
	return record, warnings, nil
}

// getFilename generates name from Period immutable and mutable properties.
func getFilename(period replicator.PeriodInfo) string {
	filename := ""

	for _, p := range period.Network {
		filename = strings.Join([]string{filename, p.Name, p.Value}, "_")
	}

	for _, p := range period.Properties {
		filename = strings.Join([]string{filename, p.Name, p.Value}, "_")
	}

	filename += ".json"

	if strings.HasPrefix(filename, "_") {
		filename = strings.Replace(filename, "_", "", 1)
	}

	return filename
}

func (repl Replicator) GrabRecordsByPeriod(ctx context.Context, quantiles []string, period replicator.PeriodInfo) (string, error) {
	var (
		allWarns = make([]string, 0)
		records  []RecordInfo
	)

	for _, p := range repl.ConsensusProperties {
		if p.Quantile {
			for _, q := range quantiles {
				query := fmt.Sprintf(p.Formula, q)

				record, warnings, err := repl.grabRecord(ctx, query, period.Start, period.End, p, q)
				if err != nil {
					return "", errors.Wrap(err, "failed to grab record")
				}

				allWarns = append(allWarns, warnings...)
				records = append(records, record)
			}
			continue
		}
		record, warnings, err := repl.grabRecord(ctx, p.Formula, period.Start, period.End, p, "")
		if err != nil {
			return "", errors.Wrap(err, "failed to grab record")
		}

		allWarns = append(allWarns, warnings...)
		records = append(records, record)
	}

	filename := getFilename(period)

	result := ResultData{
		Warnings:    allWarns,
		Records:     records,
		Properties:  toNetworkProperties(period.Properties),
		Network:     toNetworkProperties(period.Network),
		Description: period.Description,
		StartTime:   period.Start.UTC(),
		EndTime:     period.End.UTC(),
	}

	rawMsg, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return "", errors.Wrap(marshalErr, "failed to marshal result")
	}

	if err := repl.saveDataToFile(rawMsg, filename); err != nil {
		return "", errors.Wrap(err, "failed to save data to file")
	}
	return filename, nil
}

func (repl Replicator) GrabRecords(ctx context.Context, quantiles []string, periods []replicator.PeriodInfo) ([]string, []string, error) {
	var files []string
	for _, p := range periods {
		filename, err := repl.GrabRecordsByPeriod(ctx, quantiles, p)
		if err != nil {
			return []string{}, []string{}, errors.Wrap(err, "failed to grab records")
		}
		files = append(files, filename)
	}

	var charts []string
	for _, p := range repl.ConsensusProperties {
		charts = append(charts, p.Name)
	}
	return files, charts, nil
}
