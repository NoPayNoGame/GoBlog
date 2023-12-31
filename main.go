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
	router.HandleFunc("/about", aboutHandler).Methods("GET") //.Name("about")

	//	get指定id返回对应内容
	router.HandleFunc("/articles/{id:[0-9]+}", articlesShowHandler).Methods("GET").Name("articles.show")
	//	get方法
	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	//	POST方法
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")

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

/*
POST 方法,发布内容 存入数据库
*/
func articlesStoreHandler(writer http.ResponseWriter, request *http.Request) {
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
		toDB, err := saveArticlesToDB(title, body)
		if toDB > 0 {
			fmt.Fprintf(writer, "插入成功,ID为:%d", toDB)
		} else {
			checkError(err)
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(writer, "500 服务器内部错误!")
		}

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

/*
传入	文章标题和内容

	数据插入数据库

返回	插入行数,错误
*/
func saveArticlesToDB(title string, body string) (int64, error) {
	//	变量初始化
	var (
		id   int64
		err  error
		rs   sql.Result
		stmt *sql.Stmt
	)

	//	获取一个prepare 声明语句
	stmt, err = db.Prepare("INSERT INTO articles (title,body)values (?,?)")
	//	例行错误检测
	if err != nil {
		return 0, err
	}

	// 在此函数运行结束后关闭此语句,防止占用 SQL 链接
	defer stmt.Close()

	// 执行请求,传参进入绑定的内容
	rs, err = stmt.Exec(title, body)
	if err != nil {
		return 0, err
	}

	//	插入成功的话 会返回自增ID
	id, err = rs.LastInsertId()
	if err != nil {
		return 0, err
	}

	if id > 0 {
		return id, nil
	}

	return 0, err
}

func articlesIndexHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "GET访问文章列表")
}

//	必须大写,小写前端无法读取

type Article struct {
	Body, Title string
	ID          int64
}

func articlesShowHandler(writer http.ResponseWriter, request *http.Request) {
	//	1.获取 URL 参数(获取请求的ID)
	vars := mux.Vars(request)
	id := vars["id"]

	//	2.读取对应的文章数据
	article := Article{}
	query := "Select * from articles WHERE  id = ?"

	//	QueryRow执行一次查询，并期望返回最多一行结果（即Row）。QueryRow总是返回非nil的值，直到返回值的Scan方法被调用时，才会返回被延迟的错误。（如：未找到结果）
	err := db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)
	//	3.如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
			//	3.1 数据未找到
			writer.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(writer, "404 文章未找到")
		} else {
			//	3.2 数据库错误
			checkError(err)
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(writer, "500 服务器内部错误")
		}
	} else {
		//	4 读取成功 显示文章
		//fmt.Fprintf(writer, "读取成功<br>文章标题: "+article.title)
		//fmt.Fprintf(writer, "<br>文章内容: "+article.body)
		tmpl, err := template.ParseFiles("resources/views/articles/show.gohtml")
		if err != nil {
			checkError(err)
		}

		err = tmpl.Execute(writer, article)
		checkError(err)
	}
}

func aboutHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "此博客是用以记录编程笔记，如您有反馈或建议，请联系 <a href=\"mailto:summer@example.com\">summer@example.com</a>")
}

func homeHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprint(writer, "<h1>Hello, 欢迎来到 goblog！</h1>")
}
