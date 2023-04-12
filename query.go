package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func Insert(table string, colm []string, values []interface{}) error {
	dataBase, err := sql.Open("sqlite3", "database/forumDB.sqlite")
	if err != nil {
		log.Fatal(err, "Error opening database")
	}
	defer dataBase.Close() // close connection

	// Build the SQL statement using placeholders for values
	var placeholders []string
	for range colm {
		placeholders = append(placeholders, "?")
	}
	queryStmnt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(colm, ","), strings.Join(placeholders, ","))

	_, err = dataBase.Exec(queryStmnt, values...)
	if err != nil {
		return err
	}
	return nil
}

func Select(table string, columns []string, condition string) ([]map[string]interface{}, error) {
	// Open a connection to the SQL database
	database, err := sql.Open("sqlite3", "database/forumDB.sqlite")
	if err != nil {
		log.Fatal(err, "Error opening database")
	}
	defer database.Close()

	// Build the SQL statement
	var colm string
	if len(columns) == 0 {
		colm = "*"
	} else {
		colm = strings.Join(columns, ", ")
	}
	queryStmnt := fmt.Sprintf("SELECT %s FROM %s", colm, table)
	if condition != "" {
		queryStmnt += fmt.Sprintf(" WHERE %s", condition)
	}

	// Execute the SQL statement
	rows, err := database.Query(queryStmnt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var finalResults []map[string]interface{}

	// iterate over each row in table and store the results in a slice of maps
	for rows.Next() {
		//rows.column returns column name in table
		rowValues, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		//to store all values
		var values []interface{}

		for range rowValues {
			//to store value for each(one) row in table
			var value interface{}
			//append to all values interface
			values = append(values, &value)
		}
		//scan values of row into interface
		err = rows.Scan(values...)
		if err != nil {
			return nil, err
		}
		//to store row data
		rowMap := make(map[string]interface{})

		for i, column := range rowValues {
			//matches column names to their values
			rowMap[column] = *(values[i].(*interface{}))
		}
		finalResults = append(finalResults, rowMap)
	}

	return finalResults, nil
}

func updateOrDelet(table string, operation string, columns []string, values []interface{}, condition string) (int64, error) {
	// Open a connection to the SQLite database
	db, err := sql.Open("sqlite3", "database/forumDB.sqlite")
	if err != nil {
		return 0, err
	}
	defer db.Close() // Make sure to close the connection at the end

	// Build the SQL statement using the table name, column names, condition, and operation
	var queryStmnt string
	var vals []interface{}
	if operation == "update" {
		// For the "update" operation, construct a comma-separated list of column names
		// and corresponding placeholders for each value to be updated.
		setColumns := make([]string, len(columns))
		for i, col := range columns {
			//for amount of columns there are, set placeholders for values
			setColumns[i] = fmt.Sprintf("%s = ?", col)
			//for amount of columns there are, append each value into the args interface
			vals = append(vals, values[i])
			//now setColumns and vals are corresponding placeholder and values
		}
		queryStmnt = fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(setColumns, ", "), condition)
	} else if operation == "delete" {
		queryStmnt = fmt.Sprintf("DELETE FROM %s WHERE %s", table, condition)
	} else {
		return 0, fmt.Errorf("unsupported operation: %s", operation)
	}

	// Execute the SQL statement with the given values
	result, err := db.Exec(queryStmnt, vals...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func maiin() {
	// err := Insert("users", []string{"email", "username", "password"}, []interface{}{"exp@.com4691", "exmp108", "newuser4823"})
	// if err != nil {
	// 	log.Fatal(err, "Error inserting data")
	// }
	// fmt.Println("Data inserted successfully")

	// answer, err := Select("categories", []string{"category"}, "id = 2")
	// if err != nil {
	// 	log.Fatal(err, "Error selecting data")
	// }
	// fmt.Println(answer)
	// fmt.Println("Data selected successfully")

	// Update the email address for user with ID 1
	// rowsAffected, err := updateOrDelete("users", "update", []string{"email", "username"}, []interface{}{"johndoe@example.com", "johndoe"}, "id = 2")
	// if err != nil {
	// 	fmt.Println("Error updating data:", err)
	// } else {
	// 	fmt.Printf("Updated %d row(s) successfully.\n", rowsAffected)
	// }

	// Delete all users with a null email address
	// rowsAffected, err := updateOrDelete("users", "delete", []string{}, nil, "id = 9")
	// if err != nil {
	// 	fmt.Println("Error deleting data:", err)
	// } else {
	// 	fmt.Printf("Deleted %d row(s) successfully.\n", rowsAffected)
	// }
}
