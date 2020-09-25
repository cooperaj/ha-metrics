package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
)

type Cpu struct {
	metric       *metric
	reporter     reporter
	pollInterval *time.Duration
}

func NewCpu(reporter reporter, pollInterval *time.Duration) *Cpu {
	cpu := new(Cpu)

	metric := NewMetric()
	metric.Attributes["friendly_name"] = "CPU Usage"
	metric.Attributes["unit_of_measurement"] = "%"
	metric.Attributes["icon"] = "mdi:memory"

	cpu.metric = metric
	cpu.reporter = reporter
	cpu.pollInterval = pollInterval

	log.Infof("CPU metric collector with %s polling interval created", pollInterval.String())

	return cpu
}

func (c *Cpu) Monitor(wg *sync.WaitGroup) {
	info, err := cpu.Info()
	if err != nil {
		log.Fatalf("could not get cpuinfo: %s", err)
	}

	count, err := cpu.Counts(true)
	if err != nil {
		log.Fatalf("could not get cpu count: %s", err)
	}

	if len(info) > 0 {
		c.metric.Attributes["model"] = info[0].ModelName
		c.metric.Attributes["mhz"] = info[0].Mhz
		c.metric.Attributes["core_count"] = count
	}

	wg.Add(1)
	go func() {
		for {
			percent, err := cpu.Percent(0, false)
			if err != nil {
				log.Fatalf("could not get cpu usage: %s", err)
			}
			c.metric.State = percent[0]

			percents, err := cpu.Percent(0, true)
			if err != nil {
				log.Fatalf("could not get cpu usage: %s", err)
			}
			for index, stat := range percents {
				c.metric.Attributes[fmt.Sprintf("core_%d_usage", index)] = stat
			}

			pids, err := process.Pids()
			if err != nil {
				log.Fatalf("could not get pids: %s", err)
			}
			c.metric.Attributes["no_running_processes"] = len(pids)

			c.reporter.Report("cpu_usage", c.metric)

			time.Sleep(*c.pollInterval)
		}
	}()
}
