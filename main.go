package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/screepsplus/screepsplus/db"
	"github.com/screepsplus/screepsplus/routes"
)

func main() {
	host := ""
	port := "8000"
	if v, ok := os.LookupEnv("HOST"); ok {
		host = v
	}
	if v, ok := os.LookupEnv("PORT"); ok {
		port = v
	}
	db.Init()
	defer db.DB().Close()
	r := routes.NewRouter()
	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", host, port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
