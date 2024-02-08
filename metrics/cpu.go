package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
)

type Cpu struct {
	usageMetric     *metric
	tempMetric      *metric
	reporter        reporter
	tempSensorIndex *int
	pollInterval    *time.Duration
}

func NewCpu(reporter reporter, tempSensorIndex *int, pollInterval *time.Duration) *Cpu {
	cpu := new(Cpu)

	usageMetric := NewMetric()
	usageMetric.Attributes["friendly_name"] = "CPU Usage"
	usageMetric.Attributes["unit_of_measurement"] = "%"
	usageMetric.Attributes["icon"] = "mdi:cpu-64-bit"
	cpu.usageMetric = usageMetric

	tempMetric := NewMetric()
	tempMetric.Attributes["friendly_name"] = "CPU Temperature"
	tempMetric.Attributes["device_class"] = "temperature"
	tempMetric.Attributes["unit_of_measurement"] = "°C"
	tempMetric.Attributes["icon"] = "mdi:thermometer"
	cpu.tempMetric = tempMetric

	cpu.reporter = reporter
	cpu.tempSensorIndex = tempSensorIndex
	cpu.pollInterval = pollInterval

	log.Infof("CPU metric collector with %s polling interval created", pollInterval.String())

	temps, _ := host.SensorsTemperatures()
	if len(temps) > 0 {
		log.Infof(
			"Found %d CPU temperature sensors, using index %d (%s)",
			len(temps),
			*tempSensorIndex,
			temps[*tempSensorIndex].SensorKey,
		)

		for index, temp := range temps {
			log.Infof("    %d: %s", index, temp.SensorKey)
		}
	}

	return cpu
}

func (c *Cpu) Monitor(wg *sync.WaitGroup) {
	info, err := cpu.Info()
	if err != nil {
		log.Fatalf("could not get cpuinfo: %s", err)
	}

	if len(info) > 0 {
		c.usageMetric.Attributes["mhz"] = info[0].Mhz
	}

	count, err := cpu.Counts(true)
	if err != nil {
		log.Fatalf("could not get cpu count: %s", err)
	}

	c.usageMetric.Attributes["core_count"] = count

	wg.Add(1)
	go func() {
		for {
			percent, err := cpu.Percent(0, false)
			if err != nil {
				log.Fatalf("could not get cpu usage: %s", err)
			}
			c.usageMetric.State = percent[0]

			percents, err := cpu.Percent(0, true)
			if err != nil {
				log.Fatalf("could not get cpu usage: %s", err)
			}
			for index, stat := range percents {
				c.usageMetric.Attributes[fmt.Sprintf("core_%d_usage", index)] = stat
			}

			pids, err := process.Pids()
			if err != nil {
				log.Fatalf("could not get pids: %s", err)
			}
			c.usageMetric.Attributes["running_process_count"] = len(pids)

			temps, err := host.SensorsTemperatures()
			if len(temps) > 0 {
				c.tempMetric.State = temps[*c.tempSensorIndex].Temperature
				c.reporter.Report("cpu_temperature", c.tempMetric)
			}

			c.reporter.Report("cpu_usage", c.usageMetric)

			time.Sleep(*c.pollInterval)
		}
	}()
}
