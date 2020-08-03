package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/insolar/insconfig"
	"log"

	"github.com/insolar/consensus-reports/pkg/metricreplicator"
	"github.com/insolar/consensus-reports/pkg/middleware"
	"github.com/insolar/consensus-reports/pkg/replicator"
)

func main() {

	removeAfter := flag.Bool("rm", true, "Option to remove tmp dir after work")
	cfg := middleware.Config{}
	params := insconfig.Params{
		EnvPrefix:       "metricreplicator",
		FileNotRequired: true,
		ConfigPathGetter: &insconfig.FlagPathGetter{
			GoFlags: flag.CommandLine,
		},
	}
	insConfigurator := insconfig.New(params)
	err := insConfigurator.Load(&cfg)
	checkError(err)

	err = insconfig.NewYamlDumper(cfg).DumpTo(log.Writer())
	checkError(err)

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

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}