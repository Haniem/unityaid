package files

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	uploadsDir string
}

func NewHandler(uploadsDir string) *Handler {
	return &Handler{uploadsDir: uploadsDir}
}

func (h *Handler) UploadNewsImage(c *gin.Context) {
	h.uploadImage(c, "news")
}

func (h *Handler) UploadOrganizationLogo(c *gin.Context) {
	h.uploadImage(c, "organizations")
}

func (h *Handler) uploadImage(c *gin.Context, folder string) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_file", "message": "Файл не передан"})
		return
	}

	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_too_large", "message": "Максимальный размер файла 5 МБ"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" && ext != ".gif" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_extension", "message": "Поддерживаются jpg, png, webp и gif"})
		return
	}

	dir := filepath.Join(h.uploadsDir, folder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось подготовить папку загрузок"})
		return
	}

	name := randomHex(16) + ext
	destination := filepath.Join(dir, name)
	if err := c.SaveUploadedFile(file, destination); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось сохранить файл"})
		return
	}

	url := requestBaseURL(c) + "/uploads/" + folder + "/" + name
	c.JSON(http.StatusCreated, gin.H{"url": url})
}

func randomHex(size int) string {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "file"
	}
	return hex.EncodeToString(bytes)
}

func requestBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if forwardedProto := c.GetHeader("X-Forwarded-Proto"); forwardedProto != "" {
		scheme = forwardedProto
	}
	return scheme + "://" + c.Request.Host
}
