package main

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("SFSforum.01"))

// index handles the forum's main page.
func index(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "forum-session")
	posts := getAllPosts()
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
			session, _ := store.Get(r, "forum-session")
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
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}

	tpl.ExecuteTemplate(w, "register.html", nil)
}

// logout handles user logout.
func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "forum-session")
	delete(session.Values, "user_id")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// newPost handles creating a new forum post.
func newPost(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "forum-session")
	userID, loggedIn := session.Values["user_id"].(int)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		createPost(userID, title, content)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "newpost.html", nil)
}

// viewPost handles displaying a single forum post.
func viewPost(w http.ResponseWriter, r *http.Request) {
	postID := r.FormValue("id")
	post, err := getPost(postID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	tpl.ExecuteTemplate(w, "viewpost.html", post)
}
