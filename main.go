package main

import (
	"net/http"
	"time"

	"github.com/metaclips/big-brother/controller"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
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
			controller.MonitorSwitches()
		}
	}()

	router := httprouter.New()

	router.POST("/signin", controller.SignIn)
	router.GET("/query", controller.QuerySwitches)
	router.GET("/all", controller.QueryLogs)

	handler := cors.Default().Handler(router)

	http.ListenAndServe(":3000", handler)
}
