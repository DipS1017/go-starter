package dto

import "mime/multipart"

type MediaUploadResponse struct {
	Url  string `json:"url"`
	Id   string `json:"id"`
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

type MediaUploadForm struct {
	File      *multipart.FileHeader `form:"file" validate:"required" swaggertype:"string" format:"binary"`
	Thumbnail *multipart.FileHeader `form:"thumbnail" swaggertype:"string" format:"binary"`
	IsPublic  bool                  `form:"isPublic" example:"false"`
}
