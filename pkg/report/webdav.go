package report

import (
	"encoding/json"
	"path"

	"github.com/studio-b12/gowebdav"

	"github.com/insolar/consensus-reports/pkg/middleware"
)


type WebdavClient struct {
	cfg middleware.WebDavConfig
	directory string
	client *gowebdav.Client
}

func CreateWebdavClient(cfg middleware.WebDavConfig, directory string) *WebdavClient {
	client := gowebdav.NewClient(cfg.URL, cfg.User, cfg.Password)
	client.SetTimeout(cfg.Timeout)

	return &WebdavClient{cfg, directory, client}
}

func(w* WebdavClient) ReadReportData() (*ReportTemplateConfig, error) {


	var reportCfg ConfigFileJson
	buf, err := w.client.Read(path.Join(w.directory, "/config.json"))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(buf, &reportCfg)
	if err != nil {
		return nil, err
	}

	files, err := w.client.ReadDir(w.directory)
	if err != nil {
		return nil, err
	}

	filesData := make(map[string]MetricFileJson, 0)
	for _, file := range files {
		if file.Name() == "config.json" || file.IsDir() {
			continue
		}

		buf, err = w.client.Read(path.Join(w.directory, file.Name()))
		if err != nil {
			return nil, err
		}

		var f MetricFileJson
		err = json.Unmarshal(buf, &f)
		if err != nil {
			return nil, err
		}
		filesData[file.Name()] = f
		// fmt.Println(file.Name()) // network_size_10.json
	}

	// transform data to
	result := &ReportTemplateConfig{}
	result.HtmlTitle = "title doc"
	result.ChartConfig = reportCfg

	return result, nil
}

func(w* WebdavClient) WriteReport(data []byte) error {
	return w.client.Write(path.Join(w.directory, "index.html"), data, 644)
}