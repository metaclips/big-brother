package main

import (
	"net/http"
	"time"

	"github.com/metaclips/big-brother/controller"

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
			controller.MonitorSwitches()
		}
	}()

	router := httprouter.New()

	router.GET("/signin", controller.SignIn)
	router.GET("/", controller.HomePage)

	router.POST("/logout", controller.Logout)
	router.POST("/signin", controller.SignInPost)
	router.POST("/", controller.HomePost)

	router.ServeFiles("/assets/*filepath", http.Dir("./templates/assets"))

	http.ListenAndServe(":8080", router)
}
