package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lpoulain/PaaKS/paaks"
	"github.com/lpoulain/PaaKS/paaksdb"
)

var validationError = ""

var test paaks.User

////////////////////////////////////////////////

func sqlHandler(w http.ResponseWriter, r *http.Request) {
	token, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusBadRequest)
		return
	}

	tenantShort := token.Tenant[:8]

	body, _ := ioutil.ReadAll(r.Body)
	database := fmt.Sprintf("tnt_%s", tenantShort)

	qry, err := ParseSql(string(body), database)
	fmt.Println(qry)

	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusBadRequest)
		return
	}

	firstRecord := true
	columns := make([]interface{}, 0)

	constructor := func(rows *sql.Rows) ([]interface{}, error) {
		if firstRecord {
			cols, _ := rows.ColumnTypes()
			for _, col := range cols {
				columns = append(columns, map[string]string{
					"name": col.Name(),
					"type": col.ScanType().Name(),
				})
			}
			firstRecord = false
		}
		nbColumns := len(columns)
		valuePtrs := make([]interface{}, nbColumns)

		result := make([]interface{}, nbColumns)

		for i := 0; i < nbColumns; i++ {
			valuePtrs[i] = &result[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		return valuePtrs, nil
	}

	fmt.Printf("SQL: %s\n", qry)
	if r.Header.Get("Test") != "" {
		return
	}

	rows, err := paaksdb.QueryDb[[]interface{}](qry, constructor)

	if err != nil {
		paaks.IssueError(w, "Error executing: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Results: %v\n", rows)
	w.Header().Set("Content-Type", "text/json")

	//	data, _ := json.Marshal(rows)

	result := map[string]interface{}{
		"columns": columns,
		"data":    rows,
	}

	jsonResult, _ := json.Marshal(result)

	w.Write(jsonResult)

}

///////////////////////////////////////////////////////////

func tableConstructor(rows *sql.Rows) (string, error) {
	var name string

	if err := rows.Scan(&name); err != nil {
		return "", err
	}

	return name, nil
}

func tablesHandler(w http.ResponseWriter, r *http.Request) {
	token, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, "No token", http.StatusUnauthorized)
		return
	}

	objects, err := paaksdb.QueryDb(fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema = 'tnt_%s'", token.Tenant[:8]), tableConstructor)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	result, _ := json.Marshal(objects)
	w.Write(result)
}

///////////////////////////////////////////////////////////

func columnConstructor(rows *sql.Rows) (interface{}, error) {
	var name string
	var ordinal_position int
	var is_nullable bool
	var data_type string

	if err := rows.Scan(&name, &ordinal_position, &is_nullable, &data_type); err != nil {
		return "", err
	}

	return map[string]interface{}{
		"name":             name,
		"ordinal_position": ordinal_position,
		"is_nullable":      is_nullable,
		"data_type":        data_type,
	}, nil
}

func columnHandler(w http.ResponseWriter, r *http.Request) {
	token, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, "No token", http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(r.Body)

	paaksdb.QueryDbToResponse(w, fmt.Sprintf("SELECT column_name, ordinal_position, CASE WHEN is_nullable = 'YES' THEN 'true' ELSE 'false' END, data_type FROM information_schema.columns WHERE table_schema = 'tnt_%s' AND table_name = $1",
		token.Tenant[:8]),
		columnConstructor,
		string(body))
}

///////////////////////////////////////////////////////////

func check(w http.ResponseWriter, r *http.Request) {
	_, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, "No token", http.StatusUnauthorized)
		return
	}
	fmt.Fprint(w, "Health Check")
}

///////////////////////////////////////////////////////////

func main() {
	http.HandleFunc("/sql", sqlHandler)
	http.HandleFunc("/tables", tablesHandler)
	http.HandleFunc("/columns", columnHandler)
	http.HandleFunc("/health_check", check)
	fmt.Println("Server starting...")
	http.ListenAndServe(":3000", nil)
}
