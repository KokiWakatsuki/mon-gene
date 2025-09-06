package main

import (
  "net/http"
  "fmt"
)

func main() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello, world!")
  })

  server := http.Server{
    Addr:    ":8080",
    Handler: nil,
  }
  server.ListenAndServe()
}