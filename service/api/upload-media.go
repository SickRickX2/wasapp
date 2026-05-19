package api

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) uploadMedia(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	const maxUploadSize = 5 << 20 // 5mb
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {

		http.Error(w, `{"message": "File too big"}`, http.StatusRequestEntityTooLarge)
		return
	}

	// prendi il file
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"description": "Invalid file key (use 'file')"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// vedi se è un immagine
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	mimeType := http.DetectContentType(buff)
	if !strings.HasPrefix(mimeType, "image/") {
		http.Error(w, "Only images are allowed", http.StatusBadRequest)
		return
	}
	// resetta il puntatore
	_, _ = file.Seek(0, 0)

	// genera id e nome
	mediaId := generateMediaId()
	// metti l'estensione
	fileExt := filepath.Ext(fileHeader.Filename)
	if fileExt == "" {
		fileExt = ".jpg"
	}
	newFilename := string(mediaId) + fileExt
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		rt.baseLogger.WithError(err).Error("Error creating media directory")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	savePath := filepath.Join(mediaDir, newFilename)

	// salva su disco
	dst, err := os.Create(savePath)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error creating file on disk")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		rt.baseLogger.WithError(err).Error("Error saving file content")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// salva nel db
	mediaObj := schemas.Media{
		URL:      "/images/" + newFilename,
		Filename: fileHeader.Filename,
		MimeType: mimeType,
		Size:     int(fileHeader.Size),
	}

	if err := rt.db.SaveMedia(mediaObj, string(mediaId)); err != nil {
		rt.baseLogger.WithError(err).Error("Error saving media info to DB")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// rispondi con il json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := struct {
		MediaId string `json:"mediaId"`
		Url     string `json:"url"`
	}{
		MediaId: string(mediaId),
		Url:     mediaObj.URL,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in uploadMedia")
	}
}

func generateMediaId() schemas.MessageId {
	const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = charSet[rand.Intn(len(charSet))]
	}
	return schemas.MessageId("med_" + string(b))
}
