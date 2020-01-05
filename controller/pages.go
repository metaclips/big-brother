package controller

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"

	"github.com/metaclips/big-brother/model"
)

const (
	key    = "Hello there Unilag"
	expire = 30
)

func signPageError(err string, w http.ResponseWriter) {
	data := map[string]interface{}{
		"Error": err,
	}

	tmpl, terr := template.New("login.html").Delims("(%", "%)").ParseFiles("templates/login.html", "templates/logo.html")
	if terr != nil {
		log.Println("Error at refund.html", terr)
		return
	}
	if terr = tmpl.Execute(w, data); terr != nil {
		log.Println(terr)
	}
}

func QueryLogs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var data []model.DownTimeLogger
	fmt.Println(model.Db.Find("Date", time.Now().Format("2006-01-02"), &data))

	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Fatalln(err)
	}

	w.Write(jsonData)
}

func QuerySwitches(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := decodeCookie(r, w)
	if err != nil {
		w.WriteHeader(425)
		return
	}

	data, err := json.MarshalIndent(servers, "", "\t")
	if err != nil {
		w.Write([]byte(fmt.Sprintf("server error: %s", err.Error())))
		return
	}

	w.Write(data)
}

func HomePage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := decodeCookie(r, w)
	if err != nil {
		http.Redirect(w, r, "/signin", 302)
		return
	}

	data := map[string]interface{}{
		"Servers": servers,
	}

	var serverData []model.DownTimeLogger
	err = model.Db.Find("Date", time.Now().Format("2006-01-02"), &serverData)
	if err == nil {
		data["Logs"] = serverData
	}

	tmpl, terr := template.New("home.html").Delims("(%", "%)").ParseFiles("templates/home.html", "templates/logo.html")
	if terr != nil {
		log.Println("Error at refund.html", terr)
		return
	}

	if terr = tmpl.Execute(w, data); terr != nil {
		log.Println(terr)
	}
}

func SignIn(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tmpl, terr := template.New("login.html").Delims("(%", "%)").ParseFiles("templates/login.html", "templates/logo.html")
	if terr != nil {
		log.Println("Error at refund.html", terr)
		return
	}

	if terr = tmpl.Execute(w, nil); terr != nil {
		log.Println(terr)
	}
}

func SignInPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := decodeCookie(r, w)
	if err == nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	r.ParseForm()

	username := r.FormValue("username")
	pass := r.FormValue("password")

	var user model.User
	err = model.Db.One("Name", username, &user)
	if err != nil {
		signPageError("Wrong sign in credentials", w)
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(pass)); err != nil {
		signPageError("Wrong sign in credentials", w)
		return
	}

	err = createCookie(username, w, r)
	if err != nil {
		http.Redirect(w, r, "/signin", 301)
		return
	}

	http.Redirect(w, r, "/", 302)
}

func Logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.SetCookie(
		w,
		&http.Cookie{
			Value:   "token",
			Expires: time.Now(), MaxAge: -1})

	http.Redirect(w, r, "/signin", 301)
}

func createCookie(email string, w http.ResponseWriter, r *http.Request) error {
	token := jwt.New(jwt.SigningMethodHS512)
	claims := make(jwt.MapClaims)
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Minute * 30)
	token.Claims = claims

	uniqueKey, err := token.SignedString([]byte(key))
	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     "token",
		Value:    uniqueKey,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
		Expires:  time.Now().Add(time.Minute * expire),
		Path:     "/",
	}
	http.SetCookie(w, &cookie)
	cc := r.Cookies()
	fmt.Println(cc)

	return nil
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
