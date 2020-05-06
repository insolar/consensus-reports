package prometheus

type ConsensusProperty struct {
	Name        string
	Formula     string
	Description string
	Unit        string
	Quantile    bool
}

var (
	sentTrafficPerNode = ConsensusProperty{
		Name:        "sent_traffic_per_node",
		Formula:     "quantile(%s, sum(rate(insolar_consensus_packets_sent_bytes[%s])) by (instance))",
		Description: "Sent bytes by node per second",
		Unit:        "bytes/sec",
		Quantile:    true,
	}
	recvTrafficPerNode = ConsensusProperty{
		Name:        "recv_traffic_per_node",
		Formula:     "quantile(%s, sum(rate(insolar_consensus_packets_recv_bytes[%s])) by (instance))",
		Description: "Received bytes by node per second",
		Unit:        "bytes/sec",
		Quantile:    true,
	}
	sentTrafficOverall = ConsensusProperty{
		Name:        "sent_traffic",
		Formula:     "sum(rate(insolar_consensus_packets_sent_bytes[%s]))",
		Description: "Overall network sent bytes per second",
		Unit:        "bytes/sec",
		Quantile:    false,
	}
	recvTrafficOverall = ConsensusProperty{
		Name:        "recv_traffic",
		Formula:     "sum(rate(insolar_consensus_packets_recv_bytes[%s]))",
		Description: "Overall network received bytes per second",
		Unit:        "bytes/sec",
		Quantile:    false,
	}
	sentConsensusPackets = ConsensusProperty{
		Name:        "sent_consensus_packets",
		Formula:     "quantile(%s, sum(rate(insolar_consensus_packets_sent_count[%s])) by (instance))",
		Description: "Sent consensus packets by node per second",
		Unit:        "packets/sec",
		Quantile:    true,
	}
	recvConsensusPackets = ConsensusProperty{
		Name:        "recv_consensus_packets",
		Formula:     "quantile(%s, sum(rate(insolar_consensus_packets_recv_count[%s])) by (instance))",
		Description: "Received consensus packets by node per second",
		Unit:        "packets/sec",
		Quantile:    true,
	}
	phase01Duration = ConsensusProperty{
		Name:        "phase01_duration",
		Formula:     "histogram_quantile(%s, sum(rate(insolar_phase01_latency_bucket[%s])) by (le))",
		Description: "Duration of consensus phase01",
		Unit:        "ms",
		Quantile:    true,
	}
	phase2Duration = ConsensusProperty{
		Name:        "phase2_duration",
		Formula:     "histogram_quantile(%s, sum(rate(insolar_phase2_latency_bucket[%s])) by (le))",
		Description: "Duration of consensus phase2",
		Unit:        "ms",
		Quantile:    true,
	}
	phase3Duration = ConsensusProperty{
		Name:        "phase3_duration",
		Formula:     "histogram_quantile(%s, sum(rate(insolar_phase3_latency_bucket[%s])) by (le))",
		Description: "Duration of consensus phase3",
		Unit:        "ms",
		Quantile:    true,
	}

	properties = []ConsensusProperty{
		sentTrafficPerNode, sentTrafficOverall,
		recvTrafficPerNode, recvTrafficOverall,
		sentConsensusPackets, recvConsensusPackets,
		phase01Duration, phase2Duration, phase3Duration,
	}
)
