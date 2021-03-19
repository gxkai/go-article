package main

import (
	bytes "bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	rand "math/rand"
	"net/http"
	"time"
)

// 请求体结构体
type RequestBody struct {
	Key    string `json:"key"`
	Info   string `json:"info"`
	UserId string `json:"userid"`
}

// 结果体结构体
type ResponseBody struct {
	Code int      `json:"code"`
	Text string   `json:"text"`
	List []string `json:"list"`
	Url  string   `json:"url"`
}

func robot(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var reqData RequestBody
	json.Unmarshal(reqBody, &reqData)
	// 构建请求体
	reqData.Key = "792bcf45156d488c92e9d11da494b085"
	reqData.UserId = string(rand.Int63())
	// 转义为json
	byteData, _ := json.Marshal(&reqData)
	// 请求聊天机器人接口
	req, err := http.NewRequest("POST",
		"http://www.tuling123.com/openapi/api",
		bytes.NewReader(byteData))

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	} else {
		// 将结果从json中解析并输出到命令行
		body, _ := ioutil.ReadAll(resp.Body)
		var respData ResponseBody
		json.Unmarshal(body, &respData)
		fmt.Println("AI: " + respData.Text)
		json.NewEncoder(w).Encode(respData.Text)
	}
}
