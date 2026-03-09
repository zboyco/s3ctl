package s3client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockS3Client 模拟 S3 客户端用于测试
type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) ListObjects(bucketName, prefix string, recursive bool, onlyFolders bool, maxKeys ...int) <-chan ObjectInfo {
	args := m.Called(bucketName, prefix, recursive, onlyFolders, maxKeys)
	return args.Get(0).(<-chan ObjectInfo)
}

// ObjectInfo 测试用的对象信息结构
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	Err          error
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
		wantErr  bool
	}{
		{
			name:     "normal path",
			path:     "folder/file.txt",
			expected: "folder/file.txt",
			wantErr:  false,
		},
		{
			name:     "path with dot dot",
			path:     "../../../etc/passwd",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "absolute path",
			path:     "/etc/passwd",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "current directory",
			path:     "./file.txt",
			expected: "file.txt",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizePath(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestProgressReader(t *testing.T) {
	t.Run("progress tracking", func(t *testing.T) {
		totalSize := int64(1000)
		pr := newProgressReader(totalSize)

		assert.Equal(t, totalSize, pr.totalSize)
		assert.Equal(t, int64(0), pr.bytesRead)
		assert.Equal(t, int64(DefaultProgressPercent), pr.printPercent)
		assert.Equal(t, DefaultMinBarWidth, pr.minBarWidth)
	})

	t.Run("concurrent safety", func(t *testing.T) {
		pr := newProgressReader(1000)

		// 模拟并发读取
		done := make(chan bool, 2)

		go func() {
			for i := 0; i < 50; i++ {
				pr.Read(make([]byte, 10))
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 50; i++ {
				pr.Write(make([]byte, 10))
			}
			done <- true
		}()

		// 等待两个 goroutine 完成
		<-done
		<-done

		// 验证最终状态 - 由于并发安全，应该正确累加
		assert.Equal(t, int64(1000), pr.bytesRead) // 50*10 + 50*10
	})
}
