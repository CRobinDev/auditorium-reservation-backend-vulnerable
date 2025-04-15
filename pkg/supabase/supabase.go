package supabase

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/nathakusuma/auditorium-reservation-backend/domain/dto"
	storage_go "github.com/supabase-community/storage-go"
)

type Supabase struct {
	client storage_go.Client
}

type ISupabase interface {
	UploadFile(file *multipart.FileHeader, dir string) (string, error)
}

func New() ISupabase {
	url := fmt.Sprintf("%s/storage/v1", os.Getenv("SUPABASE_PROJECT_URL"))
	client := storage_go.NewClient(url, os.Getenv("SUPABASE_KEY"), nil)

	return Supabase{
		client: *client,
	}
}

func (s Supabase) UploadFile(file *multipart.FileHeader, dir string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	path := dir + "/" + uuid.NewString() + filepath.Ext(file.Filename)
	contentType, err := dto.GetImageType(file)
	if err != nil {
		return "", err
	}

	_, err = s.client.UploadFile(
		os.Getenv("SUPABASE_BUCKET_NAME"),
		path,
		src,
		storage_go.FileOptions{
			ContentType: &contentType,
		},
	)

	if err != nil {
		return "", err
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		os.Getenv("SUPABASE_PROJECT_URL"),
		os.Getenv("SUPABASE_BUCKET_NAME"),
		path,
	)

	return publicURL, nil
}
