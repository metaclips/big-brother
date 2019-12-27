package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/metaclips/big-brother/model"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

const (
	key    = "Hello there Unilag"
	host   = "http://127.0.0.1:8080"
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
	err := decodeCookie(r, w)
	if err != nil {
		http.Redirect(w, r, host+"/signin", http.StatusTemporaryRedirect)
		return
	}

	data, err := json.MarshalIndent(servers, "", "\t")
	if err != nil {
		w.Write([]byte(fmt.Sprintf("server error: %s", err.Error())))
		return
	}

	w.Write(data)
}

func SignIn(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Println(r.ParseForm())

	username := r.FormValue("username")
	pass := r.FormValue("password")

	var user model.User
	err := model.Db.One("Name", username, &user)
	if err != nil {
		w.WriteHeader(422)
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(pass)); err != nil {
		w.WriteHeader(422)
		return
	}

	createCookie(username, w, r)
	http.Redirect(w, r, host, 301)
}

func createCookie(email string, w http.ResponseWriter, r *http.Request) {
	token := jwt.New(jwt.SigningMethodHS512)
	claims := make(jwt.MapClaims)
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Minute * 30)
	token.Claims = claims

	key, err := token.SignedString([]byte(key))
	if err != nil {
		http.Redirect(w, r, host+"/signin", 301)
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

func decodeCookie(r *http.Request, w http.ResponseWriter) error {
	//Get cookie
	cookie, err := r.Cookie("token")
	if err != nil {
		return err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	if err == nil && token.Valid {
		return nil
	}

	return err
}
