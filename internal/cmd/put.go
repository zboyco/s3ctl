package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zboyco/s3ctl/internal/s3client"
	"github.com/zboyco/s3ctl/internal/utils"
)

var isPublic bool

var putCmd = &cobra.Command{
	Use:   "put [file/directory] [s3://bucketname/newpath/file.jpg]",
	Short: "上传文件或目录到 S3 存储",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 创建 S3 客户端
		client, err := s3client.NewClient(cmd.Context(), false)
		if err != nil {
			return err
		}

		// 获取文件或目录路径
		localPath := args[0]

		// 解析 S3 路径
		s3Path := args[1]
		bucketName, objectPath, err := utils.ParseS3Path(s3Path)
		if err != nil {
			return err
		}

		// 判断是文件还是目录
		isDir, err := isDirectory(localPath)
		if err != nil {
			return err
		}

		if isDir {
			// 上传目录
			fmt.Printf("正在上传目录 %s 到 %s/%s...\n", localPath, bucketName, objectPath)
			if err := client.UploadDirectory(bucketName, localPath, objectPath, isPublic); err != nil {
				return err
			}
			fmt.Println("目录上传成功")
		} else {
			// 上传文件
			if err := client.UploadFile(bucketName, localPath, objectPath, isPublic); err != nil {
				return err
			}
			fmt.Println("文件上传成功")
		}

		return nil
	},
}

func init() {
	putCmd.Flags().BoolVarP(&isPublic, "public", "p", false, "上传为公开文件")
}

// isDirectory 判断路径是否为目录
func isDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("获取路径信息失败: %w", err)
	}
	return info.IsDir(), nil
}
