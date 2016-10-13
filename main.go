package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func main() {
	http.HandleFunc("/", HelloHandler)

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
