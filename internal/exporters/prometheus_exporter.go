package exporters

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	DeletedPVC *prometheus.CounterVec
}

func NewCollector() *Collector {
	return &Collector{
		DeletedPVC: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "pvc_deleted",
				Help: "Represents the number of successful PVC deletions.",
			},
			[]string{"dryrun"},
		),
	}
}

// Collect implements Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.DeletedPVC.Collect(ch)
}

// Describe implements Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.DeletedPVC.Describe(ch)
}
