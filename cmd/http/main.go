package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./static"))
	log.Println("Serving on http://localhost:8080/json-to-go")
	_ = http.ListenAndServe(":8080", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Cache-Control", "no-cache")
		fs.ServeHTTP(resp, req)
	}))
}
