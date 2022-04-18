package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func error(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusBadRequest)
}

func router(w http.ResponseWriter, r *http.Request) {
	fmt.Println("URL: ", r.URL.Path, ", method: ", r.Method)

	token := r.Header.Get("Authorization")
	fmt.Println(token)
	if !strings.HasPrefix(token, "Bearer ") {
		fmt.Fprintf(w, "No token")
		return
	}

	companyId := token[7:15]

	regex, _ := regexp.Compile("^/([^/]+)/(.*)$")
	results := regex.FindStringSubmatch(r.URL.Path)

	if len(results) < 3 {
		error(w, "Invalid path")
		return
	}

	service := fmt.Sprintf("tnt-%s-%s", companyId, results[1])
	path := results[2]

	fmt.Println("Service: ", service, "File: ", path)

	switch r.Method {
	case "GET":
		load(w, service, path)
		break
	case "POST":
		save(w, r, service, path)
		break
	}
}

func load(w http.ResponseWriter, service string, path string) {
	filename := fmt.Sprintf("/tmp/storage/%s/%s", service, path)

	file, err := os.Open(filename)

	if err != nil {
		error(w, "Invalid path")
		return
	}

	fileInfo, err := file.Stat()
	if err != nil {
		error(w, "Invalid path")
		return
	}

	if fileInfo.IsDir() {
		loadDir(w, filename)
	} else {
		loadFile(w, filename)
	}
}

func loadFile(w http.ResponseWriter, path string) {
	data, err := os.ReadFile(path)

	if err != nil {
		error(w, "Cannot read the file")
		return
	}

	w.Write(data)
}

type File struct {
	name string
	dir  bool
}

func loadDir(w http.ResponseWriter, path string) {
	data, err := os.ReadDir(path)

	if err != nil {
		error(w, "Cannot read the directory")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var files = []map[string]interface{}{}
	for _, dir := range data {
		f := make(map[string]interface{})
		f["name"] = dir.Name()
		f["dir"] = dir.IsDir()
		files = append(files, f)
	}

	fmt.Println(files)

	result, _ := json.Marshal(files)

	w.Write(result)
}

func save(w http.ResponseWriter, req *http.Request, service string, path string) {
	filename := fmt.Sprintf("/tmp/storage/%s/%s", service, path)

	err := req.ParseForm()
	if err != nil {
		return
	}

	body := req.FormValue("body")

	file, err := os.Create(filename)

	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(body)
}

func check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Health Check</h1>")
}

func main() {
	http.HandleFunc("/", router)
	http.HandleFunc("/health_check", check)
	fmt.Println("Server starting...")
	http.ListenAndServe(":3000", nil)
}
