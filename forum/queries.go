package forum

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func InsertInfo(table string, colm []string, values []interface{}) error {
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
		log.Fatal(err, "error executing insert statement")
	}
	return nil
}

func SelectInfo(table string, columns []string, condition string) ([]map[string]interface{}, error) {
	// Open a connection to the SQL database
	database, err := sql.Open("sqlite3", "forum/database/forumDB.sqlite")
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
		log.Fatal(err, " error selecting information")
	}
	defer rows.Close()

	var finalResults []map[string]interface{}

	// iterate over each row in table and store the results in a slice of maps
	for rows.Next() {
		//rows.column returns column name in table
		rowValues, err := rows.Columns()
		if err != nil {
			log.Fatal(err, "error getting column names")
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
			log.Fatal(err, "error scanning row values")
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

func updateOrDeleteInfo(table string, operation string, columns []string, values []interface{}, condition string) (int64, error) {
	// Open a connection to the SQLite database
	db, err := sql.Open("sqlite3", "database/forumDB.sqlite")
	if err != nil {
		log.Fatal(err, "Error opening database")
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
		log.Fatal(err, "error executing update/delete statement")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err, "error returning rows affected")
	}

	return rowsAffected, nil
}
