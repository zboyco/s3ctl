package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zboyco/s3ctl/internal/s3client"
	"github.com/zboyco/s3ctl/internal/utils"
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
		bucketName, err := utils.ParseS3BucketPath(s3Path)
		if err != nil {
			return err
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
