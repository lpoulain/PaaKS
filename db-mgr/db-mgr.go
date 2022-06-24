package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/lpoulain/PaaKS/paaks"
	"github.com/lpoulain/PaaKS/paaksdb"
)

var validationError = ""

var test paaks.User

////////////////////////////////////////////////

func sqlHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

func tablesHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
	var not_null bool
	var data_type string
	var column_default string

	if err := rows.Scan(&name, &not_null, &data_type, &column_default); err != nil {
		return "", err
	}

	return map[string]interface{}{
		"name":           name,
		"not_null":       not_null,
		"data_type":      data_type,
		"column_default": column_default,
	}, nil
}

func readTableHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	token, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, "No token", http.StatusUnauthorized)
		return
	}

	tableName := ps.ByName("tableName")

	objects, err := paaksdb.QueryDb(fmt.Sprintf("SELECT column_name, CASE WHEN is_nullable = 'YES' THEN 'false' ELSE 'true' END, UPPER(udt_name) AS data_type, COALESCE(column_default, '') FROM information_schema.columns WHERE table_schema = 'tnt_%s' AND table_name = $1 ORDER BY ordinal_position",
		token.Tenant[:8]),
		columnConstructor,
		tableName)

	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"table":   tableName,
		"columns": objects,
	}

	w.Header().Set("Content-Type", "text/json")
	result, _ := json.Marshal(response)
	w.Write(result)
}

func downloadTableHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	token, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, "No token", http.StatusUnauthorized)
		return
	}

	tableName := ps.ByName("tableName")

	objects, err := paaksdb.QueryDb(fmt.Sprintf("SELECT column_name, CASE WHEN is_nullable = 'YES' THEN 'false' ELSE 'true' END, UPPER(udt_name) AS data_type, COALESCE(column_default, '') FROM information_schema.columns WHERE table_schema = 'tnt_%s' AND table_name = $1 ORDER BY ordinal_position",
		token.Tenant[:8]),
		columnConstructor,
		tableName)

	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"table":   tableName,
		"columns": objects,
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\"paaks_"+tableName+"_definition.json\"")
	result, _ := json.Marshal(response)
	w.Write(result)

}

func alterTableHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, "No token", http.StatusUnauthorized)
		return
	}

	tenantShort := token.Tenant[:8]

	//	body, _ := ioutil.ReadAll(r.Body)
	database := fmt.Sprintf("tnt_%s", tenantShort)

	var alter TableDescription
	err = json.NewDecoder(r.Body).Decode(&alter)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusBadRequest)
		return
	}

	subSql := []string{}

	for _, col := range alter.Columns {

		line := ""
		if !checkValue(col.Name) {
			paaks.IssueError(w, "Illegal value "+col.Name, http.StatusBadRequest)
			return
		}

		switch col.Action {
		case "DROP":
			line = fmt.Sprintf("DROP COLUMN %s", col.Name)
		case "ADD":
			if !checkValue(col.DataType) {
				paaks.IssueError(w, "Illegal value "+col.DataType, http.StatusBadRequest)
				return
			}

			line = fmt.Sprintf("ADD COLUMN %s %s", col.Name, col.DataType)

			if col.CannotBeNull {
				line += " NOT NULL"
			}

			if col.Default != "" {
				if !checkValue(col.Default) {
					paaks.IssueError(w, "Illegal value "+col.Default, http.StatusBadRequest)
					return
				}

				line += fmt.Sprintf(" DEFAULT %s", col.Default)
			}
		default:
			paaks.IssueError(w, "Unknown action: "+col.Action, http.StatusBadRequest)
			return
		}
		subSql = append(subSql, line)
	}

	var sql = fmt.Sprintf("ALTER TABLE %s.%s\n%s", database, alter.Table, strings.Join(subSql, ",\n"))
	fmt.Println(sql)

	err = paaksdb.ExecDb(sql)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Success")
}

func checkValue(value string) bool {
	tokens, err := GetTokens(value)
	if err != nil {
		return false
	}
	res, tokensLeft, _ := ParseAst("value", tokens)
	if len(tokensLeft) > 0 {
		return false
	}
	return res
}

type TableDescription struct {
	Table   string   `json:"table"`
	Columns []Column `json:"columns"`
}

type Column struct {
	Action       string `json:"action"`
	Name         string `json:"name"`
	DataType     string `json:"data_type"`
	CannotBeNull bool   `json:"not_null"`
	Default      string `json:"default"`
}

///////////////////////////////////////////////////////////

func check(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, "No token", http.StatusUnauthorized)
		return
	}
	fmt.Fprint(w, "Health Check")
}

func createTableHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	token, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, "No token", http.StatusUnauthorized)
		return
	}

	tenantShort := token.Tenant[:8]

	//	body, _ := ioutil.ReadAll(r.Body)
	database := fmt.Sprintf("tnt_%s", tenantShort)

	f, _, _ := r.FormFile("file")
	metadata, _ := ioutil.ReadAll(f)
	fmt.Println(string(metadata))

	var create TableDescription
	err = json.Unmarshal(metadata, &create)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusBadRequest)
		return
	}

	subSql := []string{}

	for _, col := range create.Columns {
		if !checkValue(col.DataType) {
			paaks.IssueError(w, "Illegal datatype value: ["+col.DataType+"]", http.StatusBadRequest)
			return
		}

		if col.Default != "" && !checkValue(col.Default) {
			paaks.IssueError(w, "Illegal default value: ["+col.Default+"]", http.StatusBadRequest)
			return
		}

		line := fmt.Sprintf("%s %s", col.Name, col.DataType)

		if col.CannotBeNull {
			line += fmt.Sprintf(" NOT NULL")
		}

		if col.Default != "" {
			line += fmt.Sprintf(" DEFAULT %s", col.Default)
		}

		subSql = append(subSql, line)
	}

	sql := fmt.Sprintf("CREATE TABLE %s.%s(\n  %s\n)", database, create.Table, strings.Join(subSql, ",\n  "))
	fmt.Println(sql)
	err = paaksdb.ExecDb(sql)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Success")
}

func deleteTableHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	token, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, "No token", http.StatusUnauthorized)
		return
	}

	tenantShort := token.Tenant[:8]
	database := fmt.Sprintf("tnt_%s", tenantShort)

	sql := fmt.Sprintf("DROP TABLE %s.%s", database, ps.ByName("tableName"))
	fmt.Println(sql)
	err = paaksdb.ExecDb(sql)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Success")
}

///////////////////////////////////////////////////////////

func main() {
	router := httprouter.New()
	router.POST("/sql", sqlHandler)
	router.GET("/tables", tablesHandler)
	router.GET("/tables/:tableName", readTableHandler)
	router.GET("/tables/:tableName/download", downloadTableHandler)
	router.POST("/tables", createTableHandler)
	router.DELETE("/tables/:tableName", deleteTableHandler)
	router.PUT("/tables/:tableName", alterTableHandler)
	router.GET("/health_check", check)

	/*	http.HandleFunc("/sql", sqlHandler)
		http.HandleFunc("/tables", tablesHandler)
		router.HandleFunc("/tables/{tableName}", columnHandler).Methods("GET")
		router.HandleFunc("/tables/{tableName}", createTableHandler).Methods("PUT")
		router.HandleFunc("/tables/{tableName}", deleteTableHandler).Methods("DELETE")
		http.HandleFunc("/columns", columnHandler)
		http.HandleFunc("/alter-table", alterTableHandler)
		http.HandleFunc("/health_check", check)
	*/
	fmt.Println("Server starting...")

	http.ListenAndServe(":3000", router)
}
