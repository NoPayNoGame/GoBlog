package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

// ArticlesFormData 创建博文表单数据
type ArticlesFormData struct {
	Title, Body string
	URL         *url.URL
	Errors      map[string]string
}

// 创建mux路由
var router = mux.NewRouter()

// 创建数据库连接池
var db *sql.DB

func main() {

	//	初始化数据库
	initDB()

	//	创建数据表
	createTables()

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

func createTables() {
	createArticlesSQL := `CREATE TABLE IF NOT EXISTS articles(
    id bigint(20) PRIMARY KEY AUTO_INCREMENT NOT NULL,
    title varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
    body longtext COLLATE utf8mb4_unicode_ci
); `

	_, err := db.Exec(createArticlesSQL)
	checkError(err)
}

func initDB() {
	var err error

	//	创建连接配置信息
	config := mysql.Config{
		User:                 "root",
		Passwd:               "a651651651",
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "goBlog",
		AllowNativePasswords: true,
	}

	//	准备数据库连接池
	db, err = sql.Open("mysql", config.FormatDSN())
	checkError(err)

	//	设置最大连接数
	db.SetMaxOpenConns(25)

	//	设置最大空闲连接数
	db.SetMaxIdleConns(25)

	//	设置每个连接的过期时间
	db.SetConnMaxIdleTime(5 * time.Minute)

	//	尝试连接,失败会报错
	err = db.Ping()
	checkError(err)
}

// 错误处理
func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

/*
创建博文表单
*/
func articlesCreateHandler(writer http.ResponseWriter, request *http.Request) {

	storeURL, _ := router.Get("articles.store").URL()

	data := ArticlesFormData{
		Title:  "",
		Body:   "",
		URL:    storeURL,
		Errors: nil,
	}

	files, err := template.ParseFiles("resources/views/articles/create.gohtml")
	if err != nil {
		panic(err)
	}

	err = files.Execute(writer, data)
	if err != nil {
		panic(err)
	}
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

	//	获取标题和内容
	title := request.FormValue("MyTitle")
	body := request.FormValue("MyBody")

	//	创建map存储错误
	errors := make(map[string]string)

	//	验证标题
	if title == "" {
		errors["title"] = "标题不能为空"
	} else if utf8.RuneCountInString(title) < 2 || utf8.RuneCountInString(title) > 40 {
		errors["title"] = " 标题内容要大于两个字符,且小于40个字符"
	}

	//	验证内容
	if body == "" {
		errors["body"] = "内容不能为空"
	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容长度要大于10个字符"
	}

	//	检查是否有错误
	if len(errors) == 0 {
		fmt.Fprintf(writer, "验证通过!<br><br>")
		fmt.Fprintf(writer, "title的值为:%v<br>", title)
		fmt.Fprintf(writer, "title的长度为:%v<br><br>", utf8.RuneCountInString(title))

		fmt.Fprintf(writer, "body的值为%v<br>", body)
		fmt.Fprintf(writer, "body的长度为%v<br>", utf8.RuneCountInString(body))

		fmt.Println(1)
	} else {
		fmt.Println(2)

		//	通过路由name获取提交后的URL
		storeURL, _ := router.Get("articles.store").URL()

		data := ArticlesFormData{
			Title:  title,
			Body:   body,
			URL:    storeURL,
			Errors: errors,
		}

		tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
		if err != nil {
			//panic(err)
			fmt.Fprintf(writer, err.Error())
			fmt.Fprintf(writer, "<br>")
		}

		err = tmpl.Execute(writer, data)
		if err != nil {
			//panic(err)
			fmt.Fprintf(writer, err.Error())
		}
	}
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
