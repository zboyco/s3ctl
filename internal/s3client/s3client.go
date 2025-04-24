package s3client

import (
	"fmt"

	"github.com/minio/minio-go/v7"
)

// IsDirectory 检查路径是否为目录
func (c *Client) IsDirectory(objectPath string) bool {
	if objectPath == "" {
		return false
	}
	// path 结尾必须是 /
	if objectPath[len(objectPath)-1] == '/' {
		return true
	}
	return false
}

// DeleteDirectory 递归删除目录下的所有对象
func (c *Client) DeleteDirectory(bucketName, objectPath string) error {
	objects := c.ListObjects(bucketName, objectPath, true, false)

	for object := range objects {
		if object.Err != nil {
			return object.Err
		}

		if err := c.DeleteObject(bucketName, object.Key); err != nil {
			return err
		}
	}
	return nil
}

// DeleteObject 删除指定对象
func (c *Client) DeleteObject(bucketName, objectPath string) error {
	if objectPath == "" {
		return fmt.Errorf("对象路径不能为空")
	}
	// 实现删除对象的逻辑
	fmt.Printf("正在删除对象 %s/%s...\n", bucketName, objectPath)

	// 检查参数
	if bucketName == "" {
		return fmt.Errorf("桶名称不能为空")
	}

	if objectPath == "" {
		return fmt.Errorf("对象路径不能为空")
	}

	// 使用 minio 客户端删除对象
	err := c.client.RemoveObject(c.ctx, bucketName, objectPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("删除对象失败: %w", err)
	}

	return nil
}
