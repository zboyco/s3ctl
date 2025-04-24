package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zboyco/s3ctl/internal/s3client"
)

var (
	expiry time.Duration
	useV2  bool
)

var urlCmd = &cobra.Command{
	Use:   "url [s3://bucket/object]",
	Short: "生成对象的访问 URL",
	Long: `生成指定对象的访问 URL。
- 指定对象路径时生成访问 URL
- 可选设置 URL 有效期和签名协议`,
	Args: cobra.ExactArgs(1),
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

		// 生成访问 URL
		// 确保 GenerateURL 方法的参数与签名匹配
		url, err := client.GenerateURL(bucketName, objectPath, expiry)
		if err != nil {
			return err
		}

		fmt.Println(url)
		return nil
	},
}

func init() {
	urlCmd.Flags().DurationVarP(&expiry, "expiry", "e", 24*time.Hour, "URL 有效期（例如：1h, 24h, 7d）")
	urlCmd.Flags().BoolVarP(&useV2, "v2", "2", false, "使用 V2 签名协议（默认使用 V4）")
}
