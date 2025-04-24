package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zboyco/s3ctl/internal/config"
)

func init() {
	// 添加子命令
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configUseCmd) // 添加 use 子命令
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理 S3 配置",
	Long:  `管理 S3 配置，包括初始化配置文件、查看配置信息等。`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "创建默认配置文件",
	Long:  `创建默认配置文件到 ~/.s3ctl，如果文件已存在则不会覆盖`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取用户主目录
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("获取用户主目录失败: %w", err)
		}

		// 设置配置文件路径
		configFile := filepath.Join(home, ".s3ctl")

		// 检查配置文件是否存在
		if _, err := os.Stat(configFile); err == nil {
			fmt.Printf("配置文件已存在: %s\n", configFile)
			return nil
		}

		// 创建默认配置文件
		if err := config.CreateDefaultConfig(configFile); err != nil {
			return fmt.Errorf("创建默认配置文件失败: %w", err)
		}

		fmt.Printf("已创建默认配置文件: %s，请修改后再使用\n", configFile)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "显示 S3 配置信息",
	Long:  `显示当前 S3 配置信息，包括端点、访问密钥等，并列出所有可用的配置名称。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取配置
		cfg, err := config.GetCurrentS3ConfigItem()
		if err != nil {
			fmt.Println("获取配置失败:", err)
			return
		}
		fmt.Printf("配置文件位置: %s\n\n", viper.ConfigFileUsed())

		// 获取并显示配置信息
		fmt.Println("当前 S3 配置信息:")
		fmt.Println("-------------------")
		fmt.Printf("端点 (Endpoint): %s\n", cfg.Endpoint)
		fmt.Printf("访问密钥 ID (AccessKey): %s\n", cfg.AccessKeyID)

		// 不直接显示 SecretKey，而是显示掩码后的字符串
		fmt.Printf("密钥 (SecretKey): %s\n", maskString(cfg.SecretAccessKey))

		fmt.Printf("使用 SSL: %v\n", cfg.UseSSL)

		// 获取所有配置名称
		allConfig, err := config.GetS3Config()
		if err != nil {
			fmt.Println("获取所有配置失败:", err)
			return
		}

		// 打印所有配置名称
		fmt.Println("\n所有可用配置:")
		fmt.Println("-------------------")
		for name := range allConfig.Services {
			// 标记当前使用的配置
			if name == allConfig.Current {
				fmt.Printf("* %s (当前使用)\n", name)
			} else {
				fmt.Printf("  %s\n", name)
			}
		}
	},
}

var configUseCmd = &cobra.Command{
	Use:   "use [配置名]",
	Short: "设置当前使用的配置",
	Long:  `设置当前使用的配置，切换到指定的配置名称。`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configName := args[0]

		// 获取所有配置
		s3Config, err := config.GetS3Config()
		if err != nil {
			return fmt.Errorf("获取配置失败: %w", err)
		}

		// 检查配置是否存在
		if _, ok := s3Config.Services[configName]; !ok {
			return fmt.Errorf("配置 '%s' 不存在，请使用 's3ctl config list' 查看可用的配置", configName)
		}

		// 设置当前配置
		s3Config.Current = configName

		// 将配置写入文件
		viper.Set("current", configName)
		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}

		fmt.Printf("已切换到配置: %s\n", configName)
		return nil
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
