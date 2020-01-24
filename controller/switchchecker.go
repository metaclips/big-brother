package controller

import (
	"net"
	"sync"
	"time"

	"github.com/metaclips/big-brother/model"
)

type serverInfo struct {
	macAddress   string
	lastTimeDown string
	lastTimeUp   string
}

type downServerStruct struct {
	server []serverInfo
	mux    sync.Mutex
}

var alarmOn bool

var serversInfo = make(map[string]bool)
var downServers downServerStruct
var syncWg sync.WaitGroup

func monitor() {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}

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
			_, err := d.Dial("tcp", "google.com:80")

			if err != nil { // server is down
				exists, _ := isInArray(iface.HardwareAddr.String(), downServers.server)

				if exists { // if it exists, it's already in goroutine
					continue
				}

				downServer := serverInfo{
					macAddress: iface.HardwareAddr.String(),
					lastTimeUp: time.Now().Format("2006-01-02 3:4 AM")}

				// add to down servers
				downServers.server = append(downServers.server, downServer)

				locationOnStack := len(downServers.server) - 1
				removeAddress := func() {
					downServers.server[locationOnStack] = downServers.server[len(downServers.server)-1] // Copy last element to index i.
					downServers.server[len(downServers.server)-1] = serverInfo{}                        // Erase last element (write zero value).
					downServers.server = downServers.server[:len(downServers.server)-1]                 // Truncate slice.
				}
				go func() {
					downServers.mux.Lock()

					for _, err := d.Dial("tcp", "google.com:80"); err != nil; {
						time.Sleep(5 * time.Second)
					}

					// if network is up
					networkInfo := model.NetworkInfo{
						LastTimeUp:   downServer.lastTimeDown,
						LastTimeDown: time.Now().Format("2006-01-02 3:4 AM"),
						MacAddress:   []string{downServer.macAddress},
					}

					var networkLogger model.DownTimeLogger

					err = model.Db.One("Date", time.Now().Format("2006-01-02 Mon"), &networkLogger)
					if err != nil || len(networkLogger.NetworkInfo) == 0 { // no reg for that date
						model.SaveToDatabase(networkInfo)
						downServers.mux.Unlock()
						removeAddress()
						return
					}

					pos := len(networkLogger.NetworkInfo) - 1
					if networkLogger.NetworkInfo[pos].LastTimeUp == networkInfo.LastTimeUp {
						networkLogger.NetworkInfo[pos].LastTimeDown = networkInfo.LastTimeDown
						networkLogger.NetworkInfo[pos].MacAddress = append(networkLogger.NetworkInfo[pos].MacAddress, downServer.macAddress)
						model.Db.Update(networkLogger)
					} else {
						model.SaveToDatabase(networkInfo)
					}

					removeAddress()
					downServers.mux.Unlock()
				}()
			}
		}
	}
}

// func MonitorSwitches() {
// 	ifaces, err := net.Interfaces()
// 	if err != nil {
// 		return
// 	}
// 	iServer := make(map[string]bool)

// 	var wassDown []string
// 	for _, iface := range ifaces {
// 		// use a sync go routine here to check all interface at once
// 		if iface.Flags&net.FlagLoopback != 0 {
// 			continue // if loopback interface
// 		}

// 		if iface.Flags&net.FlagUp == 0 {
// 			continue // if interface down
// 		}

// 		addrs, err := iface.Addrs()
// 		if err != nil {
// 			continue
// 		}

// 		for _, a := range addrs {
// 			tcpAddr := &net.TCPAddr{
// 				IP: a.(*net.IPNet).IP,
// 			}

// 			var ip net.IP
// 			switch v := a.(type) {
// 			case *net.IPNet:
// 				ip = v.IP
// 			case *net.IPAddr:
// 				ip = v.IP
// 			}

// 			if ip == nil || ip.IsLoopback() {
// 				continue
// 			}

// 			ip = ip.To4()
// 			if ip == nil {
// 				continue // not an ipv4 address
// 			}

// 			d := net.Dialer{
// 				LocalAddr: tcpAddr,
// 				Timeout:   15 * time.Second,
// 			}

// 			// This basically checks if the network is up
// 			// If down, add to server
// 			_, err := d.Dial("tcp", "google.com:80")

// 			if err != nil {
// 				iServer[iface.HardwareAddr.String()] = false

// 				if len(downServers.MacAddress) > 0 {
// 					// save to database
// 					model.SaveToDatabase(downServers)
// 				}

// 				// add to server down list
// 				exists, _ := isInArray(iface.HardwareAddr.String(), downServers.MacAddress)
// 				if !exists {
// 					downServers.MacAddress = append(downServers.MacAddress, iface.HardwareAddr.String())
// 				}
// 				downServers.LastTimeUp = time.Now().Format("15:04:05 Mon")

// 			} else {
// 				iServer[iface.HardwareAddr.String()] = true

// 				exists, pos := isInArray(iface.HardwareAddr.String(), downServers.MacAddress)
// 				if exists {
// 					wassDown = append(wassDown, iface.HardwareAddr.String())
// 					copy(downServers.MacAddress[pos:], downServers.MacAddress[pos+1:])
// 					downServers.LastTimeUp = ""
// 					downServers.LastTimeDown = time.Now().Format("15:04:05.000")
// 					model.SaveToDatabase(downServers)
// 				}
// 			}
// 		}
// 	}
// 	serversInfo = iServer
// }

func isInArray(str string, array []serverInfo) (bool, int) {
	for pos, word := range array {
		if word.macAddress == str {
			return true, pos
		}
	}

	return false, 0
}
