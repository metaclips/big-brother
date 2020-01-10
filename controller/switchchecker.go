package controller

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/metaclips/big-brother/model"
)

var alarmOn bool

var serversInfo = make(map[string]bool)
var downServers model.NetworkInfo
var syncWg sync.WaitGroup

func MonitorSwitches() {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	iServer := make(map[string]bool)

	var wassDown []string
	for _, iface := range ifaces {
		// use a sync go routine here to check all interface at once
		if iface.Flags&net.FlagLoopback != 0 {
			continue // if loopback interface
		}

		if iface.Flags&net.FlagUp == 0 {
			continue // if interface down
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, a := range addrs {
			tcpAddr := &net.TCPAddr{
				IP: a.(*net.IPNet).IP,
			}

			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			d := net.Dialer{
				LocalAddr: tcpAddr,
				Timeout:   15 * time.Second,
			}

			// This basically checks if the network is up
			// If down, add to server
			_, err := d.Dial("tcp", "google.com:80")

			if err != nil {
				iServer[iface.HardwareAddr.String()] = false

				if len(downServers.MacAddress) > 0 {
					// save to database
					model.SaveToDatabase(downServers)
				}

				// add to server down list
				exists, _ := isInArray(iface.HardwareAddr.String(), downServers.MacAddress)
				if !exists {
					downServers.MacAddress = append(downServers.MacAddress, iface.HardwareAddr.String())
				}
				downServers.LastTimeUp = time.Now().Format("15:04:05.000")

			} else {
				iServer[iface.HardwareAddr.String()] = true

				exists, pos := isInArray(iface.HardwareAddr.String(), downServers.MacAddress)
				if exists {
					wassDown = append(wassDown, iface.HardwareAddr.String())
					copy(downServers.MacAddress[pos:], downServers.MacAddress[pos+1:])
					downServers.LastTimeUp = ""
					downServers.LastTimeDown = time.Now().Format("15:04:05.000")
					model.SaveToDatabase(downServers)
				}
			}
		}
	}
	serversInfo = iServer
}

func blockingAlarm() {
	fmt.Println("Call to on alarm made")
	syncWg.Wait()
	fmt.Println("Call to off alarm made")
	alarmOn = false
	fmt.Println("done")
}

func isInArray(str string, array []string) (bool, int) {
	for pos, word := range array {
		if word == str {
			return true, pos
		}
	}

	return false, 0
}
