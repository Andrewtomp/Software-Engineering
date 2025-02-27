package imageStore

import (
	"errors"
	"fmt"
	"front-runner/internal/coredbutils"
	"front-runner/internal/login"
	"io"
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

	if _, err := os.Stat("data"); os.IsNotExist(err) {
		os.Mkdir("data", 0755)
	}

	if _, err := os.Stat("data/images"); os.IsNotExist(err) {
		os.Mkdir("data/images", 0755)
	}
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

// Retrieves the specified image.
//
// @Summary      Retrive an image
// @Description  Fetches an image if it exists and they are authorized.
//
// @Tags         images
// @Param        filename formData string true "Filepath of image"
// @Success      200 {string} binary
// @Failure      401 {string} string "User is not logged in"
// @Failure      403 {string} string "Permission denied"
// @Failure      404 {string} string "Requested image does not exist"
// @Failure      500 {string} string "Unable to retrieve User ID"
// @Router       /api/data/upload [post]
func UploadImage(w http.ResponseWriter, r *http.Request) {

	if !login.IsLoggedIn(r) {
		http.Error(w, "User is not logged in", http.StatusUnauthorized)
		return
	}

	userID, err := login.GetUserID(r)
	if err != nil {
		http.Error(w, "Unable to retrieve User ID", http.StatusInternalServerError)
		return
	}

	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("filename")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("data/images", "upload-*"+filepath.Ext(handler.Filename))
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	tempFile.Write(fileBytes)

	image := ProductImage{
		Filename: filepath.Base(tempFile.Name()),
		ID:       userID,
	}

	if err := db.Create(&image).Error; err != nil {
		http.Error(w, "File already exists.", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully Uploaded File\n")
}
