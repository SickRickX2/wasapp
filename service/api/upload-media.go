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
	// 1. Autenticazione (bisogna essere loggati per caricare file)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	const maxUploadSize = 5 << 20 // 5 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {

		http.Error(w, `{"message": "File too big"}`, http.StatusRequestEntityTooLarge)
		return
	}

	// 3. Leggi il file dal form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"description": "Invalid file key (use 'file')"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 4. Controlla che sia un'immagine (lettura dei primi 512 byte)
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	mimeType := http.DetectContentType(buff)
	// Controlliamo che sia un'immagine
	if !strings.HasPrefix(mimeType, "image/") {
		http.Error(w, "Only images are allowed", http.StatusBadRequest)
		return
	}
	// Resettiamo il puntatore del file all'inizio dopo aver letto i 512 byte
	_, _ = file.Seek(0, 0)

	// 5. Genera ID e Nome File
	mediaId := generateMediaId()
	// Estensione (es. .jpg)
	fileExt := filepath.Ext(fileHeader.Filename)
	if fileExt == "" {
		// Fallback brutale se non c'è estensione
		fileExt = ".jpg"
	}
	newFilename := string(mediaId) + fileExt
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		rt.baseLogger.WithError(err).Error("Error creating media directory")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Percorso salvataggio: /tmp/wasa-images/media_xxxx.jpg
	savePath := filepath.Join(mediaDir, newFilename)

	// 6. Salva su Disco
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

	// 7. Salva nel DB
	mediaObj := schemas.Media{
		URL:      "/images/" + newFilename, // Questo sarà l'URL per scaricarla
		Filename: fileHeader.Filename,
		MimeType: mimeType,
		Size:     int(fileHeader.Size),
	}

	if err := rt.db.SaveMedia(mediaObj, string(mediaId)); err != nil {
		rt.baseLogger.WithError(err).Error("Error saving media info to DB")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 8. Rispondi col JSON
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
