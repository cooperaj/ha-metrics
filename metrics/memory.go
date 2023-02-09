package metrics

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
	log "github.com/sirupsen/logrus"
)

type Memory struct {
	memoryMetric *metric
	reporter     reporter
	pollInterval *time.Duration
}

func NewMemory(reporter reporter, pollInterval *time.Duration) *Memory {
	memory := new(Memory)

	memoryMetric := NewMetric()
	memoryMetric.Attributes["friendly_name"] = "Memory Usage"
	memoryMetric.Attributes["device_class"] = "data_size"
	memoryMetric.Attributes["unit_of_measurement"] = "MiB"
	memoryMetric.Attributes["icon"] = "mdi:memory"
	memory.memoryMetric = memoryMetric

	memory.reporter = reporter
	memory.pollInterval = pollInterval

	log.Infof("Memory metric collector with %s polling interval created", pollInterval.String())

	return memory
}

func (m *Memory) Monitor(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		for {
			memory, err := mem.VirtualMemory()
			if err != nil {
				log.Fatalf("could not get memory information: %s", err)
			}

			m.memoryMetric.State = float64(memory.Used) / 1048576.0                       //MiB
			m.memoryMetric.Attributes["total_amount"] = float64(memory.Total) / 1048576.0 //MiB
			m.memoryMetric.Attributes["used_percent"] = memory.UsedPercent

			m.reporter.Report("memory_usage", m.memoryMetric)

			time.Sleep(*m.pollInterval)
		}
	}()
}
