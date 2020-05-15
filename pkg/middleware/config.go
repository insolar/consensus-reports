package middleware

import (
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"

	"github.com/insolar/consensus-reports/pkg/replicator"
)

type PropertyConfig struct {
	Name  string `mapstructure:"name" validate:"required"`
	Value string `mapstructure:"value" validate:"required"`
}

type RangeConfig struct {
	StartTime  int64            `mapstructure:"start_time" validate:"required"`
	Interval   time.Duration    `mapstructure:"interval" validate:"required"`
	Properties []PropertyConfig `mapstructure:"props" validate:"min=1,dive,required"`
}

type WebDavConfig struct {
	Host      string        `mapstructure:"host" validate:"required"`
	Username  string        `mapstructure:"username" validate:"required"`
	Password  string        `mapstructure:"password" validate:"required"`
	Timeout   time.Duration `mapstructure:"timeout" validate:"required"`
	Directory string        `directory:"host"`
}

type GroupConfig struct {
	Description string           `mapstructure:"description" validate:"required"`
	Network     []PropertyConfig `mapstructure:"network" validate:"omitempty"`
	Ranges      []RangeConfig    `mapstructure:"ranges" validate:"min=1,dive,required"`
}

type PrometheusConfig struct {
	Host string `mapstructure:"host" validate:"required"`
}

type Config struct {
	Quantiles  []string         `mapstructure:"quantiles" validate:"min=1,dive,required"`
	TmpDir     string           `mapstructure:"tmp_directory" validate:"required"`
	Prometheus PrometheusConfig `mapstructure:"prometheus" validate:"required"`
	Groups     []GroupConfig    `mapstructure:"groups" validate:"min=1,dive,required"`
	WebDav     WebDavConfig     `mapstructure:"webdav" validate:"required"`
	Commit     string           `mapstructure:"commit_hash" validate:"required"`
}

func NewConfig(cfgPath string) (Config, error) {
	rawData, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return Config{}, errors.Wrap(err, "failed to read cfg file")
	}

	var cfg Config
	if err := yaml.Unmarshal(rawData, &cfg); err != nil {
		return Config{}, errors.Wrap(err, "failed to unmarshal cfg file")
	}

	return cfg, nil
}

func (cfg *Config) Validate() error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return err
	}
	return nil
}

func (cfg Config) LoaderConfig() replicator.LoaderConfig {
	return replicator.LoaderConfig{
		URL:           cfg.WebDav.Host,
		User:          cfg.WebDav.Username,
		Password:      cfg.WebDav.Password,
		RemoteDirName: cfg.Commit,
		Timeout:       cfg.WebDav.Timeout,
	}
}

func GroupsToReplicatorPeriods(groups []GroupConfig) []replicator.PeriodInfo {
	props := make([]replicator.PeriodInfo, 0, len(groups))
	for _, g := range groups {
		for _, r := range g.Ranges {
			props = append(props, replicator.PeriodInfo{
				Start:       time.Unix(r.StartTime, 0),
				End:         time.Unix(r.StartTime, 0).Add(r.Interval),
				Interval:    r.Interval,
				Properties:  toPeriodProperties(r.Properties),
				Network:     toPeriodProperties(g.Network),
				Description: g.Description,
			})
		}
	}
	return props
}

func toPeriodProperties(props []PropertyConfig) []replicator.PeriodProperty {
	replProps := make([]replicator.PeriodProperty, 0, len(props))
	for _, p := range props {
		replProps = append(replProps, replicator.PeriodProperty{
			Name:  p.Name,
			Value: p.Value,
		})
	}
	return replProps
}
