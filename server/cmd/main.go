package main

import (
	"net/http"

	"github.com/styltsou/url-shortener/server/pkg/api"
)

func main() {
	r := api.NewRouter()

	err := http.ListenAndServe(":5000", r)
	if err != nil {
		panic("Error listeing on port 5000")
	}
}
