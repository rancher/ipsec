package monitor

import (
	"fmt"
	"time"

	"github.com/bronze1man/goStrongswanVici"
	"github.com/leodotcloud/log"
	"github.com/rancher/go-rancher-metadata/metadata"
)

// SAsMonitor ...
type SAsMonitor struct {
	mc metadata.Client
}

const (
	startDelay         = time.Duration(60) * time.Second
	monitorSAsInterval = time.Duration(5) * time.Second
)

// Watch monitors the IPSec SAs and intiates the tunnels if missing
func Watch(mc metadata.Client) {
	sm := SAsMonitor{
		mc: mc,
	}

	go sm.monitorSAs()
}
func getClient() (*goStrongswanVici.ClientConn, error) {
	var err error
	for i := 0; i < 3; i++ {
		var client *goStrongswanVici.ClientConn
		client, err = goStrongswanVici.NewClientConnFromDefaultSocket()
		if err == nil {
			return client, nil
		}

		if i > 0 {
			log.Errorf("Failed to connect to charon: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	return nil, err
}

func buildHostsMap(hosts []metadata.Host, selfHost metadata.Host) map[string]bool {
	hostsMap := map[string]bool{}

	for _, aHost := range hosts {
		if aHost.UUID == selfHost.UUID {
			continue
		}
		hostsMap[aHost.AgentIP] = false
	}

	return hostsMap
}

// This function is used to check the IPSec SAs
// to be present for the existing hosts
func (sm *SAsMonitor) monitorSAs() {
	log.Infof("samonitor: sleeping initially for %v", startDelay)
	time.Sleep(startDelay)
	log.Infof("samonitor: started monitoring IPSec SAs")
	for {
		time.Sleep(monitorSAsInterval)
		selfService, err := sm.mc.GetSelfService()
		if err != nil {
			log.Errorf("samonitor: error fetching self service: %v", err)
			continue
		}
		if selfService.State != "active" {
			log.Infof("samonitor: skipping as service is not active but in % state", selfService.State)
			continue
		}

		hosts, err := sm.mc.GetHosts()
		if err != nil {
			log.Errorf("samonitor: error fetching hosts: %v", err)
			continue
		}

		selfHost, err := sm.mc.GetSelfHost()
		if err != nil {
			log.Errorf("samonitor: error fetching self host: %v", err)
			continue
		}

		hostsMap := buildHostsMap(hosts, selfHost)
		log.Debugf("samonitor: hostsMap: %v", hostsMap)

		client, err := getClient()
		if err != nil {
			log.Errorf("samonitor: error getting strongswan client: %v", err)
			continue
		}
		defer client.Close()

		sas, err := client.ListSas("", "")
		if err != nil {
			log.Errorf("samonitor: error getting list of sas from strongswan: %v", err)
			continue
		}

		for _, aSA := range sas {
			log.Debugf("samonitor: sa: %+v", aSA)
			for k, v := range aSA {
				log.Debugf("samonitor: sa details k=%v, v=%v", k, v)
				hostsMap[v.Remote_host] = true
			}
		}

		for host, saFound := range hostsMap {
			if !saFound {
				log.Infof("samonitor: expected SA for host: %v, but not found.", host)
				childSA := fmt.Sprintf("child-%v", host)
				err := client.Initiate(childSA, "")
				if err != nil {
					log.Errorf("samonitor: error initiating missing SA %v: %v", childSA, err)
				}
			}
		}
	}
}
