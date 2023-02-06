package metrics

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
	log "github.com/sirupsen/logrus"
)

type Disk struct {
	usageUnit    *metric
	mountPoint   string
	deviceName   string
	reporter     reporter
	pollInterval *time.Duration
}

func NewDisk(mountPoint string, reporter reporter, pollInterval *time.Duration) *Disk {
	const usageFriendlyName = "Disk Usage"
	const icon = "mdi:harddisk"

	disk := new(Disk)

	unit := NewMetric()
	unit.Attributes["friendly_name"] = usageFriendlyName
	unit.Attributes["unit_of_measurement"] = "GB"
	unit.Attributes["icon"] = icon

	disk.usageUnit = unit
	disk.mountPoint = mountPoint
	disk.reporter = reporter
	disk.pollInterval = pollInterval

	log.Infof("Disk metric collector for %s with %s polling interval created", mountPoint, pollInterval.String())

	return disk
}

func (d *Disk) Monitor(wg *sync.WaitGroup) {
	partitions, err := disk.Partitions(true)
	if err != nil {
		log.Fatalf("could not get disk partitions: %s", err)
	}

	mount, err := filterMountsByName(partitions, d.mountPoint)
	if err != nil {
		log.Fatalf("%s", err)
	}

	d.usageUnit.Attributes["mountpoint"] = mount.Mountpoint
	d.usageUnit.Attributes["filesystem"] = mount.Fstype
	d.deviceName = mount.Device

	sanitisedDeviceName := sanitiseDeviceName(mount.Device)

	wg.Add(1)
	go func() {
		for {
			stat, err := disk.Usage(d.mountPoint)
			if err != nil {
				log.Warnf("Unable to fetch disk usage for mount at %s: %s", d.mountPoint, err)
			}

			d.usageUnit.State = float64(stat.Used) / 1000000000.0                     //GB
			d.usageUnit.Attributes["total_size"] = float64(stat.Total) / 1000000000.0 //GB
			d.usageUnit.Attributes["used_percent"] = stat.UsedPercent

			d.reporter.Report(fmt.Sprintf("disk_usage_%s", sanitisedDeviceName), d.usageUnit)

			time.Sleep(*d.pollInterval)
		}
	}()
}

func filterMountsByName(mounts []disk.PartitionStat, mountPoint string) (*disk.PartitionStat, error) {
	for _, device := range mounts {
		if device.Mountpoint == mountPoint {
			return &device, nil
		}
	}

	return nil, fmt.Errorf("Unable to find device mounted at %s", mountPoint)
}

func sanitiseDeviceName(deviceName string) string {
	r := strings.NewReplacer("/", "_", "-", "_")
	return strings.TrimPrefix(r.Replace(deviceName), "_")
}
