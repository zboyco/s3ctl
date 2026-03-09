package config

import (
	"fmt"
	"os"
	"text/template"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// S3Config S3 配置
type S3Config struct {
	Current  string                   `mapstructure:"current"`
	Services map[string]*S3ConfigItem `mapstructure:"services"`
}

type S3ConfigItem struct {
	Endpoint        string `mapstructure:"endpoint" validate:"required,url"`
	AccessKeyID     string `mapstructure:"access_key_id" validate:"required,min=3"`
	SecretAccessKey string `mapstructure:"secret_access_key" validate:"required,min=8"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	Region          string `mapstructure:"region"`
	Timeout         int    `mapstructure:"timeout" validate:"min=1,max=300"`
}

// Validate 验证配置项
func (c *S3ConfigItem) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
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
			// 支持环境变量覆盖
			if endpoint := os.Getenv("S3CTL_ENDPOINT"); endpoint != "" {
				item.Endpoint = endpoint
			}
			if accessKey := os.Getenv("S3CTL_ACCESS_KEY"); accessKey != "" {
				item.AccessKeyID = accessKey
			}
			if secretKey := os.Getenv("S3CTL_SECRET_KEY"); secretKey != "" {
				item.SecretAccessKey = secretKey
			}
			if region := os.Getenv("S3CTL_REGION"); region != "" {
				item.Region = region
			}

			// 验证配置
			if err := item.Validate(); err != nil {
				return nil, fmt.Errorf("配置验证失败: %w", err)
			}

			return item, nil
		}
		// 如果没有任何配置项，返回错误
		return nil, nil
	}

	item := config.Services[config.Current]

	// 支持环境变量覆盖
	if endpoint := os.Getenv("S3CTL_ENDPOINT"); endpoint != "" {
		item.Endpoint = endpoint
	}
	if accessKey := os.Getenv("S3CTL_ACCESS_KEY"); accessKey != "" {
		item.AccessKeyID = accessKey
	}
	if secretKey := os.Getenv("S3CTL_SECRET_KEY"); secretKey != "" {
		item.SecretAccessKey = secretKey
	}
	if region := os.Getenv("S3CTL_REGION"); region != "" {
		item.Region = region
	}

	// 验证配置
	if err := item.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return item, nil
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

	// 创建配置文件，设置安全权限（仅用户可读写）
	file, err := os.OpenFile(configFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入模板内容
	return tmpl.Execute(file, nil)
}
