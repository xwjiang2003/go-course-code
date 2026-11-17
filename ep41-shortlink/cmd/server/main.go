package main

import (
	"log"
	"net/http"

	"shortlink/internal/cache"
	"shortlink/internal/handler"
	"shortlink/internal/middleware"
	"shortlink/internal/service"
	"shortlink/internal/store"
)

func main() {
	c := cache.New()
	st := store.NewMemory() // 生产换成 store.NewMySQL(db)
	svc := service.New(st, c)
	h := handler.New(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", h.Shorten)
	mux.HandleFunc("/", h.Redirect)

	log.Println("shortlink listening on :8080")
	log.Fatal(http.ListenAndServe(":8080",
		middleware.Logger(middleware.Recover(mux))))
}
