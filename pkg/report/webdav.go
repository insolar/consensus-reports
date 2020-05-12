package report

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/studio-b12/gowebdav"

	"github.com/insolar/consensus-reports/pkg/middleware"
)


type WebdavClient struct {
	cfg middleware.WebDavConfig
	directory string
}

func CreateWebdavClient(cfg middleware.WebDavConfig, directory string) *WebdavClient {
	return &WebdavClient{cfg, directory}
}

func(w* WebdavClient) ReadReportData() (*ReportTemplateConfig, error) {
	client := gowebdav.NewClient(w.cfg.URL, w.cfg.User, w.cfg.Password)
	client.SetTimeout(w.cfg.Timeout)

	var reportCfg ConfigFileJson
	buf, err := client.Read(path.Join(w.directory, "/config.json"))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(buf, &reportCfg)
	if err != nil {
		return nil, err
	}

	files, err := client.ReadDir(w.directory)
	if err != nil {
		return nil, err
	}

	filesData := make(map[string]MetricFileJson, 0)
	for _, file := range files {
		if file.Name() == "config.json" || file.IsDir() {
			continue
		}

		buf, err = client.Read(path.Join(w.directory, file.Name()))
		if err != nil {
			return nil, err
		}

		var f MetricFileJson
		err = json.Unmarshal(buf, &f)
		if err != nil {
			return nil, err
		}
		filesData[file.Name()] = f
		fmt.Println(file.Name()) // network_size_10.json
	}

	// transform data to

	return nil, nil
}