package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zboyco/s3ctl/internal/s3client"
	"github.com/zboyco/s3ctl/internal/utils"
)

var delCmd = &cobra.Command{
	Use:   "del [s3://bucketname/path/file]",
	Short: "删除 S3 存储中的对象",
	Long:  `删除指定的 S3 对象或文件夹。`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 创建 S3 客户端
		client, err := s3client.NewClient(cmd.Context(), false)
		if err != nil {
			return err
		}

		// 解析 S3 路径
		s3Path := args[0]
		bucketName, objectPath, err := utils.ParseS3Path(s3Path)
		if err != nil {
			return err
		}

		// 检查对象是否为文件夹
		if client.IsDirectory(s3Path) {
			// 递归删除文件夹中的所有对象
			fmt.Printf("正在递归删除文件夹 %s/%s...\n", bucketName, objectPath)
			if err := client.DeleteDirectory(bucketName, objectPath); err != nil {
				return err
			}
			fmt.Println("文件夹删除成功")
		} else {
			// 删除单个对象
			if err := client.DeleteObject(bucketName, objectPath); err != nil {
				return err
			}
			fmt.Println("对象删除成功")
		}

		return nil
	},
}
