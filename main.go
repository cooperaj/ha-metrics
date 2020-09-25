package main

import (
	"sync"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"go.acpr.dev/ha-metrics/metrics"
)

var (
	wg         sync.WaitGroup
	collectors []collector
)

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	log.SetLevel(log.DebugLevel)

	reporter := NewReporter(cfg.HAEndpoint, cfg.HAToken, cfg.SensorPrefix)
	reporter.Run(&wg)

	collectors = append(collectors, metrics.NewCpu(reporter, cfg.CPUPollInterval))
	collectors = append(collectors, metrics.NewDisk("/", reporter, cfg.DiskPollInterval))

	for _, collector := range collectors {
		collector.Monitor(&wg)
	}

	log.Printf("Started %d collector/s", len(collectors))

	wg.Wait()
}
