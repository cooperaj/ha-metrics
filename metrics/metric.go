package metrics

type metric struct {
	State      interface{}            `json:"state"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

func NewMetric() *metric {
	metric := new(metric)
	metric.Attributes = make(map[string]interface{})

	return metric
}
