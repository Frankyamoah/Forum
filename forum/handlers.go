package forum

import (
	"fmt"
	"net/http"

	"log"

	_ "github.com/mattn/go-sqlite3"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var allUsernames []string
	var allPasswords []string
	var allEmails []string
	var allIds []int64

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		allUserInfo, err := SelectInfo("users", []string{"*"}, "")
		if err != nil {
			log.Fatal(err, "error retrieving info from table")
		}

		for _, mp := range allUserInfo {
			for key, value := range mp {
				if key == "username" {
					allUsernames = append(allUsernames, value.(string))
				}
				if key == "password" {
					allPasswords = append(allPasswords, value.(string))
				}
				if key == "email" {
					allEmails = append(allEmails, value.(string))
				}
				if key == "id" {
					allIds = append(allIds, value.(int64))
				}
			}
		}

		var loggedIn bool
		for i, u := range allUsernames {
			if username == u && password == allPasswords[i] {
				fmt.Println("username and password correct")
				loggedIn = true
				http.Redirect(w, r, "/static/leg.html", http.StatusSeeOther)
				return
			}
		}
		if !loggedIn {
			fmt.Fprintf(w, "Invalid username or password")
		}
	} else {
		http.NotFound(w, r)
	}
}
