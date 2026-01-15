package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Manager 構造体（メソッドを持たせるための箱）
type R2Manager struct {
	Client        *s3.Client
	PresignClient *s3.PresignClient
	BucketName    string
}

// 初期化関数
func NewR2Manager() (*R2Manager, error) {
	accountId := os.Getenv("R2_ACCOUNT_ID")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("R2_BUCKET_NAME")

	// R2のエンドポイントURLを作成
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId)

	// 固定の認証情報を作成
	creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")

	// 設定をロード
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(creds),
		config.WithRegion("auto"), // R2はリージョン指定不要（auto）
	)
	if err != nil {
		return nil, err
	}

	// S3クライアントを作成
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2Endpoint)
	})

	// 署名生成用のクライアントを作成
	presignClient := s3.NewPresignClient(client)

	return &R2Manager{
		Client:        client,
		PresignClient: presignClient,
		BucketName:    bucketName,
	}, nil
}

// GenerateUploadURL : アップロード用の署名付きURLを発行する
func (m *R2Manager) GenerateUploadURL(objectKey string) (string, error) {
	// 15分間だけ有効なPUT許可証を作成
	req, err := m.PresignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(m.BucketName),
		Key:    aws.String(objectKey),
		// ContentType: aws.String("video/mp4"), // 必要に応じて制限可能
	}, func(o *s3.PresignOptions) {
		o.Expires = 15 * time.Minute
	})

	if err != nil {
		return "", err
	}

	return req.URL, nil
}

// DeleteFile : R2上のファイルを削除する
func (m *R2Manager) DeleteFile(objectKey string) error {
	_, err := m.Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput {
		Bucket: aws.String(m.BucketName),
		Key: aws.String(objectKey),
	})
	return err
}
