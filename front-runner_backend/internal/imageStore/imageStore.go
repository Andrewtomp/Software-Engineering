package imageStore

import (
	"errors"
	"front-runner/internal/login"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

// Retrieves the specified image.
//
// @Summary      Retrive an image
// @Description  Fetches an image if it exists and they are authorized.
//
// @Tags         images
// @Produce      image
// @Success      200 {binary} binary
// @Failure      401 {string} string "User is not logged in"
// @Failure      403 {string} string "Permission denied"
// @Failure      404 {string} string "Requested image does not exist"
// @Failure      500 {string} string "Unable to retrieve User ID"
// @Router       /api/data/image/{filename} [get]
func LoadImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imagePath := filepath.Join("data/images", vars["imagePath"])

	if !login.IsLoggedIn(r) {
		http.Error(w, "User is not logged in", http.StatusUnauthorized)
		return
	}

	userID, err := login.GetUserID(r)
	if err != nil {
		http.Error(w, "Unable to retrieve User ID", http.StatusInternalServerError)
		return
	}

	if userID == 0 {

	}

	//TODO: check if user has access to file

	if _, err := os.Stat(imagePath); errors.Is(err, os.ErrNotExist) {
		http.Error(w, "Requested image does not exist", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, imagePath)
}
