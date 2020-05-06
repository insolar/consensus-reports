package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/insolar/consensus-reports/pkg/prometheus"
	"github.com/insolar/consensus-reports/pkg/replicator"
)

func main() {
	cfgPath := pflag.String("cfg", "", "config for replicator")
	pflag.Parse()

	if *cfgPath == "" {
		log.Fatalln("empty path to cfg file")
	}

	cfg, err := prometheus.NewConfig(*cfgPath)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("failed to validate config: %v", err)
	}

	repl, err := prometheus.NewReplicator(cfg.PrometheusHost, cfg.TmpDir)
	if err != nil {
		log.Fatalf("failed to init replicator: %v", err)
	}

	if err := Run(repl, cfg); err != nil {
		log.Fatalf("failed to replicate metrics: %v", err)
	}

	fmt.Println("Done!")
}

func Run(repl replicator.Replicator, cfg prometheus.Config) error {
	cleanDir, err := ToTmpDir(cfg.TmpDir)
	defer cleanDir()
	if err != nil {
		return err
	}

	ctx := context.Background()

	files, charts, err := repl.GrabRecords(ctx, cfg.Quantiles, prometheus.RangesToReplicatorPeriods(cfg.Ranges))
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

func ToTmpDir(dirname string) (func(), error) {
	if err := os.Mkdir(dirname, 0777); err != nil {
		return func() {}, errors.Wrap(err, "failed to create tmp dir")
	}

	removeFunc := func() {
		if err := os.RemoveAll(dirname); err != nil {
			log.Printf("failed to remove tmp dir: %v", err)
		}
	}

	if err := os.Chdir(dirname); err != nil {
		return removeFunc, errors.Wrap(err, "cant change dir to tmp directory")
	}
	return removeFunc, nil
}
