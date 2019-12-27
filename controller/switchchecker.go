package controller

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/asdine/storm/q"

	"github.com/metaclips/big-brother/model"
)

var alarmOn bool
var servers = make(map[string]model.NetworkInfo)
var syncWg sync.WaitGroup

func MonitorSwitches() {
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

	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}

	// remove old data from map, if a server was unplugged
	var iServers = make(map[string]model.NetworkInfo)
	var downServerMacAddress []string
	for _, i := range ifaces {
		if servers[i.HardwareAddr.String()].LastTimeUp != "" {
			iServers[i.HardwareAddr.String()] = servers[i.HardwareAddr.String()]
			if iServers[i.HardwareAddr.String()].Up == false {
				downServerMacAddress = append(downServerMacAddress, i.HardwareAddr.String())
			}
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

			// This basically checks if the network is up, if down Add to sync
			// if initially down and network comes up, remove from sync group.
			conn, err := d.Dial("tcp", "google.com:80")
			if err != nil {
				// if network isn't registered or if network was recently up and internet isn't available
				if servers[i.HardwareAddr.String()].LastTimeDown == "" || servers[i.HardwareAddr.String()].Up == true {
					servers[i.HardwareAddr.String()] = model.NetworkInfo{
						Up:           false,
						LastTimeDown: time.Now().Format(time.Kitchen),
					}

					addToSyncWaitGroup()
				}

				log.Println("Could not ping addr", i, err, a)
			} else {

				// check if network is registered or was initially down
				if servers[i.HardwareAddr.String()].LastTimeUp == "" || servers[i.HardwareAddr.String()].Up == false { //if network isn't registered
					servers[i.HardwareAddr.String()] = model.NetworkInfo{
						Up:         true,
						LastTimeUp: time.Now().Format(time.Kitchen),
					}

					if servers[i.HardwareAddr.String()].Up == false {
						syncWg.Done()
					}

					fmt.Println("Done one")

					// check if a recent uptime was saved in database
					var user model.DownTimeLogger

					checkRecentDownTime := model.Db.Select(
						q.And(
							q.Eq("Date", time.Now().Format("2006-01-02")),
							q.Eq("LastTimeDown", time.Now().Format(time.Kitchen)),
						),
					)

					err := checkRecentDownTime.First(&user)
					if err != nil {

						// only log to db when server switches from down-up or up-down state
						user = model.DownTimeLogger{
							Date:        time.Now().Format("2006-01-02"),
							NetworkInfo: servers[i.HardwareAddr.String()],
						}
						user.MacAddress = append(user.MacAddress, i.HardwareAddr.String())

					} else {
						user.MacAddress = append(user.MacAddress, i.HardwareAddr.String())
					}

					if err := model.Db.Update(&user); err != nil {
						fmt.Println(model.Db.Save(&user))
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
