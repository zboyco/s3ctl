package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/cobra"
	"github.com/zboyco/s3ctl/internal/s3client"
)

var (
	recursive    bool
	onlyFolders  bool
	showFullPath bool // 新增的布尔标志参数
)

var listCmd = &cobra.Command{
	Use:   "ls [s3://bucket/prefix]",
	Short: "列出 S3 存储桶或对象",
	Long: `列出 S3 存储桶或指定桶中的对象。
- 不带参数时列出所有存储桶
- 指定桶名时列出该桶中的对象
- 可选指定前缀筛选对象`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 创建 S3 客户端
		client, err := s3client.NewClient(cmd.Context(), false)
		if err != nil {
			return err
		}

		// 如果没有参数，列出所有桶
		if len(args) == 0 {
			return listAllBuckets(client, "")
		}

		input := args[0]

		if !strings.HasPrefix(input, "s3://") {
			return errors.New("参数格式错误, 请使用 s3://bucket/prefix 格式")
		}

		// 如果除了s3://,没有一个/，则认为是桶名，过滤桶
		if !strings.Contains(strings.Replace(input, "s3://", "", 1), "/") {
			return listAllBuckets(client, input)
		}

		// 解析桶名和前缀
		var bucketName, prefix string
		if strings.HasPrefix(input, "s3://") {
			trimmedInput := strings.TrimPrefix(input, "s3://")
			parts := strings.SplitN(trimmedInput, "/", 2)
			bucketName = parts[0]
			if len(parts) > 1 {
				prefix = parts[1]
			}
		} else {
			bucketName = input
		}

		// 列出桶中的对象
		err = listBucketObjects(client, bucketName, prefix, recursive, onlyFolders, showFullPath)
		if minio.ToErrorResponse(err).Code == "NoSuchBucket" {
			fmt.Printf("存储桶 %s 不存在\n", bucketName)
			return nil
		}
		return err
	},
}

func init() {
	listCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "递归列出所有对象")
	listCmd.Flags().BoolVarP(&onlyFolders, "folders", "f", false, "只列出文件夹")
	listCmd.Flags().BoolVarP(&showFullPath, "full-path", "p", false, "显示完整路径") // 注册新参数
}

// listAllBuckets 列出所有桶
func listAllBuckets(client *s3client.Client, prefix string) error {
	// 列出所有桶
	buckets, err := client.ListBuckets()
	if err != nil {
		return err
	}

	// 输出桶列表
	if len(buckets) == 0 {
		fmt.Println("没有找到存储桶")
		return nil
	}

	for _, bucket := range buckets {
		// 添加 s3:// 前缀
		bucketName := fmt.Sprintf("s3://%s/", bucket.Name)
		if strings.HasPrefix(bucketName, prefix) {
			// 打印修改时间，大小，路径，占用固定宽度
			fmt.Printf("%-22s %-11s %s\n", "", "BUCKET", bucketName)
		}
	}

	return nil
}

// listBucketObjects 列出桶中的对象
func listBucketObjects(client *s3client.Client, bucketName, prefix string, recursive, onlyFolders, showFullPath bool) error {
	fullPrefix := fmt.Sprintf("s3://%s/%s", bucketName, prefix)
	if !strings.HasSuffix(fullPrefix, "/") {
		// fullPrefix 保留最后一个 /前的部分
		fullPrefix = fullPrefix[:strings.LastIndex(fullPrefix, "/")+1]
	}
	// 检查桶是否存在

	// 列出对象
	objects := client.ListObjects(bucketName, prefix, recursive, onlyFolders)

	for object := range objects {
		if object.Err != nil {
			// fmt.Printf("列出对象失败: %v\n", object.Err)
			return object.Err
		}

		if object.Key[len(object.Key)-1] != '/' || object.Key != prefix {

			// 完整路径
			fullPath := fmt.Sprintf("s3://%s/%s", bucketName, object.Key)

			// 不显示完整路径
			if !showFullPath {
				fullPath = strings.Replace(fullPath, fullPrefix, "", 1)
			}

			date := ""
			size := "DIR"
			if !client.IsDirectory(object.Key) {
				size = formatSize(object.Size)
				date = object.LastModified.Format("2006-01-02 15:04:05")
			}

			// 打印修改时间，大小，路径，占用固定宽度
			fmt.Printf("%-22s %-11s %s\n", date, size, fullPath)
		}
	}

	return nil
}

// 大小转换为带单位的字符串
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
