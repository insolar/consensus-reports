package metricreplicator

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/studio-b12/gowebdav"

	"github.com/insolar/consensus-reports/pkg/replicator"
)

const (
	fileMode = 0644
)

func (repl Replicator) UploadFiles(ctx context.Context, cfg replicator.LoaderConfig, files []string) error {
	client := gowebdav.NewClient(cfg.URL, cfg.User, cfg.Password)
	client.SetTimeout(cfg.Timeout)

	if err := client.Mkdir(cfg.RemoteDirName, fileMode); err != nil {
		return errors.Wrap(err, "failed to create remote dir")
	}

	for _, f := range files {
		localFilePath := repl.TmpDir + "/" + f
		data, err := ioutil.ReadFile(localFilePath)
		if err != nil {
			return errors.Wrap(err, "failed to read local file")
		}

		remoteFilePath := cfg.RemoteDirName + "/" + f
		if err := client.Write(remoteFilePath, data, fileMode); err != nil {
			return errors.Wrap(err, "failed to write data to remote file")
		}
	}
	return nil
}

func (repl Replicator) saveDataToFile(data []byte, filename string) error {
	filePath := repl.TmpDir + "/" + filename

	if _, err := os.Stat(filePath); err == nil {
		return errors.Errorf("file already exists: %s", filePath)
	}

	recordFile, createErr := os.Create(filePath)
	if createErr != nil {
		return errors.Wrap(createErr, "failed to create file")
	}
	defer recordFile.Close()

	_, writeErr := recordFile.Write(data)
	if writeErr != nil {
		return errors.Wrap(writeErr, "failed to write file")
	}
	return nil
}

func (repl Replicator) MakeConfigFile(ctx context.Context, cfg replicator.OutputConfig, filename string) error {
	indexData, err := json.Marshal(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal output config")
	}

	if err := repl.saveDataToFile(indexData, filename); err != nil {
		return errors.Wrap(err, "failed to save config file")
	}
	return nil
}

func MakeTmpDir(dirname string) (func(), error) {
	if err := os.Mkdir(dirname, 0777); err != nil {
		return func() {}, errors.Wrap(err, "failed to create tmp dir")
	}

	removeFunc := func() {
		if err := os.RemoveAll(dirname); err != nil {
			log.Printf("failed to remove tmp dir: %v", err)
		}
	}

	return removeFunc, nil
}
