package forum

import (
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Post struct {
	PostTitle    string
	UserId       int64
	PostCategory int64
	PostContent  string
}

func populatePost() []Post {

	postdata, err := SelectInfo("posts", []string{"*"}, "")
	if err != nil {
		log.Fatal(err, "error retrieving post postdata")
	}

	posts := make([]Post, 0, len(postdata))

	for _, mp := range postdata {
		post := Post{
			PostTitle:    mp["title"].(string),
			UserId:       mp["user_id"].(int64),
			PostCategory: mp["category_id"].(int64),
			PostContent:  mp["content"].(string),
		}
		posts = append(posts, post)
	}

	return posts
}
func main() {
	fmt.Println(populatePost(), "h")
}

// func Handler(w http.ResponseWriter, r *http.Request) {
// 	posts := populatePost()
// 	//	fmt.Println(posts)

// 	tmpl, err := template.ParseFiles("forum/static/dashboard.html")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	err = tmpl.Execute(w, posts)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }
