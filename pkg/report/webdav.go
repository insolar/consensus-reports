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

	"github.com/pkg/errors"
	"github.com/studio-b12/gowebdav"

	"github.com/insolar/consensus-reports/pkg/metricreplicator"
	"github.com/insolar/consensus-reports/pkg/middleware"
	"github.com/insolar/consensus-reports/pkg/replicator"
)

const DefaultReportFileName = "index.html"
const NetworkSizePrefix = "network_size_"
const JSONFileExtension = ".json"
const ReadTemplateDataErrorMessage = "Failed to read template data"

type MetricFileJSON metricreplicator.ResultData

// ConfigFileJSON read from config.json
type ConfigFileJSON struct {
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
	reportCfg, err := w.readConfigJSON()
	if err != nil {
		return nil, errors.Wrap(err, ReadTemplateDataErrorMessage)
	}

	filenames, err := w.scanWebdavFiles()
	if err != nil {
		return nil, errors.Wrap(err, ReadTemplateDataErrorMessage)
	}

	return w.collectTemplateData(filenames, reportCfg)
}

func (w *WebdavClient) readConfigJSON() (*ConfigFileJSON, error) {
	var reportCfg ConfigFileJSON
	buf, err := w.fs.Read(path.Join(w.cfg.Webdav.Directory, "/", replicator.DefaultConfigFilename))
	if err != nil {
		return nil, errors.Wrap(err, ReadTemplateDataErrorMessage)
	}

	err = json.Unmarshal(buf, &reportCfg)
	if err != nil {
		return nil, errors.Wrap(err, ReadTemplateDataErrorMessage)
	}

	return &reportCfg, nil
}

func (w *WebdavClient) scanWebdavFiles() ([]fileInfo, error) {
	files, err := w.fs.ReadDir(w.cfg.Webdav.Directory)
	if err != nil {
		return nil, errors.Wrap(err, ReadTemplateDataErrorMessage)
	}

	parseNumber := func(filename string) int {
		trimmed := strings.TrimPrefix(filename, NetworkSizePrefix)
		numStr := strings.TrimSuffix(trimmed, JSONFileExtension)

		res, err := strconv.Atoi(numStr)
		if err != nil {
			panic(err)
		}
		return res
	}

	filenames := make([]fileInfo, 0)
	for _, file := range files {
		if strings.HasPrefix(file.Name(), NetworkSizePrefix) && strings.HasSuffix(file.Name(), JSONFileExtension) {
			filenames = append(filenames, fileInfo{file.Name(), parseNumber(file.Name()), "network_size"})
		}
	}

	return filenames, nil
}

func (w *WebdavClient) collectTemplateData(filenames []fileInfo, reportCfg *ConfigFileJSON) (*TemplateData, error) {
	sort.Slice(filenames, func(i, j int) bool {
		return filenames[i].networkPropertyValue < filenames[j].networkPropertyValue
	})

	xValues := make([]int, 0)
	for _, n := range filenames {
		xValues = append(xValues, n.networkPropertyValue)
	}

	filesData := make([]MetricFileJSON, 0, len(filenames))
	for _, file := range filenames {
		buf, err := w.fs.Read(path.Join(w.cfg.Webdav.Directory, file.filename))
		if err != nil {
			return nil, errors.Wrap(err, ReadTemplateDataErrorMessage)
		}

		var f MetricFileJSON
		err = json.Unmarshal(buf, &f)
		if err != nil {
			return nil, errors.Wrap(err, ReadTemplateDataErrorMessage)
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
		switch {
		case v.Quantile == reportCfg.Quantiles[0]:
			quantiles = append(quantiles, reportCfg.Quantiles...)
		case v.Quantile == "":
			quantiles = append(quantiles, "")
		default:
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
