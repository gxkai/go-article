// main.go
package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wuwenbao/gcors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// Article - Our struct for all articles
type Article struct {
	Id          string
	Owner       string
	Title       string
	Description string
	Content     string
	Image       string
	CreatedTime string
	UpdatedTime string
	LikeNum     int
}
type Message struct {
	Id          string
	Owner       string
	Title       string
	Description string
	Content     string
	Image       string
	CreatedTime string
	UpdatedTime string
	LikeNum     int
	Logo        string
}
type User struct {
	Username string
	Password string
	Logo     string
	Token    string
}
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var database *sql.DB
var users = [...]User{
	{
		Username: "gxk",
		Password: "1024",
		Logo:     "https://images.unsplash.com/photo-1561141249-189f1b37e6e6?ixlib=rb-1.2.1&ixid=MXwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHw%3D&auto=format&fit=crop&w=556&q=80",
	},
	{
		Username: "lqy",
		Password: "1207",
		Logo:     "https://images.unsplash.com/photo-1613988958430-75932db23fbf?ixid=MXwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHw%3D&ixlib=rb-1.2.1&auto=format&fit=crop&w=564&q=80",
	},
}
var jwtKey = []byte("666")
var username string
var user User

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reg, _ := regexp.Compile("^/static")
		if r.RequestURI == "/login" || reg.MatchString(r.RequestURI) {
			next.ServeHTTP(w, r)
			return
		}
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			tokenString = r.FormValue("Authorization")
		}
		if tokenString == "" {
			return401(w, errors.New("token is empty"))
			return
		}
		tokenString, err := strconv.Unquote(tokenString)
		if err != nil {
			return401(w, err)
			return
		}
		claims := &Claims{}
		tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			return401(w, err)
			return
		}
		if !tkn.Valid {
			return401(w, err)
			return
		}
		isExist := false
		for _, v := range users {
			if v.Username == claims.Username {
				user = v
				isExist = true
				break
			}
		}
		if isExist == true {
			username = claims.Username
			next.ServeHTTP(w, r)
		} else {
			return401(w, err)
		}
	})
}
func ws(w http.ResponseWriter, r *http.Request) {

	serveWs(hub, w, r)
}
func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(Middleware)
	myRouter.HandleFunc("/login", returnJwt).Methods(http.MethodPost)
	myRouter.HandleFunc("/articles", returnAllArticles).Methods(http.MethodGet)
	myRouter.HandleFunc("/article", createNewArticle).Methods(http.MethodPost)
	myRouter.HandleFunc("/article", updateArticle).Methods(http.MethodPut)
	myRouter.HandleFunc("/article/{id}", deleteArticle).Methods(http.MethodDelete)
	myRouter.HandleFunc("/article/{id}", returnSingleArticle).Methods(http.MethodGet)
	myRouter.HandleFunc("/upload", uploadFile).Methods(http.MethodPost)
	myRouter.HandleFunc("/like/{id}", like).Methods(http.MethodPost)
	myRouter.HandleFunc("/robot", robot).Methods(http.MethodPost)
	myRouter.HandleFunc("/ws", ws).Methods(http.MethodGet)
	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	myRouter.PathPrefix("/static/").Handler(s)
	cors := gcors.New(
		myRouter,
		gcors.WithOrigin("*"),
		gcors.WithMethods("*"),
		gcors.WithHeaders("*"),
	)
	log.Fatal(http.ListenAndServe(":10000", cors))
}

func returnJwt(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var user User
	json.Unmarshal(reqBody, &user)
	isExist := false
	for _, v := range users {
		if v.Username == user.Username && v.Password == user.Password {
			user.Password = ""
			user.Logo = v.Logo
			isExist = true
			break
		}
	}
	if isExist == true {
		expirationTime := time.Now().Add(60 * 60 * 20 * time.Minute)
		claims := &Claims{
			Username: user.Username,
			StandardClaims: jwt.StandardClaims{
				// In JWT, the expiry time is expressed as unix milliseconds
				ExpiresAt: expirationTime.Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			return500(w, err)
			return
		}
		user.Token = tokenString
		json.NewEncoder(w).Encode(user)
		return
	}
	return401(w, errors.New("The user does not exist"))
}

var hub = newHub()

func main() {
	var err error
	database, err =
		sql.Open("sqlite3", "./ionic.db")
	if err != nil {
		log.Fatal(err)
		return
	}
	Statement, err :=
		database.Prepare("CREATE TABLE IF NOT EXISTS Article (Id INTEGER PRIMARY KEY, Owner TEXT DEFAULT '', Title TEXT DEFAULT '', Description TEXT DEFAULT '', Content Text DEFAULT '', Image Text DEFAULT '',CreatedTime Text DEFAULT '',UpdatedTime Text DEFAULT '', LikeNum INTEGER DEFAULT 0)")
	if err != nil {
		log.Fatal(err)
		return
	}
	Statement.Exec()
	Statement, err = database.Prepare("CREATE TABLE IF NOT EXISTS Message (Id INTEGER PRIMARY KEY, Owner TEXT DEFAULT '', Title TEXT DEFAULT '', Description TEXT DEFAULT '', Content Text DEFAULT '', Image Text DEFAULT '',CreatedTime Text DEFAULT '',UpdatedTime Text DEFAULT '', LikeNum INTEGER DEFAULT 0)")
	if err != nil {
		log.Fatal(err)
		return
	}
	Statement.Exec()
	go hub.run()
	handleRequests()
}
