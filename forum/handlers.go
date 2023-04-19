package forum

import (
	"fmt"
	"html/template"
	"net/http"

	"log"

	_ "github.com/mattn/go-sqlite3"
)

type PostData struct {
	PostTitle      string
	PostContent    string
	PostCategory   string
	UserName       string
	CommentContent []CommentData
}
type CommentData struct {
	CommentContent string
	// Add other fields if needed, e.g., User ID, Timestamp, etc.
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var allUsernames []string
	var allPasswords []string

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			fmt.Fprintf(w, "Please enter a username and password")
			return
		}

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
	} else if r.Method == "GET" {
		http.ServeFile(w, r, "forum/static/index.html")
	} else {
		http.NotFound(w, r)
	}
}

// func PostHandler(w http.ResponseWriter, r *http.Request) {
// 	wd, err := os.Getwd()
// 	if err != nil {
// 		fmt.Println("Error getting working directory:", err)
// 		return
// 	}

// 	dashboardFilePath := filepath.Join(wd, "forum", "static", "dashboard.html")
// 	file, err := os.Open(dashboardFilePath) // Change this line
// 	if err != nil {
// 		fmt.Println("Error opening dashboard.html:", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	defer file.Close()

// 	var allTitles []string
// 	var allContent []string
// 	var allCategoryNames []string
// 	var allUserNames []string
// 	var allComments []string
// 	postInfo, err := SelectInfo("posts", []string{"*"}, "")
// 	if err != nil {
// 		log.Fatal(err, "error retrieving posts data")
// 	}
// 	categoryInfo, err := SelectInfo("categories", []string{"*"}, "")
// 	if err != nil {
// 		log.Fatal(err, "error retrieving category data")
// 	}
// 	userInfo, err := SelectInfo("users", []string{"*"}, "")
// 	if err != nil {
// 		log.Fatal(err, "error retrieving user data")
// 	}
// 	commentsInfo, err := SelectInfo("comments", []string{"*"}, "")
// 	if err != nil {
// 		log.Fatal(err, "error retrieving comment data")
// 	}
// 	for _, mp := range commentsInfo {
// 		for key, value := range mp {
// 			if key == "content" {
// 				allComments = append(allComments, value.(string))
// 			}
// 		}
// 	}
// 	for _, mp := range categoryInfo {
// 		for key, value := range mp {
// 			if key == "name" {
// 				allCategoryNames = append(allCategoryNames, value.(string))
// 			}
// 		}
// 	}
// 	//fmt.Println(allCategoryNames)

// 	for _, mp := range userInfo {
// 		for key, value := range mp {
// 			if key == "username" {
// 				allUserNames = append(allUserNames, value.(string))
// 			}
// 		}
// 	}
// 	//fmt.Println(allUserNames)

// 	for _, mp := range postInfo {
// 		for key, value := range mp {
// 			if key == "title" {
// 				allTitles = append(allTitles, value.(string))
// 			}
// 			if key == "content" {
// 				allContent = append(allContent, value.(string))
// 			}

// 		}
// 	}

// 	posts := make([]PostData, len(postInfo))

// 	for i := 0; i < len(postInfo); i++ {
// 		postID := int(postInfo[i]["id"].(int64))
// 		posts[i] = PostData{
// 			PostTitle:      allTitles[i],
// 			PostContent:    allContent[i],
// 			PostCategory:   allCategoryNames[i],
// 			UserName:       allUserNames[i],
// 			CommentContent: filterComments(postID, commentsInfo),
// 		}
// 		//fmt.Println(posts[i], "POST")
// 	}

// 	tmpl, err := template.ParseFiles(dashboardFilePath)
// 	if err != nil {
// 		panic(err)
// 	}

// 	err = tmpl.ExecuteTemplate(w, "post.html", map[string]interface{}{
// 		"Posts": posts,
// 	})
// 	if err != nil {
// 		return
// 	}

// }
func filterComments(postID int, allComments []map[string]interface{}) []CommentData {
	var comments []CommentData
	for _, comment := range allComments {
		if postIDValue, ok := comment["post_id"].(int64); ok && postIDValue == int64(postID) {
			comments = append(comments, CommentData{
				CommentContent: comment["content"].(string),
				// Populate other fields if needed
			})
		}
	}
	return comments
}
func PostHandler(w http.ResponseWriter, r *http.Request) {

	posts := populatePost()
	//	fmt.Println(posts)

	tmpl, err := template.ParseFiles("forum/static/dashboard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
