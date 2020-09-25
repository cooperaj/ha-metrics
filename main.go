package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/peterbourgon/ff/v3"
	log "github.com/sirupsen/logrus"
	"go.acpr.dev/ha-metrics/metrics"
)

var (
	wg         sync.WaitGroup
	collectors []collector
)

func main() {
	fs := flag.NewFlagSet("ha-metrics", flag.ExitOnError)
	var (
		endpoint = fs.String(
			"ha-endpoint",
			"",
			"REQUIRED Your home assistant api url (also via HA_ENDPOINT)",
		)
		token = fs.String(
			"ha-token",
			"",
			"REQUIRED The authorisation token you've created for your home assistant user account (also via HA_TOKEN)",
		)
		sensorPrefix = fs.String(
			"sensor-prefix",
			"ha_metrics_",
			"A prefix added to all sensor entities created (also via SENSOR_PREFIX)",
		)
		CPUPollInterval = fs.Duration(
			"cpu-poll-interval",
			20*time.Second,
			"The poll time for CPU metrics i.e. 20s, 5m, 1h (also via CPU_POLL_INTERVAL)",
		)
		DiskPollInterval = fs.Duration(
			"disk-poll-interval",
			30*time.Minute,
			"The poll time for Disk metrics i.e. 20s, 5m, 1h (also via DISK_POLL_INTERVAL)",
		)
		disks diskSlice
		debug = fs.Bool("debug", false, "Enable debug logging")
	)
	fs.Var(&disks, "disk", "A mountpoint to be reported as a disk, repeatable")
	fs.Usage = usage

	ff.Parse(fs, os.Args[1:], ff.WithEnvVarNoPrefix())

	if *endpoint == "" || *token == "" {
		fmt.Println("Missing -ha-endpoint or -ha-token configuration value")
		fmt.Println("")
		usage()
		os.Exit(1)
	}

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	reporter := NewReporter(*endpoint, *token, *sensorPrefix)
	reporter.Run(&wg)

	collectors = append(collectors, metrics.NewCpu(reporter, CPUPollInterval))

	for _, disk := range disks {
		collectors = append(collectors, metrics.NewDisk(disk, reporter, DiskPollInterval))
	}

	for _, collector := range collectors {
		collector.Monitor(&wg)
	}

	log.Printf("Started %d collector/s", len(collectors))

	wg.Wait()
}

func usage() {
	fmt.Printf("Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}
