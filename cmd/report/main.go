package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/pflag"

	"github.com/insolar/consensus-reports/pkg/middleware"
	"github.com/insolar/consensus-reports/pkg/report"
)

// type Config = middleware.WebDavConfig

func main() {
	cfgPath := pflag.String("cfg", "", "Path to cfg file")
	serveAddress := pflag.String("serve", "", "Serve html on address")

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

	client := report.CreateWebdavClient(cfg.WebDav, cfg.Commit)


	if serveAddress != nil && *serveAddress != "" {
		fmt.Println("listen at http://" + *serveAddress)
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			err := report.MakeReport(client, w)
			if err != nil {
				panic(err)
			}
		})

		err = http.ListenAndServe(*serveAddress, nil)
		if err != nil {
			panic(err)
		}
	} else {
		// write to file
	}
}