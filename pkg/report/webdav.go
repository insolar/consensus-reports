// Copyright 2020 Insolar Network Ltd.
// All rights reserved.
// This material is licensed under the Insolar License version 1.0,
// available at https://github.com/insolar/assured-ledger/blob/master/LICENSE.md.

package report

import (
	"encoding/json"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/studio-b12/gowebdav"

	"github.com/insolar/consensus-reports/pkg/metricreplicator"
	"github.com/insolar/consensus-reports/pkg/middleware"
	"github.com/insolar/consensus-reports/pkg/replicator"
)

const DefaultReportFileName = "index.html"
const NetworkSizePrefix = "network_size_"
const JsonFileExtension = ".json"

type MetricFileJson metricreplicator.ResultData

// ConfigFileJson read from config.json
type ConfigFileJson struct {
	ChartNames []string `json:"charts"`
	Quantiles  []string `json:"quantiles"` // series
}

type filesystem interface {
	ReadDir(path string) ([]os.FileInfo, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte, _ os.FileMode) error
}

type Config struct {
	Webdav middleware.WebDavConfig
	Git    struct {
		Branch string
		Hash   string
	}
}

type WebdavClient struct {
	cfg Config
	fs  filesystem
}

func CreateWebdavClient(cfg Config) *WebdavClient {
	client := gowebdav.NewClient(cfg.Webdav.Host, cfg.Webdav.Username, cfg.Webdav.Password)
	client.SetTimeout(cfg.Webdav.Timeout)

	return &WebdavClient{cfg, client}
}

type fileInfo struct {
	filename             string
	networkPropertyValue int
	networkPropertyUnit  string // just a thought for future properties like latency_50ms
}

func (w *WebdavClient) ReadTemplateData() (*TemplateData, error) {
	var reportCfg ConfigFileJson
	buf, err := w.fs.Read(path.Join(w.cfg.Webdav.Directory, "/", replicator.DefaultConfigFilename))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(buf, &reportCfg)
	if err != nil {
		return nil, err
	}

	files, err := w.fs.ReadDir(w.cfg.Webdav.Directory)
	if err != nil {
		return nil, err
	}

	parseNumber := func(filename string) int {
		trimmed := strings.TrimPrefix(filename, NetworkSizePrefix)
		numStr := strings.TrimSuffix(trimmed, JsonFileExtension)

		res, err := strconv.Atoi(numStr)
		if err != nil {
			panic(err)
		}
		return res
	}

	filenames := make([]fileInfo, 0)
	for _, file := range files {
		if strings.HasPrefix(file.Name(), NetworkSizePrefix) && strings.HasSuffix(file.Name(), JsonFileExtension) {
			filenames = append(filenames, fileInfo{file.Name(), parseNumber(file.Name()), "network_size"})
		}
	}

	sort.Slice(filenames, func(i, j int) bool {
		return filenames[i].networkPropertyValue < filenames[j].networkPropertyValue
	})

	xValues := make([]int, 0)
	for _, n := range filenames {
		xValues = append(xValues, n.networkPropertyValue)
	}

	filesData := make([]MetricFileJson, 0, len(filenames))
	for _, file := range filenames {
		buf, err = w.fs.Read(path.Join(w.cfg.Webdav.Directory, file.filename))
		if err != nil {
			return nil, err
		}

		var f MetricFileJson
		err = json.Unmarshal(buf, &f)
		if err != nil {
			return nil, err
		}
		filesData = append(filesData, f)
	}

	result := &TemplateData{}
	result.GitBranch = w.cfg.Git.Branch
	result.GitCommitHash = w.cfg.Git.Hash
	result.xAxis.Name = "Nodes count"
	result.xAxis.Data = append(result.xAxis.Data, xValues...)

	var ct ChartTemplate
	for i, v := range filesData[0].Records {
		var quantiles []string
		if v.Quantile == reportCfg.Quantiles[0] {
			quantiles = append(quantiles, reportCfg.Quantiles...)
		} else if v.Quantile == "" {
			quantiles = append(quantiles, "")
		} else {
			continue
		}

		ct = ChartTemplate{
			Name:        v.Chart,
			Description: v.Description,
			YAxisName:   v.Unit,
		}

		for j, q := range quantiles {
			serie1 := SeriesTemplate{
				Name: q,
				Data: make([]float64, 0),
			}
			for k := 0; k < len(xValues); k++ {
				// index [i+j] is used because records  for all quantiles go consistently
				serie1.Data = append(serie1.Data, filesData[k].Records[i+j].Value)
			}
			ct.Series = append(ct.Series, serie1)
		}

		result.ChartConfig = append(result.ChartConfig, ct)
	}

	return result, nil
}

func (w *WebdavClient) WriteReport(data []byte) error {
	return w.fs.Write(path.Join(w.cfg.Webdav.Directory, DefaultReportFileName), data, 0644)
}
