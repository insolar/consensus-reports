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
	Name  string `yaml:"name" validate:"required"`
	Value string `yaml:"value" validate:"required"`
}

type RangeConfig struct {
	StartTime  int64            `yaml:"start_time" validate:"required"`
	Interval   time.Duration    `yaml:"interval" validate:"required"`
	Properties []PropertyConfig `yaml:"props" validate:"min=1,dive,required"`
}

type WebDavConfig struct {
	URL      string        `yaml:"url" validate:"required"`
	User     string        `yaml:"user" validate:"required"`
	Password string        `yaml:"password" validate:"required"`
	Timeout  time.Duration `yaml:"timeout" validate:"required"`
}

type GroupConfig struct {
	Description string           `yaml:"description" validate:"required"`
	Network     []PropertyConfig `yaml:"network" validate:"omitempty"`
	Ranges      []RangeConfig    `yaml:"ranges" validate:"min=1,dive,required"`
}

type Config struct {
	Quantiles      []string      `yaml:"quantiles" validate:"min=1,dive,required"`
	TmpDir         string        `yaml:"tmp_directory" validate:"required"`
	PrometheusHost string        `yaml:"prometheus_host" validate:"required"`
	Groups         []GroupConfig `yaml:"groups" validate:"min=1,dive,required"`
	WebDav         WebDavConfig  `yaml:"webdav" validate:"required"`
	Commit         string        `yaml:"commit_hash" validate:"required"`
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
		URL:           cfg.WebDav.URL,
		User:          cfg.WebDav.User,
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
