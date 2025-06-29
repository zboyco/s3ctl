package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zboyco/s3ctl" // 导入 s3ctl 包
)

var rootCmd = &cobra.Command{
	Use:   "s3ctl",
	Short: "s3ctl 是一个用于操作 S3 存储的命令行工具",
	Long: `s3ctl 是一个基于 minio-go 的命令行工具，用于操作 S3 兼容的对象存储。
支持文件上传、生成访问链接和查看文件列表等功能。`,
	// 添加 Version 字段，Cobra 会自动处理 --version 标志
	// Version: s3ctl.Version(), // 直接在这里设置也可以，但通常在 init 中设置
}

// Execute 执行根命令
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	// 设置版本号
	rootCmd.Version = s3ctl.Version() // 从 s3ctl 包获取版本信息

	// 添加子命令
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(urlCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(delCmd)
	rootCmd.AddCommand(mbCmd)
	rootCmd.AddCommand(rbCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(downloadCmd)

	// 禁用 help 和 completion 命令
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	// 禁用自动生成的 completion 命令
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// 添加根命令的 PersistentPreRunE 函数
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// 如果是 config init 命令，跳过配置文件检查
		if cmd.Name() == "config" && len(args) > 0 && args[0] == "init" {
			return nil
		}
		// 如果是 init 子命令，跳过配置文件检查
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
	configFile := filepath.Join(home, ".s3ctl")
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，提示用户使用 init 命令创建
			return fmt.Errorf("配置文件不存在: %s\n请使用 's3ctl config init' 命令创建默认配置文件", configFile)
		} else if pathErr, ok := err.(*os.PathError); ok {
			// 检查是否是因为文件不存在
			if os.IsNotExist(pathErr) {
				return fmt.Errorf("配置文件不存在: %s\n请使用 's3ctl config init' 命令创建默认配置文件", configFile)
			}
			return fmt.Errorf("读取配置文件失败: %w", err)
		} else {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	return nil
}
