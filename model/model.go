package model

import (
	"log"

	"github.com/asdine/storm"
	"golang.org/x/crypto/bcrypt"
)

type NetworkInfo struct {
	Up           bool
	LastTimeUp   string
	LastTimeDown string `storm:"index"`
}

type DownTimeLogger struct {
	ID          int `storm:"increment"`
	Date        string
	MacAddress  []string
	NetworkInfo `storm:"inline"`
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
}

func database() {

}
