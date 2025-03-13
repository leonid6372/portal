package minioServer

import (
	"context"
	"fmt"
	"os"
	"portal/internal/structs/models"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioProvider struct {
	minioAuthData
	client *minio.Client
}

type minioAuthData struct {
	url      string
	user     string
	password string
	bucket   string
	token    string
	ssl      bool
}

func New(minioURL, bucket string, ssl bool) (*MinioProvider, error) {
	const op = "storage.minioServer.New"

	srvUsername, exists := os.LookupEnv("MINIO_USERNAME")
	if !exists {
		return nil, fmt.Errorf("%s: username for MINIO does not exists in env", op)
	}
	srvPassword, exists := os.LookupEnv("MINIO_PASSWORD")
	if !exists {
		return nil, fmt.Errorf("%s: password for MINIO does not exists in env", op)
	}

	return &MinioProvider{
		minioAuthData: minioAuthData{
			password: srvPassword,
			bucket:   bucket,
			url:      minioURL,
			user:     srvUsername,
			ssl:      ssl,
		}}, nil
}

func (m *MinioProvider) Connect() error {
	const op = "storage.minioServer.Connect"

	var err error
	m.client, err = minio.New(m.url, &minio.Options{
		Creds:  credentials.NewStaticV4(m.user, m.password, m.token),
		Secure: m.ssl,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// UploadImage - Отправляет файл в minio
func (m *MinioProvider) UploadImage(ctx context.Context, image models.Image) error {
	const op = "storage.minioServer.UploadImage"

	_, err := m.client.PutObject(
		ctx,
		m.bucket, // Константа с именем бакета
		image.Name,
		image.Payload,
		image.Size,
		minio.PutObjectOptions{ContentType: fmt.Sprintf("image/%s", image.Extension)},
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// DownloadImage - Возвращает файл из minio. Полученный файл надо close() после использования
func (m *MinioProvider) DownloadImage(ctx context.Context, imageName string) (*minio.Object, error) {
	const op = "storage.minioServer.DownloadImage"

	object, err := m.client.GetObject(
		ctx,
		m.bucket, // Константа с именем бакета
		imageName,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return object, nil
}

// RemoveImage - Удаляет файл в minio.
func (m *MinioProvider) RemoveImage(ctx context.Context, imageName string) error {
	const op = "storage.minioServer.RemoveImage"

	err := m.client.RemoveObject(
		ctx,
		m.bucket,
		imageName,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
