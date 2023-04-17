package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type PostData struct {
	PostTitle    string
	PostContent  string
	PostCategory string
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/dashbaord", postHandler)

	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "forum/static/index.html")
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "forum/static/dashboard.html")
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	data := PostData{
		PostTitle:    "My Post Title",
		PostContent:  "This is the content of my post.",
		PostCategory: "Technology",
	}

	tmpl, err := template.ParseFiles("dashboard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
