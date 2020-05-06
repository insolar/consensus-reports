package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
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
	Warnings          []string          `json:"warnings"`
	Records           []RecordInfo      `json:"records"`
	NetworkProperties []NetworkProperty `json:"network_properties"`
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

func (repl Replicator) queryVector(ctx context.Context, query string, ts time.Time) (float64, []string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	result, warnings, queryErr := repl.APIClient.Query(queryCtx, query, ts)
	if queryErr != nil {
		return 0, []string{}, errors.Wrap(queryErr, "failed to query prometheus")
	}

	records, ok := result.(model.Vector)
	if !ok {
		return 0, []string{}, errors.Errorf("failed to cast result type %T to %T", result, model.Vector{})
	}

	return float64(records[0].Value), warnings, nil
}

func (repl Replicator) grabRecord(ctx context.Context, query string, ts time.Time, property ConsensusProperty, quantile string) (RecordInfo, []string, error) {
	value, warnings, grabErr := repl.queryVector(ctx, query, ts)
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

func getFilename(period replicator.PeriodInfo) string {
	filename := ""

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

	interval := period.Interval.String()
	if strings.Contains(interval, "0s") {
		interval = strings.Replace(interval, "0s", "", 1)
	}

	for _, p := range repl.ConsensusProperties {
		if p.Quantile {
			for _, q := range quantiles {
				query := fmt.Sprintf(p.Formula, q, interval)

				record, warnings, err := repl.grabRecord(ctx, query, period.End, p, q)
				if err != nil {
					return "", errors.Wrap(err, "failed to grab record")
				}

				allWarns = append(allWarns, warnings...)
				records = append(records, record)
			}
			continue
		}
		query := fmt.Sprintf(p.Formula, interval)

		record, warnings, err := repl.grabRecord(ctx, query, period.End, p, "")
		if err != nil {
			return "", errors.Wrap(err, "failed to grab record")
		}

		allWarns = append(allWarns, warnings...)
		records = append(records, record)
	}

	filename := getFilename(period)

	result := ResultData{
		Warnings:          allWarns,
		Records:           records,
		NetworkProperties: toNetworkProperties(period.Properties),
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
