package main

import (
	"log"
	"net/http"
	"shopping-api/api"

	"github.com/go-chi/chi/v5"
)

func main() {

	r := chi.NewRouter()
	r.Get("/items", api.GetItemsList)
	r.Post("/items", api.CreateItem)
	r.Get("/items/{id}", api.GetItem)
	r.Put("/items/{id}", api.FullUpdateItem)
	r.Patch("/items/{id}", api.PartlyUpdateItems)
	r.Delete("/items/{id}", api.DeleteItem)
	r.Get("/items/category/{category}", api.GetCategoryItems)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
