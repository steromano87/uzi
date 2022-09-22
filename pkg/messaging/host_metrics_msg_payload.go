package messaging

const HostMetricsMsgId = "HOST_METRICS"

type HostMetricsPayload struct {
	CPU    float64
	Memory struct {
		Total uint64
		Used  uint64
	}
	Storage struct {
		Total uint64
		Used  uint64
	}
	Network struct {
		UpSpeed   float64
		DownSpeed float64
	}
}

func (h *HostMetricsPayload) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}
