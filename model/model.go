package model

import (
	"log"
	"time"

	"github.com/asdine/storm"
)

type NetworkInfo struct {
	Up           bool
	Last_Time_Up time.Time
	TimeDown     time.Time
	TimeUp       time.Time
}

type DownTimeLogger struct {
	Date        string `storm:"id"`
	SwitchNo    []int  `storm:"index"` // multiple switches could be down at the same time
	MacAddress  string
	NetworkInfo `storm:"inline"`
}

var Db *storm.DB

func init() {
	var err error

	Db, err = storm.Open("log.db")

	if err != nil {
		log.Fatalln("Could not start db err: ", err.Error())
	}
}

func database() {

}
