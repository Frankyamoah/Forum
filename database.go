package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type User struct {
	ID       int
	Username string
}

type Post struct {
	ID        int
	Title     string
	Content   string
	CreatedAt time.Time
	Author    User
	Category  string
	Likes     int
	Dislikes  int
}

type Comment struct {
	ID        int
	Content   string
	CreatedAt time.Time
	Author    User
	PostID    int
	Likes     int
	Dislikes  int
}

func getAllPosts() []Post {
	rows, err := db.Query(`SELECT p.id, p.title, p.content, p.created_at, u.id, u.username, p.category
		FROM posts p JOIN users u ON p.author_id = u.id
		ORDER BY p.created_at DESC`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.Author.ID, &p.Author.Username, &p.Category)
		if err != nil {
			panic(err)
		}
		posts = append(posts, p)
	}

	return posts
}

func authenticateUser(username, password string) (int, error) {
	var id int
	var hashedPassword string
	err := db.QueryRow("SELECT id, password FROM users WHERE username = ?", username).Scan(&id, &hashedPassword)
	if err != nil {
		return 0, err
	}

	if password == hashedPassword {
		return id, nil
	}
	return 0, errors.New("invalid password")
}

func registerUser(username, password, email string) error {
	_, err := db.Exec("INSERT INTO users (username, password, email) VALUES (?, ?, ?)", username, password, email)
	return err
}

func getUsernameFromSession(session map[string]interface{}) string {
	userID, loggedIn := session["user_id"].(int)
	if !loggedIn {
		return ""
	}

	var username string
	err := db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		panic(err)
	}

	return username
}
func createPost(userID int, title, content string, categories []string) (int, error) {
	// Your code to create the post in the database

	// Execute the INSERT statement to create the post
	result, err := db.Exec("INSERT INTO posts (title, content, author_id, created_at) VALUES (?, ?, ?, ?)",
		title, content, userID, time.Now())
	if err != nil {
		return 0, fmt.Errorf("error creating post: %w", err)
	}

	// Get the post ID of the newly inserted post
	postID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting post ID: %w", err)
	}

	// Call the postCategoryRelation function to populate the post_categories table
	err = postCategoryRelation(int(postID), categories)
	if err != nil {
		log.Printf("Error inserting post category relations: %v", err)
		// Handle the error accordingly
		return 0, err
	}

	return int(postID), nil
}

// postCategoryRelation populates the post_categories table based on post ID and categories.
func postCategoryRelation(postID int, categories []string) error {
	stmt, err := db.Prepare("INSERT INTO posts_categories (post_id, category) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, category := range categories {
		_, err = stmt.Exec(postID, category)
		if err != nil {
			return fmt.Errorf("error inserting category: %w", err)
		}
	}

	return nil
}

func getPost(postID string) (Post, error) {
	row := db.QueryRow(`SELECT p.id, p.title, p.content, p.category, p.created_at, u.id, u.username
			FROM posts p JOIN users u ON p.author_id = u.id
			WHERE p.id = ?`, postID)

	var p Post
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.Category, &p.CreatedAt, &p.Author.ID, &p.Author.Username)
	if err != nil {
		return Post{}, err
	}

	return p, nil
}

func createComment(userID, postID int, content string) error {
	_, err := db.Exec("INSERT INTO comments (content, author_id, post_id, created_at) VALUES (?, ?, ?, ?)", content, userID, postID, time.Now())
	return err
}

func getPostComments(postID int) ([]Comment, error) {
	rows, err := db.Query(`SELECT c.id, c.content, c.created_at, u.id, u.username, c.post_id
			FROM comments c JOIN users u ON c.author_id = u.id
			WHERE c.post_id = ?
			ORDER BY c.created_at DESC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		err := rows.Scan(&c.ID, &c.Content, &c.CreatedAt, &c.Author.ID, &c.Author.Username, &c.PostID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}

func getPostsByCategory(categories []string) []Post {
	query := `SELECT p.id, p.title, p.content, p.created_at, u.id, u.username, p.category
			  FROM posts p JOIN users u ON p.author_id = u.id
			  WHERE p.category IN (?)`

	// Prepare the placeholders for the category values
	placeholders := strings.TrimRight(strings.Repeat("?,", len(categories)), ",")
	query = strings.Replace(query, "(?)", "("+placeholders+")", -1)

	// Create the arguments slice to pass to the query
	args := make([]interface{}, len(categories))
	for i, category := range categories {
		args[i] = category
		//fmt.Println(category, i, "args")
	}

	// Create a map to store posts and avoid duplicates
	postsMap := make(map[int]Post)

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.Author.ID, &p.Author.Username, &p.Category); err != nil {
			log.Fatal(err)
		}
		postsMap[p.ID] = p
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Convert the map of posts to a slice
	posts := make([]Post, 0, len(postsMap))
	for _, post := range postsMap {
		posts = append(posts, post)
	}

	return posts
}

func toggleLikePost(userID, postID int) error {
	_, err := db.Exec(`INSERT INTO post_likes (user_id, post_id, liked, disliked) VALUES (?, ?, true, false) ON CONFLICT (user_id, post_id) DO UPDATE SET liked = NOT post_likes.liked, disliked = false`, userID, postID)
	if err != nil {
		log.Printf("Error toggling like: %v", err)
		return err
	}
	return nil
}

func toggleDislikePost(userID, postID int) error {
	_, err := db.Exec(`INSERT INTO post_likes (user_id, post_id, liked, disliked) VALUES (?, ?, false, true) ON CONFLICT (user_id, post_id) DO UPDATE SET disliked = NOT post_likes.disliked, liked = false`, userID, postID)
	if err != nil {
		log.Printf("Error toggling dislike: %v", err)
		return err
	}
	return nil
}

func toggleLikeComment(userID, postID int) error {
	_, err := db.Exec(`INSERT INTO comment_likes (user_id, comment_id, liked, disliked) VALUES (?, ?, true, false) ON CONFLICT (user_id, comment_id) DO UPDATE SET liked = NOT comment_likes.liked, disliked = false`, userID, postID)
	if err != nil {
		log.Printf("Error toggling like: %v", err)
		return err
	}
	return nil
}

func toggleDislikeComment(userID, postID int) error {
	_, err := db.Exec(`INSERT INTO comment_likes (user_id, comment_id, liked, disliked) VALUES (?, ?, false, true) ON CONFLICT (user_id, comment_id) DO UPDATE SET disliked = NOT comment_likes.disliked, liked = false`, userID, postID)
	if err != nil {
		log.Printf("Error toggling dislike: %v", err)
		return err
	}
	return nil
}

func GetLikeCount(id int, contentType string) int {
	var count int
	if contentType == "post" {
		err := db.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND liked = 1", id).Scan(&count)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error fetching like count for post %d: %v", id, err)
		}
	} else if contentType == "comment" {
		err := db.QueryRow("SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND liked = 1", id).Scan(&count)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error fetching like count for comment %d: %v", id, err)
		}
	}
	return count
}

func GetDislikeCount(id int, contentType string) int {
	var count int
	if contentType == "post" {
		err := db.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND disliked = 1", id).Scan(&count)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error fetching dislike count for post %d: %v", id, err)
		}
	} else if contentType == "comment" {
		err := db.QueryRow("SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND disliked = 1", id).Scan(&count)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error fetching dislike count for comment %d: %v", id, err)
		}
	}
	return count
}

func getCommentByID(commentID int) (Comment, error) {
	row := db.QueryRow(`SELECT c.id, c.content, c.created_at, u.id, u.username, c.post_id
			FROM comments c JOIN users u ON c.author_id = u.id
			WHERE c.id = ?`, commentID)

	var c Comment
	err := row.Scan(&c.ID, &c.Content, &c.CreatedAt, &c.Author.ID, &c.Author.Username, &c.PostID)
	if err != nil {
		return Comment{}, err
	}

	return c, nil
}

func getPostIDByCommentID(commentID int) (int, error) {
	comment, err := getCommentByID(commentID)
	if err != nil {
		return 0, err
	}
	return comment.PostID, nil
}
