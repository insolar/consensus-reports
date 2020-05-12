package report

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"text/template"

	"github.com/markbates/pkger"

	"github.com/insolar/consensus-reports/pkg/metricreplicator"
)

type ConfigFileJson struct {
	ChartNames []string `json:"charts"`
	Quantiles []string `json:"quantiles"` // series
}

type MetricFileJson metricreplicator.ResultData

// reportTemplateConfig passes to template
type ReportTemplateConfig struct {
	HtmlTitle string
	ChartConfig ConfigFileJson
}

type ReportDataReader interface {
	ReadReportData() (*ReportTemplateConfig, error)
}

func MakeReport(reader ReportDataReader, wr io.Writer)  error {
	c, err := reader.ReadReportData()
	if err != nil {
		return err
	}

	chartConfigBuf, err := json.Marshal(c.ChartConfig)
	if err != nil {
		return err
	}

	templateData := struct {
		HtmlTitle string
		ChartConfig string
	} {
		HtmlTitle: c.HtmlTitle,
		ChartConfig: string(chartConfigBuf),
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

	// cfg := ReportConfig{"my report", ""}
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

