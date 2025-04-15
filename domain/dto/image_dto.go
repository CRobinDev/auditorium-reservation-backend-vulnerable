package dto

import (
	"mime/multipart"
	"net/http"
)

type Image struct {
	File *multipart.FileHeader `form:"file" validate:"required,image_type,image_size"`
}

func GetImageType(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}

	defer src.Close()

	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buffer), nil
}
