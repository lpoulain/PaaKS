package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

func queryToResponse(w http.ResponseWriter, sqlQuery string, constructor func(*sql.Rows) (interface{}, error), args ...interface{}) {
	objects, err := query(sqlQuery, constructor, args...)
	if err != nil {
		issueError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	result, _ := json.Marshal(objects)
	w.Write(result)
}

func query[K interface{}](sqlQuery string, constructor func(*sql.Rows) (K, error), args ...interface{}) ([]K, error) {
	connStr := getConnectionString()
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("ERROR connecting to the database:", err.Error())
		return nil, err
	}

	var rows *sql.Rows

	if len(args) > 0 {
		stmt, err := conn.Prepare(sqlQuery)
		if err != nil {
			return nil, err
		}
		if constructor == nil {
			_, err = stmt.Exec(args...)
			return nil, err
		} else {
			rows, err = stmt.Query(args...)
		}
	} else {
		rows, err = conn.Query(sqlQuery)
	}

	if err != nil {
		fmt.Println("ERROR querying data:", err.Error())
		return nil, err
	}

	if constructor == nil {
		return nil, nil
	}

	objects := []K{}

	defer conn.Close()
	defer rows.Close()
	for rows.Next() {
		object, err := constructor(rows)
		if err != nil {
			return nil, err
		}

		objects = append(objects, object)
	}

	return objects, nil
}

func exec(sqlQuery string, args ...interface{}) error {
	connStr := getConnectionString()
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("ERROR connecting to the database:", err.Error())
		return err
	}

	if len(args) > 0 {
		stmt, err := conn.Prepare(sqlQuery)
		if err != nil {
			return err
		}

		_, err = stmt.Exec(args...)
		return err
	} else {
		_, err = conn.Exec(sqlQuery)
		return err
	}
}
