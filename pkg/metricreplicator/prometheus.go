package metricreplicator

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"

	"github.com/insolar/consensus-reports/pkg/replicator"
)

type Replicator struct {
	Address             string
	TmpDir              string
	APIClient           v1.API
	ConsensusProperties []consensusProperty
}

func New(address, tmpDir string) (replicator.Replicator, error) {
	properties := []consensusProperty{
		sentTrafficPerNode, sentTrafficOverall,
		recvTrafficPerNode, recvTrafficOverall,
		sentConsensusPackets, recvConsensusPackets,
		phase01Duration, phase2Duration, phase3Duration,
	}

	repl := Replicator{
		Address:             address,
		TmpDir:              tmpDir,
		ConsensusProperties: properties,
	}

	client, err := api.NewClient(api.Config{Address: repl.Address})
	if err != nil {
		return Replicator{}, errors.Wrap(err, "failed to create prometheus client")
	}

	repl.APIClient = v1.NewAPI(client)
	return repl, nil
}
