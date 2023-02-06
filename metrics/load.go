package metrics

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/load"
	log "github.com/sirupsen/logrus"
)

type Load struct {
	one          *metric
	five         *metric
	fifteen      *metric
	reporter     reporter
	pollInterval *time.Duration
}

func NewLoad(reporter reporter, pollInterval *time.Duration) *Load {
	load := new(Load)

	oneMetric := NewMetric()
	oneMetric.Attributes["friendly_name"] = "1 minute load avg."
	oneMetric.Attributes["icon"] = "mdi:memory"
	load.one = oneMetric

	fiveMetric := NewMetric()
	fiveMetric.Attributes["friendly_name"] = "5 minute load avg."
	fiveMetric.Attributes["icon"] = "mdi:memory"
	load.five = fiveMetric

	fifteenMetric := NewMetric()
	fifteenMetric.Attributes["friendly_name"] = "15 minute load avg."
	fifteenMetric.Attributes["icon"] = "mdi:memory"
	load.fifteen = fifteenMetric

	load.reporter = reporter
	load.pollInterval = pollInterval

	log.Infof("System Load metric collector with %s polling interval created", pollInterval.String())

	return load
}

func (l *Load) Monitor(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		for {
			load, err := load.Avg()
			if err != nil {
				log.Fatalf("could not get cpu load: %s", err)
			}

			l.one.State = load.Load1
			l.five.State = load.Load5
			l.fifteen.State = load.Load15

			l.reporter.Report("system_load_1", l.one)
			l.reporter.Report("system_load_5", l.five)
			l.reporter.Report("system_load_15", l.fifteen)

			time.Sleep(*l.pollInterval)
		}
	}()
}
