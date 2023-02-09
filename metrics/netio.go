package metrics

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/net"
	log "github.com/sirupsen/logrus"
)

var errInterfaceNotFound = errors.New("Interface not found on system")

type BitRate int

const (
	Bit  BitRate = 1
	Kbit         = Bit * 1000
	Mbit         = Kbit * 1000
	Gbit         = Mbit * 1000
)

func (b BitRate) String() string {
	switch b {
	case Bit:
		return "b"
	case Kbit:
		return "kb"
	case Mbit:
		return "Mb"
	case Gbit:
		return "Gb"
	}
	return "unknown"
}

type NetIO struct {
	iface      string
	rate       BitRate
	networkIn  *metric
	networkOut *metric

	networkInPrevious  uint64
	networkOutPrevious uint64

	reporter     reporter
	pollInterval *time.Duration
}

func NewNetIO(iface string, rate BitRate, reporter reporter, pollInterval *time.Duration) *NetIO {
	netIO := new(NetIO)

	networkIn := NewMetric()
	networkIn.Attributes["friendly_name"] = "Network IO (Bits Received)"
	networkIn.Attributes["device_class"] = "data_rate"
	networkIn.Attributes["unit_of_measurement"] = fmt.Sprintf("%s/s", rate)
	networkIn.Attributes["icon"] = "mdi:download-network"

	networkOut := NewMetric()
	networkOut.Attributes["friendly_name"] = "Network IO (Bits Sent)"
	networkOut.Attributes["device_class"] = "data_rate"
	networkOut.Attributes["unit_of_measurement"] = fmt.Sprintf("%s/s", rate)
	networkOut.Attributes["icon"] = "mdi:upload-network"

	netIO.iface = iface
	netIO.rate = rate
	netIO.networkIn = networkIn
	netIO.networkOut = networkOut
	netIO.reporter = reporter
	netIO.pollInterval = pollInterval

	log.Infof("NetIO metric collector for %s with %s polling interval created", iface, pollInterval.String())

	return netIO
}

func (n *NetIO) Monitor(wg *sync.WaitGroup) {
	sanitisedDeviceName := sanitiseDeviceName(n.iface)

	wg.Add(1)
	go func() {
		for {
			ifaces, err := net.IOCounters(true)
			if err != nil {
				log.Fatalf("could not get net io: %s", err)
			}

			iface, err := filterIfaces(ifaces, n.iface)
			if err != nil {
				log.Fatalf("failed to find interface on system: %s", err)
			}

			bitps := (iface.BytesRecv - n.networkInPrevious) / uint64(n.pollInterval.Seconds()) * 8
			n.networkIn.State = float64(bitps) / float64(n.rate)
			n.networkIn.Attributes["cumulative_bytes"] = iface.BytesRecv
			n.networkInPrevious = iface.BytesRecv

			bitps = (iface.BytesSent - n.networkOutPrevious) / uint64(n.pollInterval.Seconds()) * 8
			n.networkOut.State = float64(bitps) / float64(n.rate)
			n.networkOut.Attributes["cumulative_bytes"] = iface.BytesSent
			n.networkOutPrevious = iface.BytesSent

			n.reporter.Report(fmt.Sprintf("netio_in_%s", sanitisedDeviceName), n.networkIn)
			n.reporter.Report(fmt.Sprintf("netio_out_%s", sanitisedDeviceName), n.networkOut)

			time.Sleep(*n.pollInterval)
		}
	}()
}

func filterIfaces(ifaces []net.IOCountersStat, name string) (*net.IOCountersStat, error) {
	for _, iface := range ifaces {
		if iface.Name == name {
			return &iface, nil
		}
	}

	return nil, errInterfaceNotFound
}
