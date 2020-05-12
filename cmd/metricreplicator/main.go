package main

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/insolar/consensus-reports/pkg/metricreplicator"
	"github.com/insolar/consensus-reports/pkg/middleware"
	"github.com/insolar/consensus-reports/pkg/replicator"
)

func main() {
	cfgPath := pflag.String("cfg", "", "Path to cfg file")
	removeAfter := pflag.Bool("rm", true, "Option to remove tmp dir after work")
	pflag.Parse()

	if *cfgPath == "" {
		log.Fatalln("empty path to cfg file")
	}

	// insConfig := insconfig.New(insconfig.Params{
	// 	EnvPrefix:        "reports_webdav",
	// 	ConfigPathGetter: &insconfig.PFlagPathGetter{PFlags: pflag.CommandLine},
	// })
	// var cfg middleware.Config
	// if err := insConfig.Load(&cfg); err != nil {
	// 	log.Fatalf("failed to load config: %v", err)
	// }

	vp := viper.New()
	vp.SetConfigFile(*cfgPath)

	vp.SetEnvPrefix("reports_webdav") // will be uppercased automatically
	if err := vp.BindEnv("host"); err != nil {
		log.Fatalf("failed to get webdav host: %v", err)
	}
	if err := vp.BindEnv("password"); err != nil {
		log.Fatalf("failed to get webdav password: %v", err)
	}
	if err := vp.BindEnv("username"); err != nil {
		log.Fatalf("failed to get webdav username: %v", err)
	}

	if err := vp.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	var cfg middleware.Config
	if err := vp.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	cfg.WebDav.Host = vp.GetString("host")
	cfg.WebDav.Username = vp.GetString("username")
	cfg.WebDav.Password = vp.GetString("password")

	if err := cfg.Validate(); err != nil {
		log.Fatalf("failed to validate config: %v\ncfg: %+v", err, cfg)
	}

	repl, err := metricreplicator.New(cfg.Prometheus.Host, cfg.TmpDir)
	if err != nil {
		log.Fatalf("failed to init replicator: %v", err)
	}

	if err := Run(repl, cfg, *removeAfter); err != nil {
		log.Fatalf("failed to replicate metrics: %v", err)
	}

	fmt.Println("Done!")
}

func Run(repl replicator.Replicator, cfg middleware.Config, removeAfter bool) error {
	cleanDir, err := metricreplicator.MakeTmpDir(cfg.TmpDir)
	if removeAfter {
		defer cleanDir()
	}
	if err != nil {
		return err
	}

	ctx := context.Background()

	files, charts, err := repl.GrabRecords(ctx, cfg.Quantiles, middleware.GroupsToReplicatorPeriods(cfg.Groups))
	if err != nil {
		return err
	}

	indexFilename := replicator.DefaultConfigFilename
	outputCfg := replicator.OutputConfig{
		Charts:    charts,
		Quantiles: cfg.Quantiles,
	}
	if err := repl.MakeConfigFile(ctx, outputCfg, indexFilename); err != nil {
		return err
	}

	files = append(files, indexFilename)

	loaderCfg := cfg.LoaderConfig()
	if err := repl.UploadFiles(ctx, loaderCfg, files); err != nil {
		return err
	}
	return nil
}
