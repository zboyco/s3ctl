package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zboyco/s3ctl/internal/s3client"
	"github.com/zboyco/s3ctl/internal/utils"
)

var rbCmd = &cobra.Command{
	Use:   "rb s3://bucketname",
	Short: "删除 S3 存储桶",
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

		// 删除存储桶 (客户端方法内部会检查是否为空)
		if err := client.RemoveBucket(bucketName); err != nil {
			return err
		}

		fmt.Printf("存储桶 %s 删除成功\n", bucketName)
		return nil
	},
}
