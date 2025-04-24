package config

import (
	"os"
	"text/template"

	"github.com/spf13/viper"
)

// S3Config S3 配置
type S3Config struct {
	Current  string                   `mapstructure:"current"`
	Services map[string]*S3ConfigItem `mapstructure:"services"`
}

type S3ConfigItem struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
}

// GetS3Config 获取 S3 配置
func GetS3Config() (*S3Config, error) {
	var config S3Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// GetCurrentS3ConfigItem 获取当前使用的 S3 配置项
func GetCurrentS3ConfigItem() (*S3ConfigItem, error) {
	config, err := GetS3Config()
	if err != nil {
		return nil, err
	}

	// 如果没有设置当前配置或当前配置不存在，则使用第一个配置
	if config.Current == "" || config.Services[config.Current] == nil {
		// 遍历 map 获取第一个配置
		for name, item := range config.Services {
			config.Current = name
			return item, nil
		}
		// 如果没有任何配置项，返回错误
		return nil, nil
	}

	return config.Services[config.Current], nil
}

// CreateDefaultConfig 创建默认配置文件
func CreateDefaultConfig(configFile string) error {
	// 配置文件模板，支持多配置项
	const configTemplate = `# s3ctl 配置文件
current: default
services:
  default:
    endpoint: "play.min.io"
    access_key_id: "THISISKEYID"
    secret_access_key: "THISISSECRETKEY"
    use_ssl: true
  example:
    endpoint: "s3.example.com"
    access_key_id: "EXAMPLEKEYID"
    secret_access_key: "EXAMPLESECRETKEY"
    use_ssl: true
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
