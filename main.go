package main

import (
	"fmt"
	"forumProject/forum"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "forum/static/index.html")
}

func main() {

	path := "forum/static/"
	fs := http.FileServer(http.Dir(path))
	http.Handle("/forum/static/", http.StripPrefix("/forum/static/", fs))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", forum.LoginHandler)
	http.HandleFunc("/dashboard", forum.PostHandler)
	http.HandleFunc("/post", forum.PostHandler)

	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}
