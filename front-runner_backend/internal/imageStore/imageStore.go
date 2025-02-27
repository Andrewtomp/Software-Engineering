package imageStore

import (
	"errors"
	"fmt"
	"front-runner/internal/login"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func LoadImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imagePath := filepath.Join("data/images", vars["imagePath"])
	session, err := login.SessionStore.Get(r, "auth")
	if err != nil {
		http.Error(w, "Error getting session", http.StatusInternalServerError)
		return
	}

	if auth, ok := session.Values["authenticated"].(bool); !(ok && auth) {
		fmt.Fprintf(w, "User is not logged in.")
		return
	}

	//TODO: check if user has access to file

	if _, err := os.Stat(imagePath); errors.Is(err, os.ErrNotExist) {
		http.Error(w, "Requested image does not exist", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, imagePath)
}
