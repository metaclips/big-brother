package controller

import (
	"errors"
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
	cost   = 15
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

func HomePage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name, err := decodeCookie(r, w)
	if err != nil {
		http.Redirect(w, r, "/signin", 302)
		return
	}

	data := map[string]interface{}{
		"Servers": serversInfo,
		"Name":    name,
	}

	var serverData []model.DownTimeLogger
	log.Println(model.Db.All(&serverData))
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

func ChangePass(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()

	oldPassword := r.FormValue("changedOldPassword")
	newPassword := r.FormValue("changedPassword")

	name, err := decodeCookie(r, w)
	if err != nil {
		http.Redirect(w, r, "/signin", 302)
		return
	}

	var user model.User
	err = model.Db.One("Name", name, &user)
	if err != nil {
		signPageError("Wrong sign in credentials", w)
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(oldPassword)); err != nil {
		// todo show wrong password
		return
	}
	user.Password, err = bcrypt.GenerateFromPassword([]byte(newPassword), cost)
	if err != nil {
		// todo show could not store password
		return
	}

	err = model.Db.Update(&user)
	if err != nil {
		// todo show could not store password
		return
	}
	// todo show home page noting password has been changed
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
	_, err := decodeCookie(r, w)
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
			Expires: time.Now(),
			MaxAge:  -1,
			Path:    "/"})

	http.Redirect(w, r, "/signin", 301)
}

func createCookie(name string, w http.ResponseWriter, r *http.Request) error {
	// token := jwt.New(jwt.SigningMethodHS512)
	// claims := make(jwt.MapClaims)
	// claims["name"] = name
	// claims["exp"] = time.Now().Add(time.Minute * 30)
	// token.Claims = claims

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": name,
		"exp":  time.Now().Add(time.Minute * 30),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(key))

	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
		Expires:  time.Now().Add(time.Minute * expire),
		Path:     "/",
	}
	http.SetCookie(w, &cookie)

	return nil
}

func decodeCookie(r *http.Request, w http.ResponseWriter) (string, error) {
	//Get cookie
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("could not detect cookie validity")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["name"].(string), nil
	}

	return "", err
}
