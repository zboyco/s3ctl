package s3client

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zboyco/s3ctl/internal/config"
)

// Client S3 客户端
type Client struct {
	client *minio.Client
}

// NewClient 创建 S3 客户端
func NewClient() (*Client, error) {
	// 获取配置
	cfg, err := config.GetS3Config()
	if err != nil {
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}

	// 创建 minio 客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 S3 客户端失败: %w", err)
	}

	return &Client{
		client: client,
	}, nil
}

// ListBuckets 列出所有桶
func (c *Client) ListBuckets() ([]minio.BucketInfo, error) {
	buckets, err := c.client.ListBuckets(context.Background())
	if err != nil {
		return nil, fmt.Errorf("列出桶失败: %w", err)
	}
	return buckets, nil
}

// UploadFile 上传文件
func (c *Client) UploadFile(bucketName, filePath, objectName string, isPublic bool) error {
	fmt.Printf("正在上传文件 %s 到 %s/%s...\n", filePath, bucketName, objectName)
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 如果未指定对象名称，则使用文件名
	if objectName == "" {
		objectName = filepath.Base(filePath)
	}

	// 设置对象选项
	opts := minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	}

	// 如果是公开文件，设置权限
	if isPublic {
		opts.UserMetadata = map[string]string{"x-amz-acl": "public-read"}
	}

	// 上传文件
	_, err = c.client.PutObject(
		context.Background(),
		bucketName,
		objectName,
		file,
		fileInfo.Size(),
		opts,
	)
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	return nil
}

// UploadDirectory 上传目录
func (c *Client) UploadDirectory(bucketName, dirPath, prefix string, isPublic bool) error {
	// 检查目录是否存在
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("获取目录信息失败: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 不是一个目录", dirPath)
	}

	// 遍历目录
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 计算对象名称
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return fmt.Errorf("计算相对路径失败: %w", err)
		}

		// 替换 Windows 路径分隔符
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		// 构建对象名称
		objectName := relPath
		if prefix != "" {
			objectName = filepath.Join(prefix, relPath)
			// 替换 Windows 路径分隔符
			objectName = strings.ReplaceAll(objectName, "\\", "/")
		}

		// 上传文件
		return c.UploadFile(bucketName, path, objectName, isPublic)
	})
}

// GenerateURL 生成访问 URL
func (c *Client) GenerateURL(bucketName, objectName string, expires time.Duration, useV2 bool) (string, error) {
	// 检查对象是否存在
	_, err := c.client.StatObject(context.Background(), bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("对象不存在: %w", err)
	}

	var reqParams url.Values
	if useV2 {
		// minio-go/v7 不支持 V2 签名，使用 V4 代替并添加警告
		fmt.Println("警告: 当前版本不支持 V2 签名，将使用 V4 签名代替")
	}

	// 使用 V4 签名
	presignedURL, err := c.client.PresignedGetObject(context.Background(), bucketName, objectName, expires, reqParams)
	if err != nil {
		return "", fmt.Errorf("生成签名 URL 失败: %w", err)
	}
	return presignedURL.String(), nil
}

// ListObjects 列出对象
func (c *Client) ListObjects(bucketName, prefix string, recursive bool, onlyFolders bool) <-chan minio.ObjectInfo {
	ctx := context.Background()

	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
		UseV1:     true,
		MaxKeys:   1000,
	}

	objects := make(chan minio.ObjectInfo, 1)
	go func() {
		defer close(objects)

		for object := range c.client.ListObjects(ctx, bucketName, opts) {
			if object.Err != nil {
				objects <- object
				return
			}

			// 如果只查看文件夹，则跳过文件
			if onlyFolders && !strings.HasSuffix(object.Key, "/") {
				continue
			}

			objects <- object
		}
	}()

	return objects
}
