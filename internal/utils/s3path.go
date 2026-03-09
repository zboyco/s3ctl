package utils

import (
	"fmt"
	"strings"
)

// ParseS3Path 解析 S3 路径格式 s3://bucket/object
func ParseS3Path(s3Path string) (bucket, object string, err error) {
	if !strings.HasPrefix(s3Path, "s3://") {
		return "", "", fmt.Errorf("无效的 S3 路径格式: %s\n正确格式: s3://bucketname/path/file\n示例: s3://mybucket/photos/image.jpg", s3Path)
	}
	
	trimmed := strings.TrimPrefix(s3Path, "s3://")
	parts := strings.SplitN(trimmed, "/", 2)
	
	bucket = parts[0]
	if bucket == "" {
		return "", "", fmt.Errorf("存储桶名称不能为空")
	}
	
	// 验证存储桶名称格式
	if err := validateBucketName(bucket); err != nil {
		return "", "", fmt.Errorf("无效的存储桶名称: %w", err)
	}
	
	if len(parts) > 1 {
		object = parts[1]
	}
	
	return bucket, object, nil
}

// ParseS3BucketPath 解析存储桶路径，确保不包含对象路径
func ParseS3BucketPath(s3Path string) (bucket string, err error) {
	if !strings.HasPrefix(s3Path, "s3://") {
		return "", fmt.Errorf("无效的 S3 路径格式，请使用 s3://bucketname")
	}
	
	bucket = strings.TrimPrefix(s3Path, "s3://")
	// 移除末尾的 / (如果有)
	bucket = strings.TrimSuffix(bucket, "/")
	
	if strings.Contains(bucket, "/") {
		return "", fmt.Errorf("无效的存储桶名称，不能包含 '/'")
	}
	
	if bucket == "" {
		return "", fmt.Errorf("存储桶名称不能为空")
	}
	
	// 验证存储桶名称格式
	if err := validateBucketName(bucket); err != nil {
		return "", fmt.Errorf("无效的存储桶名称: %w", err)
	}
	
	return bucket, nil
}

// validateBucketName 验证存储桶名称是否符合规范
func validateBucketName(bucket string) error {
	if len(bucket) < 3 || len(bucket) > 63 {
		return fmt.Errorf("存储桶名称长度必须在 3-63 个字符之间")
	}
	
	// 简单的名称验证（可以根据需要扩展）
	if strings.HasPrefix(bucket, "-") || strings.HasSuffix(bucket, "-") {
		return fmt.Errorf("存储桶名称不能以连字符开头或结尾")
	}
	
	return nil
}
