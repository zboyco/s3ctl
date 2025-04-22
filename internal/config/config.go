package config

import (
	"os"
	"text/template"

	"github.com/spf13/viper"
)

// S3Config S3 配置
type S3Config struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	// Bucket 字段是可选的
	Bucket string `mapstructure:"bucket"`
}

// GetS3Config 获取 S3 配置
func GetS3Config() (*S3Config, error) {
	var config S3Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateDefaultConfig 创建默认配置文件
func CreateDefaultConfig(configFile string) error {
	// 配置文件模板，移除了 bucket 字段
	const configTemplate = `# s3ctl 配置文件
endpoint: "play.min.io"
access_key_id: "Q3AM3UQ867SPQQA43P2F"
secret_access_key: "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
use_ssl: true
# bucket: "your-bucket-name"  # 可选，不指定则列出所有桶
`

	// 解析模板
	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return err
	}

	// 创建配置文件
	file, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入模板内容
	return tmpl.Execute(file, nil)
}
