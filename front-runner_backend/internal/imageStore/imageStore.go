package imageStore

import (
	"errors"
	"front-runner/internal/coredbutils"
	"front-runner/internal/login"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

type ProductImage struct {
	Filename string `gorm:"primaryKey;not null"`
	ID       uint   `gorm:"not null"`
}

func init() {
	db = coredbutils.GetDB()
}

// Retrieves the specified image.
//
// @Summary      Retrive an image
// @Description  Fetches an image if it exists and they are authorized.
//
// @Tags         images
// @Produce      image/*
// @Param        filename path string true "Filepath of image"
// @Success      200 {string} binary
// @Failure      401 {string} string "User is not logged in"
// @Failure      403 {string} string "Permission denied"
// @Failure      404 {string} string "Requested image does not exist"
// @Failure      500 {string} string "Unable to retrieve User ID"
// @Router       /api/data/image/{filename} [get]
func LoadImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["imagePath"]
	imagePath := filepath.Join("data/images", filename)

	if !login.IsLoggedIn(r) {
		http.Error(w, "User is not logged in", http.StatusUnauthorized)
		return
	}

	userID, err := login.GetUserID(r)
	if err != nil {
		http.Error(w, "Unable to retrieve User ID", http.StatusInternalServerError)
		return
	}

	var image ProductImage
	if err := db.Where("filename = ?", filename).First(&image).Error; err != nil {
		http.Error(w, "Invalid filename", http.StatusUnauthorized)
		return
	}

	if userID != image.ID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	if _, err := os.Stat(imagePath); errors.Is(err, os.ErrNotExist) {
		http.Error(w, "Requested image does not exist", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, imagePath)
}
