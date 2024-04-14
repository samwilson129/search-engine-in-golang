package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

var apiKey = "AIzaSyC_1FL1-QsEKFr7IoeajLiWICEvd7Y-VWE"
var customSearchID = "3324b3fe8b5594718"

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handler)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		filePath := "static/index.html"
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Error opening file: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		_, err = io.Copy(w, file)
		if err != nil {
			log.Printf("Error writing response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		query := r.FormValue("query")
		if query == "" {
			http.Error(w, "Please provide a search query.", http.StatusBadRequest)
			return
		}

		results, err := processQuery(query)
		if err != nil {
			log.Printf("Error processing query: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("static/result.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, struct {
			Query   string
			Results []*customsearch.Result
		}{
			Query:   query,
			Results: results,
		})
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func processQuery(query string) ([]*customsearch.Result, error) {
	ctx := context.Background()
	client, err := customsearch.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Custom Search client: %v", err)
	}
	search := client.Cse.List().Q(query).Cx(customSearchID)
	results, err := search.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to perform search: %v", err)
	}
	return results.Items, nil
}
