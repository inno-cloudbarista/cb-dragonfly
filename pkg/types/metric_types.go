package types

// TODO: implements
type Metric struct{}

type Metrics struct {
	metrics []Metric
}

const (
	MONCONFIG = "config"
	COLLECTORGROUPTOPIC = "collectorGroupTopic"
	TOPIC = "topic"
)

const (
	NSID = "nsId"
	MCISID  = "mcisId"
	VMID = "vmId"
	OSTYPE = "osType"
	CSPTYPE = "cspType"
)
