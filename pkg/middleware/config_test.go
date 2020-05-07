package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		cfg := Config{
			Quantiles:      []string{"0.5", "0.8"},
			TmpDir:         "/tmp",
			PrometheusHost: "localhost",
			Ranges: []RangeConfig{
				{
					StartTime: 10,
					Interval:  time.Minute * 5,
					NetworkProperties: []NetworkPropertyConfig{
						{Name: "network_size", Value: "5"},
					},
				},
			},
			WebDav: WebDavConfig{
				URL:      "test",
				User:     "user",
				Password: "pwd",
				Timeout:  time.Minute,
			},
			Commit: "hash",
		}
		err := cfg.Validate()
		require.NoError(t, err)
	})
	t.Run("without ranges", func(t *testing.T) {
		cfg := Config{
			Quantiles:      []string{"0.5", "0.8"},
			TmpDir:         "/tmp",
			PrometheusHost: "localhost",
			Ranges:         []RangeConfig{},
			WebDav: WebDavConfig{
				URL:      "test",
				User:     "user",
				Password: "pwd",
			},
			Commit: "hash",
		}
		err := cfg.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "validation for 'Ranges' failed")
	})
	t.Run("without some fields", func(t *testing.T) {
		cfg := Config{
			Quantiles:      []string{"0.5", "0.8"},
			TmpDir:         "/tmp",
			PrometheusHost: "",
			Ranges: []RangeConfig{
				{
					Interval: time.Minute * 5,
					NetworkProperties: []NetworkPropertyConfig{
						{Name: "network_size"},
					},
				},
			},
			WebDav: WebDavConfig{
				URL:  "test",
				User: "user",
			},
			Commit: "hash",
		}
		err := cfg.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "validation for 'PrometheusHost' failed")
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
		require.Contains(t, err.Error(), "failed to read cfg file")
	})
}
