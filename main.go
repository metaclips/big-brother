package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/asdine/storm"
)

// todo: get switch name.
// https://danielmiessler.com/study/manually-set-ip-linux/
// set a default ip to lookup data
// todo: get a logger that stores uptime and downtime logs

type Networks struct {
	Up           bool
	Last_Time_Up string
}

type DownTimeLogger struct {
	Date       string `storm:"id"`
	MacAddress string `storm:"index"`
	Networks   `storm:"inline"`
	Downtimes  []string
}

var alarmOn bool
var Servers = make(map[string]Networks)
var syncWg sync.WaitGroup
var downtime = make(map[string]DownTimeLogger)
var Db *storm.DB
var err error

func init() {
	Db, err = storm.Open("log.db")
	defer func() {
		Db.Close()
	}()

	if err != nil {
		log.Fatalln("Could not start db err: ", err.Error())
	}
}

func localAddresses() {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}

	// remove old data from map, if an old server was unplugged
	var iServers = make(map[string]Networks)
	for _, i := range ifaces {
		if Servers[i.HardwareAddr.String()].Last_Time_Up != "" {
			iServers[i.HardwareAddr.String()] = Servers[i.HardwareAddr.String()]
		}
	}
	Servers = iServers

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
					// call alarm off blocking function
					go callOffAlarm()
					fmt.Println("Alarm on")
					alarmOn = true
				}
			}

			// This basically checks if the network is up, if down Add to sync
			// if initially down and network comes up, remove from sync group.
			conn, err := d.Dial("tcp", "google.com:80")
			if err != nil {
				// if network isn't registered
				if Servers[i.HardwareAddr.String()].Last_Time_Up == "" {
					Servers[i.HardwareAddr.String()] = Networks{
						Up:           false,
						Last_Time_Up: time.Now().Format("Mon Jan 2 15:04:05"),
					}
					addToSyncWaitGroup()
				}

				// add to a sync waitgroup if network was recently up and internet isn't available
				// todo, create a logger that logs timeup and timedown
				if Servers[i.HardwareAddr.String()].Up == true {
					iServer := Servers[i.HardwareAddr.String()]
					iServer.Up = false
					Servers[i.HardwareAddr.String()] = iServer
					addToSyncWaitGroup()
				}

				log.Println("Could not ping addr", i, err, a)
			} else {

				// check if network is registered or was initially down
				if Servers[i.HardwareAddr.String()].Last_Time_Up == "" { //if network isn't registered
					Servers[i.HardwareAddr.String()] = Networks{Up: true}
				}

				if Servers[i.HardwareAddr.String()].Up == false {
					syncWg.Done()
					fmt.Println("Done one")

					iServer := Servers[i.HardwareAddr.String()]
					iServer.Up = true
					Servers[i.HardwareAddr.String()] = iServer

					// log to db
					downtime[i.HardwareAddr.String()] = DownTimeLogger{
						Date:       time.Now().Format("2006-01-02"),
						MacAddress: i.HardwareAddr.String(),
						Networks:   Servers[i.HardwareAddr.String()],
						Downtimes:  append(downtime[i.HardwareAddr.String()].Downtimes, Servers[i.HardwareAddr.String()].Last_Time_Up+" - "+time.Now().Format("Mon Jan 2 15:04:05")),
					}

					iServer.Last_Time_Up = time.Now().Format("Mon Jan 2 15:04:05")
					Servers[i.HardwareAddr.String()] = iServer

					data := downtime[i.HardwareAddr.String()]
					if err := Db.Update(&data); err != nil {
						fmt.Println(Db.Save(&data))
						fmt.Println("There was an err")
					}
				}

				fmt.Println("success", conn.LocalAddr(), i, a.String(), i.Flags)
			}
		}
	}
}

func callOffAlarm() {
	fmt.Println("Call to off alarm made")
	syncWg.Wait()
	alarmOn = false
	fmt.Println("done")
}

func main() {
	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for {
			<-ticker.C
			localAddresses()
		}
	}()

	http.HandleFunc("/query", querySwitches)
	http.HandleFunc("/all", queryLogs)
	fmt.Println(http.ListenAndServe(":80", nil))
}

func querySwitches(w http.ResponseWriter, r *http.Request) {
	data, err := json.MarshalIndent(Servers, "", "\t")
	if err != nil {
		w.Write([]byte(fmt.Sprintf("server error: %s", err.Error())))
		return
	}
	w.Write(data)
}

func queryLogs(w http.ResponseWriter, r *http.Request) {
	var data []DownTimeLogger
	fmt.Println(Db.Find("Date", time.Now().Format("2006-01-02"), &data))
	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(w.Write(jsonData))
}
