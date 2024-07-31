# 多目录并发 MP3 静音分割程序

这个程序可以并发处理多个目录中的 MP3 文件，检测静音段并将音频文件分割成多个部分。

## 功能特点

- 支持多目录并发扫描
- 使用 ffmpeg 进行静音检测和音频分割
- 并发处理多个 MP3 文件
- 自动检测系统 CPU 核心数，优化并发性能

## 依赖

- Go 1.15+
- ffmpeg

## 安装

1. 确保你的系统已安装 Go 1.15 或更高版本。
2. 安装 ffmpeg:

   - 对于 Ubuntu/Debian: `sudo apt-get install ffmpeg`
   - 对于 MacOS: `brew install ffmpeg`
   - 对于 Windows: 下载 ffmpeg 并将其添加到系统 PATH 中

3. 克隆此仓库：
   程序会递归扫描指定的目录，处理所有找到的 MP3 文件。

4. 编译程序：

```bash
go build -o mp3splitter main.go
```

## 使用方法

运行编译后的程序，并指定一个或多个要处理的目录：

```bash
./mp3splitter <directory1> [directory2] [directory3] ...
```

例如：

```bash
./mp3splitter /path/to/music1 /path/to/music2
```

## 配置和调优

1. 并发数量：
   程序默认使用系统的 CPU 核心数作为并发工作线程数。如果需要调整，可以修改 `main()` 函数中的 `numWorkers` 变量。

2. Channel 缓冲大小：
   如果处理大量文件，可能需要增加 `jobs` 和 `results` channels 的缓冲大小。这些可以在 `main()` 函数中调整。

3. 静音检测参数：
   静音检测的参数（如噪音阈值和持续时间）可以在 `detectSilence()` 函数中的 ffmpeg 命令中调整。

## 调试

1. 日志输出：
   程序会打印处理每个文件的状态。如需更详细的日志，可以在关键位置添加更多的 `fmt.Printf()` 语句。

2. 错误处理：
   所有的错误都会被捕获并打印到控制台。检查这些错误信息可以帮助诊断问题。

3. 单文件调试：
   如果遇到特定文件的问题，可以修改 `processMP3File()` 函数，添加更多的调试输出。

4. ffmpeg 命令调试：
   要调试 ffmpeg 相关的问题，可以在 `detectSilence()` 和 `splitAudio()` 函数中打印完整的 ffmpeg 命令，然后在命令行中手动运行这些命令。

5. 并发问题：
   如果怀疑有并发相关的问题，可以尝试减少 `numWorkers` 的数量，甚至设置为 1 来进行测试。

6. 内存使用：
   如果程序消耗过多内存，可以使用 Go 的性能分析工具 (pprof) 来诊断内存使用情况。

## 注意事项

- 确保有足够的磁盘空间来存储分割后的音频文件。
- 处理大量文件时，注意系统的文件描述符限制。
- 该程序会覆盖已存在的输出文件，请注意备份重要数据。

## 贡献

欢迎提交 issues 和 pull requests 来改进这个项目。

## 许可

[MIT License](LICENSE)
