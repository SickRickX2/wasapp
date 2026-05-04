package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SickRickX2/wasapp/service/api/schemas"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Autenticazione
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix+" ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// estrai userId dal token (simulato)
	userId := schemas.UserId(strings.TrimPrefix(authHeader, bearerPrefix+" "))

	// path params
	vars := ps
	convId := schemas.ConversationId(vars.ByName("convId"))

	// controllo sicurezza
	isInGroup, err := rt.db.IsUserInConversation(convId, userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error checking group membership")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isInGroup {
		http.Error(w, "Forbidden: You are not a participant of this group", http.StatusForbidden)
		return
	}

	const maxUploadSize = 5 << 20 // 5MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file key (use 'file')", http.StatusBadRequest)
		return
	}
	defer file.Close()

	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	mimeType := http.DetectContentType(buff)
	if !strings.HasPrefix(mimeType, "image/") {
		http.Error(w, "Only images are allowed", http.StatusBadRequest)
		return
	}
	if _, err := file.Seek(0, 0); err != nil {
		http.Error(w, "Error processing file", http.StatusInternalServerError)
		return
	}

	mediaId := generateMediaId()
	fileExt := filepath.Ext(fileHeader.Filename)
	if fileExt == "" {
		fileExt = ".jpg"
	}
	newFilename := string(mediaId) + fileExt
	savePath := filepath.Join(mediaDir, newFilename)

	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		rt.baseLogger.WithError(err).Error("Error creating images directory")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

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

	photoUrl := "/images/" + newFilename
	mediaObj := schemas.Media{
		URL:      photoUrl,
		Filename: fileHeader.Filename,
		MimeType: mimeType,
		Size:     int(fileHeader.Size),
	}

	if err := rt.db.SaveMedia(mediaObj, string(mediaId)); err != nil {
		rt.baseLogger.WithError(err).Error("Error saving media info to DB")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Aggiorna DB
	err = rt.db.SetGroupPhoto(convId, photoUrl)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error setting group photo")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	updatedConv, err := rt.db.GetConversation(convId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("Error retrieving updated group")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(updatedConv); err != nil {
		rt.baseLogger.WithError(err).Error("failed to encode response in setGroupPhoto")
	}
}
