package imgstorage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client        *minio.Client
	bucketName    string
	useSSL        bool
	publicAddress string
}

func New(address, user, password, bucket string, useSSL bool, publicAddress string) (*MinioClient, error) {
	cli, err := minio.New(address, &minio.Options{
		Creds:  credentials.NewStaticV4(user, password, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio create: %w", err)
	}

	return &MinioClient{
		client:        cli,
		bucketName:    bucket,
		useSSL:        useSSL,
		publicAddress: publicAddress,
	}, nil
}

func (m *MinioClient) ensureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return err
	}
	if !exists {
		return m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{})
	}
	return nil
}

func (m *MinioClient) Save(ctx context.Context, objectKey string, data []byte) (string, error) {
	if err := m.ensureBucket(ctx); err != nil {
		return "", fmt.Errorf("ensure bucket: %w", err)
	}

	reader := bytes.NewReader(data)
	_, err := m.client.PutObject(ctx, m.bucketName, objectKey, reader, int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	scheme := "http"
	if m.useSSL {
		scheme = "https"
	}
	publicURL := (&url.URL{
		Scheme: scheme,
		Host:   m.publicAddress,
		Path:   fmt.Sprintf("%s/%s", m.bucketName, objectKey),
	}).String()

	return publicURL, nil
}
