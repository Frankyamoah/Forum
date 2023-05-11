package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/securecookie"
)

var (
	hashKey  = []byte("01SFSforum")
	blockKey = []byte("vienneseBiscuits")

	store = securecookie.New(hashKey, blockKey)
)

//var store = sessions.NewCookieStore([]byte("ForumProject"))

// index handles the forum's main page.
func index(w http.ResponseWriter, r *http.Request) {
	session := make(map[string]interface{})
	if cookie, err := r.Cookie("forum-session"); err == nil {
		if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
			log.Printf("Error decoding session: %v", err)
		}
	}

	// set the session's max age to 2 minutes
	//session.Options.MaxAge = 60 // in seconds

	// update the session's last activity time
	session["last_activity"] = time.Now().Unix()
	encodedSession, err := store.Encode("forum-session", session)
	if err != nil {
		log.Printf("Error encoding session: %v", err)
	} else {
		cookie := &http.Cookie{
			Name:     "forum-session",
			Value:    encodedSession,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   300 * 60, // 5 mins, equivalent to session.Options.MaxAge
		}
		http.SetCookie(w, cookie)
	}

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
			session := make(map[string]interface{})
			if cookie, err := r.Cookie("forum-session"); err == nil {
				if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
					log.Printf("Error decoding session: %v", err)
				}
			}

			// set the session's max age to 2 minutes
			// session.Options.MaxAge = 300 * 60 // i5 secsonds

			// update the session's last activity time
			session["last_activity"] = time.Now().Unix()
			session["user_id"] = userID
			encodedSession, err := store.Encode("forum-session", session)
			if err != nil {
				log.Printf("Error encoding session: %v", err)
			} else {
				cookie := &http.Cookie{
					Name:     "forum-session",
					Value:    encodedSession,
					Path:     "/",
					HttpOnly: true,
					MaxAge:   300 * 60, // 5 mins, equivalent to session.Options.MaxAge
				}
				http.SetCookie(w, cookie)
			}
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
			session := make(map[string]interface{})
			if cookie, err := r.Cookie("forum-session"); err == nil {
				if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
					log.Printf("Error decoding session: %v", err)
				}
			}

			// set the session's max age to 2 minutes
			// session.Options.MaxAge = 300 * 60 // i5 secsonds

			// update the session's last activity time
			session["last_activity"] = time.Now().Unix()
			encodedSession, err := store.Encode("forum-session", session)
			if err != nil {
				log.Printf("Error encoding session: %v", err)
			} else {
				cookie := &http.Cookie{
					Name:     "forum-session",
					Value:    encodedSession,
					Path:     "/",
					HttpOnly: true,
					MaxAge:   300 * 60, // 5 mins, equivalent to session.Options.MaxAge
				}
				http.SetCookie(w, cookie)
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}

	tpl.ExecuteTemplate(w, "register.html", nil)
}

// logout handles user logout.
func logout(w http.ResponseWriter, r *http.Request) {
	session := make(map[string]interface{})
	if cookie, err := r.Cookie("forum-session"); err == nil {
		if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
			log.Printf("Error decoding session: %v", err)
		}
	}

	// set the session's max age to 2 minutes
	// session.Options.MaxAge = 300 * 60 // i5 secsonds

	// update the session's last activity time
	session["last_activity"] = time.Now().Unix()

	delete(session, "user_id")

	encodedSession, err := store.Encode("forum-session", session)
	if err != nil {
		log.Printf("Error encoding session: %v", err)
	} else {
		cookie := &http.Cookie{
			Name:     "forum-session",
			Value:    encodedSession,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   300 * 60, // 5 mins, equivalent to session.Options.MaxAge
		}
		http.SetCookie(w, cookie)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// newPost handles creating a new forum post.
func newPost(w http.ResponseWriter, r *http.Request) {
	session := make(map[string]interface{})
	if cookie, err := r.Cookie("forum-session"); err == nil {
		if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
			log.Printf("Error decoding session: %v", err)
		}
	}

	// set the session's max age to 2 minutes
	// session.Options.MaxAge = 300 * 60 // i5 secsonds

	// update the session's last activity time
	session["last_activity"] = time.Now().Unix()

	userID, loggedIn := session["user_id"].(int)
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

	session := make(map[string]interface{})
	if cookie, err := r.Cookie("forum-session"); err == nil {
		if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
			log.Printf("Error decoding session: %v", err)
		}
	}

	// session.Options.MaxAge = 300 * 60 // i5 secsonds
	session["last_activity"] = time.Now().Unix()
	encodedSession, err := store.Encode("forum-session", session)
	if err != nil {
		log.Printf("Error encoding session: %v", err)
	} else {
		cookie := &http.Cookie{
			Name:     "forum-session",
			Value:    encodedSession,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   300 * 60, // 5 mins, equivalent to session.Options.MaxAge
		}
		http.SetCookie(w, cookie)
	}

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
	session := make(map[string]interface{})
	if cookie, err := r.Cookie("forum-session"); err == nil {
		if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
			log.Printf("Error decoding session: %v", err)
		}
	}

	// set the session's max age to 2 minutes
	// session.Options.MaxAge = 300 * 60 // i5 secsonds

	// update the session's last activity time
	session["last_activity"] = time.Now().Unix()
	encodedSession, err := store.Encode("forum-session", session)
	if err != nil {
		log.Printf("Error encoding session: %v", err)
	} else {
		cookie := &http.Cookie{
			Name:     "forum-session",
			Value:    encodedSession,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   300 * 60, // 5 mins, equivalent to session.Options.MaxAge
		}
		http.SetCookie(w, cookie)
	}

	userID, ok := session["user_id"].(int)
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
	session := make(map[string]interface{})
	if cookie, err := r.Cookie("forum-session"); err == nil {
		if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
			log.Printf("Error decoding session: %v", err)
		}
	}

	userID, loggedIn := session["user_id"].(int)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	postID, postErr := strconv.Atoi(r.FormValue("post_id"))
	commentID, commentErr := strconv.Atoi(r.FormValue("comment_id"))

	if postErr != nil && commentErr != nil {
		http.Error(w, "Invalid post or comment ID", http.StatusBadRequest)
		return
	}
	var err error
	if postErr == nil { // Liking or disliking a post
		if isLike {
			err = toggleLikePost(userID, postID)
		} else {
			err = toggleDislikePost(userID, postID)
		}
	} else { // Liking or disliking a comment
		if isLike {
			err = toggleLikeComment(userID, commentID)
		} else {
			err = toggleDislikeComment(userID, commentID)
		}
	}

	if err != nil {
		http.Error(w, "Error processing like or dislike", http.StatusInternalServerError)
		return
	}

	redirectPath := "/"
	if postErr != nil { // When it's a comment, redirect to the comment's post
		postID, err := getPostIDByCommentID(commentID)
		if err != nil {
			http.Error(w, "Error finding associated post for comment", http.StatusInternalServerError)
			return
		}
		redirectPath = "/viewpost?id=" + strconv.Itoa(postID)
	}

	http.Redirect(w, r, redirectPath, http.StatusSeeOther)
}
