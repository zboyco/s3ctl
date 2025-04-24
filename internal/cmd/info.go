package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "显示当前 S3 配置信息",
	Long:  `显示当前 S3 配置信息，包括端点、访问密钥等。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 确保配置已加载
		if err := initConfig(); err != nil {
			fmt.Println(err)
			return
		}

		// 获取并显示配置信息
		fmt.Println("当前 S3 配置信息:")
		fmt.Println("-------------------")
		fmt.Printf("端点 (Endpoint): %s\n", viper.GetString("endpoint"))
		fmt.Printf("访问密钥 ID (AccessKey): %s\n", viper.GetString("access_key_id"))

		// 不直接显示 SecretKey，而是显示掩码
		secretKey := viper.GetString("secret_access_key")
		maskedKey := maskString(secretKey)
		fmt.Printf("密钥 (SecretKey): %s\n", maskedKey)

		fmt.Printf("使用 SSL: %v\n", viper.GetBool("useSSL"))
		fmt.Printf("配置文件位置: %s\n", viper.ConfigFileUsed())
	},
}

// maskString 将字符串中间部分替换为星号，保留首尾字符
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}

	// 保留前两个和后两个字符，中间用星号代替
	masked := s[:2]
	for i := 0; i < len(s)-4; i++ {
		masked += "*"
	}
	masked += s[len(s)-2:]

	return masked
}
