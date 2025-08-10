package main

import (
	"database/sql"
	"log"
	"net/http"
)

func StartWebServer(db *sql.DB) {
	http.HandleFunc("/", MainPage(db))
	log.Println("Web server starting")
	http.ListenAndServe(":8080", nil)
}

func MainPage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		w.Write(nil)
	}
}
