// main.go
package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
}
type User struct {
	Username string
	Password string
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
	},
	{
		Username: "lqy",
		Password: "1207",
	},
}
var jwtKey = []byte("666")
var username string

func return500(w http.ResponseWriter, err error) {
	fmt.Println(err)
	http.Error(w, "System Error", 500)
}

func return401(w http.ResponseWriter, err error) {
	fmt.Println(err)
	http.Error(w, "Forbidden", 401)
}

func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllArticles")
	Owner := r.FormValue("Owner")
	sqlStr := "SELECT Id, Owner, Title, Description, Content, UpdatedTime FROM Article WHERE 1=1"
	if strings.Compare(Owner, "Mine") == 0 {
		sqlStr = sqlStr + " and Owner='" + username + "'"
	}
	if strings.Compare(Owner, "External") == 0 {
		sqlStr = sqlStr + " and Owner not in ('" + username + "')"
	}
	rows, err :=
		database.Query(sqlStr)
	if err != nil {
		return500(w, err)
		return
	}
	article := Article{}
	articles := []Article{}
	for rows.Next() {
		err := rows.Scan(&article.Id, &article.Owner, &article.Title, &article.Description, &article.Content, &article.UpdatedTime)
		if err != nil {
			return500(w, err)
			return
		}
		articles = append(articles, article)
	}
	fmt.Println(articles)
	json.NewEncoder(w).Encode(articles)
}

func returnSingleArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	rows, err :=
		database.Query("SELECT Id, Owner, Title, Description, Content FROM Article WHERE  Id = $1", id)
	if err != nil {
		return500(w, err)
		return
	}
	article := Article{}
	for rows.Next() {
		err := rows.Scan(&article.Id, &article.Owner, &article.Title, &article.Description, &article.Content)
		if err != nil {
			return500(w, err)
			return
		}
		if article.Id == id {
			json.NewEncoder(w).Encode(article)
		}
	}
}

func createNewArticle(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var article Article
	json.Unmarshal(reqBody, &article)
	article.Owner = username
	tm := time.Unix(time.Now().Unix(), 0).Format("2006-01-02 03:04:05")
	article.CreatedTime = tm
	article.UpdatedTime = tm
	Statement, err :=
		database.Prepare("INSERT INTO Article (Owner, Title, Description, Content,CreatedTime ,UpdatedTime) VALUES (?, ?, ? ,?, ? , ?)")
	if err != nil {
		return500(w, err)
		return
	}
	result, err := Statement.Exec(article.Owner, article.Title, article.Description, article.Content, article.CreatedTime, article.UpdatedTime)
	if err != nil {
		return500(w, err)
		return
	}
	newId, err := result.LastInsertId()
	if err != nil {
		return500(w, err)
		return
	}
	article.Id = strconv.FormatInt(newId, 10)
	fmt.Println(article)
	json.NewEncoder(w).Encode(article)
}

func deleteArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	Statement, err :=
		database.Prepare("DELETE FROM Article WHERE  Id = ?")
	if err != nil {
		return500(w, err)
		return
	}
	Statement.Exec(id)
	json.NewEncoder(w)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return500(w, err)
		return
	}
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("file")
	if err != nil {
		return500(w, err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)
	name := uuid.NewString()
	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("temp-images", "upload-"+name+"-*.png")
	if err != nil {
		return500(w, err)
		return
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return500(w, err)
		return
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)

	endpoint := "oss-cn-beijing.aliyuncs.com"
	client, err := oss.New(endpoint, "LTAI4G5MtT6bYPhngu5EwoR1", "i7vwZ3KZdBCHz2S61HBJ1dRQkhmRA6")
	if err != nil {
		return500(w, err)
		return
	}
	bucketName := "go-img"
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return500(w, err)
		return
	}
	err = bucket.PutObjectFromFile(name, tempFile.Name())
	if err != nil {
		return500(w, err)
		return
	}
	url := bucketName + "." + endpoint + "/" + name

	json.NewEncoder(w).Encode(url)
}
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reg, _ := regexp.Compile("^/static")
		if r.RequestURI == "/login" || reg.MatchString(r.RequestURI) {
			next.ServeHTTP(w, r)
			return
		}
		tokenString := r.Header.Get("Authorization")
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
func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(Middleware)
	myRouter.HandleFunc("/login", returnJwt).Methods(http.MethodPost)
	myRouter.HandleFunc("/articles", returnAllArticles).Methods(http.MethodGet)
	myRouter.HandleFunc("/article", createNewArticle).Methods(http.MethodPost)
	myRouter.HandleFunc("/article/{id}", deleteArticle).Methods(http.MethodDelete)
	myRouter.HandleFunc("/article/{id}", returnSingleArticle).Methods(http.MethodGet)
	myRouter.HandleFunc("/upload", uploadFile).Methods(http.MethodPost)
	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	myRouter.PathPrefix("/static/").Handler(s)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func returnJwt(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var user User
	json.Unmarshal(reqBody, &user)
	isExist := false
	for _, v := range users {
		if v.Username == user.Username && v.Password == user.Password {
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
		json.NewEncoder(w).Encode(tokenString)
		return
	}
	return401(w, errors.New("The user does not exist"))
}

func main() {
	var err error
	database, err =
		sql.Open("sqlite3", "./ionic.db")
	if err != nil {
		log.Fatal(err)
		return
	}
	Statement, err :=
		database.Prepare("CREATE TABLE IF NOT EXISTS Article (Id INTEGER PRIMARY KEY, Owner TEXT, Title TEXT, Description TEXT, Content Text, Image, Text)")
	if err != nil {
		log.Fatal(err)
		return
	}
	Statement.Exec()
	handleRequests()
}
