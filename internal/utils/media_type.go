package utils

import (
	"path/filepath"
	"strings"
)

// Known media type extensions
var mediaExtensions = map[string]string{
	// Image
	".jpg":  "image",
	".jpeg": "image",
	".png":  "image",
	".gif":  "image",
	".bmp":  "image",
	".webp": "image",
	".tiff": "image",
	".svg":  "image",

	// Audio
	".mp3":  "audio",
	".wav":  "audio",
	".aac":  "audio",
	".flac": "audio",
	".ogg":  "audio",
	".m4a":  "audio",

	// Video
	".mp4":  "video",
	".mov":  "video",
	".avi":  "video",
	".mkv":  "video",
	".webm": "video",
	".flv":  "video",
	".wmv":  "video",

	// PDF
	".pdf": "pdf",
}

// DetectMediaCategoryByExtension returns "image", "audio", "video", "pdf" or "unknown"
func DetectMediaCategoryByExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if mediaType, ok := mediaExtensions[ext]; ok {
		return mediaType
	}
	return ""
}
