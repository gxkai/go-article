package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllArticles")
	Owner := r.FormValue("Owner")
	sqlStr := "SELECT Id, Owner, Title, Description, Content, Image, UpdatedTime, LikeNum FROM Article WHERE 1=1"
	if strings.Compare(Owner, "Mine") == 0 {
		sqlStr = sqlStr + " and Owner='" + username + "'"
	}
	if strings.Compare(Owner, "External") == 0 {
		sqlStr = sqlStr + " and Owner not in ('" + username + "')"
	}
	sqlStr += " ORDER BY UpdatedTime DESC "
	rows, err :=
		database.Query(sqlStr)
	if err != nil {
		return500(w, err)
		return
	}
	article := Article{}
	articles := []Article{}
	for rows.Next() {
		err := rows.Scan(&article.Id, &article.Owner, &article.Title, &article.Description, &article.Content, &article.Image, &article.UpdatedTime, &article.LikeNum)
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
		database.Query("SELECT Id, Owner, Title, Description, Content, Image FROM Article WHERE  Id = $1", id)
	if err != nil {
		return500(w, err)
		return
	}
	article := Article{}
	for rows.Next() {
		err := rows.Scan(&article.Id, &article.Owner, &article.Title, &article.Description, &article.Content, &article.Image)
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
		database.Prepare("INSERT INTO Article (Owner, Title, Description, Content, Image ,CreatedTime ,UpdatedTime) VALUES (?, ?, ? , ?,?, ? , ?)")
	if err != nil {
		return500(w, err)
		return
	}
	result, err := Statement.Exec(article.Owner, article.Title, article.Description, article.Content, article.Image, article.CreatedTime, article.UpdatedTime)
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

func updateArticle(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var article Article
	json.Unmarshal(reqBody, &article)
	tm := time.Unix(time.Now().Unix(), 0).Format("2006-01-02 03:04:05")
	article.UpdatedTime = tm
	Statement, err :=
		database.Prepare("UPDATE Article SET Content = ? , UpdatedTime = ?, Image = ? WHERE Id = ?")
	if err != nil {
		return500(w, err)
		return
	}
	_, err = Statement.Exec(article.Content, article.UpdatedTime, article.Image, article.Id)
	if err != nil {
		return500(w, err)
		return
	}
	fmt.Println(article)
	json.NewEncoder(w).Encode(article)
}

func like(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	Statement, err :=
		database.Prepare("UPDATE Article SET LikeNum = LikeNum + 1 WHERE Id = ?")
	if err != nil {
		return500(w, err)
		return
	}
	_, err = Statement.Exec(id)
	if err != nil {
		return500(w, err)
		return
	}
	json.NewEncoder(w)
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
