package metricreplicator

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/insolar/consensus-reports/pkg/replicator"
)

const testTmpDir = "/tmp/test_replicator"

func TestReplicator_MakeConfigFile(t *testing.T) {
	repl := Replicator{TmpDir: testTmpDir}

	clean, err := MakeTmpDir(repl.TmpDir)
	defer clean()
	require.NoError(t, err, "failed to create tmp dir")

	cfg := replicator.OutputConfig{
		Charts:    []string{"sent_traffic_per_node", "phase2_duration", "sent_traffic"},
		Quantiles: []string{"0.8", "0.9"},
	}
	filename := "config.json"
	err = repl.MakeConfigFile(context.Background(), cfg, filename)
	require.NoError(t, err)

	data, err := ioutil.ReadFile(repl.TmpDir + "/" + filename)
	require.NoError(t, err)

	var fileInfo map[string][]string
	err = json.Unmarshal(data, &fileInfo)
	require.NoError(t, err)

	charts, ok := fileInfo["charts"]
	require.True(t, ok)
	require.Equal(t, []string{"sent_traffic_per_node", "phase2_duration", "sent_traffic"}, charts)

	quantiles, ok := fileInfo["quantiles"]
	require.True(t, ok)
	require.Equal(t, []string{"0.8", "0.9"}, quantiles)
}

func TestReplicator_UploadFiles(t *testing.T) {
	repl := Replicator{TmpDir: testTmpDir}

	clean, err := MakeTmpDir(repl.TmpDir)
	defer clean()
	require.NoError(t, err, "failed to create tmp dir")

	cfg := replicator.OutputConfig{
		Charts:    []string{"sent_traffic_per_node", "phase2_duration", "sent_traffic"},
		Quantiles: []string{"0.8", "0.9"},
	}
	filename := "config.json"
	err = repl.MakeConfigFile(context.Background(), cfg, filename)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	loaderCfg := replicator.LoaderConfig{
		URL:           ts.URL,
		RemoteDirName: "fake",
	}
	err = repl.UploadFiles(context.Background(), loaderCfg, []string{filename})
	require.NoError(t, err)

}
