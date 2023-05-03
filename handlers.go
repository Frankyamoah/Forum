package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("ForumProject"))

// index handles the forum's main page.
func index(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "forum-session")
	if err != nil {
		log.Fatal(err)
	}

	// set the session's max age to 2 minutes
	session.Options.MaxAge = 60 // in seconds

	// update the session's last activity time
	session.Values["last_activity"] = time.Now().Unix()
	session.Save(r, w)

	categoryFilter := r.URL.Query().Get("category")
	var posts []Post
	if categoryFilter == "" {
		posts = getAllPosts()
	} else {
		posts = getPostsByCategory(categoryFilter)
	}

	for _, post := range posts {
		post.Likes = GetLikeCount(post.ID, "post")
		post.Dislikes = GetDislikeCount(post.ID, "post")
	}

	data := struct {
		Username string
		Posts    []Post
	}{
		Username: getUsernameFromSession(session),
		Posts:    posts,
	}

	tpl.ExecuteTemplate(w, "index.html", data)
}

// login handles user login.
func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		userID, err := authenticateUser(username, password)
		if err == nil {
			session, err := store.Get(r, "forum-session")
			if err != nil {
				log.Fatal(err)
			}

			// set the session's max age to 2 minutes
			session.Options.MaxAge = 60 * 60 // in seconds

			// update the session's last activity time
			session.Values["last_activity"] = time.Now().Unix()
			session.Values["user_id"] = userID
			session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	tpl.ExecuteTemplate(w, "login.html", nil)
}

// register handles user registration.
func register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := registerUser(username, password)
		if err == nil {
			session, err := store.Get(r, "forum-session")
			if err != nil {
				log.Fatal(err)
			}

			// set the session's max age to 2 minutes
			session.Options.MaxAge = 60 * 60 // in seconds

			// update the session's last activity time
			session.Values["last_activity"] = time.Now().Unix()
			session.Save(r, w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}

	tpl.ExecuteTemplate(w, "register.html", nil)
}

// logout handles user logout.
func logout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "forum-session")
	if err != nil {
		log.Fatal(err)
	}

	// set the session's max age to 2 minutes
	session.Options.MaxAge = 60 * 60 // in seconds

	// update the session's last activity time
	session.Values["last_activity"] = time.Now().Unix()

	delete(session.Values, "user_id")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// newPost handles creating a new forum post.
func newPost(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "forum-session")
	if err != nil {
		log.Fatal(err)
	}

	// set the session's max age to 2 minutes
	session.Options.MaxAge = 60 * 60 // in seconds

	// update the session's last activity time
	session.Values["last_activity"] = time.Now().Unix()

	userID, loggedIn := session.Values["user_id"].(int)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		category := r.FormValue("category")          // Add this line
		createPost(userID, title, content, category) // Pass category to createPost
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "newpost.html", nil)
}

func viewPost(w http.ResponseWriter, r *http.Request) {
	postID := r.FormValue("id")
	post, err := getPost(postID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	comments, err := getPostComments(post.ID)
	if err != nil {
		log.Printf("Error fetching comments: %v", err)
		http.Error(w, "Error fetching comments", http.StatusInternalServerError)
		return
	}

	session, err := store.Get(r, "forum-session")
	if err != nil {
		log.Fatal(err)
	}

	session.Options.MaxAge = 60 * 60 // in seconds
	session.Values["last_activity"] = time.Now().Unix()
	session.Save(r, w)

	for _, comment := range comments {
		comment.Likes = GetLikeCount(comment.ID, "comment")
		comment.Dislikes = GetDislikeCount(comment.ID, "comment")
	}

	data := struct {
		Post     Post
		Comments []Comment
	}{
		Post:     post,
		Comments: comments,
	}

	// // Add this print statement to see if the function is being executed and check the data
	//fmt.Printf("ViewPost function called. Post: %+v, Comments: %+v\n", post, comments)

	err = tpl.ExecuteTemplate(w, "viewpost.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
	}
}

func addComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the current user ID from the session
	session, err := store.Get(r, "forum-session")
	if err != nil {
		log.Fatal(err)
	}

	// set the session's max age to 2 minutes
	session.Options.MaxAge = 60 * 60 // in seconds

	// update the session's last activity time
	session.Values["last_activity"] = time.Now().Unix()
	session.Save(r, w)

	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	// Get the post ID and comment content from the form
	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")

	// Create the comment using the createComment function
	err = createComment(userID, postID, content)
	if err != nil {
		http.Error(w, "Error creating comment", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post page
	http.Redirect(w, r, "/viewpost?id="+strconv.Itoa(postID), http.StatusSeeOther)
}

func like(w http.ResponseWriter, r *http.Request) {
	//log.Printf("Like/Dislike called, request: %+v\n", r)

	handleLikeOrDislike(w, r, true)
}

func dislike(w http.ResponseWriter, r *http.Request) {
	//log.Printf("Like/Dislike called, request: %+v\n", r)

	handleLikeOrDislike(w, r, false)
}

func handleLikeOrDislike(w http.ResponseWriter, r *http.Request, isLike bool) {
	session, err := store.Get(r, "forum-session")
	if err != nil {
		log.Fatal(err)
	}

	userID, loggedIn := session.Values["user_id"].(int)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	isPost := err == nil

	var id int
	if isPost {
		id = postID
	} else {
		id, err = strconv.Atoi(r.FormValue("comment_id"))
		if err != nil {
			http.Error(w, "Invalid post or comment ID", http.StatusBadRequest)
			return
		}
	}

	if isLike {
		if isPost {
			err = likePost(userID, id)
		} else {
			err = likeComment(userID, id)
		}
	} else {
		if isPost {
			err = dislikePost(userID, id)
		} else {
			err = dislikeComment(userID, id)
		}
	}

	if err != nil {
		http.Error(w, "Error processing like or dislike", http.StatusInternalServerError)
		return
	}

	redirectPath := "/"
	if !isPost {
		redirectPath = "/viewpost?id=" + r.FormValue("post_id")
	}
	http.Redirect(w, r, redirectPath, http.StatusSeeOther)
}
