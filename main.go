package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

// 创建mux路由
var router = mux.NewRouter()

func main() {

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

	//	创建博文表单
	router.HandleFunc("/articles/create", articlesCreateHandler).Methods("GET").Name("articles.create")

	//	重写404
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	//	中间件:强制内容为html
	router.Use(forceHTMLMiddleware)

	http.ListenAndServe(":3000", removeTrailingSlash(router))

}

/*
创建博文表单
*/
func articlesCreateHandler(writer http.ResponseWriter, request *http.Request) {
	html := `
	<!DOCTYPE html>
	<html lang='en'>
	<head>
		<title>创建文章 -- 我的技术博客</title>
	</head>
	<form action="%s?test=data" method="post">
		<p><input type="text" name="MyTitle"></p>
		<p><textarea name = "MyBody" cols ="30" rows="10"></textarea></p>
        <p><button type="submit">提交</button></p>
	</form>
	</body>
	</html>
`
	storeURL, _ := router.Get("articles.store").URL()

	fmt.Fprintf(writer, html, storeURL)
	//fmt.Fprintf(writer, html)
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
	//	如果解析错误,处理错误.
	err := request.ParseForm()
	if err != nil {
		fmt.Fprintf(writer, "请提供正确的数据")
	}

	title := request.PostForm.Get("MyTitle")
	//
	//	打印PostForm
	fmt.Fprintf(writer, "PostFrom:%v<br>", request.PostForm)
	fmt.Fprintf(writer, "Form:%v<br>", request.Form)
	fmt.Fprintf(writer, "MyTitle:%v<br><br><br>", title)

	fmt.Fprintf(writer, "r.FormValue 中 MyTitle 的值为:%v<br>", request.FormValue("MyTitle"))
	fmt.Fprintf(writer, "r.PostFormValue 中 MyTitle 的值为:%v<br><br><br>", request.PostFormValue("MyTitle"))

	fmt.Fprintf(writer, "r.FormVlue 中 test 的值为:%v<br>", request.FormValue("test"))
	fmt.Fprintf(writer, "r.PostFormValue 中 test 的值为:%v<br>", request.PostFormValue("test"))
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
