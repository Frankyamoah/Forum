package main

import (
	"database/sql"
	"fmt"
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

	tpl = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
	http.HandleFunc("/register", register)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/newpost", newPost)
	http.HandleFunc("/viewpost", viewPost)

	staticFileServer := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static", staticFileServer))

	fmt.Println("Listening on :8080...")
	http.ListenAndServe(":8080", nil)
}
