package model

import (
	"fmt"
	"log"
	"time"

	"github.com/asdine/storm"
	"golang.org/x/crypto/bcrypt"
)

type NetworkInfo struct {
	LastTimeUp   string
	LastTimeDown string
	MacAddress   []string
}

type DownTimeLogger struct {
	ID          int           `storm:"increment"`
	Date        string        `storm:"index"`
	NetworkInfo []NetworkInfo `storm:"inline"`
}

type User struct {
	ID       int    `storm:"increment"`
	Name     string `storm:"unique"`
	Password []byte
}

var Db *storm.DB

const DefaultCost = 15

func init() {
	var err error

	Db, err = storm.Open("log.db")
	if err != nil {
		log.Fatalln("Could not start db err: ", err.Error())
	}

	// create admin login if not created
	var admin User
	err = Db.One("Name", "admin", &admin)
	if err != nil { // admin hasn't been created
		admin.Name = "admin"
		admin.Password, err = bcrypt.GenerateFromPassword([]byte("admin"), 15)
		if err != nil {
			log.Fatalln("Could not generate default admin password", err)
		}

		err = Db.Save(&admin)
		if err != nil {
			log.Fatalln("Could not save default admin password to db", err)
		}
	}
	var rr []User
	fmt.Println(Db.All(&rr))
	fmt.Println(rr)
}

func SaveToDatabase(networkInfo NetworkInfo) {
	// check if any data for that day
	Date := time.Now().Format("2006-01-02 Mon")
	var logger DownTimeLogger

	err := Db.One("Date", Date, &logger)
	logger.Date = Date
	logger.NetworkInfo = append(logger.NetworkInfo, networkInfo)

	if err == nil {
		Db.Update(&logger)
	} else if err == storm.ErrNotFound {
		Db.Save(&logger)
	}
}
