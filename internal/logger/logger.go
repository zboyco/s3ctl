package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	
	// 设置日志格式
	if os.Getenv("S3CTL_LOG_FORMAT") == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
	
	// 设置日志级别
	if level := os.Getenv("S3CTL_LOG_LEVEL"); level != "" {
		if l, err := logrus.ParseLevel(level); err == nil {
			log.SetLevel(l)
		}
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
	
	// 设置输出
	log.SetOutput(os.Stderr)
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	return log
}

// Info 记录信息日志
func Info(msg string, fields ...logrus.Fields) {
	entry := log.WithFields(mergeFields(fields...))
	entry.Info(msg)
}

// Debug 记录调试日志
func Debug(msg string, fields ...logrus.Fields) {
	entry := log.WithFields(mergeFields(fields...))
	entry.Debug(msg)
}

// Warn 记录警告日志
func Warn(msg string, fields ...logrus.Fields) {
	entry := log.WithFields(mergeFields(fields...))
	entry.Warn(msg)
}

// Error 记录错误日志
func Error(msg string, err error, fields ...logrus.Fields) {
	entry := log.WithFields(mergeFields(fields...)).WithError(err)
	entry.Error(msg)
}

// Fatal 记录致命错误日志并退出
func Fatal(msg string, err error, fields ...logrus.Fields) {
	entry := log.WithFields(mergeFields(fields...)).WithError(err)
	entry.Fatal(msg)
}

// WithFields 创建带字段的日志条目
func WithFields(fields logrus.Fields) *logrus.Entry {
	return log.WithFields(fields)
}

// WithField 创建带单个字段的日志条目
func WithField(key string, value interface{}) *logrus.Entry {
	return log.WithField(key, value)
}

// mergeFields 合并多个字段映射
func mergeFields(fields ...logrus.Fields) logrus.Fields {
	result := logrus.Fields{}
	for _, f := range fields {
		for k, v := range f {
			result[k] = v
		}
	}
	return result
}

// S3Operation 记录 S3 操作日志
func S3Operation(operation, bucket, object string, err error) {
	fields := logrus.Fields{
		"operation": operation,
		"bucket":    bucket,
	}
	
	if object != "" {
		fields["object"] = object
	}
	
	if err != nil {
		log.WithFields(fields).WithError(err).Error("S3 operation failed")
	} else {
		log.WithFields(fields).Info("S3 operation completed")
	}
}

// ConfigOperation 记录配置操作日志
func ConfigOperation(operation string, config interface{}, err error) {
	fields := logrus.Fields{
		"operation": operation,
		"config":    config,
	}
	
	if err != nil {
		log.WithFields(fields).WithError(err).Error("Config operation failed")
	} else {
		log.WithFields(fields).Info("Config operation completed")
	}
}

// Performance 记录性能指标
func Performance(operation string, duration interface{}, fields ...logrus.Fields) {
	entry := log.WithFields(mergeFields(fields...)).WithFields(logrus.Fields{
		"operation": operation,
		"duration":  duration,
	})
	entry.Info("Performance metric")
}
