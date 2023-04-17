package forum

import (
	"fmt"
	"net/http"
	"time"

	"log"

	_ "github.com/mattn/go-sqlite3"
)


type PostData struct {
	PostTitle    string
	PostContent  string
	PostCategory string
}


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
			log.Fatal(err, "error retrieving all users")
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
				http.Redirect(w, r, "forum/static/dashboard.html", http.StatusSeeOther)
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

func PostHandler(w http.ResponseWriter, r *http.Request) {
	var allTitles []string
	var allContent []string
	var allcategory []int
	var allUserId []int
	var allTimeStamp []time.Time
	allPosts, err := SelectInfo("posts", []string{"*"}, "")
	if err != nil {
		log.Fatal(err, "error retrieving all posts")
	}

	func(w http.ResponseWriter, r *http.Request) {
		data := PostData{
			PostTitle:    "My Post Title",
			PostContent:  "This is the content of my post.",
			PostCategory: "Technology",
		}

		tmpl, err := template.ParseFiles("index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	for _, mp := range allPosts {
		for key, value := range mp {
			if key == "title" {
				allTitles = append(allTitles, value.(string))
			}
			if key == "content" {
				allContent = append(allContent, value.(string))
			}
			if key == "category_id" {
				allcategory = append(allcategory, value.(int))
			}
			if key == "user_id" {
				allUserId = append(allUserId, value.(int))
			}
			if key == "create_at" {
				allTimeStamp = append(allTimeStamp, value.(time.Time))
			}
		}
	}
}
