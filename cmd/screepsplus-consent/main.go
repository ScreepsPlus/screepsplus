package main

import (
	"log"
	"net/http"
	"time"

	"github.com/screepsplus/screepsplus/routes/auth"
)

func main() {
	r := auth.NewRouter()
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
