package imageStore

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func LoadImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imagePath := filepath.Join("data/images", vars["imagePath"])

	if _, err := os.Stat(imagePath); errors.Is(err, os.ErrNotExist) {
		http.Error(w, "Requested image does not exist", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, imagePath)
}
