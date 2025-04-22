package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zboyco/s3ctl/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "创建默认配置文件",
	Long:  `创建默认配置文件到 ~/.s3ctl.yaml，如果文件已存在则不会覆盖`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取用户主目录
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("获取用户主目录失败: %w", err)
		}

		// 设置配置文件路径
		configFile := filepath.Join(home, ".s3ctl.yaml")

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
