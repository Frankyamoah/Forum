package main

import (
	"fmt"
	"forumProject/forum"
	"net/http"
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", forum.LoginHandler)

	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "forum/static/index.html")
}
