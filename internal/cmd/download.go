package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zboyco/s3ctl/internal/s3client"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download <s3://bucket/path> [local-path]",
	Short: "下载 S3 对象或目录",
	Long: `从 S3 中下载对象或目录到本地文件系统。

示例:
  下载单个文件到当前目录
  s3ctl download s3://mybucket/path/to/file.txt

  下载单个文件到指定目录
  s3ctl download s3://mybucket/path/to/file.txt ./local/dir/
  
  下载单个文件到指定路径并指定文件名
  s3ctl download s3://mybucket/path/to/file.txt ./local/path/new-filename.txt

  下载目录到指定目录
  s3ctl download s3://mybucket/path/to/dir/ ./local/dir/
`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 创建 S3 客户端
		client, err := s3client.NewClient(cmd.Context(), useV2)
		if err != nil {
			return err
		}

		// 解析桶名和对象路径
		input := args[0]
		var bucketName, objectPath string
		if strings.HasPrefix(input, "s3://") {
			trimmedInput := strings.TrimPrefix(input, "s3://")
			parts := strings.SplitN(trimmedInput, "/", 2)
			bucketName = parts[0]
			if len(parts) > 1 {
				objectPath = parts[1]
			}
		} else {
			return fmt.Errorf("无效的路径格式，请使用 s3://bucket/object")
		}

		// 确定本地路径
		localPath := "."
		if len(args) == 2 {
			localPath = args[1]
		}

		// 确保目录存在
		if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}

		// 判断是文件还是目录
		isDir := strings.HasSuffix(objectPath, "/")

		if isDir {
			// 下载目录
			if err := client.DownloadDirectory(bucketName, objectPath, localPath); err != nil {
				return err
			}
			fmt.Printf("目录下载成功")
		} else {
			// 下载文件

			// 如果 localPath 是一个目录，则使用对象名称作为文件名。
			if info, err := os.Stat(localPath); err == nil && info.IsDir() {
				fileName := filepath.Base(objectPath)
				localPath = filepath.Join(localPath, fileName)
			}

			if err := client.DownloadFile(bucketName, objectPath, localPath); err != nil {
				return err
			}
			fmt.Printf("文件下载成功")
		}
		return nil
	},
}
