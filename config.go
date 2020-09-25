package main

// Config provides the application configuration parameters
type Config struct {
	SensorPrefix     string `envconfig:"SENSOR_PREFIX" default:"ha_metrics_"`
	HAEndpoint       string `envconfig:"HA_ENDPOINT" default:"localhost:8123"`
	HAToken          string `envconfig:"HA_TOKEN" required:"true"`
	CPUPollInterval  int    `envconfig:"CPU_POLLING_INTERVAL" default:"10"`
	DiskPollInterval int    `envconfig:"DISK_POLLING_INTERVAL" default:"1800"`
}
