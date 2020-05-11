package report

import (
	"io"
	"io/ioutil"
	"text/template"

	"github.com/markbates/pkger"
)

type ChartConfig struct {
	Charts []string `json:"charts"`
	Quantiles []string `json:"quantiles"` // series
}

type RecordInfo struct {
	Chart string `json:"chart"`
	Unit string  `json:"unit"`
	Quantile string  `json:"quantile"`
	Description string `json:"description"`
	OriginalFormula string  `json:"original_formula"`
	Value  float64 `json:"value"`
}

type ResultData struct {
	Warnings    []string     `json:"warnings"`
	Records     []RecordInfo `json:"records"`
	NetworkSize uint         `json:"network_size"`
}


type ReportConfig struct {
	HtmlTitle string
	ChartConfig string
	// data
}

type Inventory struct {
	Material string
	Count    uint
}

type ReportDataReader interface {
	ReadReportData()
}

func MakeReport(cfg ReportConfig, wr io.Writer)  error {
	f, err := pkger.Open("/pkg/report/template.html")
	if err != nil {
		return err
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	// cfg := ReportConfig{"my report", ""}
	tmpl, err := template.New("report").Parse(string(buf))
	if err != nil {
		return err
	}
	err = tmpl.Execute(wr, cfg)
	if err != nil {
		return err
	}

	return nil
}

