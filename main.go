package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftoml"
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
		cpuPollInterval = fs.Duration(
			"cpu-poll-interval",
			20*time.Second,
			"The poll time for CPU metrics i.e. 20s, 5m, 1h (also via CPU_POLL_INTERVAL)",
		)
		cpuTempSensor = fs.Int(
			"cpu-temp-sensor",
			0,
			"The specific temperature sensor that you want reported. Possible index values reported on application start. (also via CPU_TEMP_SENSOR)",
		)
		systemLoadPollInterval = fs.Duration(
			"system-load-poll-interval",
			20*time.Second,
			"The poll time for system load i.e. 20s, 5m, 1h (also via SYSTEM_LOAD_POLL_INTERVAL)",
		)
		memoryPollInterval = fs.Duration(
			"memory-poll-interval",
			20*time.Second,
			"The poll time for system memory usage i.e. 20s, 5m, 1h (also via MEMORY_POLL_INTERVAL)",
		)
		diskPollInterval = fs.Duration(
			"disk-poll-interval",
			30*time.Minute,
			"The poll time for Disk metrics i.e. 20s, 5m, 1h (also via DISK_POLL_INTERVAL)",
		)
		disks             diskSlice
		networkIOInterval = fs.Duration(
			"netio-poll-interval",
			20*time.Second,
			"The poll time for network IO i.e. 20s, 5m, 1h (also via NETIO_POLL_INTERVAL)",
		)
		netIOInterfaces netIOIfaceSlice
		debug           = fs.Bool("debug", false, "Enable debug logging")
		_               = fs.String("config", "", "An optional toml formatted configuration file")
	)
	fs.Var(&disks, "disk", "A mountpoint to be reported as a disk, repeatable")
	fs.Var(&netIOInterfaces, "iface", "An network interface to monitor, repeatable")

	fs.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fs.PrintDefaults()
	}

	ff.Parse(fs, os.Args[1:],
		ff.WithEnvVars(),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(fftoml.Parser),
	)

	if *endpoint == "" || *token == "" {
		fmt.Println("Missing -ha-endpoint or -ha-token configuration value")
		fmt.Println("")
		fs.Usage()
		os.Exit(1)
	}

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	reporter := NewReporter(*endpoint, *token, *sensorPrefix)
	reporter.Run(&wg)

	collectors = append(collectors, metrics.NewCpu(reporter, cpuTempSensor, cpuPollInterval))
	collectors = append(collectors, metrics.NewLoad(reporter, systemLoadPollInterval))
	collectors = append(collectors, metrics.NewMemory(reporter, memoryPollInterval))

	for _, disk := range disks {
		collectors = append(collectors, metrics.NewDisk(disk, reporter, diskPollInterval))
	}

	for _, iface := range netIOInterfaces {
		collectors = append(collectors, metrics.NewNetIO(iface, metrics.Mbit, reporter, networkIOInterval))
	}

	for _, collector := range collectors {
		collector.Monitor(&wg)
	}

	log.Printf("Started %d collector/s", len(collectors))

	wg.Wait()
}
