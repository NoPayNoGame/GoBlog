package main

import (
	"fmt"
	"net/http"
)

func main() {
	//http.HandleFunc("/", defaultHandlerfunc)
	//http.HandleFunc("/about", aboutHandlerfunc)
	router := http.NewServeMux()
	router.HandleFunc("/", defaultHandlerfunc)
	router.HandleFunc("/about", aboutHandlerfunc)
	http.ListenAndServe(":3000", router)
}

func aboutHandlerfunc(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(writer, "\n<a href=\"http://www.baidu.com\">官网</a>\n<a href=\"\\mailto:zz_@live.cn\\\">联系我们</a>")
}

func defaultHandlerfunc(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "<h1>主页</h1>")
}

func handlerfunc(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if request.URL.Path == "/" {
		fmt.Fprintf(writer, "<h1>主页</h1>现在会自动重载了")
	} else if request.URL.Path == "/about" {
		fmt.Fprintf(writer, "Orico天下无敌,一支穿云箭,千军万马来相见.")
		fmt.Fprintf(writer, "\r\n<a href=\"http://www.baidu.com\">官网</a>\r\n<a href=\"\\mailto:zz_@live.cn\\\">联系我们</a>")
	} else {
		writer.WriteHeader(404)
		fmt.Fprintf(writer, "你到底要哦该咯???")
	}
}
