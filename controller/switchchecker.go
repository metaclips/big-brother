package controller

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/metaclips/big-brother/model"
)

var alarmOn bool
var servers = make(map[string]model.NetworkInfo)
var syncWg sync.WaitGroup
var downtime = make(map[string]model.DownTimeLogger)

func localAddresses() {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}

	// remove old data from map, if a server was unplugged
	var iServers = make(map[string]model.NetworkInfo)
	for _, i := range ifaces {
		if !servers[i.HardwareAddr.String()].Last_Time_Up.IsZero() {
			iServers[i.HardwareAddr.String()] = servers[i.HardwareAddr.String()]
		}
	}

	servers = iServers

	for _, i := range ifaces {
		if i.Flags&net.FlagLoopback != 0 {
			continue // if loopback interface
		}

		if i.Flags&net.FlagUp == 0 {
			continue // if interface down
		}

		addrs, err := i.Addrs()
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
				Timeout:   5 * time.Second, // timeout at 5sec, most times network could be down
			}

			addToSyncWaitGroup := func() {
				syncWg.Add(1)
				fmt.Println("About to")

				if !alarmOn {
					// call alarm off blocking alarm only if not on
					go blockingAlarm()
					fmt.Println("Alarm on")
					alarmOn = true
				}
			}

			// This basically checks if the network is up, if down Add to sync
			// if initially down and network comes up, remove from sync group.
			conn, err := d.Dial("tcp", "google.com:80")
			if err != nil {
				// if network isn't registered
				if servers[i.HardwareAddr.String()].Last_Time_Up.IsZero() {
					servers[i.HardwareAddr.String()] = model.NetworkInfo{
						Up:           false,
						Last_Time_Up: time.Now(),
					}

					addToSyncWaitGroup()
				}

				// add to a sync waitgroup if network was recently up and internet isn't available
				// todo, create a logger that logs timeup and timedown
				if servers[i.HardwareAddr.String()].Up == true {
					iServer := servers[i.HardwareAddr.String()]
					iServer.Up = false
					servers[i.HardwareAddr.String()] = iServer

					addToSyncWaitGroup()
				}

				log.Println("Could not ping addr", i, err, a)
			} else {

				// check if network is registered or was initially down
				if servers[i.HardwareAddr.String()].Last_Time_Up.IsZero() { //if network isn't registered
					servers[i.HardwareAddr.String()] = model.NetworkInfo{Up: true}
				}

				if servers[i.HardwareAddr.String()].Up == false {
					syncWg.Done()
					fmt.Println("Done one")

					iServer := servers[i.HardwareAddr.String()]
					iServer.Up = true
					servers[i.HardwareAddr.String()] = iServer

					// log to db
					downtime[i.HardwareAddr.String()] = model.DownTimeLogger{
						Date:        time.Now().Format("2006-01-02"),
						MacAddress:  i.HardwareAddr.String(),
						NetworkInfo: servers[i.HardwareAddr.String()],
						// Downtimes:  append(downtime[i.HardwareAddr.String()].Downtimes, Servers[i.HardwareAddr.String()].Last_Time_Up+" - "+time.Now().Format("Mon Jan 2 15:04:05")),
					}

					iServer.Last_Time_Up = time.Now()
					servers[i.HardwareAddr.String()] = iServer

					data := downtime[i.HardwareAddr.String()]
					if err := model.Db.Update(&data); err != nil {
						fmt.Println(model.Db.Save(&data))
						fmt.Println("There was an err")
					}
				}

				fmt.Println("success", conn.LocalAddr(), i, a.String(), i.Flags)
			}
		}
	}
}

func blockingAlarm() {
	fmt.Println("Call to on alarm made")
	syncWg.Wait()
	fmt.Println("Call to off alarm made")
	alarmOn = false
	fmt.Println("done")
}
