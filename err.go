package main

import (
	"fmt"
	"net/http"
)

func return500(w http.ResponseWriter, err error) {
	fmt.Println(err)
	http.Error(w, "System Error", 500)
}

func return401(w http.ResponseWriter, err error) {
	fmt.Println(err)
	http.Error(w, "Forbidden", 401)
}
