package metricreplicator

import (
	"context"
	"errors"
	"testing"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"

	"github.com/insolar/consensus-reports/pkg/replicator"
)

func TestReplicator_GrabRecords(t *testing.T) {
	repl := Replicator{
		ConsensusProperties: []consensusProperty{sentTrafficPerNode, phase2Duration, sentTrafficOverall},
		TmpDir:              testTmpDir,
	}
	repl.APIClient = APIMock{QueryMock: func(ctx context.Context, query string, ts time.Time) (value model.Value, warnings v1.Warnings, err error) {
		if time.Until(ts) > time.Minute*30 {
			return nil, nil, errors.New("fake query error")
		}

		result := []*model.Sample{
			{
				Metric: map[model.LabelName]model.LabelValue{"Name": "metric1"},
				Value:  model.SampleValue(10),
			},
		}
		return model.Vector(result), nil, nil
	}}

	ctx := context.Background()
	ranges := []replicator.PeriodInfo{
		{
			Start:      time.Now(),
			End:        time.Now().Add(5 * time.Second),
			Interval:   5 * time.Second,
			Network:    []replicator.PeriodProperty{{Name: "latency", Value: "50ms"}},
			Properties: []replicator.PeriodProperty{{Name: "network_size", Value: "5"}},
		},
		{
			Start:       time.Now().Add(10 * time.Second),
			End:         time.Now().Add(10 * time.Second).Add(5 * time.Second),
			Interval:    5 * time.Second,
			Properties:  []replicator.PeriodProperty{{Name: "network_size", Value: "10"}},
			Description: "descr",
		},
	}

	clean, err := MakeTmpDir(repl.TmpDir)
	defer clean()
	require.NoError(t, err, "failed to create tmp dir")

	t.Run("positive", func(t *testing.T) {
		files, charts, err := repl.GrabRecords(ctx, []string{"0.8", "0.9"}, ranges)
		require.NoError(t, err)
		require.Equal(t, []string{"network_size_5.json", "network_size_10.json"}, files)
		require.Equal(t, []string{"sent_traffic_per_node", "phase2_duration", "sent_traffic"}, charts)
	})
	t.Run("query error", func(t *testing.T) {
		params := []replicator.PeriodInfo{
			{
				Start:      time.Now(),
				End:        time.Now().Add(time.Hour),
				Interval:   5 * time.Minute,
				Properties: []replicator.PeriodProperty{{Name: "network_size", Value: "5"}},
			},
		}
		_, _, err := repl.GrabRecords(ctx, []string{"0.8", "0.9"}, params)
		require.Error(t, err)
		require.Contains(t, err.Error(), "fake query error")
	})
}
