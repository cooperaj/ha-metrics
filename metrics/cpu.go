package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/procfs"
	log "github.com/sirupsen/logrus"
)

type cpuStatTracker struct {
	prevIdle  float64
	prevTotal float64
}

type Cpu struct {
	metric       *metric
	reporter     reporter
	pollInterval int

	totalStat *cpuStatTracker
	coreStat  []*cpuStatTracker
}

func NewCpu(reporter reporter, pollInterval int) *Cpu {
	cpu := new(Cpu)

	metric := NewMetric()
	metric.Attributes["friendly_name"] = "CPU Usage"
	metric.Attributes["unit_of_measurement"] = "%"
	metric.Attributes["icon"] = "mdi:memory"

	cpu.metric = metric
	cpu.reporter = reporter
	cpu.pollInterval = pollInterval
	cpu.totalStat = &cpuStatTracker{
		prevIdle:  0.0,
		prevTotal: 0.0,
	}

	log.Infof("CPU metric collector with %ds polling interval created", pollInterval)

	return cpu
}

func (c *Cpu) Monitor(wg *sync.WaitGroup) {
	mount, err := procfs.NewDefaultFS()
	if err != nil {
		log.Fatalf("could not get proc mount: %s", err)
	}

	wg.Add(1)
	go func() {
		cpu, err := mount.Stat()
		if err != nil {
			log.Errorf("could not get cpu info: %s", err)
			wg.Done()
			return
		}

		info, _ := mount.CPUInfo()
		if err == nil && len(info) > 0 {
			c.metric.Attributes["model"] = info[1].ModelName
			c.metric.Attributes["speed"] = info[1].CPUMHz
		}

		if len(c.coreStat) != len(cpu.CPU) {
			c.coreStat = []*cpuStatTracker{}
			for i := 1; i <= len(cpu.CPU); i++ {
				c.coreStat = append(c.coreStat, &cpuStatTracker{
					prevIdle:  0.0,
					prevTotal: 0.0,
				})
			}
		}

		// polling loop
		for {
			cpu, _ = mount.Stat()

			c.metric.State = c.totalStat.calculateUsage(&cpu.CPUTotal)
			c.metric.Attributes["total_running_processes"] = cpu.ProcessesRunning

			for index, stat := range cpu.CPU {
				prevStat := c.coreStat[index]
				c.metric.Attributes[fmt.Sprintf("core_%d_usage", index)] = prevStat.calculateUsage(&stat)
			}

			c.reporter.Report("cpu_usage", c.metric)

			time.Sleep(time.Second * time.Duration(c.pollInterval))
		}
	}()
}

func (stat *cpuStatTracker) calculateUsage(update *procfs.CPUStat) float64 {
	var idle float64 = update.Idle + update.Iowait
	var nonIdle float64 = update.User + update.Nice + update.System + update.IRQ + update.SoftIRQ + update.Steal
	var total float64 = idle + nonIdle

	cpuUsage := ((total - stat.prevTotal) - (idle - stat.prevIdle)) / (total - stat.prevTotal) * 100

	stat.prevIdle = idle
	stat.prevTotal = total

	return cpuUsage
}
