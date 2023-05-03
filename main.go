package main

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var tpl *template.Template

func init() {
	var err error
	db, err = sql.Open("sqlite3", "forum.db")
	if err != nil {
		panic(err)
	}

	tpl = template.New("").Funcs(template.FuncMap{
		"GetLikeCount":    GetLikeCount,
		"GetDislikeCount": GetDislikeCount,
		"sub":             func(a, b int) int { return a - b },
		"abs":             func(a int) int { return int(math.Abs(float64(a))) },
	})

	tpl, err = tpl.ParseGlob("templates/*.html")
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
	http.HandleFunc("/register", register)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/newpost", newPost)
	http.HandleFunc("/viewpost", viewPost)
	http.HandleFunc("/addcomment", addComment)
	http.HandleFunc("/like", like)
	http.HandleFunc("/dislike", dislike)

	staticFileServer := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static", staticFileServer))

	fmt.Println("Listening on :8080...")
	http.ListenAndServe(":8080", nil)
}
