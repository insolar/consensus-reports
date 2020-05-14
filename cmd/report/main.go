// Copyright 2020 Insolar Network Ltd.
// All rights reserved.
// This material is licensed under the Insolar License version 1.0,
// available at https://github.com/insolar/assured-ledger/blob/master/LICENSE.md.

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/insolar/insconfig"
	"go.uber.org/zap/buffer"

	"github.com/insolar/consensus-reports/pkg/report"
)

func main() {

	var serveAddress = flag.String("serve", "", "Serve html on address")
	cfg := report.Config{}
	params := insconfig.Params{
		EnvPrefix:       "report",
		FileNotRequired: true,
		ConfigPathGetter: &insconfig.FlagPathGetter{
			GoFlags: flag.CommandLine,
		},
	}
	insConfigurator := insconfig.New(params)
	err := insConfigurator.Load(&cfg)
	checkError(err)

	client := report.CreateWebdavClient(cfg)

	if serveAddress != nil && *serveAddress != "" {
		serveReport(*serveAddress, client)
	} else {
		saveReport(client)
	}
}

func saveReport(client *report.WebdavClient) {
	buff := &buffer.Buffer{}
	err := report.MakeReport(client, buff)
	checkError(err)

	err = client.WriteReport(buff.Bytes())
	checkError(err)
}

func serveReport(serveAddress string, dataReader report.TemplateDataReader) {
	log.Println("listen at http://" + serveAddress)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := report.MakeReport(dataReader, w)
		checkError(err)
	})

	err := http.ListenAndServe(serveAddress, nil)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
