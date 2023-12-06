package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", handlerfunc)
	http.ListenAndServe(":3000", nil)
}

func handlerfunc(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "<h1>1这是goblog</h1>")
}
