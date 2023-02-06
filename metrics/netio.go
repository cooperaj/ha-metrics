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

type NetIO struct {
	iface      string
	networkIn  *metric
	networkOut *metric

	reporter     reporter
	pollInterval *time.Duration
}

func NewNetIO(iface string, reporter reporter, pollInterval *time.Duration) *NetIO {
	netIO := new(NetIO)

	networkIn := NewMetric()
	networkIn.Attributes["friendly_name"] = "Network IO (Bytes Received)"
	networkIn.Attributes["unit_of_measurement"] = "Bytes"
	networkIn.Attributes["icon"] = "mdi:memory"

	networkOut := NewMetric()
	networkIn.Attributes["friendly_name"] = "Network IO (Bytes Sent)"
	networkIn.Attributes["unit_of_measurement"] = "Bytes"
	networkIn.Attributes["icon"] = "mdi:memory"

	netIO.iface = iface
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

			n.networkIn.State = iface.BytesRecv
			n.networkOut.State = iface.BytesSent

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
