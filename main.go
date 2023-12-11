package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

func main() {
	//	创建mux路由
	router := mux.NewRouter()

	//	home
	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	//	about
	router.HandleFunc("/about", aboutHandeler).Methods("GET") //.Name("about")

	//	get指定id返回对应内容
	router.HandleFunc("/articles{id:[0-9]+}", articlesShowHandeler).Methods("GET").Name("articles.show")
	//	get方法
	router.HandleFunc("/articles", articlesIndexHandeler).Methods("GET").Name("articles.index")
	//	POST方法
	router.HandleFunc("/articles", articlesStoreHandeler).Methods("POST").Name("articles.store")

	//	重写404
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	//	中间件:强制内容为html
	router.Use(forceHTMLMiddleware)

	//	通过命名路由获取 URL 示例
	homeURL, _ := router.Get("home").URL()
	fmt.Println("homeURL:", homeURL)
	articleURL, _ := router.Get("articles.show").URL("id", "23")
	fmt.Println("articleURL:", articleURL)

	http.ListenAndServe(":3000", removeTrailingSlash(router))

}

func removeTrailingSlash(router *mux.Router) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			request.URL.Path = strings.TrimSuffix(request.URL.Path, "/")
		}
		router.ServeHTTP(writer, request)
	})
}

func forceHTMLMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		//	设置标头
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")

		// 2. 继续处理请求
		handler.ServeHTTP(writer, request)
	})
}

func notFoundHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
	fmt.Fprint(writer, "<h1>请求页面未找到 :(</h1><p>如有疑惑，请联系我们。</p>")
}

func articlesStoreHandeler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "POST创建新的文章")
}

func articlesIndexHandeler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "GET访问文章列表")
}

func articlesShowHandeler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id := vars["id"]
	fmt.Fprintf(writer, "请求的文章ID:"+id)
}

func aboutHandeler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "此博客是用以记录编程笔记，如您有反馈或建议，请联系 <a href=\"mailto:summer@example.com\">summer@example.com</a>")
}

func homeHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprint(writer, "<h1>Hello, 欢迎来到 goblog！</h1>")
}
