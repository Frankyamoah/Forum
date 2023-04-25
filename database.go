package main

import (
	"errors"
	"time"

	"github.com/gorilla/sessions"
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
}

// ... other code ...

func getAllPosts() []Post {
	rows, err := db.Query(`SELECT p.id, p.title, p.content, p.created_at, u.id, u.username
		FROM posts p JOIN users u ON p.author_id = u.id
		ORDER BY p.created_at DESC`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.Author.ID, &p.Author.Username)
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

func registerUser(username, password string) error {
	_, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, password)
	return err
}

func getUsernameFromSession(session *sessions.Session) string {
	userID, loggedIn := session.Values["user_id"].(int)
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

func createPost(userID int, title, content string) error {
	_, err := db.Exec("INSERT INTO posts (title, content, author_id, created_at) VALUES (?, ?, ?, ?)", title, content, userID, time.Now())
	return err
}

func getPost(postID string) (Post, error) {
	row := db.QueryRow(`SELECT p.id, p.title, p.content, p.created_at, u.id, u.username
		FROM posts p JOIN users u ON p.author_id = u.id
		WHERE p.id = ?`, postID)

	var p Post
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.Author.ID, &p.Author.Username)
	if err != nil {
		return Post{}, err
	}

	return p, nil
}
