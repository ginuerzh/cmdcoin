// controller
package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func writeResponse(resp http.ResponseWriter, data interface{}) {
	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	b, _ := json.Marshal(data)
	fmt.Println("<<<", string(b))
	resp.Write(b)
}

type requestBody struct {
	bytes.Buffer
}

func (rb *requestBody) Close() error {
	rb.Reset()
	return nil
}

func DumpReqBodyHandler(r *http.Request) {
	fmt.Println("###", r.URL)

	if r.Method == "GET" || r.Body == nil {
		return
	}
	rb := &requestBody{}
	if _, err := io.Copy(rb, r.Body); err != nil {
		log.Println(err)
	}
	fmt.Println(">>>", rb.String())
	r.Body = rb
}
