// main.go
package main

import (
	"fmt"
	"net/http"

	login "front-runner/internal/login"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/register", login.RegisterUser).Methods("POST")
	r.HandleFunc("/login", login.LoginUser).Methods("POST")
	r.HandleFunc("/logout", login.LogoutUser).Methods("GET")

	http.Handle("/", r)
	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}
