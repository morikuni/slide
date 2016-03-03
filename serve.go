package main

import "net/http"

func main() {
	panic(http.ListenAndServe("127.0.0.1:8080", http.FileServer(http.Dir("./"))))
}
