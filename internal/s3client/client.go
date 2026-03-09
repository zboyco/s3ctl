package s3client

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zboyco/s3ctl/internal/config"
	"golang.org/x/term"
)

// 常量定义
const (
	DefaultMaxKeys         = 1000
	DefaultProgressPercent = 1
	DefaultMinBarWidth     = 20
	DefaultBufferSize      = 100
	MaxBufferSize          = 1000
)

// Client S3 客户端
type Client struct {
	client *minio.Client

	ctx context.Context
}

// NewClient 创建 S3 客户端
func NewClient(ctx context.Context, v2 bool) (*Client, error) {
	// 获取配置
	cfg, err := config.GetCurrentS3ConfigItem()
	if err != nil {
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}

	var creds *credentials.Credentials
	if v2 {
		creds = credentials.NewStaticV2(cfg.AccessKeyID, cfg.SecretAccessKey, "")
	} else {
		creds = credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, "")
	}

	// 创建 minio 客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  creds,
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 S3 客户端失败: %w", err)
	}

	return &Client{
		client: client,
		ctx:    ctx,
	}, nil
}

// ListBuckets 列出所有桶
func (c *Client) ListBuckets() ([]minio.BucketInfo, error) {
	buckets, err := c.client.ListBuckets(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("列出桶失败: %w", err)
	}
	return buckets, nil
}

// MakeBucket 创建存储桶
func (c *Client) MakeBucket(bucketName string) error {
	err := c.client.MakeBucket(c.ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// 检查桶是否已存在
		exists, errBucketExists := c.client.BucketExists(c.ctx, bucketName)
		if errBucketExists == nil && exists {
			fmt.Printf("存储桶 '%s' 已存在\n", bucketName)
			return nil
		}
		return fmt.Errorf("创建存储桶失败: %w", err)
	}
	fmt.Printf("存储桶 '%s' 创建成功\n", bucketName)
	return nil
}

// RemoveBucket 删除存储桶
func (c *Client) RemoveBucket(bucketName string) error {
	// 检查桶是否为空
	isEmpty, err := c.IsBucketEmpty(bucketName)
	if err != nil {
		return fmt.Errorf("检查存储桶状态失败: %w", err)
	}
	if !isEmpty {
		fmt.Printf("存储桶 '%s' 不为空，无法删除\n", bucketName)
		return nil
	}

	err = c.client.RemoveBucket(c.ctx, bucketName)
	if err != nil {
		return fmt.Errorf("删除存储桶失败: %w", err)
	}
	fmt.Printf("存储桶 '%s' 删除成功\n", bucketName)
	return nil
}

// IsBucketEmpty 检查存储桶是否为空
func (c *Client) IsBucketEmpty(bucketName string) (bool, error) {
	// 使用 ListObjects 只获取一个对象来判断是否为空
	objectsCh := c.ListObjects(bucketName, "", false, false, 1)

	select {
	case obj, ok := <-objectsCh:
		if !ok {
			// 通道已关闭，桶为空
			return true, nil
		}
		if obj.Err != nil {
			return false, fmt.Errorf("检查存储桶状态失败: %w", obj.Err)
		}
		// 有对象，桶不为空
		return false, nil
	case <-time.After(5 * time.Second):
		return false, fmt.Errorf("检查存储桶状态超时")
	}
}

// UploadFile 上传文件
func (c *Client) UploadFile(bucketName, filePath, objectName string, isPublic bool) error {
	fmt.Printf("上传 %s 到 %s/%s...\n", filePath, bucketName, objectName)
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 如果未指定对象名称，则使用文件名
	if objectName == "" {
		objectName = filepath.Base(filePath)
	}

	// 设置对象选项
	opts := minio.PutObjectOptions{
		ContentType: getContentType(filePath),
	}

	// 如果是公开文件，设置权限
	if isPublic {
		opts.UserMetadata = map[string]string{"x-amz-acl": "public-read"}
	}

	// 添加上传进度跟踪
	opts.Progress = newProgressReader(fileInfo.Size())

	// 上传文件
	_, err = c.client.PutObject(
		c.ctx,
		bucketName,
		objectName,
		file,
		fileInfo.Size(),
		opts,
	)
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	return nil
}

// UploadDirectoryConcurrent 并发上传目录中的所有文件
func (c *Client) UploadDirectoryConcurrent(bucketName, dirPath, prefix string, isPublic bool, maxWorkers int) error {
	if maxWorkers <= 0 {
		maxWorkers = 4 // 默认 4 个工作协程
	}

	files := make(chan string, 100)
	errors := make(chan error, maxWorkers)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range files {
				if err := c.uploadSingleFile(bucketName, filePath, dirPath, prefix, isPublic); err != nil {
					errors <- err
					return
				}
			}
		}()
	}

	// 遍历目录并发送文件路径
	go func() {
		defer close(files)
		filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			files <- path
			return nil
		})
	}()

	// 等待完成
	go func() {
		wg.Wait()
		close(errors)
	}()

	// 检查错误
	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

// uploadSingleFile 上传单个文件的辅助方法
func (c *Client) uploadSingleFile(bucketName, filePath, dirPath, prefix string, isPublic bool) error {
	// 计算对象名称
	relPath, err := filepath.Rel(dirPath, filePath)
	if err != nil {
		return fmt.Errorf("计算相对路径失败: %w", err)
	}

	// 安全检查路径
	safeRelPath, err := sanitizePath(relPath)
	if err != nil {
		return fmt.Errorf("路径安全检查失败: %w", err)
	}

	// 替换 Windows 路径分隔符
	safeRelPath = strings.ReplaceAll(safeRelPath, "\\", "/")

	// 构建对象名称
	objectName := safeRelPath
	if prefix != "" {
		objectName = filepath.Join(prefix, safeRelPath)
		// 替换 Windows 路径分隔符
		objectName = strings.ReplaceAll(objectName, "\\", "/")
	}

	return c.UploadFile(bucketName, filePath, objectName, isPublic)
}

// getContentType 根据文件扩展名获取对应的 Content-Type
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}
	return "application/octet-stream"
}

// sanitizePath 清理路径，防止路径遍历攻击
func sanitizePath(path string) (string, error) {
	// 清理路径，防止路径遍历
	cleaned := filepath.Clean(path)

	// 检查是否包含危险字符
	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("路径包含非法字符: %s", path)
	}

	// 确保路径不以 / 开头（相对路径）
	if strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("不允许绝对路径: %s", path)
	}

	return cleaned, nil
}

// progressReader 实现进度跟踪
type progressReader struct {
	mu           sync.Mutex
	totalSize    int64
	bytesRead    int64
	lastPrint    int64
	printPercent int64
	minBarWidth  int       // 最小进度条宽度
	completed    bool      // 标记是否已完成
	startTime    time.Time // 开始时间
	lastBytes    int64     // 上次统计的字节数
	lastTime     time.Time // 上次统计的时间
}

func newProgressReader(totalSize int64) *progressReader {
	return &progressReader{
		totalSize:    totalSize,
		printPercent: DefaultProgressPercent,
		minBarWidth:  DefaultMinBarWidth,
		startTime:    time.Now(),
		lastTime:     time.Now(),
	}
}

func (pr *progressReader) getTerminalWidth() int {
	// 尝试获取终端宽度
	if width, err := getTerminalWidth(); err == nil {
		// 确保不小于最小宽度
		if width < pr.minBarWidth {
			return pr.minBarWidth
		}
		return width
	}
	// 如果获取失败，使用默认宽度
	return 50
}

func getTerminalWidth() (int, error) {
	// 在Unix-like系统上获取终端宽度
	if fd := int(os.Stdout.Fd()); term.IsTerminal(fd) {
		width, _, err := term.GetSize(fd)
		if err != nil {
			return 0, err
		}
		return width, nil
	}
	return 0, fmt.Errorf("not a terminal")
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	// 如果已经完成，直接返回
	if pr.completed {
		return 0, io.EOF
	}

	// 模拟读取数据
	n = len(p)
	if n > 0 {
		pr.bytesRead += int64(n)
		pr.updateProgress()
	}
	return n, nil
}

func (pr *progressReader) Write(p []byte) (n int, err error) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	// 如果已经完成，直接返回
	if pr.completed {
		return 0, io.EOF
	}

	n = len(p)
	pr.bytesRead += int64(n)
	pr.updateProgress()
	return n, nil
}

func (pr *progressReader) updateProgress() {
	// 计算当前进度百分比
	percent := int64(float64(pr.bytesRead) / float64(pr.totalSize) * 100)

	// 如果进度达到或超过了下一个打印点
	if percent >= pr.lastPrint+pr.printPercent || percent == 100 {
		now := time.Now()
		elapsed := now.Sub(pr.lastTime).Seconds()

		// 计算上传速度 (bytes/sec)
		speed := float64(pr.bytesRead-pr.lastBytes) / elapsed

		// 计算剩余时间
		remainingBytes := pr.totalSize - pr.bytesRead
		var remainingTime time.Duration
		if speed > 0 {
			remainingTime = time.Duration(float64(remainingBytes)/speed) * time.Second
		}

		// 获取动态宽度，留出空间给百分比和字节数显示
		barWidth := max(pr.getTerminalWidth()-58, 10)

		// 计算进度条填充长度
		filled := int(float64(barWidth) * float64(percent) / 100)
		empty := barWidth - filled

		// 构建进度条字符串
		bar := "[" + strings.Repeat("=", filled) + strings.Repeat(" ", empty) + "]"

		// 打印进度信息
		fmt.Printf("\r%s %3d%% %-23s %-12s ETA:%-8s",
			bar,
			percent,
			fmt.Sprintf("(%s/%s)", formatBytes(pr.bytesRead), formatBytes(pr.totalSize)),
			fmt.Sprintf("%s/s", formatBytes(int64(speed))),
			formatDuration(remainingTime))

		if percent == 100 {
			fmt.Println()       // 上传完成后换行
			pr.completed = true // 标记为已完成
		}
		pr.lastPrint = percent
	}
}

// formatDuration 格式化时间为易读格式
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "--:--:--"
	}

	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// formatBytes 格式化字节数为易读格式
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// UploadDirectory 上传目录
func (c *Client) UploadDirectory(bucketName, dirPath, prefix string, isPublic bool) error {
	// 检查目录是否存在
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("获取目录信息失败: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 不是一个目录", dirPath)
	}

	// 遍历目录
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 计算对象名称
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return fmt.Errorf("计算相对路径失败: %w", err)
		}

		// 安全检查路径
		safeRelPath, err := sanitizePath(relPath)
		if err != nil {
			return fmt.Errorf("路径安全检查失败: %w", err)
		}

		// 替换 Windows 路径分隔符
		safeRelPath = strings.ReplaceAll(safeRelPath, "\\", "/")

		// 构建对象名称
		objectName := safeRelPath
		if prefix != "" {
			objectName = filepath.Join(prefix, safeRelPath)
			// 替换 Windows 路径分隔符
			objectName = strings.ReplaceAll(objectName, "\\", "/")
		}

		// 上传文件
		return c.UploadFile(bucketName, path, objectName, isPublic)
	})
}

// DownloadFile 下载文件
func (c *Client) DownloadFile(bucketName, objectName, filePath string) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 获取对象信息以获取大小
	objInfo, err := c.client.StatObject(c.ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("获取对象信息失败: %w", err)
	}

	// 下载对象
	fmt.Printf("下载 %s/%s 到 %s...\n", bucketName, objectName, filePath)
	object, err := c.client.GetObject(c.ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("获取对象失败: %w", err)
	}
	defer object.Close()
	defer object.Close()

	// 使用进度跟踪
	progress := newProgressReader(objInfo.Size)
	_, err = io.Copy(file, io.TeeReader(object, progress))
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}

	return nil
}

// DownloadDirectory 下载目录
func (c *Client) DownloadDirectory(bucketName, prefix, dirPath string) error {
	// 列出所有对象
	objects := c.ListObjects(bucketName, prefix, true, false)
	for object := range objects {
		if object.Err != nil {
			return fmt.Errorf("列出对象失败: %w", object.Err)
		}

		// 跳过目录标记
		if strings.HasSuffix(object.Key, "/") {
			continue
		}

		// 计算本地文件路径
		relPath := strings.TrimPrefix(object.Key, prefix)
		localPath := filepath.Join(dirPath, relPath)

		// 下载文件
		if err := c.DownloadFile(bucketName, object.Key, localPath); err != nil {
			return fmt.Errorf("下载文件 %s 失败: %w", object.Key, err)
		}
	}

	return nil
}

// GenerateURL 生成访问 URL
func (c *Client) GenerateURL(bucketName, objectName string, expires time.Duration) (string, error) {
	// 检查对象是否存在
	_, err := c.client.StatObject(c.ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("对象不存在: %w", err)
	}

	var reqParams url.Values
	// 签名
	presignedURL, err := c.client.PresignedGetObject(c.ctx, bucketName, objectName, expires, reqParams)
	if err != nil {
		return "", fmt.Errorf("生成签名 URL 失败: %w", err)
	}
	return presignedURL.String(), nil
}

// ListObjects 列出对象
func (c *Client) ListObjects(bucketName, prefix string, recursive bool, onlyFolders bool, maxKeys ...int) <-chan minio.ObjectInfo {
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
		UseV1:     true, // 保持 UseV1 以便与现有逻辑兼容
	}
	// 如果提供了 maxKeys，则设置它
	if len(maxKeys) > 0 && maxKeys[0] > 0 {
		opts.MaxKeys = maxKeys[0]
	} else {
		opts.MaxKeys = DefaultMaxKeys
	}

	// 根据 maxKeys 动态调整缓冲区大小
	bufferSize := DefaultBufferSize
	if len(maxKeys) > 0 && maxKeys[0] > 0 {
		bufferSize = min(maxKeys[0], MaxBufferSize)
	}

	objects := make(chan minio.ObjectInfo, bufferSize)
	go func() {
		defer close(objects)

		for object := range c.client.ListObjects(c.ctx, bucketName, opts) {
			if object.Err != nil {
				objects <- object // 发送错误信息
				return            // 出错后停止
			}

			// 如果只查看文件夹，则跳过文件
			if onlyFolders && !strings.HasSuffix(object.Key, "/") {
				continue
			}

			objects <- object
		}
	}()

	return objects
}
