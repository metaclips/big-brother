package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/metaclips/big-brother/model"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

const (
	key    = "Hello there Unilag"
	expire = 259200
)

func QueryLogs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var data []model.DownTimeLogger
	fmt.Println(model.Db.Find("Date", time.Now().Format("2006-01-02"), &data))

	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(w.Write(jsonData))
}

func QuerySwitches(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data, err := json.MarshalIndent(servers, "", "\t")
	if err != nil {
		w.Write([]byte(fmt.Sprintf("server error: %s", err.Error())))
		return
	}

	w.Write(data)
}

func SignIn(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()
	username := r.FormValue("username")
	pass := r.FormValue("password")

	var user model.User
	err := model.Db.One("Name", username, &user)
	if err != nil {
		w.WriteHeader(422)
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(pass)); err != nil {
		// tell users login is false
		// http.Redirect(w, r, "http://127.0.0.1:8080/signin", http.StatusMovedPermanently)
		w.WriteHeader(422)
		return
	}

	fmt.Println(username, pass)
	json.NewEncoder(w).Encode("OKOK")
}

func createCookie(email string, w http.ResponseWriter, r *http.Request) {
	token := jwt.New(jwt.SigningMethodHS512)
	claims := make(jwt.MapClaims)
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Minute * 30)
	claims["gen"] = "Test" //change this later
	token.Claims = claims

	key, err := token.SignedString([]byte(key))
	if err != nil {
		http.Redirect(w, r, ":8080/signin", 302)
		return
	}

	cookie := http.Cookie{
		Name:     "token",
		Value:    key,
		MaxAge:   expire,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)

	fmt.Println("token err", err)
}
