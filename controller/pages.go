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

func homePageWithAlert(alertType string, alertMessage string, r *http.Request, w http.ResponseWriter) {
	name, err := decodeCookie(r, w)
	if err != nil {
		http.Redirect(w, r, "/signin", 302)
		return
	}

	data := map[string]interface{}{
		"Alert":        true,
		"AlertType":    alertType,
		"AlertMessage": alertMessage,
		"Servers":      serversInfo,
		"Name":         name,
	}

	var serverData []model.DownTimeLogger
	log.Println(model.Db.All(&serverData))
	if err == nil {
		data["Logs"] = serverData
	}

	tmpl, terr := template.New("home.html").Delims("(%", "%)").ParseFiles("templates/home.html", "templates/logo.html")
	if terr != nil {
		log.Println("Error at home page alert page", terr)
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

func HomePost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()

	oldPassword := r.FormValue("changedOldPassword")
	username := r.FormValue("username")

	if oldPassword != "" {
		changePass(oldPassword, w, r)
	} else if username != "" {
		registerUser(username, w, r)
	} else {
		homePageWithAlert("error", "Form not correctly filled", r, w)
		return
	}
}

func registerUser(username string, w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("regPassword")
	if err := model.Db.One("Name", username, &model.User{}); err == nil {
		homePageWithAlert("error", "Username already taken.", r, w)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		homePageWithAlert("error", "Unable to hash password.", r, w) //ToDo show we show errors like this?
		return
	}

	var user model.User
	user.Name = username
	user.Password = hashedPassword

	err = model.Db.Save(&user)
	if err != nil {
		homePageWithAlert("error", "Could not save user to database", r, w) //ToDo show we show errors like this?
		return
	}

	homePageWithAlert("success", "Registered new user", r, w)
}

func changePass(oldPassword string, w http.ResponseWriter, r *http.Request) {

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
		homePageWithAlert("error", "Password is not recognized.", r, w)
		return
	}
	user.Password, err = bcrypt.GenerateFromPassword([]byte(newPassword), cost)
	if err != nil {
		homePageWithAlert("error", "Password could not be saved", r, w)
		log.Println("Hash password.", err)
		return
	}

	err = model.Db.Update(&user)
	if err != nil {
		homePageWithAlert("error", "Password could not be saved", r, w)
		log.Println("Could not save password to database.", err)
		return
	}

	homePageWithAlert("success", "Password changed", r, w)
}

func SignIn(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := decodeCookie(r, w)
	if err == nil {
		http.Redirect(w, r, "/", 302)
		return
	}

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
	cookie := http.Cookie{
		Name:     "token",
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
		Path:     "/",
	}

	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/signin", 301)
}

func createCookie(name string, w http.ResponseWriter, r *http.Request) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": name,
		"exp":  time.Now().Add(time.Minute * expire),
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
