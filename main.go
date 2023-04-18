package main

import (
	"fmt"
	"forumProject/forum"
	"net/http"
)

func main() {
	http.HandleFunc("/", forum.LoginHandler)
	http.HandleFunc("/dashboard", forum.PostHandler)
	http.HandleFunc("/post", forum.PostHandler)

	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)

}

// func dashboardHandler(w http.ResponseWriter, r *http.Request) {
// 	http.ServeFile(w, r, "forum/static/dashboard.html")
// }

// func postHandler(w http.ResponseWriter, r *http.Request) {
// 	http.ServeFile(w, r, "forum/static/dashboard.html")

// }
