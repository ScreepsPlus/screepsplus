package main

import (
	"log"
	"net/http"
	"time"

	"github.com/screepsplus/screepsplus/db"
	"github.com/screepsplus/screepsplus/routes/auth"
)

func main() {
	defer db.DB().Close()
	r := auth.NewRouter()
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
