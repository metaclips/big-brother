package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/metaclips/big-brother/controller"
	"github.com/metaclips/big-brother/model"
)

// todo: get switch name.
// https://danielmiessler.com/study/manually-set-ip-linux/
// set a default ip to lookup data
// todo: get a logger that stores uptime and downtime logs

func main() {
	defer model.Db.Close()
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

	fmt.Println(http.ListenAndServe(":8080", router))
}
