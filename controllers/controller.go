// controller
package controllers

import (
	//"fmt"
	"encoding/json"
	"net/http"
)

func writeResponse(resp http.ResponseWriter, data interface{}) {
	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	b, _ := json.Marshal(data)
	//fmt.Println(string(b))
	resp.Write(b)
}
