package main

import (
	"fmt"
	"net/http"
	"strings"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/", def)
	router.HandleFunc("/about", about)

	router.HandleFunc("/articles/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.SplitN(r.URL.Path, "/", 3)[2]
		fmt.Fprint(w, "文章 ID："+id)
	})

	router.HandleFunc("/mothod", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "POST":
			fmt.Fprintf(writer, "这是一个POST请求")
		case "GET":
			fmt.Fprintf(writer, "这是一个GET请求")
		default:
			fmt.Fprintf(writer, "这是一个"+request.Method+"请求")
		}

	})

	http.ListenAndServe(":3000", router)
}

func about(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(writer, "此博客是用以记录编程笔记，如您有反馈或建议，请联系 "+
		"<a href=\"mailto:summer@example.com\">summer@example.com</a>")
}

func def(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if request.URL.Path == "/" {
		fmt.Fprintf(writer, "<h1>这是GoBlog</h1>")
	} else {
		writer.WriteHeader(404)
		fmt.Fprint(writer, "<h1>请求页面未找到 :(</h1>"+
			"<p>如有疑惑，请联系我们。</p>")
	}
}
