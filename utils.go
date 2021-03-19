package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

//结构体转为map
func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}
	return data
}
func Map2Str(m map[string]interface{}) string {
	str, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
	}
	return string(str)
}

func Struct2Str(obj interface{}) string {
	return Map2Str(Struct2Map(obj))
}
