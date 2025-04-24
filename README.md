# s3ctl

`s3ctl` 是一个用于操作 S3 兼容对象存储的命令行工具。它基于 [minio-go](https://github.com/minio/minio-go) 构建，提供了方便的文件上传、下载、列表、URL 生成等功能。

## 功能

*   初始化配置  
*   管理多个 S3 配置
*   列出存储桶 (Buckets)
*   创建存储桶 (Make Bucket)
*   删除存储桶 (Remove Bucket)
*   列出存储桶中的对象 (Objects)
*   上传文件或目录
*   删除对象或目录下的对象
*   生成对象的预签名访问 URL  

## 安装

你可以使用 `go install` 来安装 `s3ctl`:

```bash
go install github.com/zboyco/s3ctl/cmd/s3ctl@latest
```

确保你的 `$GOPATH/bin` 或 `$HOME/go/bin` 目录在你的 `PATH` 环境变量中。

## 配置

首次使用前，你需要初始化配置文件。运行以下命令会在你的用户主目录下创建默认配置文件 `~/.s3ctl`:

```bash
s3ctl config init
```

命令执行后，会生成类似以下的配置文件：

```yaml:~/.s3ctl
# s3ctl 配置文件
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
```

请根据你的 S3 存储提供商信息修改此文件：

*   `endpoint`: S3 服务的地址 (例如: `s3.amazonaws.com`, `oss-cn-hangzhou.aliyuncs.com`)
*   `access_key_id`: 你的 Access Key ID
*   `secret_access_key`: 你的 Secret Access Key
*   `use_ssl`: 是否使用 HTTPS (true 或 false)

## 用法

`s3ctl` 使用 `s3://bucket/object` 的格式来指定 S3 路径。

### 1. 配置管理 (config)

*   初始化配置文件:
    ```bash
    s3ctl config init
    ```

*   查看当前配置信息:
    ```bash
    s3ctl config list
    ```

*   切换到指定配置:
    ```bash
    s3ctl config use [配置名]
    ```

### 2. 列出存储桶或对象 (ls)

*   列出所有存储桶:
    ```bash
    s3ctl ls
    ```
*   列出指定存储桶 `mybucket` 下的对象:
    ```bash
    s3ctl ls s3://mybucket/
    ```
*   列出 `mybucket` 下 `photos/` 前缀的对象:
    ```bash
    s3ctl ls s3://mybucket/photos/
    ```
*   递归列出 `mybucket` 下 `archive/` 的所有对象，并显示完整路径:
    ```bash
    s3ctl ls s3://mybucket/archive/ -r -p
    ```
*   只列出 `mybucket` 下的文件夹:
    ```bash
    s3ctl ls s3://mybucket/ -f
    ```

### 3. 创建存储桶 (mb)

创建一个名为 `new-bucket` 的存储桶:

```bash
s3ctl mb s3://new-bucket
```

### 4. 删除存储桶 (rb)

删除一个名为 `empty-bucket` 的空存储桶:

```bash
s3ctl rb s3://empty-bucket
```
**注意:** 只有当存储桶为空时才能被删除。

### 5. 上传文件或目录 (put)

*   上传本地文件 `localfile.txt` 到 `mybucket` 下的 `remote/path/file.txt`:
    ```bash
    s3ctl put localfile.txt s3://mybucket/remote/path/file.txt
    ```
*   上传本地目录 `localdir/` 下的所有内容到 `mybucket` 下的 `remote/prefix/`，并设置为公开可读:
    ```bash
    s3ctl put localdir/ s3://mybucket/remote/prefix/ -p
    ```
    `-p` 或 `--public` 标志将上传的对象设置为公开可读。

### 6. 删除对象 (del)

*   删除 `mybucket` 下的 `object/to/delete.txt` 对象:
    ```bash
    s3ctl del s3://mybucket/object/to/delete.txt
    ```
*   递归删除 `mybucket` 下 `folder/prefix/` 前缀的所有对象:
    ```bash
    s3ctl del s3://mybucket/folder/prefix/ 
    ```

### 7. 生成访问 URL (url)

为 `mybucket` 下的 `important/document.pdf` 生成一个有效期为 7 天的预签名访问 URL:

```bash
s3ctl url s3://mybucket/important/document.pdf -e 7d
```

*   `-e` 或 `--expiry`: 设置 URL 有效期 (例如: `1h`, `24h`, `7d`)，默认为 24 小时。
*   `-2` 或 `--v2`: 使用 V2 签名协议 (默认为 V4)。

## 依赖

*   [github.com/minio/minio-go/v7](https://github.com/minio/minio-go)
*   [github.com/spf13/cobra](https://github.com/spf13/cobra)
*   [github.com/spf13/viper](https://github.com/spf13/viper)