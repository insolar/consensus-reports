package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap/buffer"

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

		err := http.ListenAndServe(*serveAddress, nil)
		if err != nil {
			panic(err)
		}
	} else {
		buff := &buffer.Buffer{}
		err := report.MakeReport(client, buff)
		if err != nil {
			panic(err)
		}

		err = client.WriteReport(buff.Bytes()) // write to file
		if err != nil {
			panic(err)
		}
	}
}
