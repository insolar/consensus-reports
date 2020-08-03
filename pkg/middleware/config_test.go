package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/insolar/consensus-reports/pkg/replicator"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		cfg := Config{
			Quantiles:  []string{"0.5", "0.8"},
			TmpDir:     "/tmp",
			Prometheus: PrometheusConfig{Host: "localhost"},
			Groups: []GroupConfig{
				{
					Description: "descr",
					Network: []PropertyConfig{
						{Name: "latency", Value: "50ms"},
					},
					Ranges: []RangeConfig{
						{
							StartTime: 10,
							Interval:  time.Minute * 5,
							Properties: []PropertyConfig{
								{Name: "network_size", Value: "5"},
							},
						},
					},
				},
			},
			WebDav: WebDavConfig{
				Host:     "test",
				Username: "user",
				Password: "pwd",
				Timeout:  time.Minute,
			},
			Git: struct {
				Branch string
				Hash   string
			}{"master", "hash"},
		}
		err := cfg.Validate()
		require.NoError(t, err)
	})
	t.Run("without groups", func(t *testing.T) {
		cfg := Config{
			Quantiles:  []string{"0.5", "0.8"},
			TmpDir:     "/tmp",
			Prometheus: PrometheusConfig{Host: "localhost"},
			Groups:     []GroupConfig{},
			WebDav: WebDavConfig{
				Host:     "test",
				Username: "user",
				Password: "pwd",
			},
			Git: struct {
				Branch string
				Hash   string
			}{"master", "hash"},
		}
		err := cfg.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "validation for 'Groups' failed")
	})
	t.Run("without some fields", func(t *testing.T) {
		cfg := Config{
			Quantiles:  []string{"0.5", "0.8"},
			TmpDir:     "/tmp",
			Prometheus: PrometheusConfig{Host: ""},
			Groups: []GroupConfig{
				{
					Description: "descr",
					Ranges: []RangeConfig{
						{
							Interval: time.Minute * 5,
							Properties: []PropertyConfig{
								{Name: "network_size"},
							},
						},
					},
				},
			},
			WebDav: WebDavConfig{
				Host:     "test",
				Username: "user",
			},
			Git: struct {
				Branch string
				Hash   string
			}{"master", "hash"},
		}
		err := cfg.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "validation for 'Host' failed")
		require.Contains(t, err.Error(), "validation for 'StartTime' failed")
		require.Contains(t, err.Error(), "validation for 'Value' failed")
		require.Contains(t, err.Error(), "validation for 'Password' failed")
		require.Contains(t, err.Error(), "validation for 'Timeout' failed")
	})
	t.Run("empty fields", func(t *testing.T) {
		cfg := Config{}
		err := cfg.Validate()
		require.Error(t, err)
	})
}

func TestNewConfig(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cfgPath := "../../cmd/metricreplicator/config.yml"
		cfg, err := NewConfig(cfgPath)
		require.NoError(t, err)
		require.NotEmpty(t, cfg)
	})
	t.Run("negative", func(t *testing.T) {
		cfgPath := "fake_config.yml"
		_, err := NewConfig(cfgPath)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to load config")
	})
}

func TestGroupsToReplicatorPeriods(t *testing.T) {
	startTime := time.Now()
	groups := []GroupConfig{
		{
			Description: "network size grows with fixed latency 50ms",
			Network: []PropertyConfig{
				{Name: "latency", Value: "50ms"},
			},
			Ranges: []RangeConfig{
				{
					StartTime: startTime.Unix(),
					Interval:  time.Minute * 5,
					Properties: []PropertyConfig{
						{Name: "network_size", Value: "5"},
					},
				},
				{
					StartTime: startTime.Add(time.Minute * 10).Unix(),
					Interval:  time.Minute * 5,
					Properties: []PropertyConfig{
						{Name: "network_size", Value: "10"},
					},
				},
			},
		},
		{
			Description: "latency grows with fixed network size 10",
			Network: []PropertyConfig{
				{Name: "network_size", Value: "10"},
			},
			Ranges: []RangeConfig{
				{
					StartTime: startTime.Add(time.Minute * 20).Unix(),
					Interval:  time.Minute * 5,
					Properties: []PropertyConfig{
						{Name: "latency", Value: "50ms"},
					},
				},
				{
					StartTime: startTime.Add(time.Minute * 30).Unix(),
					Interval:  time.Minute * 3,
					Properties: []PropertyConfig{
						{Name: "latency", Value: "100ms"},
					},
				},
			},
		},
	}

	expectedStartTime := time.Unix(startTime.Unix(), 0)
	expectedPeriods := []replicator.PeriodInfo{
		{
			Start:    expectedStartTime,
			End:      expectedStartTime.Add(time.Minute * 5),
			Interval: time.Minute * 5,
			Properties: []replicator.PeriodProperty{
				{Name: "network_size", Value: "5"},
			},
			Network: []replicator.PeriodProperty{
				{Name: "latency", Value: "50ms"},
			},
			Description: "network size grows with fixed latency 50ms",
		},
		{
			Start:    expectedStartTime.Add(time.Minute * 10),
			End:      expectedStartTime.Add(time.Minute * 15),
			Interval: time.Minute * 5,
			Properties: []replicator.PeriodProperty{
				{Name: "network_size", Value: "10"},
			},
			Network: []replicator.PeriodProperty{
				{Name: "latency", Value: "50ms"},
			},
			Description: "network size grows with fixed latency 50ms",
		},
		{
			Start:    expectedStartTime.Add(time.Minute * 20),
			End:      expectedStartTime.Add(time.Minute * 25),
			Interval: time.Minute * 5,
			Properties: []replicator.PeriodProperty{
				{Name: "latency", Value: "50ms"},
			},
			Network: []replicator.PeriodProperty{
				{Name: "network_size", Value: "10"},
			},
			Description: "latency grows with fixed network size 10",
		},
		{
			Start:    expectedStartTime.Add(time.Minute * 30),
			End:      expectedStartTime.Add(time.Minute * 33),
			Interval: time.Minute * 3,
			Properties: []replicator.PeriodProperty{
				{Name: "latency", Value: "100ms"},
			},
			Network: []replicator.PeriodProperty{
				{Name: "network_size", Value: "10"},
			},
			Description: "latency grows with fixed network size 10",
		},
	}

	periods := GroupsToReplicatorPeriods(groups)
	require.Equal(t, expectedPeriods, periods)
}
