package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open the database file.
	dataBase, err := sql.Open("sqlite3", "database/forumDB.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer dataBase.Close()

	//querying info from databasee
	rows, err := dataBase.Query("SELECT * FROM users;")
	if err != nil {
		log.Fatal(err, "error querying data")
	}

	//iteraing over each row
	for rows.Next() {
		var id int
		var email string
		var username string
		var password string
		//scanning result of each row into var and checking for err
		err = rows.Scan(&id, &email, &username, &password)
		if err != nil {
			log.Fatal(err, "error scanning result")
		}
		log.Printf("ID: %d, Email : %s, Username : %s, Password : %s", id, email, username, password)

		//checks for error in iteration
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	rows.Close()
	// Test the database connection by executing a query.
	err = dataBase.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database connection successful.")
}
