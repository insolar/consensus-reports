package metricreplicator

type consensusProperty struct {
	Name        string
	Formula     string
	Description string
	Unit        string
	Quantile    bool
}

var (
	sentTrafficPerNode = consensusProperty{
		Name:        "sent_traffic_per_node",
		Formula:     "quantile(%s, sum(rate(insolar_consensus_packets_sent_bytes[20s])) by (instance))",
		Description: "Sent consensus bytes by node per second",
		Unit:        "bytes/sec",
		Quantile:    true,
	}
	recvTrafficPerNode = consensusProperty{
		Name:        "recv_traffic_per_node",
		Formula:     "quantile(%s, sum(rate(insolar_consensus_packets_recv_bytes[20s])) by (instance))",
		Description: "Received consensus bytes by node per second",
		Unit:        "bytes/sec",
		Quantile:    true,
	}
	sentTrafficOverall = consensusProperty{
		Name:        "sent_traffic",
		Formula:     "sum(rate(insolar_consensus_packets_sent_bytes[20s]))",
		Description: "Overall network sent bytes per second",
		Unit:        "bytes/sec",
		Quantile:    false,
	}
	recvTrafficOverall = consensusProperty{
		Name:        "recv_traffic",
		Formula:     "sum(rate(insolar_consensus_packets_recv_bytes[20s]))",
		Description: "Overall network received bytes per second",
		Unit:        "bytes/sec",
		Quantile:    false,
	}
	sentConsensusPackets = consensusProperty{
		Name:        "sent_consensus_packets",
		Formula:     "quantile(%s, sum(rate(insolar_consensus_packets_sent_count[20s])) by (instance))",
		Description: "Sent consensus packets by node per second",
		Unit:        "packets/sec",
		Quantile:    true,
	}
	recvConsensusPackets = consensusProperty{
		Name:        "recv_consensus_packets",
		Formula:     "quantile(%s, sum(rate(insolar_consensus_packets_recv_count[20s])) by (instance))",
		Description: "Received consensus packets by node per second",
		Unit:        "packets/sec",
		Quantile:    true,
	}
	phase01Duration = consensusProperty{
		Name:        "phase01_duration",
		Formula:     "histogram_quantile(%s, sum(rate(insolar_phase01_latency_bucket[20s])) by (le))",
		Description: "Duration of consensus phase01",
		Unit:        "ms",
		Quantile:    true,
	}
	phase2Duration = consensusProperty{
		Name:        "phase2_duration",
		Formula:     "histogram_quantile(%s, sum(rate(insolar_phase2_latency_bucket[20s])) by (le))",
		Description: "Duration of consensus phase2",
		Unit:        "ms",
		Quantile:    true,
	}
	phase3Duration = consensusProperty{
		Name:        "phase3_duration",
		Formula:     "histogram_quantile(%s, sum(rate(insolar_phase3_latency_bucket[20s])) by (le))",
		Description: "Duration of consensus phase3",
		Unit:        "ms",
		Quantile:    true,
	}
)
