package main

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/pflag"

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

	cfg, err := middleware.NewConfig(*cfgPath)
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("failed to validate config: %v", err)
	}

	repl, err := metricreplicator.New(cfg.PrometheusHost, cfg.TmpDir)
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

	indexFilename := "config.json"
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
