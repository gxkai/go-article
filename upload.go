package main

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
)

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
