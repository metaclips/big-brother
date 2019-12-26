package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/metaclips/big-brother/model"

	"github.com/julienschmidt/httprouter"
)

func queryLogs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var data []model.DownTimeLogger
	fmt.Println(model.Db.Find("Date", time.Now().Format("2006-01-02"), &data))

	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(w.Write(jsonData))
}

func querySwitches(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data, err := json.MarshalIndent(servers, "", "\t")
	if err != nil {
		w.Write([]byte(fmt.Sprintf("server error: %s", err.Error())))
		return
	}

	w.Write(data)
}
