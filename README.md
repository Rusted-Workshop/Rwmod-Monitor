# RWMod Monitor

一个用于监控文件系统变化并自动备份到 S3 的工具。

## 功能特性

- 自动监控指定目录下一层子目录的文件变化
- 延迟触发机制：目录内容 5 分钟无变化后自动打包
- 自动将目录打包为 `.rwmod` 格式（本质是 zip）
- 上传队列和失败重试机制（最多 3 次）
- 支持 S3 兼容存储
- 首次启动时为所有子目录创建初始备份

## 配置

首次运行程序会在二进制文件同目录下自动创建 `config.json` 模板：

```json
{
  "monitor_dir": "/path/to/monitor",
  "delay_minutes": 5,
  "max_retries": 3,
  "s3": {
    "endpoint": "https://s3.example.com",
    "access_key": "your-access-key-id",
    "secret_key": "your-secret-key",
    "bucket": "your-bucket-name"
  }
}
```

### 配置说明

- `monitor_dir`: 要监控的根目录路径
- `delay_minutes`: 文件无变化后多久触发备份（分钟）
- `max_retries`: 上传失败时的最大重试次数
- `s3.endpoint`: S3 服务器地址
- `s3.access_key`: S3 访问密钥 ID
- `s3.secret_key`: S3 访问密钥
- `s3.bucket`: S3 存储桶名称

## 使用方法

1. 编译程序：
```bash
go build -o rwmod-monitor
```

2. 首次运行生成配置文件：
```bash
./rwmod-monitor
```

3. 编辑 `config.json` 填入你的配置信息

4. 再次运行开始监控：
```bash
./rwmod-monitor
```

## 工作原理

1. **监控层级**：程序只监控 `monitor_dir` 下面一层的子目录
   - 例如监控 `/a`，则只监控 `/a/a`、`/a/b` 等一级子目录

2. **延迟触发**：
   - 当目录内文件发生变化时，启动 5 分钟倒计时
   - 如果期间有新的变化，重置倒计时
   - 5 分钟无变化后，自动打包该目录

3. **打包格式**：
   - 文件名格式：`目录名-时间戳.rwmod`
   - 时间戳使用 UTC Unix 时间戳
   - 例如：`a-1744444444.rwmod`
   - 保留原始目录结构在压缩包内

4. **上传队列**：
   - 打包完成后自动加入上传队列
   - 上传失败会自动重试（最多 3 次）
   - 重试失败的文件会移动到 `.failed_uploads` 目录

5. **初始备份**：
   - 程序启动时自动为所有现有子目录创建备份
   - 并立即上传到 S3

## 项目结构

```
rwmod-monitor/
├── cmd/
│   └── root.go              # 主命令和事件处理
├── internal/
│   ├── archiver/
│   │   └── archiver.go      # 文件打包模块
│   ├── config/
│   │   └── config.go        # 配置管理模块
│   ├── queue/
│   │   └── queue.go         # 上传队列和重试机制
│   ├── tracker/
│   │   └── tracker.go       # 延迟跟踪模块
│   └── uploader/
│       └── s3.go            # S3 上传模块
├── main.go
├── go.mod
└── README.md
```

## 注意事项

- 配置文件包含敏感信息（S3 密钥），请勿提交到版本控制
- 隐藏目录（以 `.` 开头）会被自动忽略
- 上传失败的文件会保存在 `monitor_dir/.failed_uploads/` 目录
- 程序需要对监控目录有读写权限
