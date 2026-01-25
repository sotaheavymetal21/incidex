package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	DefaultBucketName = "incidex-attachments"
)

type MinIOStorage struct {
	client     *minio.Client
	bucketName string
}

func NewMinIOStorage(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("MinIOクライアントの作成に失敗しました: %w", err)
	}

	storage := &MinIOStorage{
		client:     client,
		bucketName: bucketName,
	}

	// バケットが存在しない場合は作成します
	if err := storage.ensureBucketExists(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *MinIOStorage) ensureBucketExists() error {
	ctx := context.Background()
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("バケットの存在確認に失敗しました: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("バケットの作成に失敗しました: %w", err)
		}
		log.Printf("バケット %s が正常に作成されました", s.bucketName)
	}

	return nil
}

func (s *MinIOStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, fileSize int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucketName, objectKey, reader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("オブジェクトのアップロードに失敗しました: %w", err)
	}
	return nil
}

func (s *MinIOStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("オブジェクトの取得に失敗しました: %w", err)
	}
	return object, nil
}

func (s *MinIOStorage) Delete(ctx context.Context, objectKey string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("オブジェクトの削除に失敗しました: %w", err)
	}
	return nil
}

func (s *MinIOStorage) GetObjectInfo(ctx context.Context, objectKey string) (minio.ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, s.bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("オブジェクト情報の取得に失敗しました: %w", err)
	}
	return info, nil
}
