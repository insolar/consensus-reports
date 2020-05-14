// Copyright 2020 Insolar Network Ltd.
// All rights reserved.
// This material is licensed under the Insolar License version 1.0,
// available at https://github.com/insolar/assured-ledger/blob/master/LICENSE.md.

package report

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"text/template"

	"github.com/markbates/pkger"
)

type XAxis struct {
	Name string `json:"name"`
	Data []int  `json:"data"`
}

type SeriesTemplate struct {
	Name string    `json:"name"`
	Data []float64 `json:"data"`
}

type ChartTemplate struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Series      []SeriesTemplate `json:"series"`
	YAxisName   string           `json:"yAxisName"`
}

// TemplateData passes to template
type TemplateData struct {
	GitBranch     string
	GitCommitHash string
	ChartConfig   []ChartTemplate
	xAxis         XAxis
}

type TemplateDataReader interface {
	ReadTemplateData() (*TemplateData, error)
}

func mustMarshall(v interface{}) string {
	buf, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

func MakeReport(reader TemplateDataReader, wr io.Writer) error {
	c, err := reader.ReadTemplateData()
	if err != nil {
		return err
	}

	templateData := struct {
		GitBranch     string
		GitCommitHash string
		ChartConfig   string
		XAxis         string
	}{
		GitBranch:     c.GitBranch,
		GitCommitHash: c.GitCommitHash,
		ChartConfig:   mustMarshall(c.ChartConfig),
		XAxis:         mustMarshall(c.xAxis),
	}

	f, err := pkger.Open("/pkg/report/template.html")
	if err != nil {
		return err
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	tmpl, err := template.New("report").Parse(string(buf))
	if err != nil {
		return err
	}
	err = tmpl.Execute(wr, templateData)
	if err != nil {
		return err
	}

	return nil
}
