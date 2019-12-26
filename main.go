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
	"github.com/julienschmidt/httprouter"
)

// todo: get switch name.
// https://danielmiessler.com/study/manually-set-ip-linux/
// set a default ip to lookup data
// todo: get a logger that stores uptime and downtime logs





func main() {
	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for {
			<-ticker.C
			localAddresses()
		}
	}()

	router := httprouter.New()
	router.POST("/signin", signIn)
	router.GET("/query", querySwitches)
	router.GET("/all", queryLogs)
	fmt.Println(http.ListenAndServe(":3000", router))
}





func signIn(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()

	username := r.FormValue("username")
	pass := r.FormValue("password")
	fmt.Println(username, pass)
}
