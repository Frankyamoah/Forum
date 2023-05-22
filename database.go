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
	Email    string
}

type Post struct {
	ID        int
	Title     string
	Content   string
	CreatedAt time.Time
	Author    User
	Category  []string
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
	stmt := `SELECT p.id, p.title, p.content, p.created_at, u.ID, u.Username, GROUP_CONCAT(c.name) as categories 
		FROM posts p 
		JOIN users u ON p.author_id = u.ID 
		LEFT JOIN posts_categories pc ON p.id = pc.post_id
		LEFT JOIN categories c ON pc.category_id = c.id
		GROUP BY p.id 
		ORDER BY p.created_at DESC`
	rows, err := db.Query(stmt)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		var categories string
		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.Author.ID, &p.Author.Username, &categories)
		if err != nil {
			panic(err)
		}
		p.Category = strings.Split(categories, ",")
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
	// Check if the user already exists
	row := db.QueryRow("SELECT username, email FROM users WHERE Username = ? OR Email = ?", username, email)
	var existingUser string
	var existingEmail string
	err := row.Scan(&existingUser, &existingEmail)
	if err == nil {
		return fmt.Errorf("username or email already exists")
	} else if err != sql.ErrNoRows {
		return err
	}

	_, err = db.Exec("INSERT INTO users (Username, Password, Email) VALUES (?, ?, ?)", strings.ToLower(username), password, strings.ToLower(email))
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
func createPost(title string, content string, authorID int, categories []string) int {
	// Step 1: Insert the post into the 'posts' table
	query := "INSERT INTO posts (title, content, author_id, created_at) VALUES (?, ?, ?, ?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.Exec(title, content, authorID, time.Now())
	if err != nil {
		log.Fatal(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	postID := int(id)

	// Step 2: Insert the post-category relations into the 'posts_categories' table
	err = postCategoryRelation(postID, categories)
	if err != nil {
		log.Fatal(err)
	}

	return postID
}

// postCategoryRelation populates the posts_categories table based on post ID and categories.
func postCategoryRelation(postID int, categories []string) error {
	stmt, err := db.Prepare("INSERT INTO posts_categories (post_id, category_id) VALUES (?, (SELECT id FROM categories WHERE name = ?))")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()
	for _, category := range categories {
		capitalizedCategory := strings.Title(category)
		_, err = stmt.Exec(postID, capitalizedCategory)
		if err != nil {
			return fmt.Errorf("error inserting category: %w", err)
		}
	}

	return nil
}

func getPost(postID string) (Post, error) {
	row := db.QueryRow(`SELECT p.id, p.title, p.content, p.created_at, u.id, u.username
	FROM posts p JOIN users u ON p.author_id = u.id
	WHERE p.id = ?`, postID)

	var p Post
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.Author.ID, &p.Author.Username)
	if err != nil {
		fmt.Print("error scanning datbase into struct")
		return Post{}, err
	}

	categories, err := getPostCategories(p.ID)
	if err != nil {
		fmt.Print("error getting post categories")
		return Post{}, err
	}
	p.Category = categories

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
	query := `SELECT p.id, p.title, p.content, p.created_at, u.ID, u.Username, GROUP_CONCAT(pc.category)
			  FROM posts p
			  JOIN users u ON p.author_id = u.ID
			  LEFT JOIN posts_categories pc ON p.id = pc.post_id
			  JOIN categories c ON pc.category_id = c.id
			  WHERE c.name IN `

	placeholders := "(" + strings.Trim(strings.Repeat("?,", len(categories)), ",") + ")"
	query += placeholders
	query += " GROUP BY p.id"

	args := make([]interface{}, len(categories))
	for i, category := range categories {
		args[i] = category
	}
	log.Printf("Query: %v", query)
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error while executing query: %v", err)
		return nil
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		var categories string
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.Author.ID, &p.Author.Username, &categories); err != nil {
			log.Printf("Error while scanning rows: %v", err)
			continue
		}
		p.Category = strings.Split(categories, ",")
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error with rows: %v", err)
	}

	return posts
}

func toggleLikePost(userID, postID int) error {
	_, err := db.Exec(`INSERT INTO post_likes (user_id, post_id, liked, disliked) VALUES (?, ?, 1, 0) ON DUPLICATE KEY UPDATE liked = NOT liked, disliked = 0`, userID, postID)
	if err != nil {
		log.Printf("Error toggling like: %v", err)
		return err
	}
	return nil
}

func toggleDislikePost(userID, postID int) error {
	_, err := db.Exec(`INSERT INTO post_likes (user_id, post_id, liked, disliked) VALUES (?, ?, 0, 1) ON DUPLICATE KEY UPDATE disliked = NOT disliked, liked = 0`, userID, postID)
	if err != nil {
		log.Printf("Error toggling dislike: %v", err)
		return err
	}
	return nil
}

func toggleLikeComment(userID, commentID int) error {
	_, err := db.Exec(`INSERT INTO comment_likes (user_id, comment_id, liked, disliked) VALUES (?, ?, 1, 0) ON DUPLICATE KEY UPDATE liked = NOT liked, disliked = 0`, userID, commentID)
	if err != nil {
		log.Printf("Error toggling like: %v", err)
		return err
	}
	return nil
}

func toggleDislikeComment(userID, commentID int) error {
	_, err := db.Exec(`INSERT INTO comment_likes (user_id, comment_id, liked, disliked) VALUES (?, ?, 0, 1) ON DUPLICATE KEY UPDATE liked = 0, disliked = NOT disliked`, userID, commentID)
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
func addPostCategories(postID int, categories []string) {
	// Prepare the statement for fetching category ID
	catIDStmt, err := db.Prepare("SELECT id FROM categories WHERE name = ?")
	if err != nil {
		log.Fatal(err)
	}

	// Prepare the statement for inserting post-category relation
	stmt, err := db.Prepare("INSERT INTO posts_categories (post_id, category_id) VALUES (?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	for _, categoryName := range categories {
		// Fetch the category ID
		var categoryID int
		capitalizedCategory := strings.Title(categoryName) // Capitalize the category name
		err = catIDStmt.QueryRow(capitalizedCategory).Scan(&categoryID)
		if err != nil {
			log.Fatal(err)
		}

		// Insert the post-category relation
		_, err = stmt.Exec(postID, categoryID)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// This function will fetch categories for a specific post
func getPostCategories(postID int) ([]string, error) {
	rows, err := db.Query("SELECT c.name FROM categories c JOIN posts_categories pc ON c.id = pc.category_id WHERE pc.post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		err = rows.Scan(&category)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}
