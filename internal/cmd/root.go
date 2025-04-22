package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "s3ctl",
	Short: "s3ctl 是一个用于操作 S3 存储的命令行工具",
	Long: `s3ctl 是一个基于 minio-go 的命令行工具，用于操作 S3 兼容的对象存储。
支持文件上传、生成访问链接和查看文件列表等功能。`,
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 添加子命令
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(urlCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(delCmd)

	// 添加根命令的 PersistentPreRunE 函数
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// 如果是 init 命令，跳过配置文件检查
		if cmd.Name() == "init" {
			return nil
		}

		return initConfig()
	}
}

// initConfig 读取配置文件
func initConfig() error {
	// 获取用户主目录
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败: %w", err)
	}

	// 设置配置文件路径
	configFile := filepath.Join(home, ".s3ctl.yaml")
	viper.SetConfigFile(configFile)

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，提示用户使用 init 命令创建
			return fmt.Errorf("配置文件不存在: %s\n请使用 's3ctl init' 命令创建默认配置文件", configFile)
		} else {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	return nil
}
