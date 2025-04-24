package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zboyco/s3ctl/internal/s3client"
)

var mbCmd = &cobra.Command{
	Use:   "mb s3://bucketname",
	Short: "创建 S3 存储桶",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 创建 S3 客户端
		client, err := s3client.NewClient(cmd.Context(), false)
		if err != nil {
			return err
		}

		// 解析 S3 路径
		s3Path := args[0]
		var bucketName string
		if strings.HasPrefix(s3Path, "s3://") {
			bucketName = strings.TrimPrefix(s3Path, "s3://")
			// 移除末尾的 / (如果有)
			bucketName = strings.TrimSuffix(bucketName, "/")
			if strings.Contains(bucketName, "/") {
				return fmt.Errorf("无效的存储桶名称，不能包含 '/'")
			}
		} else {
			return fmt.Errorf("无效的 S3 路径格式，请使用 s3://bucketname")
		}

		if bucketName == "" {
			return fmt.Errorf("存储桶名称不能为空")
		}

		// 创建存储桶
		if err := client.MakeBucket(bucketName); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	// mbCmd 不需要额外的 flags，但 init 函数是必需的
}
