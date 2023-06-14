package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
)

var (
	hashKey  = []byte("01SFSforum")
	blockKey = []byte("vienneseBiscuits")

	store = securecookie.New(hashKey, blockKey)
)
var once sync.Once

//var store = sessions.NewCookieStore([]byte("ForumProject"))

// login handles user login.
// var (
// 	sessionMap      = make(map[int]string)
// 	sessionMapMutex = &sync.Mutex{}
// )

func index(w http.ResponseWriter, r *http.Request) {
	// Validate the session first.
	session := make(map[string]interface{})
	if cookie, err := r.Cookie("forum-session"); err == nil {
		if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
			log.Printf("Error decoding session: %v", err)
		}
	}

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
			MaxAge:   300 * 60, // 5 mins
		}
		http.SetCookie(w, cookie)
	}

	categoryFilters := r.URL.Query()["category[]"]
	var posts []Post
	if len(categoryFilters) == 0 {
		posts = getAllPosts()
	} else {
		posts = getPostsByCategory(categoryFilters)
	}
	for _, post := range posts {
		post.Likes = GetLikeCount(post.ID, "post")
		post.Dislikes = GetDislikeCount(post.ID, "post")
	}

	user, err := getUserFromSession(session)
	if err != nil {
		// Handle the error
		once.Do(func() {
			fmt.Print("User Not Logged in or needs to Register ")
		})
	}

	data := struct {
		Username string
		Posts    []Post
	}{
		Username: user.Username,
		Posts:    posts,
	}
	tpl.ExecuteTemplate(w, "index.html", data)
}

func generateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid, nil
}

func validateSession(w http.ResponseWriter, r *http.Request) (User, map[string]interface{}) {
	session := make(map[string]interface{})
	cookie, err := r.Cookie("forum-session")
	if err == nil {
		if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
			log.Printf("Error decoding session: %v", err)
		}
	}

	// Print session data for troubleshooting
	fmt.Println("Session Data Of User Logged Out:", session)

	_, loggedIn := session["user_id"].(int)
	sessionID, ok := session["session_id"].(string)

	if loggedIn && ok {
		user, err := getUserFromSession(session)
		if err != nil {
			log.Printf("Error getting user: %v", err)
		} else if user.SessionID == sessionID {
			// the session is valid
			return user, session
		}
	}

	// Remove session data
	session = make(map[string]interface{})
	encodedSession, err := store.Encode("forum-session", session)
	if err != nil {
		log.Printf("Error encoding session: %v", err)
	} else {
		cookie := &http.Cookie{
			Name:     "forum-session",
			Value:    encodedSession,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1, // Delete the cookie
		}
		http.SetCookie(w, cookie)
	}

	// Redirect the user back to the index page as a logged out user
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return User{}, session
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		userID, err := authenticateUser(username, password)
		if err == nil {
			fmt.Println("Successful login") // Print statement for successful login
			session := make(map[string]interface{})
			if cookie, err := r.Cookie("forum-session"); err == nil {
				if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
					log.Printf("Error decoding session: %v", err)
				}
			}
			// Generate a new session ID
			sessionID, err := generateUUID()
			if err != nil {
				// Handle error
				log.Printf("Error generating session ID: %v", err)
			} else {
				session["session_id"] = sessionID
				// Update session ID in the database
				err = updateUserSessionID(userID, sessionID)
				if err != nil {
					// Handle error
					log.Printf("Error updating session ID: %v", err)
				} else {
					fmt.Println("Session ID updated") // Print statement for session ID update
				}
			}

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
					MaxAge:   300 * 60, // 5 mins
				}
				http.SetCookie(w, cookie)
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		} else {
			fmt.Println("Unsuccessful login") // Print statement for unsuccessful login
		}
	}
	tpl.ExecuteTemplate(w, "login.html", nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		email := r.FormValue("email")

		// Check if the form inputs are not empty
		if username == "" || password == "" || email == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		err := registerUser(username, password, email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			session := make(map[string]interface{})
			if cookie, err := r.Cookie("forum-session"); err == nil {
				if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
					log.Printf("Error decoding session: %v", err)
				}
			}

			// Update the session's last activity time
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
					MaxAge:   300 * 60, // 5 mins
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
	user, _ := validateSession(w, r) // Second return value (session) is ignored with _

	if user.ID == 0 {
		// User is not logged in, we've already handled redirect in validateSession
		return
	}

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		categories := r.Form["category[]"]

		createPost(title, content, user.ID, categories)
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

	categories, err := getPostCategories(post.ID)
	if err != nil {
		log.Printf("Error fetching post categories: %v", err)
		http.Error(w, "Error fetching post categories", http.StatusInternalServerError)
		return
	}
	post.Category = categories

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

	user, _ := validateSession(w, r) // Second return value (session) is ignored with _

	if user.ID == 0 {
		// User is not logged in, we've already handled redirect in validateSession
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
	err = createComment(user.ID, postID, content)
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
	// Validate the session first
	user, _ := validateSession(w, r) // Second return value (session) is ignored with _

	if user.ID == 0 {
		// User is not logged in, we've already handled redirect in validateSession
		return
	}

	handleLikeOrDislike(w, r, false)
}

func handleLikeOrDislike(w http.ResponseWriter, r *http.Request, isLike bool) {
	// Validate the session first
	user, _ := validateSession(w, r) // Second return value (session) is ignored with _

	if user.ID == 0 {
		// User is not logged in, we've already handled redirect in validateSession
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
			err = toggleLikePost(user.ID, postID)
		} else {
			err = toggleDislikePost(user.ID, postID)
		}
	} else { // Liking or disliking a comment
		if isLike {
			err = toggleLikeComment(user.ID, commentID)
		} else {
			err = toggleDislikeComment(user.ID, commentID)
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

func filterHandler(w http.ResponseWriter, r *http.Request) {
	session := make(map[string]interface{})
	if cookie, err := r.Cookie("forum-session"); err == nil {
		if err := store.Decode("forum-session", cookie.Value, &session); err != nil {
			log.Printf("Error decoding session: %v", err)
		}
	}

	categories, ok := r.URL.Query()["category[]"]
	if !ok {
		log.Println("No categories found in the form data")
		// Handle the error or set a default behavior
		// For example, you can redirect to the home page or show an error message.
	}

	posts := getPostsByCategory(categories)

	for i, post := range posts {
		posts[i].Likes = GetLikeCount(post.ID, "post")
		posts[i].Dislikes = GetDislikeCount(post.ID, "post")
	}

	user, err := getUserFromSession(session)
	if err != nil {
		log.Printf("Error getting user from session: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := struct {
		Username string
		Posts    []Post
	}{
		Username: user.Username,
		Posts:    posts,
	}

	err = tpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Printf("Error while executing template: %v", err)
	}
}

func joinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// getLikedPosts handles getting all posts liked by the user.
func getLikedPosts(w http.ResponseWriter, r *http.Request) {
	// Validate the session first
	user, _ := validateSession(w, r) // Second return value (session) is ignored with _

	if user.ID == 0 {
		// User is not logged in, we've already handled redirect in validateSession
		return
	}
	posts, err := getPostsLikedByUser(user.ID)
	if err != nil {
		log.Printf("Error getting liked posts: %v", err)
		http.Error(w, "Error getting liked posts", http.StatusInternalServerError)
		return
	}

	var likedPosts []struct {
		Post Post
	}
	for _, post := range posts {
		likedPosts = append(likedPosts, struct {
			Post Post
		}{Post: post})
	}

	data := struct {
		LikedPosts []struct {
			Post Post
		}
	}{
		LikedPosts: likedPosts,
	}

	tpl.ExecuteTemplate(w, "likedposts.html", data)
}

// getCreatedPosts handles getting all posts created by the user.
func getCreatedPosts(w http.ResponseWriter, r *http.Request) {
	// Validate the session first
	user, _ := validateSession(w, r) // Second return value (session) is ignored with _

	if user.ID == 0 {
		// User is not logged in, we've already handled redirect in validateSession
		return
	}

	posts, err := getPostsCreatedByUser(user.ID)
	if err != nil {
		log.Printf("Error getting created posts: %v", err)
		http.Error(w, "Error getting created posts", http.StatusInternalServerError)
		return
	}

	var createdPosts []struct {
		Post Post
	}
	for _, post := range posts {
		createdPosts = append(createdPosts, struct {
			Post Post
		}{Post: post})
	}

	data := struct {
		CreatedPosts []struct {
			Post Post
		}
	}{
		CreatedPosts: createdPosts,
	}

	tpl.ExecuteTemplate(w, "createdposts.html", data)
}
