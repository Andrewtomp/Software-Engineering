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
	"strings"

	"github.com/google/uuid"
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

// Checks if a given file exists. Returns true if exists, false if not.
func doesFileExist(filepath string) bool {
	_, err := os.Stat(filepath)
	return !errors.Is(err, os.ErrNotExist)
}

// Retrieves the specified image.
//
//	@Summary		Retrive an image
//	@Description	Fetches an image if it exists and they are authorized.
//
//	@Tags			images
//	@Produce		image/*
//	@Param			filename	path		string	true	"Filepath of image"
//	@Success		200			{string}	binary
//	@Failure		401			{string}	string	"User is not logged in"
//	@Failure		403			{string}	string	"Permission denied"
//	@Failure		404			{string}	string	"Requested image does not exist"
//	@Failure		500			{string}	string	"Unable to retrieve User ID"
//	@Router			/api/data/image/{filename} [get]
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

	if !doesFileExist(imagePath) {
		http.Error(w, "Requested image does not exist", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, imagePath)
}

// Uploads a specified image
//
//	@Summary		Upload an image
//	@Description	Uploads an image if the user is authorized.
//
//	@Tags			images
//	@Param			filename formData file true "Filepath of image"
//	@Accept			mpfd
//	@Success		200	{string}	body	"Filename of uploaded image"
//	@Failure		401	{string}	string	"User is not logged in"
//	@Failure		403	{string}	string	"Permission denied"
//	@Failure		404	{string}	string	"Requested image does not exist"
//	@Failure		415	{string}	string	"Invalid file type"
//	@Failure		500	{string}	string	"Unable to retrieve User ID"
//	@Failure		500	{string}	string	"File already exists"
//	@Router			/api/data/upload [post]
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

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	uploadedType := http.DetectContentType(fileBytes)
	if !strings.HasPrefix(uploadedType, "image/") {
		http.Error(w, "Invalid file type", http.StatusUnsupportedMediaType)
		return
	}

	imagePath := filepath.Join("data/images", uuid.New().String()+filepath.Ext(handler.Filename))

	for doesFileExist(imagePath) {
		imagePath = filepath.Join("data/images", uuid.New().String()+filepath.Ext(handler.Filename))
	}

	tempFile, err := os.Create(imagePath)
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	tempFile.Write(fileBytes)
	localFileName := filepath.Base(tempFile.Name())
	image := ProductImage{
		Filename: localFileName,
		ID:       userID,
	}

	if err := db.Create(&image).Error; err != nil {
		http.Error(w, "File already exists", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", localFileName)
}
