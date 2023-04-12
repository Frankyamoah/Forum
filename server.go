package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/post", postHandler)
	http.ListenAndServe(":1000", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("html/index.html")
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Check username and password
		if username == "myusername" && password == "mypassword" {
			http.Redirect(w, r, "/post", http.StatusSeeOther)
		} else {
			// If the credentials are invalid, display an error message
			fmt.Fprintf(w, "Invalid username or password")
		}
	} else {
		t, err := template.ParseFiles("html/login.html")
		if err != nil {
			fmt.Println(err)
			return
		}
		t.Execute(w, nil)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Get post data
		post := r.FormValue("post")

		// Save post to database
		db, err := sql.Open("sqlite3", "database/forumDB.sqlite")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer db.Close()

		stmt, err := db.Prepare("INSERT INTO posts(content) VALUES(?);")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer stmt.Close()

		res, err := stmt.Exec(post)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(res.LastInsertId())

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		t, err := template.ParseFiles("html/post.html")
		if err != nil {
			fmt.Println(err)
			return
		}
		t.Execute(w, nil)
	}
}
